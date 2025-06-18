package sync

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ddominici/pg-sync/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

func connString(c config.DBConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

func SyncTables(cfg *config.Config, log *logrus.Logger) error {
	ctx := context.Background()

	srcConn, err := pgx.Connect(ctx, connString(cfg.Source))
	if err != nil {
		return fmt.Errorf("failed to connect to source DB: %w", err)
	}
	defer srcConn.Close(ctx)

	tgtConn, err := pgx.Connect(ctx, connString(cfg.Target))
	if err != nil {
		return fmt.Errorf("failed to connect to target DB: %w", err)
	}
	defer tgtConn.Close(ctx)

	for _, table := range cfg.Tables {
		log.Infof("Syncing table: %s", table)

		var buf bytes.Buffer
		rows, err := srcConn.PgConn().CopyTo(ctx, &buf, fmt.Sprintf("COPY %s TO STDOUT", table))
		if err != nil {
			return fmt.Errorf("copy from source failed: %w", err)
		}
		log.Debugf("Copied %d bytes from source for table %s", rows, table)

		// Truncate target table
		if _, err := tgtConn.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s", table)); err != nil {
			return fmt.Errorf("truncate failed: %w", err)
		}

		// Copy into target table
		_, err = tgtConn.PgConn().CopyFrom(ctx, &buf, fmt.Sprintf("COPY %s FROM STDIN", table))
		if err != nil {
			return fmt.Errorf("copy to target failed: %w", err)
		}

		log.Infof("Table %s synced successfully", table)
	}

	return nil
}
