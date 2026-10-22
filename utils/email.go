package utils

import "net/url"

// EncodeEmailQuery percent-encodes an email for a URL query value — ENG-250's single shared
// replacement for the seven call sites that each re-implemented this separately. A raw '+'
// decodes back to a space (browsers and net/url both do this), silently misrouting a
// plus-addressed account to no one; see email_test.go for the guard against the PathEscape trap.
func EncodeEmailQuery(email string) string {
	return url.QueryEscape(email)
}
