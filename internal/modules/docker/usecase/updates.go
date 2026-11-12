package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"trisse-home/internal/modules/docker/domain"
	"trisse-home/internal/shared/kernel"
)

// LocalImage is the image of one or more running containers.
type LocalImage struct {
	Ref         string   // as the containers reference it, e.g. nginx:1.27
	Containers  []string // container names
	RepoDigests []string
	Labels      map[string]string
}

// LocalImageReader reads the images of running containers through the
// read-only socket proxy (output port).
type LocalImageReader interface {
	RunningImages(ctx context.Context) ([]LocalImage, error)
}

// RegistryCredential authenticates against one registry host.
type RegistryCredential struct {
	Host, Username, Password string
}

// RegistryClient asks registries via the OCI distribution API (output port).
type RegistryClient interface {
	// Digest returns the manifest digest of the reference's tag.
	Digest(ctx context.Context, ref domain.ImageRef, cred *RegistryCredential) (string, error)
	// Tags lists the repository's tags.
	Tags(ctx context.Context, ref domain.ImageRef, cred *RegistryCredential) ([]string, error)
}

// ImageUpdate is a cached check result.
type ImageUpdate struct {
	Ref         string
	LocalDigest string // joined repo digests at check time
	Status      domain.UpdateStatus
	NewerTag    string
	ReleaseURL  string
	CheckedAt   time.Time
	Error       string
}

// UpdateCache stores check results (output port).
type UpdateCache interface {
	List(ctx context.Context) ([]ImageUpdate, error)
	Put(ctx context.Context, u ImageUpdate) error
	// Retain deletes results of images that no longer run.
	Retain(ctx context.Context, refs []string) error
}

// UpdateSettings configure the update check; it is off by default.
type UpdateSettings struct {
	Enabled     bool
	SemverTags  bool
	Credentials []RegistryCredential
}

// UpdateSettingsRepository persists the settings with encrypted passwords
// (output port).
type UpdateSettingsRepository interface {
	Load(ctx context.Context) (UpdateSettings, error)
	Save(ctx context.Context, s UpdateSettings) error
}

// ImageUpdateView is one image with its update state.
type ImageUpdateView struct {
	ImageUpdate
	Containers []string
}

// UpdateSettingsView never contains passwords.
type UpdateSettingsView struct {
	Enabled    bool
	SemverTags bool
	Registries []RegistryView
}

// RegistryView is a stored credential without its password.
type RegistryView struct {
	Host, Username string
	HasPassword    bool
}

// UpdateSettingsInput is the input of SaveSettings; an empty password keeps
// the stored one of the same host.
type UpdateSettingsInput struct {
	Enabled     bool
	SemverTags  bool
	Credentials []RegistryCredential
}

// ImageUpdates is the input port of the update hints (#107).
type ImageUpdates interface {
	// Check asks the registries for images not checked recently (job).
	Check(ctx context.Context) error
	List(ctx context.Context) ([]ImageUpdateView, error)
	Settings(ctx context.Context) (UpdateSettingsView, error)
	SaveSettings(ctx context.Context, in UpdateSettingsInput) (UpdateSettingsView, error)
}

// UpdateDeps bundles the ports of the update hints.
type UpdateDeps struct {
	Images   LocalImageReader
	Registry RegistryClient
	Cache    UpdateCache
	Settings UpdateSettingsRepository
	Sink     kernel.ObservationSink
	Clock    kernel.Clock
}

type imageUpdates struct{ d UpdateDeps }

// NewImageUpdates returns the update hint use cases.
func NewImageUpdates(d UpdateDeps) ImageUpdates {
	if d.Sink == nil {
		d.Sink = kernel.DiscardObservations{}
	}
	return &imageUpdates{d: d}
}

// ErrUpdatesDisabled is returned while the update check is off.
var ErrUpdatesDisabled = kernel.NewPublicError(kernel.ErrNotFound, "The image update check is disabled. Enable it in Settings → Docker updates.")

