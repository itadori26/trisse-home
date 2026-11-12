package gateway

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"trisse-home/internal/modules/backup/domain"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Cipher encrypts secrets before they are written to the database.
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// RemoteSettingsRepository stores the WebDAV target with an encrypted password.
type RemoteSettingsRepository struct {
	db     *sql.DB
	cipher Cipher
}

// NewRemoteSettingsRepository returns a repository backed by db.
func NewRemoteSettingsRepository(db *sql.DB, cipher Cipher) *RemoteSettingsRepository {
	return &RemoteSettingsRepository{db: db, cipher: cipher}
}

// Load implements usecase.RemoteSettingsRepository.
func (r *RemoteSettingsRepository) Load(ctx context.Context) (domain.RemoteSettings, error) {
	var s domain.RemoteSettings
	var enc string
	err := r.db.QueryRowContext(ctx, `SELECT url, username, password_encrypted, retention FROM backup_remote WHERE id = 1`).
		Scan(&s.URL, &s.Username, &enc, &s.Retention)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.RemoteSettings{}, domain.ErrRemoteNotConfigured
	}
	if err != nil {
		return domain.RemoteSettings{}, fmt.Errorf("select off-site settings: %w", err)
	}
	if s.Password, err = r.cipher.Decrypt(enc); err != nil {
		return domain.RemoteSettings{}, fmt.Errorf("decrypt password (was SECRET_KEY changed?): %w", err)
	}
	return s, nil
}

// Save implements usecase.RemoteSettingsRepository.
func (r *RemoteSettingsRepository) Save(ctx context.Context, s domain.RemoteSettings) error {
	enc, err := r.cipher.Encrypt(s.Password)
	if err != nil {
		return fmt.Errorf("encrypt password: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO backup_remote (id, url, username, password_encrypted, retention) VALUES (1, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET url = excluded.url, username = excluded.username,
		password_encrypted = excluded.password_encrypted, retention = excluded.retention`, s.URL, s.Username, enc, s.Retention)
	if err != nil {
		return fmt.Errorf("upsert off-site settings: %w", err)
	}
	return nil
}

// Delete implements usecase.RemoteSettingsRepository.
func (r *RemoteSettingsRepository) Delete(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM backup_remote`)
	return err
}

// WebDAV implements usecase.RemoteStore (e.g. a Nextcloud folder).
type WebDAV struct{ client *http.Client }

// NewWebDAV returns a WebDAV client.
func NewWebDAV(client *http.Client) *WebDAV { return &WebDAV{client: client} }

const (
	maxListBytes = 4 << 20
	testFile     = ".trisse-home-write-test"
)

func (w *WebDAV) request(ctx context.Context, s domain.RemoteSettings, method, name string, body io.Reader, size int64, header map[string]string) (*http.Response, error) {
	target := s.URL + url.PathEscape(name)
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if size >= 0 && body != nil {
		req.ContentLength = size
	}
	req.SetBasicAuth(s.Username, s.Password)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	res, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("webdav %s: %w", method, err)
	}
	return res, nil
}

func expect(res *http.Response, codes ...int) error {
	for _, c := range codes {
		if res.StatusCode == c {
			return nil
		}
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 64<<10))
	switch res.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return errors.New("access denied – check user name and (app) password")
	case http.StatusNotFound, http.StatusConflict:
		return errors.New("the folder does not exist")
	case http.StatusInsufficientStorage:
		return errors.New("the target is full")
	}
	return fmt.Errorf("the server answered %s", res.Status)
}

// Test implements usecase.RemoteStore: the collection must exist and a
// small file must be writable and deletable.
func (w *WebDAV) Test(ctx context.Context, s domain.RemoteSettings) error {
	res, err := w.request(ctx, s, "PROPFIND", "", nil, -1, map[string]string{"Depth": "0"})
	if err != nil {
		return err
	}
	err = expect(res, http.StatusMultiStatus)
	_ = res.Body.Close()
	if err != nil {
		return err
	}
	if err := w.Put(ctx, s, testFile, strings.NewReader("ok"), 2); err != nil {
		return err
	}
	return w.Delete(ctx, s, testFile)
}

// Put implements usecase.RemoteStore.
func (w *WebDAV) Put(ctx context.Context, s domain.RemoteSettings, name string, r io.Reader, size int64) error {
	res, err := w.request(ctx, s, http.MethodPut, name, r, size, map[string]string{"Content-Type": "application/octet-stream"})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return expect(res, http.StatusCreated, http.StatusNoContent, http.StatusOK)
}

// Delete implements usecase.RemoteStore.
func (w *WebDAV) Delete(ctx context.Context, s domain.RemoteSettings, name string) error {
	res, err := w.request(ctx, s, http.MethodDelete, name, nil, -1, nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return expect(res, http.StatusNoContent, http.StatusOK, http.StatusNotFound)
}

// Get implements usecase.RemoteStore.
func (w *WebDAV) Get(ctx context.Context, s domain.RemoteSettings, name string) (io.ReadCloser, error) {
	if !domain.ValidName(name) {
		return nil, domain.ErrBackupNotFound
	}
	res, err := w.request(ctx, s, http.MethodGet, name, nil, -1, nil)
	if err != nil {
		return nil, err
	}
	if err := expect(res, http.StatusOK); err != nil {
		_ = res.Body.Close()
		return nil, err
	}
	return res.Body, nil
}

type multistatus struct {
	Responses []struct {
		Href string `xml:"href"`
		Prop struct {
			Length   string `xml:"getcontentlength"`
			Modified string `xml:"getlastmodified"`
		} `xml:"propstat>prop"`
	} `xml:"response"`
}

const propfindBody = `<?xml version="1.0" encoding="utf-8"?><d:propfind xmlns:d="DAV:"><d:prop><d:getcontentlength/><d:getlastmodified/></d:prop></d:propfind>`

// List implements usecase.RemoteStore; only backup file names are returned.
func (w *WebDAV) List(ctx context.Context, s domain.RemoteSettings) ([]domain.File, error) {
	res, err := w.request(ctx, s, "PROPFIND", "", bytes.NewReader([]byte(propfindBody)), int64(len(propfindBody)),
		map[string]string{"Depth": "1", "Content-Type": "application/xml; charset=utf-8"})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := expect(res, http.StatusMultiStatus); err != nil {
		return nil, err
	}
	var ms multistatus
	if err := xml.NewDecoder(io.LimitReader(res.Body, maxListBytes)).Decode(&ms); err != nil {
		return nil, fmt.Errorf("decode folder listing: %w", err)
	}
	var out []domain.File
	for _, r := range ms.Responses {
		p, err := url.PathUnescape(r.Href)
		if err != nil {
			continue
		}
		name := path.Base(strings.TrimSuffix(p, "/"))
		if !domain.ValidName(name) {
			continue
		}
		f := domain.File{Name: name}
		f.Size, _ = strconv.ParseInt(r.Prop.Length, 10, 64) // optional property
		if created, ok := domain.CreatedFromName(name); ok {
			f.CreatedAt = created
		} else if t, err := http.ParseTime(r.Prop.Modified); err == nil {
			f.CreatedAt = t
		}
		out = append(out, f)
	}
	return out, nil
}
