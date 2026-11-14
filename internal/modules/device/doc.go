// Package device lets an appliance (Raspberry Pi image, ADR 0015) choose
// whether it also shows the dashboard on a connected display (#140).
//
// The app runs unprivileged: it only writes a request file; a root-owned
// systemd path unit validates it and installs or removes the kiosk.
package device
