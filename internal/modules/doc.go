// Package modules defines the plugin contract every feature module implements
// and a registry that collects the enabled modules.
//
// Ring: Interface Adapters (glue between the platform and the modules). The
// platform consumes this contract; modules implement it in their module.go.
package modules
