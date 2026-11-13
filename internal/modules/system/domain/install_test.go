package domain

import "testing"

func TestParseInstallMethod(t *testing.T) {
	tests := map[string]InstallMethod{
		"docker":    InstallDocker,
		"package":   InstallPackage,
		"appliance": InstallAppliance,
		"":          InstallUnknown,
		"snap":      InstallUnknown,
		"Docker":    InstallUnknown,
	}
	for in, want := range tests {
		if got := ParseInstallMethod(in); got != want {
			t.Errorf("ParseInstallMethod(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUpdatesAutomatically(t *testing.T) {
	for m, want := range map[InstallMethod]bool{InstallAppliance: true, InstallPackage: false, InstallDocker: false, InstallUnknown: false} {
		if got := m.UpdatesAutomatically(); got != want {
			t.Errorf("%q.UpdatesAutomatically() = %v, want %v", m, got, want)
		}
	}
}
