package dto

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyFromService_GroupHasCustomBaseURLTrue(t *testing.T) {
	src := &service.APIKey{
		ID:     1,
		UserID: 2,
		Key:    "sk-custom-group",
		Name:   "custom",
		Status: service.StatusActive,
		Group: &service.Group{
			ID:               10,
			Name:             "custom-group",
			Platform:         service.PlatformOpenAI,
			Status:           service.StatusActive,
			HasCustomBaseURL: true,
		},
	}

	out := APIKeyFromService(src)
	require.NotNil(t, out)
	require.NotNil(t, out.Group)
	require.True(t, out.Group.HasCustomBaseURL)

	raw, err := json.Marshal(out)
	require.NoError(t, err)
	jsonText := string(raw)
	require.Contains(t, jsonText, `"has_custom_base_url":true`)
	require.NotContains(t, jsonText, `"custom_base_url"`)
}

func TestAPIKeyFromService_GroupHasCustomBaseURLFalse(t *testing.T) {
	src := &service.APIKey{
		ID:     3,
		UserID: 4,
		Key:    "sk-official-group",
		Name:   "official",
		Status: service.StatusActive,
		Group: &service.Group{
			ID:               11,
			Name:             "official-group",
			Platform:         service.PlatformOpenAI,
			Status:           service.StatusActive,
			HasCustomBaseURL: false,
		},
	}

	out := APIKeyFromService(src)
	require.NotNil(t, out)
	require.NotNil(t, out.Group)
	require.False(t, out.Group.HasCustomBaseURL)

	raw, err := json.Marshal(out)
	require.NoError(t, err)
	jsonText := string(raw)
	require.Contains(t, jsonText, `"has_custom_base_url":false`)
	require.NotContains(t, jsonText, `"custom_base_url"`)
}
