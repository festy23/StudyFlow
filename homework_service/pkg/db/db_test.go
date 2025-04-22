package db_test

import (
	"homework_service/pkg/db"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostgres(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		cfg := db.Config{
			Host:     "localhost",
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			DBName:   "testdb",
			SSLMode:  "disable",
		}

		pg, err := db.NewPostgres(cfg)
		require.NoError(t, err)
		require.NotNil(t, pg)
		require.NotNil(t, pg.DB())

		err = pg.Close()
		require.NoError(t, err)
	})

	t.Run("connection error", func(t *testing.T) {
		cfg := db.Config{
			Host:     "invalid",
			Port:     9999,
			User:     "user",
			Password: "pass",
			DBName:   "db",
			SSLMode:  "disable",
		}

		_, err := db.NewPostgres(cfg)
		require.Error(t, err)
	})
}
