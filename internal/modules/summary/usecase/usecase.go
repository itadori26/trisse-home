package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"trisse-home/internal/modules/summary/domain"
	"trisse-home/internal/shared/kernel"
)

// Summarizer asks a language model (output port; Ollama today, swappable).
type Summarizer interface {
	Summarize(ctx context.Context, s domain.Settings, prompt string) (string, error)
	// Models lists the installed models; it doubles as the connection test.
	Models(ctx context.Context, url string) ([]string, error)
	// Stream asks the model and hands the answer over piece by piece; it
	// stops when ctx is cancelled or onChunk fails.
	Stream(ctx context.Context, s domain.Settings, prompt string, onChunk func(string) error) error
}

// ArticleFetcher loads the readable text of a web page (output port), for
// feed items that only carry a link.
type ArticleFetcher interface {
	Text(ctx context.Context, url string) (string, error)
}

// Cached is a stored summary.
type Cached struct {
	Text      string
	Model     string
	CreatedAt time.Time
}

// Repository stores the settings and the summaries (output port). Load
// returns domain.ErrNotConfigured when nothing is set up.
type Repository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
	Clear(ctx context.Context) error
	Cached(ctx context.Context, key string) (Cached, bool, error)
	Store(ctx context.Context, key string, c Cached) error
}

// Request asks for a summary of a text or of the page behind a link.
type Request struct {
	Text     string
	URL      string
	Language string
}

// Result is a summary.
type Result struct {
	Text      string
	Model     string
	Cached    bool
	Truncated bool
}

// SettingsView describes the setup; Configured=false hides the feature.
type SettingsView struct {
	Configured    bool
	URL           string
	Model         string
	MaxInputChars int
}

// Summary is the input port of the module.
type Summary interface {
	Settings(ctx context.Context) (SettingsView, error)
	// SaveSettings tests the connection and that the model is installed.
	SaveSettings(ctx context.Context, url, model string, maxInput int) (SettingsView, error)
	Disable(ctx context.Context) error
	// Models lists the models of an Ollama server (to pick one).
	Models(ctx context.Context, url string) ([]string, error)
	// Summarize returns the cached summary or asks the model (cache-aside).
	Summarize(ctx context.Context, r Request) (Result, error)
	// Ask answers a free question of the smart search (#162), streamed. The
	// question is validated before the first chunk, so errors before that
	// are ordinary errors.
	Ask(ctx context.Context, question, context, language string, onChunk func(string) error) error
}

// Deps bundles the output ports.
type Deps struct {
	Repo    Repository
	Model   Summarizer
	Fetcher ArticleFetcher
	Clock   kernel.Clock
}

type summary struct{ d Deps }

// NewSummary returns the use cases.
func NewSummary(d Deps) Summary { return &summary{d: d} }

func (uc *summary) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.d.Repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{MaxInputChars: domain.DefaultMaxInputChars}, nil
	}
	if err != nil {
		return SettingsView{}, err
	}
	return view(s), nil
}

func (uc *summary) SaveSettings(ctx context.Context, url, model string, maxInput int) (SettingsView, error) {
	s, err := domain.NewSettings(url, model, maxInput)
	if err != nil {
		return SettingsView{}, err
	}
	models, err := uc.d.Model.Models(ctx, s.URL)
	if err != nil {
		return SettingsView{}, kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	if !slices.Contains(models, s.Model) {
		return SettingsView{}, kernel.NewValidationErrorf("model", "The model {model} is not installed on this Ollama server.", "model", s.Model)
	}
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	return view(s), nil
}

func (uc *summary) Disable(ctx context.Context) error { return uc.d.Repo.Clear(ctx) }

func (uc *summary) Models(ctx context.Context, url string) ([]string, error) {
	s, err := domain.NewSettings(url, "probe", 0)
	if err != nil {
		return nil, err
	}
	models, err := uc.d.Model.Models(ctx, s.URL)
	if err != nil {
		return nil, kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	return models, nil
}

func (uc *summary) Summarize(ctx context.Context, r Request) (Result, error) {
	s, err := uc.d.Repo.Load(ctx)
	if err != nil {
		return Result{}, err
	}
	text := r.Text
	if r.URL != "" && text == "" {
		u, err := kernel.ParseTarget("url", r.URL, kernel.WithTargetQuery())
		if err != nil {
			return Result{}, err
		}
		if text, err = uc.d.Fetcher.Text(ctx, u.String()); err != nil {
			return Result{}, kernel.NewPublicErrorf(kernel.ErrUnavailable, "The article could not be loaded: {detail}", "detail", reason(err))
		}
	}
	in, err := domain.NewInput(text, s.MaxInputChars)
	if err != nil {
		return Result{}, err
	}
	lang := domain.ParseLanguage(r.Language)
	key := domain.CacheKey(s.Model, lang, in.Text)
	if c, ok, err := uc.d.Repo.Cached(ctx, key); err != nil {
		return Result{}, err
	} else if ok {
		return Result{Text: c.Text, Model: c.Model, Cached: true, Truncated: in.Truncated}, nil
	}
	out, err := uc.d.Model.Summarize(ctx, s, domain.Prompt(lang, in.Text))
	if err != nil {
		return Result{}, kernel.NewPublicErrorf(kernel.ErrUnavailable, "Ollama could not summarise the text: {detail}", "detail", reason(err))
	}
	out = domain.CleanSummary(out)
	if out == "" {
		return Result{}, kernel.NewPublicError(kernel.ErrUnavailable, "Ollama returned an empty summary.")
	}
	if err := uc.d.Repo.Store(ctx, key, Cached{Text: out, Model: s.Model, CreatedAt: uc.d.Clock.Now()}); err != nil {
		return Result{}, fmt.Errorf("store summary: %w", err)
	}
	return Result{Text: out, Model: s.Model, Truncated: in.Truncated}, nil
}

func (uc *summary) Ask(ctx context.Context, question, context, language string, onChunk func(string) error) error {
	s, err := uc.d.Repo.Load(ctx)
	if err != nil {
		return err
	}
	q, err := domain.NewQuestion(question, context)
	if err != nil {
		return err
	}
	return uc.d.Model.Stream(ctx, s, domain.AskPrompt(domain.ParseLanguage(language), q), onChunk)
}

func view(s domain.Settings) SettingsView {
	return SettingsView{Configured: true, URL: s.URL, Model: s.Model, MaxInputChars: s.MaxInputChars}
}

func reason(err error) string {
	var pe *kernel.PublicError
	if errors.As(err, &pe) {
		return pe.Error()
	}
	msg := err.Error()
	if len(msg) > 160 {
		msg = msg[:160] + "…"
	}
	return msg
}
