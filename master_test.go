package pgmaster_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sharpvik/pgmaster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestMaster(t *testing.T) {
	ctx := context.TODO()
	pg := PostgresContainer(t)

	host, err := pg.Host(ctx)
	require.NoError(t, err)
	port, err := pg.MappedPort(ctx, nat.Port("5432"))
	require.NoError(t, err)

	master, err := pgmaster.Find(Connect(port.Int()), 5*time.Second, []string{
		"123.45.67.89", // won't work - timeout
		host,           // master host
	})
	assert.NoError(t, err)
	assert.Equal(t, host, master)
}

func PostgresContainer(t testing.TB) *postgres.PostgresContainer {
	ctx := context.TODO()

	container, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies())
	require.NoError(t, err)
	t.Cleanup(func() { container.Terminate(ctx) })

	return container
}

func Connect(port int) pgmaster.ConnectFunc {
	return func(host string) (*sql.DB, error) {
		return sql.Open("postgres", ConnectionString(host, port))
	}
}

func ConnectionString(host string, port int) string {
	return fmt.Sprintf(
		"host=%s port=%d user=postgres password=password dbname=postgres sslmode=disable",
		host, port,
	)
}
