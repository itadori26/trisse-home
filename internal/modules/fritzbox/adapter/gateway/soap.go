package gateway

import (
	"bytes"
	"context"
	"crypto/md5" //nolint:gosec // HTTP digest authentication as required by TR-064 is defined with MD5
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const maxResponseBytes = 2 << 20

// errUnauthorized is returned when the box rejects the credentials.
var errUnauthorized = errors.New("username or password is wrong")

// soapClient calls UPnP/TR-064 actions.
type soapClient struct{ http *http.Client }

const envelope = `<?xml version="1.0" encoding="utf-8"?>` +
	`<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">` +
	`<s:Body><u:%s xmlns:u="%s">%s</u:%s></s:Body></s:Envelope>`

// call invokes action of service at baseURL+controlURL and returns the
// response arguments (element name → text). user/password enable digest auth.
func (c soapClient) call(ctx context.Context, baseURL, controlURL, service, action string, args map[string]string, user, password string) (map[string]string, error) {
	var argXML strings.Builder
	for k, v := range args {
		argXML.WriteString("<" + k + ">")
		_ = xml.EscapeText(&argXML, []byte(v))
		argXML.WriteString("</" + k + ">")
	}
	body := fmt.Sprintf(envelope, action, service, argXML.String(), action)
	do := func(auth string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+controlURL, strings.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
		req.Header.Set("SoapAction", service+"#"+action)
		if auth != "" {
			req.Header.Set("Authorization", auth)
		}
		return c.http.Do(req)
	}
	res, err := do("")
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusUnauthorized && password != "" {
		challenge := res.Header.Get("WWW-Authenticate")
		_ = res.Body.Close()
		auth, err := digestAuthorization(challenge, http.MethodPost, controlURL, user, password)
		if err != nil {
			return nil, err
		}
		if res, err = do(auth); err != nil {
			return nil, err
		}
	}
	defer res.Body.Close()
	switch {
	case res.StatusCode == http.StatusUnauthorized:
		return nil, errUnauthorized
	case res.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%s returned HTTP %d", action, res.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return responseArgs(raw, action+"Response")
}

// responseArgs collects the child elements of the <XResponse> element.
func responseArgs(raw []byte, wrapper string) (map[string]string, error) {
	dec := xml.NewDecoder(bytes.NewReader(raw))
	out := map[string]string{}
	inside, current := false, ""
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode soap response: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == wrapper {
				inside = true
			} else if inside {
				current = t.Name.Local
				out[current] = ""
			}
		case xml.CharData:
			if inside && current != "" {
				out[current] += string(t)
			}
		case xml.EndElement:
			if t.Name.Local == wrapper {
				return out, nil
			}
			current = ""
		}
	}
	if !inside {
		return nil, errors.New("unexpected soap response")
	}
	return out, nil
}

var challengeParam = regexp.MustCompile(`(\w+)="?([^",]*)"?`)

// digestAuthorization answers an RFC 7616 digest challenge (MD5, qop=auth).
func digestAuthorization(challenge, method, uri, user, password string) (string, error) {
	if !strings.HasPrefix(strings.ToLower(challenge), "digest ") {
		return "", errors.New("the FRITZ!Box did not offer digest authentication")
	}
	p := map[string]string{}
	for _, m := range challengeParam.FindAllStringSubmatch(challenge[len("digest "):], -1) {
		p[strings.ToLower(m[1])] = m[2]
	}
	h := func(s string) string {
		sum := md5.Sum([]byte(s)) //nolint:gosec // see import
		return hex.EncodeToString(sum[:])
	}
	cnonceBytes := make([]byte, 8)
	_, _ = rand.Read(cnonceBytes)
	cnonce, nc := hex.EncodeToString(cnonceBytes), "00000001"
	ha1, ha2 := h(user+":"+p["realm"]+":"+password), h(method+":"+uri)
	if p["qop"] == "" {
		return fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s"`,
			user, p["realm"], p["nonce"], uri, h(ha1+":"+p["nonce"]+":"+ha2)), nil
	}
	resp := h(ha1 + ":" + p["nonce"] + ":" + nc + ":" + cnonce + ":auth:" + ha2)
	auth := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", qop=auth, nc=%s, cnonce="%s", response="%s", algorithm=MD5`,
		user, p["realm"], p["nonce"], uri, nc, cnonce, resp)
	if p["opaque"] != "" {
		auth += `, opaque="` + p["opaque"] + `"`
	}
	return auth, nil
}
