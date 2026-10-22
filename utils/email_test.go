package utils

import "testing"

// ENG-250: the one property this helper exists for — a plus survives a query-string round trip
// rather than silently becoming a space (what both browsers and Go's own net/url do with a raw,
// un-encoded '+' in a query value).
func TestEncodeEmailQuery_PlusSurvivesRoundTrip(t *testing.T) {
	got := EncodeEmailQuery("user+tag@example.com")
	if got != "user%2Btag%40example.com" {
		t.Fatalf("EncodeEmailQuery(%q) = %q, want user%%2Btag%%40example.com", "user+tag@example.com", got)
	}

	// The red-proof this guards against: a caller swapping in url.QueryEscape's sibling
	// url.PathEscape (which leaves '+' untouched) would pass a visually-similar review but break
	// every plus-address the moment it lands in a query string rather than a path segment.
	if got == "user+tag@example.com" {
		t.Fatal("the '+' must be percent-encoded, not left as a literal plus in a query value")
	}
}

func TestEncodeEmailQuery_PlainAddressUnaffected(t *testing.T) {
	got := EncodeEmailQuery("ada@example.com")
	if got != "ada%40example.com" {
		t.Fatalf("EncodeEmailQuery(%q) = %q, want ada%%40example.com", "ada@example.com", got)
	}
}
