//go:build unit

package service

import (
	"strings"
	"testing"
)

func TestGetBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		account  Account
		expected string
	}{
		{
			name: "non-apikey type returns empty",
			account: Account{
				Type:     AccountTypeOAuth,
				Platform: PlatformAnthropic,
			},
			expected: "",
		},
		{
			name: "apikey without base_url returns default anthropic",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformAnthropic,
				Credentials: map[string]any{},
			},
			expected: "https://api.anthropic.com",
		},
		{
			name: "apikey with custom base_url",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformAnthropic,
				Credentials: map[string]any{"base_url": "https://custom.example.com"},
			},
			expected: "https://custom.example.com",
		},
		{
			name: "antigravity apikey auto-appends /antigravity",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformAntigravity,
				Credentials: map[string]any{"base_url": "https://upstream.example.com"},
			},
			expected: "https://upstream.example.com/antigravity",
		},
		{
			name: "antigravity apikey trims trailing slash before appending",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformAntigravity,
				Credentials: map[string]any{"base_url": "https://upstream.example.com/"},
			},
			expected: "https://upstream.example.com/antigravity",
		},
		{
			name: "antigravity non-apikey returns empty",
			account: Account{
				Type:        AccountTypeOAuth,
				Platform:    PlatformAntigravity,
				Credentials: map[string]any{"base_url": "https://upstream.example.com"},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.account.GetBaseURL()
			if result != tt.expected {
				t.Errorf("GetBaseURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetGeminiBaseURL(t *testing.T) {
	const defaultGeminiURL = "https://generativelanguage.googleapis.com"

	tests := []struct {
		name     string
		account  Account
		expected string
	}{
		{
			name: "apikey without base_url returns default",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformGemini,
				Credentials: map[string]any{},
			},
			expected: defaultGeminiURL,
		},
		{
			name: "apikey with custom base_url",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformGemini,
				Credentials: map[string]any{"base_url": "https://custom-gemini.example.com"},
			},
			expected: "https://custom-gemini.example.com",
		},
		{
			name: "antigravity apikey auto-appends /antigravity",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformAntigravity,
				Credentials: map[string]any{"base_url": "https://upstream.example.com"},
			},
			expected: "https://upstream.example.com/antigravity",
		},
		{
			name: "antigravity apikey trims trailing slash",
			account: Account{
				Type:        AccountTypeAPIKey,
				Platform:    PlatformAntigravity,
				Credentials: map[string]any{"base_url": "https://upstream.example.com/"},
			},
			expected: "https://upstream.example.com/antigravity",
		},
		{
			name: "antigravity oauth does NOT append /antigravity",
			account: Account{
				Type:        AccountTypeOAuth,
				Platform:    PlatformAntigravity,
				Credentials: map[string]any{"base_url": "https://upstream.example.com"},
			},
			expected: "https://upstream.example.com",
		},
		{
			name: "oauth without base_url returns default",
			account: Account{
				Type:        AccountTypeOAuth,
				Platform:    PlatformAntigravity,
				Credentials: map[string]any{},
			},
			expected: defaultGeminiURL,
		},
		{
			name: "nil credentials returns default",
			account: Account{
				Type:     AccountTypeAPIKey,
				Platform: PlatformGemini,
			},
			expected: defaultGeminiURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.account.GetGeminiBaseURL(defaultGeminiURL)
			if result != tt.expected {
				t.Errorf("GetGeminiBaseURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetEffectiveCustomBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		account  Account
		expected string
	}{
		{
			name: "reads from credentials.base_url",
			account: Account{
				Credentials: map[string]any{"base_url": "https://api.example.com/v1"},
			},
			expected: "https://api.example.com/v1",
		},
		{
			name: "reads from extra.custom_base_url with higher priority",
			account: Account{
				Extra:       map[string]any{"custom_base_url": " https://proxy.example.com/v1 "},
				Credentials: map[string]any{"base_url": "https://api.example.com/v1"},
			},
			expected: "https://proxy.example.com/v1",
		},
		{
			name: "reads from alias keys",
			account: Account{
				Extra: map[string]any{"customBaseURL": "https://alias.example.com/v1"},
			},
			expected: "https://alias.example.com/v1",
		},
		{
			name: "empty returns empty",
			account: Account{
				Credentials: map[string]any{"base_url": "   "},
				Extra:       map[string]any{"custom_base_url": ""},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.account.GetEffectiveCustomBaseURL()
			if got != tt.expected {
				t.Fatalf("GetEffectiveCustomBaseURL() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestJoinCustomBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		endpoint string
		want     string
	}{
		{
			name:     "domain without trailing slash",
			baseURL:  "https://api.example.com",
			endpoint: "chat/completions",
			want:     "https://api.example.com/chat/completions",
		},
		{
			name:     "domain with trailing slash",
			baseURL:  "https://api.example.com/",
			endpoint: "chat/completions",
			want:     "https://api.example.com/chat/completions",
		},
		{
			name:     "v1 without trailing slash",
			baseURL:  "https://api.example.com/v1",
			endpoint: "chat/completions",
			want:     "https://api.example.com/v1/chat/completions",
		},
		{
			name:     "v1 with trailing slash",
			baseURL:  "https://api.example.com/v1/",
			endpoint: "chat/completions",
			want:     "https://api.example.com/v1/chat/completions",
		},
		{
			name:     "endpoint with leading slash",
			baseURL:  "https://api.example.com/v1/",
			endpoint: "/chat/completions",
			want:     "https://api.example.com/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JoinCustomBaseURL(tt.baseURL, tt.endpoint)
			if got != tt.want {
				t.Fatalf("JoinCustomBaseURL() = %q, want %q", got, tt.want)
			}
			if containsAny(got, "/v1/v1/", "//chat/completions", "/chat/completions/chat/completions") {
				t.Fatalf("JoinCustomBaseURL() produced invalid path: %q", got)
			}
		})
	}
}

func containsAny(s string, patterns ...string) bool {
	for _, p := range patterns {
		if p != "" && strings.Contains(s, p) {
			return true
		}
	}
	return false
}
