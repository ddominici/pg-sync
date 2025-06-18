package sync

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ddominici/pg-sync/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

func connString(c config.DBConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

func connectWithTimeout(connStr string, timeout time.Duration) (*pgx.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("connection timed out")
		}
		return nil, err
	}

	return conn, nil
}

func SyncTables(cfg *config.Config, log *logrus.Logger) error {
	log.Info("Connecting to source database...")
	srcConn, err := connectWithTimeout(connString(cfg.Source), 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to source DB: %w", err)
	}
	defer srcConn.Close(context.Background())

	log.Info("Connecting to target database...")
	tgtConn, err := connectWithTimeout(connString(cfg.Target), 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to target DB: %w", err)
	}
	defer tgtConn.Close(context.Background())

	ctx := context.Background()

	for _, table := range cfg.Tables {
		log.Infof("Syncing table: %s", table)

		var buf bytes.Buffer
		rows, err := srcConn.PgConn().CopyTo(ctx, &buf, fmt.Sprintf("COPY %s TO STDOUT", table))
		if err != nil {
			return fmt.Errorf("copy from source failed: %w", err)
		}
		log.Debugf("Copied %d bytes from source for table %s", rows, table)

		if _, err := tgtConn.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s", table)); err != nil {
			return fmt.Errorf("truncate failed: %w", err)
		}

		_, err = tgtConn.PgConn().CopyFrom(ctx, &buf, fmt.Sprintf("COPY %s FROM STDIN", table))
		if err != nil {
			return fmt.Errorf("copy to target failed: %w", err)
		}

		log.Infof("Table %s synced successfully", table)
	}

	return nil
}
