package gateway

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/fritzbox/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func md5hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// fakeBox emulates IGD (optional), TR-064 with digest auth and the call list.
func fakeBox(t *testing.T, igd bool) *httptest.Server {
	t.Helper()
	const realm, nonce, user, pass = "F!Box SOAP-Auth", "ABC123", "home", "secret"
	param := regexp.MustCompile(`(\w+)="?([^",]*)"?`)
	authorized := func(r *http.Request) bool {
		h := r.Header.Get("Authorization")
		p := map[string]string{}
		for _, m := range param.FindAllStringSubmatch(strings.TrimPrefix(h, "Digest "), -1) {
			p[m[1]] = m[2]
		}
		want := md5hex(md5hex(user+":"+realm+":"+pass) + ":" + nonce + ":" + p["nc"] + ":" + p["cnonce"] + ":auth:" + md5hex(r.Method+":"+r.URL.Path))
		return p["username"] == user && p["response"] == want
	}
	soap := func(w http.ResponseWriter, action string, kv ...string) {
		var b strings.Builder
		for i := 0; i+1 < len(kv); i += 2 {
			fmt.Fprintf(&b, "<%s>%s</%s>", kv[i], kv[i+1], kv[i])
		}
		fmt.Fprintf(w, `<?xml version="1.0"?><s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><u:%sResponse xmlns:u="x">%s</u:%sResponse></s:Body></s:Envelope>`, action, b.String(), action)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /tr64desc.xml", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `<root><device><modelName>FRITZ!Box 6670 Cable</modelName></device></root>`)
	})
	mux.HandleFunc("POST /igdupnp/control/", func(w http.ResponseWriter, r *http.Request) {
		if !igd {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		switch action := r.Header.Get("SoapAction")[strings.Index(r.Header.Get("SoapAction"), "#")+1:]; action {
		case "GetStatusInfo":
			soap(w, action, "NewConnectionStatus", "Connected", "NewUptime", "3600")
		case "GetExternalIPAddress":
			soap(w, action, "NewExternalIPAddress", "203.0.113.7")
		case "GetAddonInfos":
			soap(w, action, "NewByteReceiveRate", "12500000", "NewByteSendRate", "1000")
		}
	})
	mux.HandleFunc("POST /upnp/control/", func(w http.ResponseWriter, r *http.Request) {
		if !authorized(r) {
			w.Header().Set("WWW-Authenticate", `Digest realm="`+realm+`", nonce="`+nonce+`", algorithm=MD5, qop="auth"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch action := r.Header.Get("SoapAction")[strings.Index(r.Header.Get("SoapAction"), "#")+1:]; action {
		case "GetCommonLinkProperties":
			soap(w, action, "NewLayer1DownstreamMaxBitRate", "200000000", "NewLayer1UpstreamMaxBitRate", "50000000")
		case "GetInfo":
			soap(w, action, "NewConnectionStatus", "Connected", "NewExternalIPAddress", "198.51.100.1", "NewUptime", "60")
		case "GetCallList":
			soap(w, action, "NewCallListURL", "http://192.168.178.1:49000/calllist.lua?sid=1234")
		}
	})
	mux.HandleFunc("GET /calllist.lua", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("sid") != "1234" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?><root><Call><Type>2</Type><Caller>0441123456</Caller><Called>SIP: 99</Called><Name>Oma</Name><Date>11.11.26 18:03</Date><Duration>0:00</Duration></Call>
			<Call><Type>3</Type><Caller>SIP: 99</Caller><Called>0170111</Called><Name></Name><Date>10.11.26 09:15</Date><Duration>1:05</Duration></Call></root>`)
	})
	srv := httptest.NewServer(mux)
	return srv
}

func TestBoxWithIGDAndLogin(t *testing.T) {
	srv := fakeBox(t, true)
	defer srv.Close()
	ctx := context.Background()
	b := NewBox(srv.Client(), time.UTC)
	s := domain.Settings{URL: srv.URL, Username: "home", Password: "secret"}
	w, model, err := b.WAN(ctx, s)
	if err != nil || !w.Connected || w.ExternalIP != "203.0.113.7" || w.DownstreamPercent() != 50 || !w.Known || model != "FRITZ!Box 6670 Cable" {
		t.Fatalf("wan %+v %q %v", w, model, err)
	}
	calls, err := b.Calls(ctx, s, 10)
	if err != nil || len(calls) != 2 || calls[0].Type != domain.CallMissed || calls[0].Name != "Oma" || calls[1].Number != "0170111" || calls[1].Duration != 65*time.Minute {
		t.Fatalf("calls %+v %v", calls, err)
	}
	if calls[0].At.Day() != 11 || calls[0].At.Hour() != 18 {
		t.Fatalf("date %v", calls[0].At)
	}
	s.Password = "wrong"
	if _, err := b.Calls(ctx, s, 10); !errors.Is(err, errUnauthorized) {
		t.Fatalf("wrong password: %v", err)
	}
}

func TestBoxFallbackWithoutIGD(t *testing.T) {
	srv := fakeBox(t, false)
	defer srv.Close()
	ctx := context.Background()
	b := NewBox(srv.Client(), time.UTC)
	if _, _, err := b.WAN(ctx, domain.Settings{URL: srv.URL}); err == nil {
		t.Fatal("without IGD and login there is no status")
	}
	w, _, err := b.WAN(ctx, domain.Settings{URL: srv.URL, Username: "home", Password: "secret"})
	if err != nil || !w.Connected || w.ExternalIP != "198.51.100.1" || w.Known {
		t.Fatalf("tr-064 fallback %+v %v", w, err)
	}
}

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (prefixCipher) Decrypt(s string) (string, error) { return strings.TrimPrefix(s, "enc:"), nil }

func TestSettingsRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewSettingsRepository(dbtest.Open(t, "fritzbox", Migrations()), prefixCipher{})
	if _, err := repo.Load(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatal(err)
	}
	_ = repo.Save(ctx, domain.Settings{URL: domain.DefaultURL})
	if s, err := repo.Load(ctx); err != nil || s.HasLogin() {
		t.Fatalf("no-login round trip %+v %v", s, err)
	}
	_ = repo.Save(ctx, domain.Settings{URL: domain.DefaultURL, Username: "home", Password: "pw"})
	if s, err := repo.Load(ctx); err != nil || s.Password != "pw" {
		t.Fatalf("login round trip %+v %v", s, err)
	}
}
