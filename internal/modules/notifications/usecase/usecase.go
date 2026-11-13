package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"trisse-home/internal/modules/notifications/domain"
	"trisse-home/internal/shared/kernel"
)

// deliveryTimeout bounds one push delivery.
const deliveryTimeout = 15 * time.Second

// maxDigestLines bounds the body of a quiet-hours digest.
const maxDigestLines = 20

// SettingsInput is the input of SaveSettings; an empty token keeps the stored one.
type SettingsInput struct {
	Enabled   bool
	ServerURL string
	Topic     string
	Token     string
	Rules     domain.Rules
	Policy    domain.Policy
}

// SettingsView never contains the token.
type SettingsView struct {
	Enabled       bool
	ServerURL     string
	Topic         string
	HasToken      bool
	Rules         domain.Rules
	Policy        domain.Policy
	LastError     string
	LastDelivered time.Time
}

// ChannelInput creates or updates a channel; an empty Secret keeps the stored one.
type ChannelInput struct {
	domain.Channel
}

// ChannelView describes a channel without its secret.
type ChannelView struct {
	domain.Channel
	HasSecret bool
}

// Notifier is the input port of the module. It also implements
// kernel.ObservationSink and kernel.AlertSink so other modules can report to it.
type Notifier interface {
	kernel.ObservationSink
	kernel.AlertSink
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	SendTest(ctx context.Context) error

	Channels(ctx context.Context) ([]ChannelView, error)
	SaveChannel(ctx context.Context, in ChannelInput) (ChannelView, error)
	DeleteChannel(ctx context.Context, id string) error
	TestChannel(ctx context.Context, id string) error

	// Maintenance lists active and upcoming windows.
	Maintenance(ctx context.Context) ([]domain.MaintenanceWindow, error)
	StartMaintenance(ctx context.Context, scope []string, start, end time.Time, note string) (domain.MaintenanceWindow, error)
	EndMaintenance(ctx context.Context, id string) error
	// UnderMaintenance reports whether an object is muted right now (for uptime #81).
	UnderMaintenance(ctx context.Context, kind, key, label string, at time.Time) bool

	// FlushDigest sends notifications collected during quiet hours (job).
	FlushDigest(ctx context.Context) error
}

// Repository persists the settings (output port); Load returns the defaults
// when nothing is stored.
type Repository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// ChannelRepository persists additional channels; secrets are encrypted.
type ChannelRepository interface {
	List(ctx context.Context) ([]domain.Channel, error)
	Save(ctx context.Context, c domain.Channel) error
	Delete(ctx context.Context, id string) error
}

// MaintenanceRepository persists maintenance windows.
type MaintenanceRepository interface {
	// ListEndingAfter returns windows that end after t.
	ListEndingAfter(ctx context.Context, t time.Time) ([]domain.MaintenanceWindow, error)
	Save(ctx context.Context, w domain.MaintenanceWindow) error
	Delete(ctx context.Context, id string) error
}

// PendingRepository keeps notifications deferred by quiet hours.
type PendingRepository interface {
	Add(ctx context.Context, m domain.Message, at time.Time) error
	// Take returns and removes all pending messages, oldest first.
	Take(ctx context.Context) ([]domain.Message, error)
}

// Sender delivers a message via ntfy (output port).
type Sender interface {
	Send(ctx context.Context, s domain.Settings, m domain.Message) error
}

// Publisher delivers a message to an additional channel (output port, one
// adapter per channel kind).
type Publisher interface {
	Publish(ctx context.Context, c domain.Channel, m domain.Message) error
}

// Deps bundles the output ports.
type Deps struct {
	Repo        Repository
	Sender      Sender
	Channels    ChannelRepository
	Publishers  map[domain.ChannelKind]Publisher
	Maintenance MaintenanceRepository
	Pending     PendingRepository
	Clock       kernel.Clock
	IDs         kernel.IDGenerator
	// Location is the time zone of quiet hours (TZ).
	Location *time.Location
}

type notifier struct {
	d Deps

	mu            sync.Mutex
	settings      *domain.Settings
	tracks        map[string]domain.Track
	lastError     string
	lastDelivered time.Time
	wg            sync.WaitGroup
}

// NewNotifier returns the notification use cases.
func NewNotifier(d Deps) Notifier {
	if d.Location == nil {
		d.Location = time.UTC
	}
	if d.Publishers == nil {
		d.Publishers = map[domain.ChannelKind]Publisher{}
	}
	return &notifier{d: d, tracks: map[string]domain.Track{}}
}

