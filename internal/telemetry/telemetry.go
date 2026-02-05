package telemetry

import (
	"os"
	"runtime"

	"github.com/bab-sh/bab/internal/config"
	"github.com/charmbracelet/huh"
	"github.com/google/uuid"
	"github.com/posthog/posthog-go"
	"golang.org/x/term"
)

var client posthog.Client

func Track(key, version string) {
	if key == "" || envDisabled() {
		return
	}

	cfg, err := config.Load()
	if err != nil {
		return
	}

	if cfg.Telemetry.Consent == nil {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return
		}
		consent := promptConsent()
		cfg.Telemetry.Consent = &consent
		if consent {
			cfg.Telemetry.ID = uuid.NewString()
		}
		_ = config.Save(cfg)
		if !consent {
			return
		}
	}

	if !*cfg.Telemetry.Consent {
		return
	}

	if cfg.Telemetry.ID == "" {
		cfg.Telemetry.ID = uuid.NewString()
		_ = config.Save(cfg)
	}

	maxRetries := 2
	c, err := posthog.NewWithConfig(key, posthog.Config{
		Endpoint:   "https://us.i.posthog.com",
		BatchSize:  1,
		MaxRetries: &maxRetries,
	})
	if err != nil {
		return
	}
	client = c

	_ = client.Enqueue(posthog.Capture{
		DistinctId: cfg.Telemetry.ID,
		Event:      "cli_invoked",
		Properties: map[string]interface{}{
			"$set": map[string]interface{}{
				"os":      runtime.GOOS,
				"arch":    runtime.GOARCH,
				"version": version,
			},
			"$set_once": map[string]interface{}{
				"initial_version": version,
			},
		},
	})
}

func Close() {
	if client != nil {
		_ = client.Close()
	}
}

func envDisabled() bool {
	if os.Getenv("DO_NOT_TRACK") == "1" || os.Getenv("BAB_NO_TELEMETRY") == "1" {
		return true
	}
	ci := os.Getenv("CI")
	return ci == "true" || ci == "1"
}

func promptConsent() bool {
	var consent bool
	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Help improve bab by sharing anonymous usage data?").
			Description("This includes: OS, architecture, and version.\nNo task names, file paths, or personal data is collected.").
			Value(&consent),
	)).Run()
	if err != nil {
		return false
	}
	return consent
}
