package webauthn

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// maxCBORDepth and maxCBORItems bound the decoder against hostile input.
const (
	maxCBORDepth = 8
	maxCBORItems = 1024
)

var errCBOR = errors.New("invalid CBOR")

// decodeCBOR decodes the subset of CBOR (RFC 8949) used by WebAuthn:
// integers, byte and text strings, arrays, maps and simple values. Maps
// are returned as map[any]any with int64 or string keys.
func decodeCBOR(data []byte) (any, []byte, error) {
	d := &cborDecoder{data: data}
	v, err := d.value(0)
	if err != nil {
		return nil, nil, err
	}
	return v, d.data[d.pos:], nil
}

type cborDecoder struct {
	data  []byte
	pos   int
	items int
}

func (d *cborDecoder) next(n int) ([]byte, error) {
	if n < 0 || d.pos+n > len(d.data) {
		return nil, errCBOR
	}
	b := d.data[d.pos : d.pos+n]
	d.pos += n
	return b, nil
}

func (d *cborDecoder) argument(info byte) (uint64, error) {
	switch {
	case info < 24:
		return uint64(info), nil
	case info == 24:
		b, err := d.next(1)
		if err != nil {
			return 0, err
		}
		return uint64(b[0]), nil
	case info == 25:
		b, err := d.next(2)
		if err != nil {
			return 0, err
		}
		return uint64(binary.BigEndian.Uint16(b)), nil
	case info == 26:
		b, err := d.next(4)
		if err != nil {
			return 0, err
		}
		return uint64(binary.BigEndian.Uint32(b)), nil
	case info == 27:
		b, err := d.next(8)
		if err != nil {
			return 0, err
		}
		return binary.BigEndian.Uint64(b), nil
	}
	return 0, errCBOR // indefinite lengths are not used by WebAuthn
}

func (d *cborDecoder) value(depth int) (any, error) {
	if depth > maxCBORDepth {
		return nil, errCBOR
	}
	d.items++
	if d.items > maxCBORItems {
		return nil, errCBOR
	}
	head, err := d.next(1)
	if err != nil {
		return nil, err
	}
	major, info := head[0]>>5, head[0]&0x1f
	if major == 7 {
		switch info {
		case 20:
			return false, nil
		case 21:
			return true, nil
		case 22, 23:
			return nil, nil
		}
		return nil, errCBOR
	}
	arg, err := d.argument(info)
	if err != nil {
		return nil, err
	}
	switch major {
	case 0:
		if arg > 1<<62 {
			return nil, errCBOR
		}
		return int64(arg), nil
	case 1:
		if arg > 1<<62 {
			return nil, errCBOR
		}
		return -1 - int64(arg), nil
	case 2, 3:
		if arg > uint64(len(d.data)) {
			return nil, errCBOR
		}
		b, err := d.next(int(arg)) //nolint:gosec // bounded by len(d.data) above
		if err != nil {
			return nil, err
		}
		if major == 3 {
			return string(b), nil
		}
		return append([]byte(nil), b...), nil
	case 4:
		if arg > maxCBORItems {
			return nil, errCBOR
		}
		out := make([]any, 0, arg)
		for range arg {
			v, err := d.value(depth + 1)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case 5:
		if arg > maxCBORItems {
			return nil, errCBOR
		}
		out := make(map[any]any, arg)
		for range arg {
			k, err := d.value(depth + 1)
			if err != nil {
				return nil, err
			}
			switch k.(type) {
			case int64, string:
			default:
				return nil, fmt.Errorf("%w: unsupported map key", errCBOR)
			}
			v, err := d.value(depth + 1)
			if err != nil {
				return nil, err
			}
			out[k] = v
		}
		return out, nil
	}
	return nil, errCBOR // tags and floats are not used
}
