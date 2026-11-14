package usecase

import (
	"context"
	"time"

	"trisse-home/internal/modules/parcels/domain"
)

// MailState is where the mail discovery stopped (#166).
type MailState struct {
	// UIDValidity and LastUID identify the last message read; a new
	// UIDVALIDITY means the folder was rebuilt and the first run repeats.
	UIDValidity uint32
	LastUID     uint32
	LastRun     time.Time
	LastError   string
	// Found counts the parcels discovered so far.
	Found int
}

// MailSource reads new messages of the configured folder, read-only
// (output port).
type MailSource interface {
	// Fetch returns the messages after st; on the first run (or a new
	// UIDVALIDITY) the messages since since. It returns the new position.
	Fetch(ctx context.Context, s domain.MailSettings, st MailState, since time.Time) ([]domain.Mail, MailState, error)
	// Test logs in and opens the folder.
	Test(ctx context.Context, s domain.MailSettings) error
}

// MailStore persists the mailbox settings, the position and the numbers the
// user removed (output port).
type MailStore interface {
	LoadMail(ctx context.Context) (domain.MailSettings, MailState, error)
	SaveMail(ctx context.Context, s domain.MailSettings, st MailState) error
	Dismissed(ctx context.Context, number string) (bool, error)
	Dismiss(ctx context.Context, number string) error
}

// MailSettingsView never contains the password.
type MailSettingsView struct {
	Enabled         bool
	Host            string
	Port            int
	Username        string
	HasPassword     bool
	Folder          string
	IntervalMinutes int
	LastRun         time.Time
	LastError       string
	Found           int
}

// MailSettingsInput is the input of SaveMailSettings; an empty password
// keeps the stored one for the same server and user.
type MailSettingsInput struct {
	Enabled         bool
	Host            string
	Port            int
	Username        string
	Password        string
	Folder          string
	IntervalMinutes int
}
