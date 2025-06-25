package main

import (
	"flag"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"

	"github.com/ddominici/pg-sync/internal/config"
	"github.com/ddominici/pg-sync/internal/notify"
	"github.com/ddominici/pg-sync/internal/sync"
	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config file")
	verbose := flag.Bool("v", false, "Enable verbose logging")
	logFile := flag.String("logfile", "log/pg-sync.log", "Path to log file")
	flag.Parse()

	// Setup logging
	log := logrus.New()

	logWriter := &lumberjack.Logger{
		Filename:   *logFile,
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     28,   // days
		Compress:   true, // gzip
	}

	multiWriter := io.MultiWriter(os.Stdout, logWriter)
	log.SetOutput(multiWriter)

	if *verbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	//log.SetFormatter(&logrus.TextFormatter{
	//	FullTimestamp: true,
	//})

	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
		PrettyPrint:     false, // true if you want human-readable JSON
	})

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := sync.SyncTables(cfg, log); err != nil {
		log.Errorf("Sync failed: %v", err)

		emailErr := notify.SendErrorEmail(
			cfg.Email,
			"pg-sync error: Sync Failed",
			fmt.Sprintf("The sync failed with error:\n\n%v", err),
		)

		if emailErr != nil {
			log.Errorf("Failed to send error notification email: %v", emailErr)
		}

		os.Exit(1)
	}

	log.Info("All tables synced successfully.")
}
