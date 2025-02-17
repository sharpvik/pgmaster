package pgmaster

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var (
	ErrNoHosts  = errors.New("no hosts provided")
	ErrNotFound = errors.New("master not found")
)

// Opens the [sql.DB] connection to provided host. ConnectFunc is specifically
// made generic like that to allow for different connection variants
// (e.g. different SSL modes).
type ConnectFunc func(host string) (*sql.DB, error)

// Find PostgreSQL master node with timeout using a generic [ConnectFunc].
func Find(
	connect ConnectFunc,
	timeout time.Duration,
	hosts []string,
) (string, error) {
	if len(hosts) == 0 {
		return "", ErrNoHosts
	}

	for _, host := range hosts {
		if isMasterNode(connect, timeout, host) {
			return host, nil
		}
	}

	return "", ErrNotFound
}

func isMasterNode(
	connect ConnectFunc,
	timeout time.Duration,
	host string,
) bool {
	db, err := connect(host)
	if err != nil {
		return false
	}
	defer db.Close()

	return pingWithTimeout(db, timeout)
}

func pingWithTimeout(db *sql.DB, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	okChan := make(chan bool, 1)

	go func() { okChan <- ping(ctx, db) }()

	select {
	case <-ctx.Done(): // Timeout reached
		return false
	case err := <-okChan: // Ping completed
		return err
	}
}

func ping(
	ctx context.Context,
	db *sql.DB,
) bool {
	if err := db.PingContext(ctx); err != nil {
		return false
	}

	var isReadOnlyNode bool

	err := db.
		QueryRowContext(ctx, "SELECT pg_is_in_recovery()").
		Scan(&isReadOnlyNode)

	return err == nil && !isReadOnlyNode
}