// Check implements ImageUpdates.
func (uc *imageUpdates) Check(ctx context.Context) error {
	s, err := uc.d.Settings.Load(ctx)
	if err != nil {
		return fmt.Errorf("load update settings: %w", err)
	}
	if !s.Enabled {
		return nil
	}
	images, err := uc.d.Images.RunningImages(ctx)
	if err != nil {
		return fmt.Errorf("read local images: %w", err)
	}
	cached, err := uc.d.Cache.List(ctx)
	if err != nil {
		return fmt.Errorf("load update cache: %w", err)
	}
	byRef := map[string]ImageUpdate{}
	for _, c := range cached {
		byRef[c.Ref] = c
	}
	now := uc.d.Clock.Now()
	refs := make([]string, 0, len(images))
	var errs []error
	for _, img := range images {
		refs = append(refs, img.Ref)
		local := strings.Join(img.RepoDigests, ",")
		if c, ok := byRef[img.Ref]; ok && c.LocalDigest == local && now.Sub(c.CheckedAt) < domain.UpdateCheckInterval {
			continue // rate limits: at most every 6 h per image
		}
		u := uc.check(ctx, img, s)
		u.LocalDigest, u.CheckedAt = local, now
		if err := uc.d.Cache.Put(ctx, u); err != nil {
			errs = append(errs, err)
			continue
		}
		if u.Status == domain.UpToDate || u.Status == domain.UpdateAvailable {
			o := kernel.Observation{Kind: "image", Key: img.Ref, Label: img.Ref, Healthy: u.Status == domain.UpToDate}
			if !o.Healthy {
				o.Detail = "A newer image is available for " + strings.Join(img.Containers, ", ")
			}
			uc.d.Sink.Observe(ctx, o)
		}
	}
	if err := uc.d.Cache.Retain(ctx, refs); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (uc *imageUpdates) check(ctx context.Context, img LocalImage, s UpdateSettings) ImageUpdate {
	u := ImageUpdate{Ref: img.Ref, Status: domain.UpdateUnknown, ReleaseURL: domain.ReleaseNotesURL(img.Labels)}
	ref, ok := domain.ParseImageRef(img.Ref)
	switch {
	case !ok || len(img.RepoDigests) == 0:
		u.Status = domain.UpdateLocal
		return u
	case ref.Pinned():
		u.Status = domain.UpdatePinned
		return u
	}
	cred := credentialFor(s.Credentials, ref.Registry)
	remote, err := uc.d.Registry.Digest(ctx, ref, cred)
	if err != nil {
		u.Status, u.Error = domain.UpdateError, "The registry could not be asked: "+err.Error()
		return u
	}
	u.Status = domain.CompareDigest(img.RepoDigests, remote)
	if s.SemverTags && domain.IsSemverTag(ref.Tag) {
		if tags, err := uc.d.Registry.Tags(ctx, ref, cred); err == nil {
			u.NewerTag = domain.NewerTag(ref.Tag, tags)
		}
	}
	return u
}

func credentialFor(creds []RegistryCredential, registry string) *RegistryCredential {
	for i := range creds {
		if creds[i].Host == registry {
			return &creds[i]
		}
	}
	return nil
}

// List implements ImageUpdates: running images with their last result.
func (uc *imageUpdates) List(ctx context.Context) ([]ImageUpdateView, error) {
	s, err := uc.d.Settings.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load update settings: %w", err)
	}
	if !s.Enabled {
		return nil, ErrUpdatesDisabled
	}
	images, err := uc.d.Images.RunningImages(ctx)
	if err != nil {
		return nil, kernel.NewPublicError(kernel.ErrUnavailable, "The Docker API is not reachable. Does the socket proxy allow IMAGES=1?")
	}
	cached, err := uc.d.Cache.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("load update cache: %w", err)
	}
	byRef := map[string]ImageUpdate{}
	for _, c := range cached {
		byRef[c.Ref] = c
	}
	out := make([]ImageUpdateView, 0, len(images))
	for _, img := range images {
		u, ok := byRef[img.Ref]
		if !ok {
			u = ImageUpdate{Ref: img.Ref, Status: domain.UpdateUnknown, ReleaseURL: domain.ReleaseNotesURL(img.Labels)}
		}
		out = append(out, ImageUpdateView{ImageUpdate: u, Containers: img.Containers})
	}
	// Updates first, then by reference.
	sort.SliceStable(out, func(i, j int) bool {
		ai, aj := out[i].Status == domain.UpdateAvailable, out[j].Status == domain.UpdateAvailable
		if ai != aj {
			return ai
		}
		return out[i].Ref < out[j].Ref
	})
	return out, nil
}

// Settings implements ImageUpdates.
func (uc *imageUpdates) Settings(ctx context.Context) (UpdateSettingsView, error) {
	s, err := uc.d.Settings.Load(ctx)
	if err != nil {
		return UpdateSettingsView{}, fmt.Errorf("load update settings: %w", err)
	}
	return settingsView(s), nil
}

func settingsView(s UpdateSettings) UpdateSettingsView {
	v := UpdateSettingsView{Enabled: s.Enabled, SemverTags: s.SemverTags, Registries: make([]RegistryView, 0, len(s.Credentials))}
	for _, c := range s.Credentials {
		v.Registries = append(v.Registries, RegistryView{Host: c.Host, Username: c.Username, HasPassword: c.Password != ""})
	}
	return v
}

const maxCredentials = 20

// SaveSettings implements ImageUpdates.
func (uc *imageUpdates) SaveSettings(ctx context.Context, in UpdateSettingsInput) (UpdateSettingsView, error) {
	old, err := uc.d.Settings.Load(ctx)
	if err != nil {
		return UpdateSettingsView{}, fmt.Errorf("load update settings: %w", err)
	}
	if len(in.Credentials) > maxCredentials {
		return UpdateSettingsView{}, kernel.NewValidationError("credentials", "At most 20 registries can be stored.")
	}
	s := UpdateSettings{Enabled: in.Enabled, SemverTags: in.SemverTags}
	seen := map[string]bool{}
	for _, c := range in.Credentials {
		c.Host = strings.ToLower(strings.TrimSpace(c.Host))
		c.Username = strings.TrimSpace(c.Username)
		if c.Host == "" || strings.ContainsAny(c.Host, "/ ") || c.Username == "" {
			return UpdateSettingsView{}, kernel.NewValidationError("credentials", "Each registry needs a host name (e.g. ghcr.io) and a user name.")
		}
		if seen[c.Host] {
			return UpdateSettingsView{}, kernel.NewValidationError("credentials", "Each registry may only be listed once.")
		}
		seen[c.Host] = true
		if c.Password == "" {
			if prev := credentialFor(old.Credentials, c.Host); prev != nil {
				c.Password = prev.Password
			}
		}
		if c.Password == "" {
			return UpdateSettingsView{}, kernel.NewValidationError("credentials", "A password or access token is required for new registries.")
		}
		s.Credentials = append(s.Credentials, c)
	}
	slices.SortFunc(s.Credentials, func(a, b RegistryCredential) int { return strings.Compare(a.Host, b.Host) })
	if err := uc.d.Settings.Save(ctx, s); err != nil {
		return UpdateSettingsView{}, fmt.Errorf("save update settings: %w", err)
	}
	return settingsView(s), nil
}
