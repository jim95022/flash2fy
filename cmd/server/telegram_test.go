package main

import "testing"

func TestResolveWebhookEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		baseURL      string
		configPath   string
		wantURL      string
		wantPath     string
		expectingErr bool
	}{
		{
			name:       "base url without path uses configured path",
			baseURL:    "https://example.com",
			configPath: "/telegram/webhook",
			wantURL:    "https://example.com/telegram/webhook",
			wantPath:   "/telegram/webhook",
		},
		{
			name:       "config path without leading slash is normalized",
			baseURL:    "https://example.com",
			configPath: "custom/hook",
			wantURL:    "https://example.com/custom/hook",
			wantPath:   "/custom/hook",
		},
		{
			name:       "url path takes precedence over configured path",
			baseURL:    "https://example.com/hook",
			configPath: "/ignored",
			wantURL:    "https://example.com/hook",
			wantPath:   "/hook",
		},
		{
			name:         "invalid url returns error",
			baseURL:      "://bad-url",
			configPath:   "/hook",
			expectingErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotURL, gotPath, err := resolveWebhookEndpoint(tc.baseURL, tc.configPath)
			if tc.expectingErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotURL != tc.wantURL {
				t.Fatalf("url mismatch: got %q, want %q", gotURL, tc.wantURL)
			}
			if gotPath != tc.wantPath {
				t.Fatalf("path mismatch: got %q, want %q", gotPath, tc.wantPath)
			}
		})
	}
}
