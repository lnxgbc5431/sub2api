//go:build unit

package service

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestAccountTestService_TestAccountConnection_CustomBaseURLForcesChatCompletions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		platform  string
		customURL string
	}{
		{
			name:      "anthropic custom base url",
			platform:  PlatformAnthropic,
			customURL: "https://api.linkapi.ai/v1",
		},
		{
			name:      "openai custom base url",
			platform:  PlatformOpenAI,
			customURL: "https://yunwu.ai/v1",
		},
		{
			name:      "gemini custom base url",
			platform:  PlatformGemini,
			customURL: "https://api.deepseek.com/v1",
		},
		{
			name:      "antigravity custom base url",
			platform:  PlatformAntigravity,
			customURL: "https://api.example.com/v1",
		},
		{
			name:      "custom base url with trailing slash",
			platform:  PlatformAnthropic,
			customURL: "https://api.example.com/v1/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := newTestContext()
			upstream := &queuedHTTPUpstream{
				responses: []*http.Response{
					newJSONResponse(http.StatusOK, `{"id":"ok"}`),
				},
			}
			repo := &mockAccountRepoForGemini{
				accountsByID: map[int64]*Account{
					1: {
						ID:          1,
						Platform:    tt.platform,
						Type:        AccountTypeAPIKey,
						Concurrency: 1,
						Credentials: map[string]any{
							"api_key":  "sk-test",
							"base_url": tt.customURL,
						},
					},
				},
			}

			svc := &AccountTestService{
				accountRepo:  repo,
				httpUpstream: upstream,
				cfg:          &config.Config{},
			}

			err := svc.TestAccountConnection(ctx, 1, "gpt-4o-mini", "hi", AccountTestModeDefault)
			require.NoError(t, err)
			require.Len(t, upstream.requests, 1)
			require.Equal(t, http.MethodPost, upstream.requests[0].Method)
			require.Equal(t, "/v1/chat/completions", upstream.requests[0].URL.Path)
			require.NotContains(t, upstream.requests[0].URL.Path, "/v1/v1/")
			require.NotContains(t, upstream.requests[0].URL.Path, "//chat/completions")
			require.NotContains(t, upstream.requests[0].URL.Path, "/chat/completions/chat/completions")
		})
	}
}

func TestAccountTestService_TestAccountConnection_EmptyCustomBaseURLKeepsPlatformRoutes(t *testing.T) {
	t.Parallel()

	openAISuccessBody := "data: {\"type\":\"response.completed\"}\n\n"
	claudeSuccessBody := "event: message_start\ndata: {\"type\":\"message_start\"}\n\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"ok\"}}\n\ndata: {\"type\":\"message_stop\"}\n\n"
	geminiSuccessBody := "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"ok\"}]}}]}\n\ndata: [DONE]\n\n"

	tests := []struct {
		name         string
		account      *Account
		responseBody string
		expectedPath string
	}{
		{
			name: "anthropic official path",
			account: &Account{
				ID:          11,
				Platform:    PlatformAnthropic,
				Type:        AccountTypeAPIKey,
				Concurrency: 1,
				Credentials: map[string]any{"api_key": "sk-test"},
			},
			responseBody: claudeSuccessBody,
			expectedPath: "/v1/messages",
		},
		{
			name: "openai official path",
			account: &Account{
				ID:          12,
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				Concurrency: 1,
				Credentials: map[string]any{"api_key": "sk-test"},
			},
			responseBody: openAISuccessBody,
			expectedPath: "/responses",
		},
		{
			name: "gemini official path",
			account: &Account{
				ID:          13,
				Platform:    PlatformGemini,
				Type:        AccountTypeAPIKey,
				Concurrency: 1,
				Credentials: map[string]any{"api_key": "sk-test"},
			},
			responseBody: geminiSuccessBody,
			expectedPath: "/v1beta/models",
		},
		{
			name: "antigravity official path",
			account: &Account{
				ID:          14,
				Platform:    PlatformAntigravity,
				Type:        AccountTypeAPIKey,
				Concurrency: 1,
				Credentials: map[string]any{"api_key": "sk-test"},
			},
			responseBody: claudeSuccessBody,
			expectedPath: "/v1/messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := newTestContext()
			upstream := &queuedHTTPUpstream{responses: []*http.Response{
				{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
				},
			}}
			repo := &mockAccountRepoForGemini{accountsByID: map[int64]*Account{
				tt.account.ID: tt.account,
			}}
			svc := &AccountTestService{
				accountRepo:  repo,
				httpUpstream: upstream,
				cfg:          &config.Config{},
			}

			err := svc.TestAccountConnection(ctx, tt.account.ID, "", "hi", AccountTestModeDefault)
			require.NoError(t, err)
			require.Len(t, upstream.requests, 1)
			require.Equal(t, http.MethodPost, upstream.requests[0].Method)
			require.Contains(t, upstream.requests[0].URL.Path, tt.expectedPath)
			require.NotContains(t, upstream.requests[0].URL.Path, "/v1/v1/messages")
			require.NotContains(t, upstream.requests[0].URL.Path, "/v1/v1/chat/completions")
		})
	}
}
