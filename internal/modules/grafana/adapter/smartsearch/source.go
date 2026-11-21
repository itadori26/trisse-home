// Package smartsearch answers "grafana alerts" in the smart search
// (ADR 0026).
package smartsearch

import (
	"context"
	"fmt"

	"waylight/internal/modules/grafana/domain"
	"waylight/internal/modules/grafana/usecase"
	"waylight/internal/shared/answers"
)

// maxItems is the longest list in an answer.
const maxItems = 6

var intents = []answers.Intent{
	{ID: "grafana.alerts", Source: "grafana", Prefixes: []string{"grafana", "alarm", "alerts"}, Answer: answers.List,
		WidgetKind: "grafana-alerts",
		Terms: append(answers.Terms(3, "grafana", "alarm", "alarme", "alert", "alerts", "alerting"),
			answers.Terms(1, "monitoring", "firing", "aktiv", "active")...)},
}

var messages = answers.Messages{
	answers.English: {"title": "Grafana alerts: {n}", "empty": "No alert is firing.", "pending": "pending"},
	answers.German:  {"title": "Grafana-Alarme: {n}", "empty": "Kein Alarm ist aktiv.", "pending": "ausstehend"},
}

// Source answers from the alerts of Grafana.
type Source struct{ grafana usecase.Grafana }

// New returns the source on grafana.
func New(grafana usecase.Grafana) *Source { return &Source{grafana: grafana} }

// ID implements answers.Source.
func (*Source) ID() string { return "grafana" }

// Intents implements answers.Source.
func (*Source) Intents() []answers.Intent { return intents }

// Entities implements answers.Source.
func (*Source) Entities(context.Context) ([]answers.Entity, error) { return nil, nil }

// Resolve implements answers.Source.
func (s *Source) Resolve(ctx context.Context, r answers.Request) (answers.Answer, error) {
	list, err := s.grafana.Alerts(ctx, "")
	if err != nil {
		return answers.Answer{}, fmt.Errorf("grafana alerts: %w", err)
	}
	l := r.Lang
	a := answers.Answer{Kind: answers.List, Title: messages.T(l, "title", "n", len(list)), Empty: messages.T(l, "empty"), At: r.Now}
	for i, al := range list {
		if i >= maxItems {
			break
		}
		value := al.Summary
		state := answers.StateError
		if al.State == domain.Pending {
			state = answers.StateWarning
			value = messages.T(l, "pending") + joinIf(value)
		}
		if !al.Since.IsZero() {
			value = answers.Ago(l, al.Since, r.Now) + joinIf(value)
		}
		a.Items = append(a.Items, answers.Item{Label: al.Name, Value: value, State: state})
	}
	return a, nil
}

func joinIf(s string) string {
	if s == "" {
		return ""
	}
	return " · " + s
}
