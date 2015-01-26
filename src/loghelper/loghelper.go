package loghelper

import (
	"fmt"
	stdlog "log"
	"os"

	"github.com/op/go-logging"
	json "github.com/orofarne/strict-json"
)

type logConfig struct {
	Backend string
	Level   string
	Format  string
	Options json.RawMessage
}

var logBackendMap = map[string]func(*logConfig) error{
	"Console": func(cfg *logConfig) error {
		var pcfg struct {
			Color bool
		}
		err := json.Unmarshal(cfg.Options, &pcfg)
		if err != nil {
			return err
		}

		logBackend := logging.NewLogBackend(os.Stderr, "", stdlog.LstdFlags|stdlog.Lshortfile)
		if pcfg.Color {
			logBackend.Color = true
		}

		logging.SetBackend(logBackend)

		return nil
	},
	"Syslog": func(cfg *logConfig) error {
		var pcfg struct {
			Prefix string
		}
		err := json.Unmarshal(cfg.Options, &pcfg)
		if err != nil {
			return err
		}

		syslogBackend, err := logging.NewSyslogBackend(pcfg.Prefix)
		if err != nil {
			return err
		}

		logging.SetBackend(syslogBackend)

		return nil
	},
}

func IntiLog(rawCfg json.RawMessage) error {
	cfg := new(logConfig)
	err := json.Unmarshal(rawCfg, cfg)
	if err != nil {
		return err
	}

	if initializer, found := logBackendMap[cfg.Backend]; found {
		if err = initializer(cfg); err != nil {
			return err
		}
	} else {
		return err
	}

	switch cfg.Level {
	case "Critical":
		logging.SetLevel(logging.CRITICAL, "global")
	case "Error":
		logging.SetLevel(logging.ERROR, "global")
	case "Warning":
		logging.SetLevel(logging.WARNING, "global")
	case "Notice":
		logging.SetLevel(logging.NOTICE, "global")
	case "Info":
		logging.SetLevel(logging.INFO, "global")
	case "Debug":
		logging.SetLevel(logging.DEBUG, "global")
	default:
		return fmt.Errorf("Invalid log level '%s'", cfg.Level)
	}

	if cfg.Format == "" {
		cfg.Format = "[%{level}] %{message}"
	}

	formatter, err := logging.NewStringFormatter(cfg.Format)
	if err != nil {
		return err
	}
	logging.SetFormatter(formatter)

	return nil
}
