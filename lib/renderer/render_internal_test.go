package renderer

import "testing"

// TestAcceptQuality verifies the Accept-header content negotiation helper:
// media-range matching, specificity (exact > "typ/*" > "*/*"), q-value
// parsing/defaults, whitespace handling, and unacceptable types.
func TestAcceptQuality(t *testing.T) {
	testCases := []struct {
		name   string
		accept string
		typ    string
		sub    string
		want   float64
	}{
		{
			name:   "empty header is not acceptable",
			accept: "",
			typ:    "text",
			sub:    "html",
			want:   0,
		},
		{
			name:   "exact match defaults q to 1.0",
			accept: "text/html",
			typ:    "text",
			sub:    "html",
			want:   1.0,
		},
		{
			name:   "type wildcard matches with spec 1",
			accept: "text/*",
			typ:    "text",
			sub:    "html",
			want:   1.0,
		},
		{
			name:   "full wildcard matches with spec 0",
			accept: "*/*",
			typ:    "application",
			sub:    "json",
			want:   1.0,
		},
		{
			name:   "explicit q-value is parsed",
			accept: "text/html;q=0.5",
			typ:    "text",
			sub:    "html",
			want:   0.5,
		},
		{
			name:   "most specific range wins regardless of order",
			accept: "text/*;q=0.5, text/html;q=0.1",
			typ:    "text",
			sub:    "html",
			want:   0.1,
		},
		{
			name:   "most specific range wins when listed first",
			accept: "text/html;q=0.1, text/*;q=0.5",
			typ:    "text",
			sub:    "html",
			want:   0.1,
		},
		{
			name:   "exact beats full wildcard",
			accept: "*/*;q=0.9, application/json;q=0.2",
			typ:    "application",
			sub:    "json",
			want:   0.2,
		},
		{
			name:   "non-matching type is not acceptable",
			accept: "application/json",
			typ:    "text",
			sub:    "html",
			want:   0,
		},
		{
			name:   "malformed range without slash is skipped",
			accept: "text",
			typ:    "text",
			sub:    "html",
			want:   0,
		},
		{
			name:   "json preferred over html in mixed list (html side)",
			accept: "application/json, text/html;q=0.1",
			typ:    "text",
			sub:    "html",
			want:   0.1,
		},
		{
			name:   "json preferred over html in mixed list (json side)",
			accept: "application/json, text/html;q=0.1",
			typ:    "application",
			sub:    "json",
			want:   1.0,
		},
		{
			name:   "surrounding whitespace is trimmed",
			accept: "  text/html ; q=0.8 ",
			typ:    "text",
			sub:    "html",
			want:   0.8,
		},
		{
			name:   "invalid q-value falls back to default",
			accept: "text/html;q=abc",
			typ:    "text",
			sub:    "html",
			want:   1.0,
		},
		{
			name:   "q after other params is parsed",
			accept: "text/html;level=1;q=0.7",
			typ:    "text",
			sub:    "html",
			want:   0.7,
		},
		{
			name:   "type wildcard q-value applies",
			accept: "*/*;q=0.3",
			typ:    "application",
			sub:    "json",
			want:   0.3,
		},
		{
			name:   "empty entries from stray commas are skipped",
			accept: ",, text/html ,,",
			typ:    "text",
			sub:    "html",
			want:   1.0,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			got := acceptQuality(tc.accept, tc.typ, tc.sub)
			if got != tc.want {
				t.Errorf("acceptQuality(%q, %q, %q) = %v, want %v",
					tc.accept, tc.typ, tc.sub, got, tc.want)
			}
		})
	}
}
