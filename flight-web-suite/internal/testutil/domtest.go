package testutil

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// Interface for expectations on DOM nodes.
type Expectation interface {
	check(sel *goquery.Selection) error
	describe() string
}

type expectationFunc struct {
	desc string
	fn   func(*goquery.Selection) error
}

func (e expectationFunc) check(sel *goquery.Selection) error {
	return e.fn(sel)
}

func (e expectationFunc) describe() string {
	return e.desc
}

func HasText(want string) Expectation {
	return expectationFunc{
		desc: fmt.Sprintf("have text %q", want),
		fn: func(sel *goquery.Selection) error {
			got := strings.TrimSpace(sel.Text())
			if got != want {
				return fmt.Errorf("expected text %q, got %q", want, got)
			}
			return nil
		},
	}
}

func HasAttr(name, want string) Expectation {
	return expectationFunc{
		desc: fmt.Sprintf("have attr %q=%q", name, want),
		fn: func(sel *goquery.Selection) error {
			got, ok := sel.Attr(name)
			if !ok {
				return fmt.Errorf("expected attr %q=%q, but attr was missing", name, want)
			}
			if got != want {
				return fmt.Errorf("expected attr %q=%q, got %q", name, want, got)
			}
			return nil
		},
	}
}

// Fail if the selector is not in body or in body more than once.
// Run each expectation on the selected node, fail if any of them return an error.
func AssertSelection(t *testing.T, body, selector string, exps ...Expectation) {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		t.Fatalf("failed to parse html: %v\nbody:\n%s", err, body)
	}

	sel := doc.Find(selector)
	if sel.Length() == 0 {
		var want []string
		for _, exp := range exps {
			want = append(want, exp.describe())
		}
		t.Fatalf(
			"expected element matching selector %q (%s), but none found\nbody:\n%s",
			selector,
			strings.Join(want, ", "),
			body,
		)
	}

	if sel.Length() > 1 {
		t.Fatalf(
			"expected exactly 1 element matching selector %q, found %d\nmatches:\n%s",
			selector,
			sel.Length(),
			renderSelection(sel),
		)
	}

	for _, exp := range exps {
		if err := exp.check(sel); err != nil {
			t.Errorf(
				"selector %q: %v\nnode:\n%s",
				selector,
				err,
				renderSelection(sel),
			)
		}
	}
}

// Fail if the selector is in body.
func AssertNoSelection(t *testing.T, body, selector string) {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		t.Fatalf("failed to parse html: %v\nbody:\n%s", err, body)
	}

	sel := doc.Find(selector)
	if sel.Length() != 0 {
		t.Fatalf(
			"expected no elements matching selector %q, found %d\nmatches:\n%s",
			selector,
			sel.Length(),
			renderSelection(sel),
		)
	}
}

func renderSelection(sel *goquery.Selection) string {
	var parts []string
	sel.Each(func(_ int, s *goquery.Selection) {
		html, err := goquery.OuterHtml(s)
		if err != nil {
			parts = append(parts, fmt.Sprintf("<render error: %v>", err))
			return
		}
		parts = append(parts, html)
	})
	return strings.Join(parts, "\n")
}
