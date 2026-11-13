package usecase

import (
	"context"
	"fmt"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/shared/kernel"
)

// SubscriptionView is an ICS subscription without its full URL, which may
// contain a private token.
type SubscriptionView struct {
	ID, Name, URLHint, Color string
	RefreshMinutes           int
}

// SubscriptionInput carries the editable fields; an empty URL keeps the
// stored one when updating.
type SubscriptionInput struct {
	Name, URL, Color string
	RefreshMinutes   int
}

// Subscriptions is the input port for ICS subscriptions (#98).
type Subscriptions interface {
	ListSubscriptions(ctx context.Context) ([]SubscriptionView, error)
	AddSubscription(ctx context.Context, in SubscriptionInput) (SubscriptionView, error)
	UpdateSubscription(ctx context.Context, id string, in SubscriptionInput) (SubscriptionView, error)
	DeleteSubscription(ctx context.Context, id string) error
}

// ListSubscriptions implements Subscriptions.
func (uc *calendar) ListSubscriptions(ctx context.Context) ([]SubscriptionView, error) {
	subs, err := uc.subs.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	out := make([]SubscriptionView, 0, len(subs))
	for _, s := range subs {
		out = append(out, subscriptionView(s))
	}
	return out, nil
}

// AddSubscription implements Subscriptions; the feed is loaded once so a
// wrong URL is reported immediately.
func (uc *calendar) AddSubscription(ctx context.Context, in SubscriptionInput) (SubscriptionView, error) {
	subs, err := uc.subs.List(ctx)
	if err != nil {
		return SubscriptionView{}, fmt.Errorf("list subscriptions: %w", err)
	}
	if len(subs) >= domain.MaxSubscriptions {
		return SubscriptionView{}, domain.ErrTooManySubscriptions
	}
	s, err := domain.NewSubscription(uc.ids.NewID(), in.Name, in.URL, in.Color, minutes(in.RefreshMinutes))
	if err != nil {
		return SubscriptionView{}, err
	}
	return uc.saveSubscription(ctx, s, true)
}

// UpdateSubscription implements Subscriptions.
func (uc *calendar) UpdateSubscription(ctx context.Context, id string, in SubscriptionInput) (SubscriptionView, error) {
	old, err := uc.subs.Get(ctx, id)
	if err != nil {
		return SubscriptionView{}, err
	}
	rawURL := in.URL
	if rawURL == "" {
		rawURL = old.URL
	}
	s, err := domain.NewSubscription(id, in.Name, rawURL, in.Color, minutes(in.RefreshMinutes))
	if err != nil {
		return SubscriptionView{}, err
	}
	return uc.saveSubscription(ctx, s, s.URL != old.URL)
}

// DeleteSubscription implements Subscriptions.
func (uc *calendar) DeleteSubscription(ctx context.Context, id string) error {
	if _, err := uc.subs.Get(ctx, id); err != nil {
		return err
	}
	if err := uc.subs.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	_ = uc.Refresh(ctx) // the agenda no longer shows its events
	return nil
}

func (uc *calendar) saveSubscription(ctx context.Context, s domain.Subscription, test bool) (SubscriptionView, error) {
	if test {
		now := uc.clock.Now()
		if _, err := uc.feeds.Events(ctx, s, now, now.Add(domain.Horizon)); err != nil {
			return SubscriptionView{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "Could not load the ICS calendar: {detail}", "detail", err.Error())
		}
	}
	if err := uc.subs.Save(ctx, s); err != nil {
		return SubscriptionView{}, fmt.Errorf("save subscription: %w", err)
	}
	_ = uc.Refresh(ctx) // errors are reported through AgendaResponse.Error
	return subscriptionView(s), nil
}

func minutes(n int) time.Duration { return time.Duration(n) * time.Minute }

func subscriptionView(s domain.Subscription) SubscriptionView {
	return SubscriptionView{ID: s.ID, Name: s.Name, URLHint: s.URLHint(), Color: s.Color, RefreshMinutes: int(s.RefreshInterval / time.Minute)}
}
