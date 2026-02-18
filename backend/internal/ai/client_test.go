package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIClient_Complete(t *testing.T) {
	t.Run("returns content from successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

			var body openAIRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "gpt-5-mini", body.Model)
			assert.Len(t, body.Messages, 2)
			assert.Equal(t, "system", body.Messages[0].Role)
			assert.Equal(t, "user", body.Messages[1].Role)
			assert.Equal(t, "json_object", body.ResponseFormat.Type)
			assert.Equal(t, float64(0), body.Temperature)
			assert.Equal(t, 4096, body.MaxTokens)

			resp := openAIResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{
					{Message: struct {
						Content string `json:"content"`
					}{Content: `{"workouts": []}`}},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key")
		client.httpClient = server.Client()

		// Override the URL by using a custom transport
		origURL := server.URL
		client.httpClient.Transport = &rewriteTransport{
			base:    server.Client().Transport,
			baseURL: origURL,
		}

		result, err := client.Complete(context.Background(), CompletionRequest{
			SystemPrompt: "You are a coach",
			UserPrompt:   "Create a plan",
		})
		require.NoError(t, err)
		assert.Equal(t, `{"workouts": []}`, result)
	})

	t.Run("returns error on API error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openAIResponse{
				Error: &struct {
					Message string `json:"message"`
				}{Message: "invalid api key"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewOpenAIClient("bad-key")
		client.httpClient = server.Client()
		client.httpClient.Transport = &rewriteTransport{
			base:    server.Client().Transport,
			baseURL: server.URL,
		}

		_, err := client.Complete(context.Background(), CompletionRequest{
			SystemPrompt: "test",
			UserPrompt:   "test",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid api key")
	})

	t.Run("returns error on empty choices", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openAIResponse{Choices: nil}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key")
		client.httpClient = server.Client()
		client.httpClient.Transport = &rewriteTransport{
			base:    server.Client().Transport,
			baseURL: server.URL,
		}

		_, err := client.Complete(context.Background(), CompletionRequest{
			SystemPrompt: "test",
			UserPrompt:   "test",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no choices")
	})

	t.Run("returns error on non-JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("internal server error"))
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key")
		client.httpClient = server.Client()
		client.httpClient.Transport = &rewriteTransport{
			base:    server.Client().Transport,
			baseURL: server.URL,
		}

		_, err := client.Complete(context.Background(), CompletionRequest{
			SystemPrompt: "test",
			UserPrompt:   "test",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal response")
	})
}

// rewriteTransport rewrites requests to the OpenAI URL to point at the test server.
type rewriteTransport struct {
	base    http.RoundTripper
	baseURL string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = t.baseURL[len("http://"):]
	if t.base != nil {
		return t.base.RoundTrip(req)
	}
	return http.DefaultTransport.RoundTrip(req)
}
