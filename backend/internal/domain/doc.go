// Package domain holds the calculator's business rules. It is pure: no HTTP, no Gin, no I/O.
// Adapters in internal/platform depend on it, never the other way around (enforced by depguard).
package domain
