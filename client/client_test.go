package client

import (
	"net/url"
	"strings"
	"testing"
)

func TestBuildQueryParamsEscapesSpecialCharacters(t *testing.T) {
	query := buildQueryParams(map[string]string{
		"exactUser":         "automq-connect",
		"fuzzyResourceName": "*",
		"path":              "foo/bar baz",
	})

	values, err := url.ParseQuery(query)
	if err != nil {
		t.Fatalf("ParseQuery(%q) failed: %v", query, err)
	}

	if got := values.Get("fuzzyResourceName"); got != "*" {
		t.Fatalf("fuzzyResourceName = %q, want *", got)
	}
	if got := values.Get("path"); got != "foo/bar baz" {
		t.Fatalf("path = %q, want foo/bar baz", got)
	}
	if strings.Contains(query, "fuzzyResourceName=*") {
		t.Fatalf("wildcard query value was not URL encoded: %q", query)
	}
	if !strings.Contains(query, "fuzzyResourceName=%2A") {
		t.Fatalf("wildcard query value was not encoded as %%2A: %q", query)
	}
}
