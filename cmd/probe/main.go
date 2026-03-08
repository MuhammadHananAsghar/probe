// Package main is the entrypoint for the probe CLI.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/MuhammadHananAsghar/probe/internal/analyze"
	"github.com/MuhammadHananAsghar/probe/internal/cost"
	"github.com/MuhammadHananAsghar/probe/internal/proxy"
	"github.com/MuhammadHananAsghar/probe/internal/store"
	"github.com/MuhammadHananAsghar/probe/internal/tui"
	"github.com/MuhammadHananAsghar/probe/pkg/config"
	"github.com/MuhammadHananAsghar/probe/pkg/logger"
)

// version is injected at build time via -ldflags.
var version = "dev"

func main() {
	root := buildRoot()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "probe",
		Short: "Universal LLM API debugger",
		Long:  "Probe intercepts every LLM API call your app makes — tokens, cost, latency, streaming, tool calls.",
	}

	root.AddCommand(buildListen())
	root.AddCommand(buildVersion())

	return root
}

// listenFlags groups all flags for the listen command.
type listenFlags struct {
	port          int
	dashboardPort int
	noBrowser     bool
	noTLS         bool
	filter        string
	debug         bool
}

func buildListen() *cobra.Command {
	var flags listenFlags

	cmd := &cobra.Command{
		Use:   "listen",
		Short: "Start intercepting LLM API traffic",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListen(flags)
		},
	}

	cmd.Flags().IntVarP(&flags.port, "port", "p", 0, "Proxy port (default from config, fallback 8080)")
	cmd.Flags().IntVar(&flags.dashboardPort, "dashboard-port", 0, "Dashboard port (default from config, fallback 4041)")
	cmd.Flags().BoolVar(&flags.noBrowser, "no-browser", false, "Don't auto-open browser")
	cmd.Flags().BoolVar(&flags.noTLS, "no-tls", false, "Skip TLS interception (base URL mode only)")
	cmd.Flags().StringVar(&flags.filter, "filter", "", "Only intercept matching provider or model=name")
	cmd.Flags().BoolVar(&flags.debug, "debug", false, "Enable debug logging")

	return cmd
}

func runListen(flags listenFlags) error {
	logger.Init(flags.debug)
	log := logger.Get()

	// Load config (falls back to defaults if ~/.probe/config.yaml absent).
	cfg, err := config.Load()
	if err != nil {
		log.Warn().Err(err).Msg("could not load config, using defaults")
		cfg = config.Default()
	}

	// CLI flags override config.
	if flags.port != 0 {
		cfg.Proxy.Port = flags.port
	}
	if flags.dashboardPort != 0 {
		cfg.Proxy.DashboardPort = flags.dashboardPort
	}

	// Build the pricing database.
	pricingDB, err := cost.NewDB(cfg.Pricing.Custom)
	if err != nil {
		return fmt.Errorf("loading pricing database: %w", err)
	}
	_ = pricingDB // will be threaded into proxy in a later phase

	// Set up in-memory store and session tracker.
	mem := store.NewMemory(cfg.Storage.RingBufferSize)
	tracker := analyze.NewTracker()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start tracker in background (subscribes to store events).
	tracker.Start(ctx, mem)

	// Build proxy server config.
	proxyCfg := proxy.Config{
		Port:           cfg.Proxy.Port,
		DashboardPort:  cfg.Proxy.DashboardPort,
		StallThreshold: cfg.Proxy.StallThreshold,
		NoBrowser:      flags.noBrowser,
		NoTLS:          flags.noTLS,
		Filter:         flags.filter,
	}

	srv, err := proxy.New(proxyCfg, mem, tracker)
	if err != nil {
		return fmt.Errorf("creating proxy: %w", err)
	}

	// Print startup banner.
	printBanner(cfg.Proxy.Port, cfg.Proxy.DashboardPort)

	// Start proxy in background.
	proxyErr := make(chan error, 1)
	go func() {
		proxyErr <- srv.Start(ctx)
	}()

	// Build and start TUI.
	proxyAddr := fmt.Sprintf("localhost:%d", cfg.Proxy.Port)
	dashAddr := fmt.Sprintf("localhost:%d", cfg.Proxy.DashboardPort)
	app := tui.New(proxyAddr, dashAddr, tracker, srv.Events())

	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	// TUI exited (user pressed q) — cancel context to stop proxy.
	cancel()

	// Wait for proxy to shut down or pick up its error.
	select {
	case err := <-proxyErr:
		if err != nil && err != context.Canceled {
			return fmt.Errorf("proxy: %w", err)
		}
	}

	// Print session summary.
	stats := tracker.Stats()
	fmt.Printf("\nSession complete: %d requests | %s total\n",
		stats.RequestCount, formatCostSummary(stats.TotalCost))

	return nil
}

func printBanner(port, dashPort int) {
	fmt.Printf(`
  Probe — LLM API Debugger

  Proxy:      http://localhost:%d
  Dashboard:  http://localhost:%d

  Tip: Run your app with HTTPS_PROXY=http://localhost:%d

  Waiting for requests...

`, port, dashPort, port)
}

func formatCostSummary(cost float64) string {
	if cost == 0 {
		return "$0.00"
	}
	if cost < 0.01 {
		return fmt.Sprintf("$%.4f", cost)
	}
	return fmt.Sprintf("$%.2f", cost)
}

func buildVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("probe %s\n", version)
		},
	}
}
