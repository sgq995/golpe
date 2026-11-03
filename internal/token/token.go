// Package token extends go/token with the tokens Golpe adds on top of Go.
//
// Golpe tokens are allocated after the range used by go/token so a single
// Token value can carry either a standard Go token or a Golpe one.
//
// Status: scaffold. See docs/knowledge/tasks/001-template-language/02-token-extension.md.
package token
