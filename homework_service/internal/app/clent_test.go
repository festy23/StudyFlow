package app_test

import (
	"context"
	"homework_service/internal/app"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserClient(t *testing.T) {
	t.Run("UserExists - success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := app.NewUserClient(ts.URL)
		exists := client.UserExists(context.Background(), "user1")
		require.True(t, exists)
	})

	t.Run("GetUserRole - success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"role": "tutor"}`))
		}))
		defer ts.Close()

		client := app.NewUserClient(ts.URL)
		role, err := client.GetUserRole(context.Background(), "user1")
		require.NoError(t, err)
		require.Equal(t, "tutor", role)
	})

	t.Run("IsPair - success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := app.NewUserClient(ts.URL)
		isPair := client.IsPair(context.Background(), "tutor1", "student1")
		require.True(t, isPair)
	})
}

func TestFileClient(t *testing.T) {
	t.Run("FileExists - success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := app.NewFileClient(ts.URL)
		exists := client.FileExists(context.Background(), "file1")
		require.True(t, exists)
	})

	t.Run("GetFileURL - success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"url": "http://example.com/file1"}`))
		}))
		defer ts.Close()

		client := app.NewFileClient(ts.URL)
		url, err := client.GetFileURL(context.Background(), "file1", "user1")
		require.NoError(t, err)
		require.Equal(t, "http://example.com/file1", url)
	})
}
