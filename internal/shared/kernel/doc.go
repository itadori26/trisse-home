// Package kernel is the minimal shared kernel used by every module.
//
// Ring: Entities (innermost). It contains only error kinds and tiny ports
// (Clock, IDGenerator) that all modules need. It must never import any other
// package of this project and must stay free of framework dependencies.
package kernel
