package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/shared/kernel"
)

// MailSettings implements Parcels.
func (uc *parcels) MailSettings(ctx context.Context) (MailSettingsView, error) {
	s, st, err := uc.d.MailStore.LoadMail(ctx)
	if err != nil {
		return MailSettingsView{}, fmt.Errorf("load mail settings: %w", err)
	}
	return mailView(s, st), nil
}

// SaveMailSettings implements Parcels. An enabled mailbox is tested first,
// so typos are reported right away.
func (uc *parcels) SaveMailSettings(ctx context.Context, in MailSettingsInput) (MailSettingsView, error) {
	old, st, err := uc.d.MailStore.LoadMail(ctx)
	if err != nil {
		return MailSettingsView{}, fmt.Errorf("load mail settings: %w", err)
	}
	host := in.Host
	password := in.Password
	if password == "" && old.Password != "" {
		// The stored password only goes to the same server and user (#191).
		if kernel.TargetChanged([]string{old.Host, old.Username}, []string{host, in.Username}) {
			return MailSettingsView{}, kernel.SecretRequired("password")
		}
		password = old.Password
	}
	s, err := domain.NewMailSettings(in.Enabled, host, in.Port, in.Username, password, in.Folder, time.Duration(in.IntervalMinutes)*time.Minute)
	if err != nil {
		return MailSettingsView{}, err
	}
	if s.Enabled {
		if err := uc.d.Mail.Test(ctx, s); err != nil {
			return MailSettingsView{}, kernel.TargetFailure(kernel.ErrInvalid, err)
		}
	}
	// Another mailbox or folder starts over (looking back FirstRunDays).
	if kernel.TargetChanged([]string{old.Host, old.Username, old.Folder}, []string{s.Host, s.Username, s.Folder}) {
		st = MailState{Found: st.Found}
	}
	st.LastRun, st.LastError = time.Time{}, "" // check right away
	if err := uc.d.MailStore.SaveMail(ctx, s, st); err != nil {
		return MailSettingsView{}, fmt.Errorf("save mail settings: %w", err)
	}
	return mailView(s, st), nil
}

// DiscoverParcels implements Parcels.
func (uc *parcels) DiscoverParcels(ctx context.Context) error {
	s, st, err := uc.d.MailStore.LoadMail(ctx)
	if err != nil {
		return fmt.Errorf("load mail settings: %w", err)
	}
	now := uc.d.Clock.Now()
	if !s.Enabled || (!st.LastRun.IsZero() && now.Sub(st.LastRun) < s.Interval) {
		return nil
	}
	mails, next, err := uc.d.Mail.Fetch(ctx, s, st, now.AddDate(0, 0, -domain.FirstRunDays))
	st.LastRun = now
	if err != nil {
		st.LastError = "The mailbox could not be read: " + publicReason(err)
		if serr := uc.d.MailStore.SaveMail(ctx, s, st); serr != nil {
			return errors.Join(err, serr)
		}
		return fmt.Errorf("read mailbox: %w", err)
	}
	next.LastRun, next.LastError, next.Found = now, "", st.Found
	added, addErr := uc.addCandidates(ctx, mails)
	next.Found += added
	if addErr != nil {
		next.LastError = publicReason(addErr)
	}
	if err := uc.d.MailStore.SaveMail(ctx, s, next); err != nil {
		return fmt.Errorf("save mail state: %w", err)
	}
	return addErr
}

// addCandidates adds the new numbers of the mails; known and dismissed
// numbers are skipped.
func (uc *parcels) addCandidates(ctx context.Context, mails []domain.Mail) (int, error) {
	all, err := uc.d.Repo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("list parcels: %w", err)
	}
	known := make([]string, 0, len(all))
	for _, p := range all {
		known = append(known, p.TrackingNumber)
	}
	added := 0
	for _, m := range mails {
		for _, c := range domain.ExtractCandidates(m) {
			if slices.Contains(known, c.Number) {
				continue
			}
			dismissed, err := uc.d.MailStore.Dismissed(ctx, c.Number)
			if err != nil {
				return added, fmt.Errorf("check dismissed: %w", err)
			}
			if dismissed {
				continue
			}
			if _, err := uc.add(ctx, c.Number, c.Carrier, c.Label, true); err != nil {
				return added, err
			}
			known = append(known, c.Number)
			added++
		}
	}
	return added, nil
}

func publicReason(err error) string {
	var pe *kernel.PublicError
	if errors.As(err, &pe) {
		return pe.Error()
	}
	var ve *kernel.ValidationError
	if errors.As(err, &ve) {
		return ve.Error()
	}
	msg := err.Error()
	if len(msg) > 200 {
		msg = msg[:200] + "…"
	}
	return msg
}

func mailView(s domain.MailSettings, st MailState) MailSettingsView {
	return MailSettingsView{Enabled: s.Enabled, Host: s.Host, Port: s.Port, Username: s.Username, HasPassword: s.Password != "",
		Folder: s.Folder, IntervalMinutes: int(s.Interval / time.Minute), LastRun: st.LastRun, LastError: st.LastError, Found: st.Found}
}
