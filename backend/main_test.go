package main

import (
	"net/url"
	"strings"
	"testing"
)

func TestParseListQueryDefaults(t *testing.T) {
	lq, err := parseListQuery(url.Values{})
	if err != nil {
		t.Fatal(err)
	}
	if lq.Page != 1 || lq.PageSize != 25 {
		t.Fatalf("page=%d size=%d", lq.Page, lq.PageSize)
	}
	if lq.WhereSQL != "TRUE" || len(lq.Args) != 0 {
		t.Fatalf("where=%s args=%v", lq.WhereSQL, lq.Args)
	}
	if lq.OrderSQL != "id ASC" {
		t.Fatalf("order=%s", lq.OrderSQL)
	}
}

func TestParseListQueryRejectsBadInput(t *testing.T) {
	cases := []url.Values{
		{"sort": {"id; DROP TABLE items"}},
		{"order": {"desc;--"}},
		{"page": {"0"}},
		{"page": {"-3"}},
		{"page": {"abc"}},
		{"deliveredWithBox": {"yes"}},
		{"delivered": {"maybe"}},
		{"hasPicture": {"yes"}},
		{"pageSize": {"0"}},
		{"pageSize": {"61"}},
	}
	for _, values := range cases {
		if _, err := parseListQuery(values); err == nil {
			t.Fatalf("expected error for %v", values)
		}
	}
}

func TestParseListQueryFilters(t *testing.T) {
	lq, err := parseListQuery(url.Values{
		"page":             {"2"},
		"sort":             {"text"},
		"order":            {"desc"},
		"q":                {"50%"},
		"group":            {"casa"},
		"woodType":         {"TECA"},
		"deliveredWithBox": {"true"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if lq.Page != 2 {
		t.Fatalf("page=%d", lq.Page)
	}
	if !strings.HasPrefix(lq.OrderSQL, "text DESC") {
		t.Fatalf("order=%s", lq.OrderSQL)
	}
	if len(lq.Args) != 4 {
		t.Fatalf("args=%v", lq.Args)
	}
	if lq.Args[0] != `%50\%%` {
		t.Fatalf("pattern=%q", lq.Args[0])
	}
	if lq.Args[2] != "TECA" || lq.Args[3] != true {
		t.Fatalf("args=%v", lq.Args)
	}
	if strings.Count(lq.WhereSQL, "$1") != 6 {
		t.Fatalf("search placeholder reused: %s", lq.WhereSQL)
	}
	if !strings.Contains(lq.WhereSQL, "woodType = $3") {
		t.Fatalf("where=%s", lq.WhereSQL)
	}
}

func TestParseListQueryDelivered(t *testing.T) {
	filled, err := parseListQuery(url.Values{"delivered": {"true"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(filled.WhereSQL, "deliveredTo, '')) <> ''") || len(filled.Args) != 0 {
		t.Fatalf("filled where=%s args=%v", filled.WhereSQL, filled.Args)
	}
	empty, err := parseListQuery(url.Values{"delivered": {"false"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(empty.WhereSQL, "deliveredTo, '')) = ''") {
		t.Fatalf("empty where=%s", empty.WhereSQL)
	}
}

func TestParseListQueryGalleryFilters(t *testing.T) {
	lq, err := parseListQuery(url.Values{
		"name":       {"maría"},
		"size":       {"5"},
		"woodType":   {"TECA"},
		"hasPicture": {"true"},
		"pageSize":   {"24"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if lq.PageSize != 24 {
		t.Fatalf("pageSize=%d", lq.PageSize)
	}
	if len(lq.Args) != 3 || lq.Args[0] != "%maría%" || lq.Args[1] != "5" || lq.Args[2] != "TECA" {
		t.Fatalf("args=%v", lq.Args)
	}
	for _, part := range []string{`COALESCE(text, '') ILIKE $1`, `COALESCE(size, '') = $2`, `woodType = $3`, `COALESCE(picture, '') <> ''`} {
		if !strings.Contains(lq.WhereSQL, part) {
			t.Fatalf("missing %s in %s", part, lq.WhereSQL)
		}
	}
}
