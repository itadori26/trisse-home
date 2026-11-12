package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"trisse-home/internal/modules/backupmonitor/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ m map[string]domain.Source }

func (r *memRepo) List(context.Context) ([]domain.Source, error) {
	var out []domain.Source
	for _, s := range r.m {
		out = append(out, s)
	}
	return out, nil
}
func (r *memRepo) Get(_ context.Context, id string) (domain.Source, error) {
	s, ok := r.m[id]
	if !ok {
		return s, domain.ErrSourceNotFound
	}
	return s, nil
}
func (r *memRepo) GetByToken(_ context.Context, h domain.TokenHash) (domain.Source, error) {
	for _, s := range r.m {
		if s.TokenHash == h {
			return s, nil
		}
	}
	return domain.Source{}, domain.ErrSourceNotFound
}
func (r *memRepo) Save(_ context.Context, s domain.Source) error { r.m[s.ID] = s; return nil }
func (r *memRepo) Delete(_ context.Context, id string) error     { delete(r.m, id); return nil }

type seq struct{ n int }

func (s *seq) NewID() string             { s.n++; return fmt.Sprint("id", s.n) }
func (s *seq) NewToken() (string, error) { s.n++; return fmt.Sprint("token", s.n), nil }

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { return c.t }

type recorder struct{ obs []kernel.Observation }

func (r *recorder) Observe(_ context.Context, o kernel.Observation) { r.obs = append(r.obs, o) }

func TestMonitor(t *testing.T) {
	ctx := context.Background()
	c := &clk{t: time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)}
	rec := &recorder{}
	ids := &seq{}
	uc := NewMonitor(Deps{Repo: &memRepo{m: map[string]domain.Source{}}, Tokens: ids, Clock: c, IDs: ids, Observations: rec})

	created, err := uc.Create(ctx, SourceInput{Name: "NAS", Schedule: domain.Schedule{MaxAge: 24 * time.Hour}})
	if err != nil || created.Token == "" || created.Source.State != "waiting" {
		t.Fatalf("create: %+v %v", created, err)
	}
	if err := uc.Ping(ctx, "wrong", PingSuccess, ""); !errors.Is(err, domain.ErrSourceNotFound) {
		t.Fatalf("wrong token accepted: %v", err)
	}
	if err := uc.Ping(ctx, created.Token, PingFail, "disk full"); err != nil {
		t.Fatal(err)
	}
	if len(rec.obs) != 1 || rec.obs[0].Healthy || rec.obs[0].Kind != "backup" {
		t.Fatalf("observation %+v", rec.obs)
	}
	rotated, _ := uc.RotateToken(ctx, created.Source.ID)
	if err := uc.Ping(ctx, created.Token, PingSuccess, ""); !errors.Is(err, domain.ErrSourceNotFound) {
		t.Fatal("old token still valid after rotation")
	}
	if err := uc.Ping(ctx, rotated.Token, PingSuccess, "done"); err != nil {
		t.Fatal(err)
	}
	list, _ := uc.List(ctx)
	if len(list) != 1 || list[0].State != "ok" || list[0].LastMessage != "done" {
		t.Fatalf("list %+v", list)
	}
	c.t = c.t.Add(30 * time.Hour)
	_ = uc.Report(ctx)
	if last := rec.obs[len(rec.obs)-1]; last.Healthy {
		t.Fatalf("overdue source reported healthy: %+v", last)
	}
}
