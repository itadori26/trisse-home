package gateway

import (
	"context"
	"io/fs"
	"testing"

	"trisse-home/internal/modules/dashboard/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestTabRulesRoundTrip(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "dashboard", Migrations()))
	tab, _ := domain.NewTab("t1", "Evening", 0)
	if err := repo.CreateTab(ctx, tab); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.GetTab(ctx, "t1"); len(got.Rules()) != 0 {
		t.Fatal("new tabs have no rules")
	}
	rules, err := domain.NewTabRules([]domain.RuleInput{{Days: []int{5, 6}, From: "22:00", To: "02:00", Networks: []string{"10.13.13.0/24"}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateTab(ctx, tab.WithRules(rules)); err != nil {
		t.Fatal(err)
	}
	list, err := repo.ListTabs(ctx)
	if err != nil || len(list) != 1 || len(list[0].Rules()) != 1 {
		t.Fatalf("%+v %v", list, err)
	}
	r := list[0].Rules()[0]
	if from, to := r.Range(); from != "22:00" || to != "02:00" || r.Networks()[0] != "10.13.13.0/24" || len(r.Days()) != 2 {
		t.Fatalf("%v %v %v %v", from, to, r.Networks(), r.Days())
	}
	if len(decodeRules("not json")) != 0 {
		t.Fatal("broken rules must be ignored")
	}
}

func TestMobileLayoutRoundTrip(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "dashboard", Migrations()))
	tab, _ := domain.NewTab("t1", "Home", 0)
	if err := repo.CreateTab(ctx, tab); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.GetTab(ctx, "t1"); got.Mobile() != nil {
		t.Fatal("new tabs follow the desktop layout")
	}
	m := []domain.MobileEntry{{WidgetID: "w2", Hidden: true, Rows: 3}, {WidgetID: "w1", Rows: 5}}
	if err := repo.UpdateTab(ctx, tab.WithMobile(m)); err != nil {
		t.Fatal(err)
	}
	list, _ := repo.ListTabs(ctx)
	if got := list[0].Mobile(); len(got) != 2 || got[0] != m[0] || got[1] != m[1] {
		t.Fatalf("got %+v", got)
	}
	if err := repo.UpdateTab(ctx, tab.WithMobile(nil)); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.GetTab(ctx, "t1"); got.Mobile() != nil {
		t.Fatalf("reset must store NULL, got %+v", got.Mobile())
	}
}

func TestMigrationRemovesSearchWidgets(t *testing.T) {
	// The dashboard migrations run first; then a placed search widget is put
	// back and the removal migration is applied again (it is idempotent).
	db := dbtest.Open(t, "dashboard", Migrations())
	ctx := context.Background()
	mustExec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}
	mustExec(`INSERT INTO dashboard_tabs (id, name, position) VALUES ('t1', 'Home', 0)`)
	mustExec(`INSERT INTO dashboard_widgets (id, tab_id, kind, x, y, w, h) VALUES ('w1', 't1', 'search', 0, 0, 12, 2), ('w2', 't1', 'clock', 0, 2, 3, 3)`)
	sql, err := fs.ReadFile(Migrations(), "0005_remove_search_widget.sql")
	if err != nil {
		t.Fatal(err)
	}
	mustExec(string(sql))
	var kinds []string
	rows, err := db.QueryContext(ctx, `SELECT kind FROM dashboard_widgets ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		_ = rows.Scan(&k)
		kinds = append(kinds, k)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(kinds) != 1 || kinds[0] != "clock" {
		t.Fatalf("only the search widget is removed, got %v", kinds)
	}
}
