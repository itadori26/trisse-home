package domain

// InstallMethod is how this instance was installed. It decides how updates
// reach it, so the update hint can tell the user what to do.
type InstallMethod string

// Known install methods (INSTALL_METHOD). The zero value means unknown, e.g.
// a binary started by hand.
const (
	InstallUnknown   InstallMethod = ""
	InstallDocker    InstallMethod = "docker"
	InstallPackage   InstallMethod = "package"
	InstallAppliance InstallMethod = "appliance"
)

// ParseInstallMethod returns the install method named s; anything else is
// InstallUnknown.
func ParseInstallMethod(s string) InstallMethod {
	switch m := InstallMethod(s); m {
	case InstallDocker, InstallPackage, InstallAppliance:
		return m
	default:
		return InstallUnknown
	}
}

// UpdatesAutomatically reports whether new releases are installed without any
// action of the user (the Raspberry Pi image updates itself at night).
func (m InstallMethod) UpdatesAutomatically() bool { return m == InstallAppliance }
