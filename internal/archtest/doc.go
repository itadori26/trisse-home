// Package archtest verifies the dependency rule of the clean architecture by
// parsing the imports of every package below internal/. It complements the
// depguard linter with rules that need knowledge of the module a package
// belongs to.
package archtest
