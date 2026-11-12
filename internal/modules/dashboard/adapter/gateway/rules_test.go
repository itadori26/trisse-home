package gateway

import (
	"context"
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