// current returns the cached settings, loading them once.
func (uc *notifier) current(ctx context.Context) (domain.Settings, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if uc.settings != nil {
		return *uc.settings, nil
	}
	s, err := uc.d.Repo.Load(ctx)
	if err != nil {
		return domain.Settings{}, fmt.Errorf("load notification settings: %w", err)
	}
	uc.settings = &s
	return s, nil
}

func priorityFor(s domain.Severity) int {
	switch s {
	case domain.SeverityCritical:
		return 5
	case domain.SeverityWarning:
		return domain.PriorityHigh
	}
	return domain.PriorityDefault
}

// Observe implements kernel.ObservationSink. Observations are always tracked
// (so enabling notifications later does not report old problems); a
// confirmed change is dispatched unless a maintenance window covers it.
func (uc *notifier) Observe(ctx context.Context, o kernel.Observation) {
	s, err := uc.current(ctx)
	if err != nil {
		return
	}
	relevant, healthy := s.Rules.Judge(o)
	if !relevant {
		return
	}
	key := o.Kind + ":" + o.Key
	uc.mu.Lock()
	track, changed := uc.tracks[key].Advance(healthy)
	uc.tracks[key] = track
	uc.mu.Unlock()
	if !changed || uc.UnderMaintenance(ctx, o.Kind, o.Key, o.Label, uc.d.Clock.Now()) {
		return
	}
	msg := domain.MessageFor(o, healthy)
	msg.Type, msg.Severity = o.Kind, s.Policy.SeverityOf(o.Kind)
	if healthy {
		msg.Severity = domain.SeverityInfo
	}
	uc.dispatch(ctx, s, msg)
}

// Alert implements kernel.AlertSink: one-off messages such as security alerts.
func (uc *notifier) Alert(ctx context.Context, a kernel.Alert) {
	s, err := uc.current(ctx)
	if err != nil {
		return
	}
	if a.Kind == "security" && !s.Rules.Security {
		return
	}
	msg := domain.Message{Title: a.Title, Body: a.Body, Type: a.Kind, Severity: s.Policy.SeverityOf(a.Kind), Tags: []string{domain.AlertTag(a.Kind)}}
	if a.Urgent {
		msg.Severity, msg.Tags = domain.SeverityCritical, []string{"rotating_light"}
	}
	uc.dispatch(ctx, s, msg)
}

// dispatch applies quiet hours and delivers in the background, so slow
// push servers never delay the checks that produce observations.
func (uc *notifier) dispatch(ctx context.Context, s domain.Settings, msg domain.Message) {
	if msg.Priority == 0 {
		msg.Priority = priorityFor(msg.Severity)
	}
	now := uc.d.Clock.Now()
	if s.Policy.Decide(msg.Severity, now.In(uc.d.Location)) == domain.Defer && uc.d.Pending != nil {
		if err := uc.d.Pending.Add(ctx, msg, now); err == nil {
			return
		}
	}
	uc.wg.Add(1)
	go func() {
		defer uc.wg.Done()
		dctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), deliveryTimeout)
		defer cancel()
		uc.deliver(dctx, s, msg)
	}()
}

// deliver sends msg to ntfy (when enabled) and every channel that routes it.
func (uc *notifier) deliver(ctx context.Context, s domain.Settings, msg domain.Message) {
	if s.Enabled && s.Topic != "" {
		uc.record(uc.d.Sender.Send(ctx, s, msg))
	}
	if uc.d.Channels == nil {
		return
	}
	channels, err := uc.d.Channels.List(ctx)
	if err != nil {
		uc.record(err)
		return
	}
	for _, c := range channels {
		if p, ok := uc.d.Publishers[c.Kind]; ok && c.Accepts(msg) {
			uc.record(p.Publish(ctx, c, msg))
		}
	}
}

func (uc *notifier) record(err error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if err != nil {
		uc.lastError = kernel.TargetFailure(kernel.ErrUnavailable, err).Message
		return
	}
	uc.lastError, uc.lastDelivered = "", uc.d.Clock.Now()
}

// Wait blocks until pending deliveries finished (used by tests).
func (uc *notifier) Wait() { uc.wg.Wait() }

