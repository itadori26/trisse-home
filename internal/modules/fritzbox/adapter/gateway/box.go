package gateway

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"trisse-home/internal/modules/fritzbox/domain"
)

// UPnP/IGD (no login) and TR-064 (login) services.
const (
	igdCommon       = "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1"
	igdCommonURL    = "/igdupnp/control/WANCommonIFC1"
	igdIP           = "urn:schemas-upnp-org:service:WANIPConnection:1"
	igdIPURL        = "/igdupnp/control/WANIPConn1"
	trCommon        = "urn:dslforum-org:service:WANCommonInterfaceConfig:1"
	trCommonURL     = "/upnp/control/wancommonifconfig1"
	trIP            = "urn:dslforum-org:service:WANIPConnection:1"
	trIPURL         = "/upnp/control/wanipconnection1"
	trPPP           = "urn:dslforum-org:service:WANPPPConnection:1"
	trPPPURL        = "/upnp/control/wanpppconn1"
	trOnTel         = "urn:dslforum-org:service:X_AVM-DE_OnTel:1"
	trOnTelURL      = "/upnp/control/x_contact"
	connected       = "Connected"
	callDateLayout  = "02.01.06 15:04"
	descriptionPath = "/tr64desc.xml"
)

// Box implements usecase.Box.
type Box struct {
	client *http.Client
	soap   soapClient
	loc    *time.Location

	mu     sync.Mutex
	models map[string]string
}

// NewBox returns a client; loc is the time zone of the call list dates.
func NewBox(client *http.Client, loc *time.Location) *Box {
	return &Box{client: client, soap: soapClient{http: client}, loc: loc, models: map[string]string{}}
}

// WAN implements usecase.Box. IGD is preferred because it needs no login and
// reports the throughput; TR-064 is the fallback for status and IP.
func (b *Box) WAN(ctx context.Context, s domain.Settings) (domain.WAN, string, error) {
	var w domain.WAN
	if st, err := b.soap.call(ctx, s.URL, igdIPURL, igdIP, "GetStatusInfo", nil, "", ""); err == nil {
		w.Connected = st["NewConnectionStatus"] == connected
		w.Uptime = seconds(st["NewUptime"])
		if ip, err := b.soap.call(ctx, s.URL, igdIPURL, igdIP, "GetExternalIPAddress", nil, "", ""); err == nil {
			w.ExternalIP = ip["NewExternalIPAddress"]
		}
		if add, err := b.soap.call(ctx, s.URL, igdCommonURL, igdCommon, "GetAddonInfos", nil, "", ""); err == nil {
			w.Known = true
			w.DownstreamBs, _ = strconv.ParseInt(add["NewByteReceiveRate"], 10, 64)
			w.UpstreamBs, _ = strconv.ParseInt(add["NewByteSendRate"], 10, 64)
		}
	} else if !s.HasLogin() {
		return w, "", fmt.Errorf("UPnP status: %w", err)
	} else if err := b.tr064Status(ctx, s, &w); err != nil {
		return w, "", err
	}
	if s.HasLogin() {
		if link, err := b.soap.call(ctx, s.URL, trCommonURL, trCommon, "GetCommonLinkProperties", nil, s.Username, s.Password); err == nil {
			w.DownstreamMaxBits, _ = strconv.ParseInt(link["NewLayer1DownstreamMaxBitRate"], 10, 64)
			w.UpstreamMaxBits, _ = strconv.ParseInt(link["NewLayer1UpstreamMaxBitRate"], 10, 64)
		}
	}
	return w, b.model(ctx, s.URL), nil
}

