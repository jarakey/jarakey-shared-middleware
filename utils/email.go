package utils

import "net/url"

// EncodeEmailQuery percent-encodes an email address for use as a URL query value, so that
// plus-addressed accounts (e.g. user+tag@gmail.com) survive a round trip through a link or an
// outbound service-to-service request.
//
// ENG-250: a raw '+' is decoded back to a space by both browsers and every Go HTTP server, so an
// un-encoded plus-address in a query string silently becomes a DIFFERENT, well-formed address that
// belongs to nobody — not an error, just a lookup that quietly finds the wrong (or no) account.
// Before this helper, every place in the backend that placed an email into a link or an S2S
// request call re-implemented this encoding separately; it had already been forgotten once (and
// fixed only after the fact, in property-service's own prior encodeEmailQuery). This is now the
// single place every such call site imports from, so a new one cannot silently omit it.
func EncodeEmailQuery(email string) string {
	return url.QueryEscape(email)
}
