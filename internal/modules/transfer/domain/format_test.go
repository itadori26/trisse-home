package domain

import "testing"

func TestCheckHeader(t *testing.T) {
	if CheckHeader(Format, 1) != nil || CheckHeader("other", 1) == nil || CheckHeader(Format, Version+1) == nil || CheckHeader(Format, 0) == nil {
		t.Fatal("header check wrong")
	}
}