// FlushDigest implements Notifier.
func (uc *notifier) FlushDigest(ctx context.Context) error {
	if uc.d.Pending == nil {
		return nil
	}
	s, err := uc.current(ctx)
	if err != nil {
		return err
	}
	if s.Policy.Quiet.Active(uc.d.Clock.Now().In(uc.d.Location)) {
		return nil
	}
	msgs, err := uc.d.Pending.Take(ctx)
	if err != nil || len(msgs) == 0 {
		return err
	}
	digest := domain.Message{Type: "digest", Severity: domain.SeverityInfo, Tags: []string{"bell"},
		Title: fmt.Sprintf("%d notifications during quiet hours", len(msgs))}
	lines := make([]string, 0, len(msgs))
	for i, m := range msgs {
		if m.Severity.Rank() > digest.Severity.Rank() {
			digest.Severity = m.Severity
		}
		if i < maxDigestLines {
			line := "• " + m.Title
			if m.Body != "" {
				line += ": " + m.Body
			}
			lines = append(lines, line)
		}
	}
	if len(msgs) > maxDigestLines {
		lines = append(lines, fmt.Sprintf("… and %d more", len(msgs)-maxDigestLines))
	}
	digest.Body, digest.Priority = strings.Join(lines, "\n"), priorityFor(digest.Severity)
	uc.deliver(ctx, s, digest)
	return nil
}

// Settings implements Notifier.
func (uc *notifier) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.current(ctx)
	if err != nil {
		return SettingsView{}, err
	}
	return uc.view(s), nil
}

func (uc *notifier) view(s domain.Settings) SettingsView {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return SettingsView{Enabled: s.Enabled, ServerURL: s.ServerURL, Topic: s.Topic, HasToken: s.Token != "",
		Rules: s.Rules, Policy: s.Policy, LastError: uc.lastError, LastDelivered: uc.lastDelivered}
}

// SaveSettings implements Notifier.
func (uc *notifier) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	old, err := uc.current(ctx)
	if err != nil {
		return SettingsView{}, err
	}
	token := in.Token
	if token == "" && old.Token != "" {
		server := in.ServerURL
		if strings.TrimSpace(server) == "" {
			server = domain.DefaultServerURL
		}
		if kernel.TargetChanged([]string{old.ServerURL}, []string{server}) {
			return SettingsView{}, kernel.SecretRequired("token")
		}
		token = old.Token
	}
	s, err := domain.NewSettings(in.Enabled, in.ServerURL, in.Topic, token, in.Rules)
	if err != nil {
		return SettingsView{}, err
	}
	if err := in.Policy.Quiet.Validate(); err != nil {
		return SettingsView{}, err
	}
	for t, sev := range in.Policy.Severities {
		if !sev.Valid() {
			delete(in.Policy.Severities, t)
		}
	}
	s.Policy = in.Policy
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save notification settings: %w", err)
	}
	uc.mu.Lock()
	uc.settings = &s
	uc.mu.Unlock()
	return uc.view(s), nil
}

// ErrNotConfigured is returned by SendTest without a topic.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrInvalid, "Please enter and save a topic first.")

func testMessage() domain.Message {
	return domain.Message{Title: "trisse-home test", Body: "Notifications work. 🎉", Priority: domain.PriorityDefault,
		Tags: []string{"tada"}, Type: "test", Severity: domain.SeverityInfo}
}