// tr064Status reads status, IP and uptime via TR-064 (cable: WANIPConnection,
// DSL: WANPPPConnection).
func (b *Box) tr064Status(ctx context.Context, s domain.Settings, w *domain.WAN) error {
	var lastErr error
	for _, svc := range []struct{ url, service string }{{trIPURL, trIP}, {trPPPURL, trPPP}} {
		info, err := b.soap.call(ctx, s.URL, svc.url, svc.service, "GetInfo", nil, s.Username, s.Password)
		if err != nil {
			lastErr = err
			continue
		}
		w.Connected = info["NewConnectionStatus"] == connected
		w.ExternalIP = info["NewExternalIPAddress"]
		w.Uptime = seconds(info["NewUptime"])
		if w.Connected {
			return nil
		}
	}
	if lastErr != nil && !w.Connected {
		return fmt.Errorf("TR-064 status: %w", lastErr)
	}
	return nil
}

func seconds(s string) time.Duration {
	n, _ := strconv.ParseInt(s, 10, 64)
	return time.Duration(n) * time.Second
}

// model returns the model name from the public device description (cached).
func (b *Box) model(ctx context.Context, baseURL string) string {
	b.mu.Lock()
	m, ok := b.models[baseURL]
	b.mu.Unlock()
	if ok {
		return m
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+descriptionPath, nil)
	if err != nil {
		return ""
	}
	res, err := b.client.Do(req)
	if err != nil {
		return ""
	}
	defer res.Body.Close()
	var desc struct {
		Device struct {
			ModelName string `xml:"modelName"`
		} `xml:"device"`
	}
	if res.StatusCode == http.StatusOK && xml.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(&desc) == nil {
		m = desc.Device.ModelName
	}
	b.mu.Lock()
	b.models[baseURL] = m
	b.mu.Unlock()
	return m
}

type callList struct {
	Calls []struct {
		Type     int    `xml:"Type"`
		Caller   string `xml:"Caller"`
		Called   string `xml:"Called"`
		Name     string `xml:"Name"`
		Date     string `xml:"Date"`
		Duration string `xml:"Duration"`
	} `xml:"Call"`
}

// Calls implements usecase.Box: GetCallList returns a URL with a session ID
// from which the list is downloaded.
func (b *Box) Calls(ctx context.Context, s domain.Settings, limit int) ([]domain.Call, error) {
	res, err := b.soap.call(ctx, s.URL, trOnTelURL, trOnTel, "GetCallList", nil, s.Username, s.Password)
	if err != nil {
		return nil, err
	}
	listURL, err := sameHost(s.URL, res["NewCallListURL"])
	if err != nil {
		return nil, err
	}
	q := listURL.Query()
	q.Set("max", strconv.Itoa(limit))
	listURL.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download call list: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("call list returned HTTP %d", resp.StatusCode)
	}
	var list callList
	if err := xml.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(&list); err != nil {
		return nil, fmt.Errorf("decode call list: %w", err)
	}
	out := make([]domain.Call, 0, len(list.Calls))
	for _, c := range list.Calls {
		t := domain.CallTypeFromAVM(c.Type)
		number := c.Caller
		if t == domain.CallOutgoing {
			number = c.Called
		}
		at, _ := time.ParseInLocation(callDateLayout, strings.TrimSpace(c.Date), b.loc)
		out = append(out, domain.Call{Type: t, Number: number, Name: c.Name, At: at, Duration: callDuration(c.Duration)})
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

// sameHost resolves the call list URL and makes sure it points to the
// configured box (the box reports its own address, which may differ from
// the configured host name, e.g. fritz.box vs. 192.168.178.1).
func sameHost(base, raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Path == "" {
		return nil, errors.New("invalid call list URL")
	}
	b, _ := url.Parse(base) // validated by domain.NewSettings
	u.Scheme, u.Host = b.Scheme, b.Host
	return u, nil
}

// callDuration parses "h:mm".
func callDuration(s string) time.Duration {
	h, m, ok := strings.Cut(strings.TrimSpace(s), ":")
	if !ok {
		return 0
	}
	hours, _ := strconv.Atoi(h)
	minutes, _ := strconv.Atoi(m)
	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
}
