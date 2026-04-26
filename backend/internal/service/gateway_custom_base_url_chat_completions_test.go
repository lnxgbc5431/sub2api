//go:build unit

package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type chatCompletionsUpstreamRecorder struct {
	lastReq   *http.Request
	lastBody  []byte
	responses []*http.Response
}

func (u *chatCompletionsUpstreamRecorder) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return u.record(req)
}

func (u *chatCompletionsUpstreamRecorder) DoWithTLS(req *http.Request, _ string, _ int64, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.record(req)
}

func (u *chatCompletionsUpstreamRecorder) record(req *http.Request) (*http.Response, error) {
	u.lastReq = req
	body, _ := io.ReadAll(req.Body)
	u.lastBody = body
	if len(u.responses) == 0 {
		return nil, fmt.Errorf("no mocked response")
	}
	resp := u.responses[0]
	u.responses = u.responses[1:]
	return resp, nil
}

func newGatewayCCContext(body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, rec
}

func TestGatewayService_ForwardAsChatCompletions_CustomBaseURLRoutesToChatCompletions(t *testing.T) {
	t.Parallel()

	body := []byte(`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}],"stream":false}`)
	tests := []struct {
		name      string
		platform  string
		customURL string
	}{
		{name: "anthropic custom", platform: PlatformAnthropic, customURL: "https://api.linkapi.ai/v1"},
		{name: "gemini custom", platform: PlatformGemini, customURL: "https://api.deepseek.com/v1"},
		{name: "antigravity custom", platform: PlatformAntigravity, customURL: "https://api.example.com/v1/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, rec := newGatewayCCContext(body)
			upstream := &chatCompletionsUpstreamRecorder{
				responses: []*http.Response{
					{
						StatusCode: http.StatusOK,
						Header:     make(http.Header),
						Body:       io.NopCloser(strings.NewReader(`{"id":"ok"}`)),
					},
				},
			}
			svc := &GatewayService{
				httpUpstream: upstream,
				cfg:          &config.Config{},
			}
			account := &Account{
				ID:          1,
				Platform:    tt.platform,
				Type:        AccountTypeAPIKey,
				Concurrency: 1,
				Credentials: map[string]any{
					"api_key":  "sk-test",
					"base_url": tt.customURL,
				},
			}

			result, err := svc.ForwardAsChatCompletions(context.Background(), ctx, account, body, &ParsedRequest{})
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, http.StatusOK, rec.Code)
			require.Equal(t, "/v1/chat/completions", upstream.lastReq.URL.Path)
			require.Equal(t, http.MethodPost, upstream.lastReq.Method)
			require.True(t, strings.HasPrefix(upstream.lastReq.Header.Get("Authorization"), "Bearer "))
			require.NotContains(t, upstream.lastReq.URL.Path, "/v1/v1/")
			require.NotContains(t, upstream.lastReq.URL.Path, "//chat/completions")
		})
	}
}

func TestOpenAIGatewayService_ForwardAsChatCompletions_CustomBaseURLRoutesToChatCompletions(t *testing.T) {
	t.Parallel()

	body := []byte(`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}],"stream":false}`)
	ctx, rec := newGatewayCCContext(body)
	upstream := &chatCompletionsUpstreamRecorder{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"id":"ok"}`)),
			},
		},
	}
	svc := &OpenAIGatewayService{
		httpUpstream: upstream,
		cfg:          &config.Config{},
	}
	account := &Account{
		ID:          2,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":  "sk-test",
			"base_url": "https://yunwu.ai/v1",
		},
	}

	result, err := svc.ForwardAsChatCompletions(context.Background(), ctx, account, body, "", "")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "/v1/chat/completions", upstream.lastReq.URL.Path)
	require.Equal(t, http.MethodPost, upstream.lastReq.Method)
	require.True(t, strings.HasPrefix(upstream.lastReq.Header.Get("Authorization"), "Bearer "))
	require.NotContains(t, upstream.lastReq.URL.Path, "/v1/v1/")
}
