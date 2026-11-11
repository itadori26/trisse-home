package usecase

import (
	"errors"
	"testing"

	"trisse-home/internal/modules/auth/domain"
)

func TestDemoLogin(t *testing.T) {
	if _, err := NewDemoLogin(newEnv().deps).Execute(ctx, DemoLoginRequest{ClientKey: "ip"}); !errors.Is(err, domain.ErrSetupRequired) {
		t.Fatalf("expected ErrSetupRequired, got %v", err)
	}
	res, err := NewDemoLogin(newEnv().withAccount().deps).Execute(ctx, DemoLoginRequest{ClientKey: "ip"})
	if err != nil || res.SessionToken == "" || res.CSRFToken == "" {
		t.Fatalf("%+v %v", res, err)
	}
}
