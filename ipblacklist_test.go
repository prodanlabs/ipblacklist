package ipblacklist

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	cfg := &Config{}
	cfg.IPBlacklists = []string{"127.0.0.1", "192.168.1.1"}
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := New(ctx, next, cfg, "ipblacklist")
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		header   string
		desc     string
		value    string
		expected string
	}{
		{
			header:   "X-Original-Forwarded-For",
			desc:     "good",
			value:    "10.0.1.1",
			expected: "10.0.1.1",
		},
		{
			header:   "X-Original-Forwarded-For",
			desc:     "bad",
			value:    "10.0.1.1",
			expected: "10.0.1.2",
		},
		{
			header:   "X-Forwarded-For",
			desc:     "good",
			value:    "10.0.2.2",
			expected: "10.0.2.2",
		},
		{
			header:   "X-Real-Ip",
			desc:     "good",
			value:    "10.0.3.3",
			expected: "10.0.3.3",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.expected, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set(test.header, test.value)

			handler.ServeHTTP(recorder, req)

			assertHeader(t, req, test.header, test.expected)
		})
	}
}

func assertHeader(t *testing.T, req *http.Request, key, expected string) {
	t.Helper()

	if req.Header.Get(key) != expected {
		t.Errorf("invalid header value: %s", req.Header.Get(key))
	}
}