// SendTest implements Notifier; it delivers synchronously so errors are shown.
func (uc *notifier) SendTest(ctx context.Context) error {
	s, err := uc.current(ctx)
	if err != nil {
		return err
	}
	if s.Topic == "" {
		return ErrNotConfigured
	}
	err = uc.d.Sender.Send(ctx, s, testMessage())
	uc.record(err)
	if err != nil {
		var pe *kernel.PublicError
		if errors.As(err, &pe) {
			return err
		}
		return kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	return nil
}

// ErrChannelNotFound is returned for unknown channels.
var ErrChannelNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The channel does not exist.")

// Channels implements Notifier.
func (uc *notifier) Channels(ctx context.Context) ([]ChannelView, error) {
	list, err := uc.d.Channels.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	out := make([]ChannelView, 0, len(list))
	for _, c := range list {
		out = append(out, channelView(c))
	}
	return out, nil
}

func channelView(c domain.Channel) ChannelView {
	v := ChannelView{Channel: c, HasSecret: c.Secret != ""}
	v.Secret = ""
	return v
}

func (uc *notifier) channel(ctx context.Context, id string) (domain.Channel, error) {
	list, err := uc.d.Channels.List(ctx)
	if err != nil {
		return domain.Channel{}, fmt.Errorf("list channels: %w", err)
	}
	for _, c := range list {
		if c.ID == id {
			return c, nil
		}
	}
	return domain.Channel{}, ErrChannelNotFound
}

// SaveChannel implements Notifier.
func (uc *notifier) SaveChannel(ctx context.Context, in ChannelInput) (ChannelView, error) {
	c := in.Channel
	if c.ID == "" {
		list, err := uc.d.Channels.List(ctx)
		if err != nil {
			return ChannelView{}, fmt.Errorf("list channels: %w", err)
		}
		if len(list) >= domain.MaxChannels {
			return ChannelView{}, kernel.NewPublicError(kernel.ErrInvalid, "There are already 10 channels.")
		}
		c.ID = uc.d.IDs.NewID()
	} else {
		old, err := uc.channel(ctx, c.ID)
		if err != nil {
			return ChannelView{}, err
		}
		if c.Secret == "" {
			// URL (Gotify, webhook) or SMTP server and user identify the target.
			if old.Secret != "" && kernel.TargetChanged([]string{old.URL, old.SMTPHost, old.Username},
				[]string{c.URL, c.SMTPHost, c.Username}) {
				return ChannelView{}, kernel.SecretRequired("secret")
			}
			c.Secret = old.Secret
		}
		c.Kind = old.Kind
	}
	if err := c.Validate(); err != nil {
		return ChannelView{}, err
	}
	if err := uc.d.Channels.Save(ctx, c); err != nil {
		return ChannelView{}, fmt.Errorf("save channel: %w", err)
	}
	return channelView(c), nil
}

// DeleteChannel implements Notifier.
func (uc *notifier) DeleteChannel(ctx context.Context, id string) error {
	if _, err := uc.channel(ctx, id); err != nil {
		return err
	}
	return uc.d.Channels.Delete(ctx, id)
}

// TestChannel implements Notifier; it ignores the routing rules.
func (uc *notifier) TestChannel(ctx context.Context, id string) error {
	c, err := uc.channel(ctx, id)
	if err != nil {
		return err
	}
	p, ok := uc.d.Publishers[c.Kind]
	if !ok {
		return ErrChannelNotFound
	}
	if err := p.Publish(ctx, c, testMessage()); err != nil {
		var pe *kernel.PublicError
		if errors.As(err, &pe) {
			return err
		}
		return kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	return nil
}

// Maintenance implements Notifier.
func (uc *notifier) Maintenance(ctx context.Context) ([]domain.MaintenanceWindow, error) {
	list, err := uc.d.Maintenance.ListEndingAfter(ctx, uc.d.Clock.Now())
	if err != nil {
		return nil, fmt.Errorf("list maintenance: %w", err)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Start.Before(list[j].Start) })
	return list, nil
}

// StartMaintenance implements Notifier.
func (uc *notifier) StartMaintenance(ctx context.Context, scope []string, start, end time.Time, note string) (domain.MaintenanceWindow, error) {
	w, err := domain.NewMaintenanceWindow(uc.d.IDs.NewID(), scope, start, end, note)
	if err != nil {
		return domain.MaintenanceWindow{}, err
	}
	if err := uc.d.Maintenance.Save(ctx, w); err != nil {
		return domain.MaintenanceWindow{}, fmt.Errorf("save maintenance: %w", err)
	}
	return w, nil
}

// EndMaintenance implements Notifier: an active window ends now, a future
// one is removed.
func (uc *notifier) EndMaintenance(ctx context.Context, id string) error {
	list, err := uc.Maintenance(ctx)
	if err != nil {
		return err
	}
	now := uc.d.Clock.Now()
	for _, w := range list {
		if w.ID != id {
			continue
		}
		if w.ActiveAt(now) {
			w.End = now
			return uc.d.Maintenance.Save(ctx, w)
		}
		return uc.d.Maintenance.Delete(ctx, id)
	}
	return kernel.NewPublicError(kernel.ErrNotFound, "The maintenance window does not exist.")
}

// UnderMaintenance implements Notifier.
func (uc *notifier) UnderMaintenance(ctx context.Context, kind, key, label string, at time.Time) bool {
	if uc.d.Maintenance == nil {
		return false
	}
	list, err := uc.d.Maintenance.ListEndingAfter(ctx, at)
	if err != nil {
		return false
	}
	for _, w := range list {
		if w.ActiveAt(at) && w.Covers(kind, key, label) {
			return true
		}
	}
	return false
}
