package gateway

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"
)

const maxLogBytes = 1 << 20

// Restart implements usecase.ContainerController. The API must point to the
// actions proxy, which only allows stop/restart/kill.
func (a *API) Restart(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/containers/"+url.PathEscape(name)+"/restart?t=10", nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	res, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("docker api: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("docker api returned %s", res.Status)
	}
	return nil
}

// Logs implements usecase.LogSource.
func (a *API) Logs(ctx context.Context, name string, lines int) ([]string, error) {
	u := fmt.Sprintf("%s/containers/%s/logs?stdout=1&stderr=1&tail=%d", a.baseURL, url.PathEscape(name), lines)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	res, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("docker api: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docker api returned %s", res.Status)
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, maxLogBytes))
	if err != nil {
		return nil, fmt.Errorf("read logs: %w", err)
	}
	if res.Header.Get("Content-Type") == "application/vnd.docker.multiplexed-stream" || looksMultiplexed(raw) {
		raw = demultiplex(raw)
	}
	return splitLines(raw, lines), nil
}

// looksMultiplexed detects the 8-byte frame header Docker uses for
// containers without TTY (older APIs do not set the content type).
func looksMultiplexed(b []byte) bool {
	return len(b) >= 8 && b[0] <= 2 && b[1] == 0 && b[2] == 0 && b[3] == 0
}

// demultiplex strips the stream frame headers (stream byte, 3 zero bytes,
// big-endian payload size); a truncated last frame is kept partially.
func demultiplex(b []byte) []byte {
	var out bytes.Buffer
	for len(b) >= 8 {
		size := int(binary.BigEndian.Uint32(b[4:8]))
		b = b[8:]
		if size > len(b) {
			size = len(b)
		}
		out.Write(b[:size])
		b = b[size:]
	}
	return out.Bytes()
}

func splitLines(b []byte, limit int) []string {
	var lines []string
	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(make([]byte, 64<<10), maxLogBytes)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		if !utf8.ValidString(line) {
			line = strings.ToValidUTF8(line, "�")
		}
		lines = append(lines, line)
	}
	if err := sc.Err(); err != nil && !errors.Is(err, bufio.ErrTooLong) {
		lines = append(lines, "… "+err.Error())
	}
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	return lines
}
