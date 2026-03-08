// Package main is the entrypoint for the probe CLI.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/MuhammadHananAsghar/probe/internal/analyze"
	"github.com/MuhammadHananAsghar/probe/internal/cost"
	"github.com/MuhammadHananAsghar/probe/internal/dashboard"
	"github.com/MuhammadHananAsghar/probe/internal/export"
	"github.com/MuhammadHananAsghar/probe/internal/proxy"
	"github.com/MuhammadHananAsghar/probe/internal/replay"
	"github.com/MuhammadHananAsghar/probe/internal/store"
	"github.com/MuhammadHananAsghar/probe/internal/tui"
	"github.com/MuhammadHananAsghar/probe/pkg/config"
	"github.com/MuhammadHananAsghar/probe/pkg/logger"
)

// version and buildDate are injected at build time via -ldflags.
var (
	version   = "dev"
	buildDate = "unknown"
)

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
	root.AddCommand(buildInspect())
	root.AddCommand(buildReplay())
	root.AddCommand(buildCompare())
	root.AddCommand(buildExport())
	root.AddCommand(buildAnalyze())
	root.AddCommand(buildHistory())

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
	persist       bool
	alertCost     float64
	alertLatency  time.Duration
	alertError    bool
	quiet         bool
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

	cmd.Flags().IntVarP(&flags.port, "port", "p", 0, "Proxy port (default from config, fallback 9000)")
	cmd.Flags().IntVar(&flags.dashboardPort, "dashboard-port", 0, "Dashboard port (default from config, fallback 9001)")
	cmd.Flags().BoolVar(&flags.noBrowser, "no-browser", false, "Don't auto-open browser")
	cmd.Flags().BoolVar(&flags.noTLS, "no-tls", false, "Skip TLS interception (base URL mode only)")
	cmd.Flags().StringVar(&flags.filter, "filter", "", "Only intercept matching provider or model=name")
	cmd.Flags().BoolVar(&flags.debug, "debug", false, "Enable debug logging")
	cmd.Flags().BoolVar(&flags.persist, "persist", false, "Persist requests to ~/.probe/history.db")
	cmd.Flags().Float64Var(&flags.alertCost, "alert-cost", 0, "Alert when session total cost exceeds this amount")
	cmd.Flags().DurationVar(&flags.alertLatency, "alert-latency", 0, "Alert when a request exceeds this latency")
	cmd.Flags().BoolVar(&flags.alertError, "alert-error", false, "Alert on any 4xx/5xx response")
	cmd.Flags().BoolVar(&flags.quiet, "quiet", false, "Suppress TUI, only log errors to stderr")

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

	// Fetch live pricing from LiteLLM (with 5s timeout + local cache fallback).
	// Returns nil map on network failure — embedded data is used as fallback.
	fetchCtx, fetchCancel := context.WithTimeout(context.Background(), 5*time.Second)
	livePricing, _ := cost.FetchLiteLLM(fetchCtx)
	fetchCancel()
	if livePricing != nil {
		log.Debug().Msgf("loaded %d model prices from LiteLLM", len(livePricing))
	} else {
		log.Debug().Msg("using embedded pricing data (LiteLLM fetch unavailable)")
	}

	// Build the pricing database.
	pricingDB, err := cost.NewDB(cfg.Pricing.Custom, livePricing)
	if err != nil {
		return fmt.Errorf("loading pricing database: %w", err)
	}

	// Set up store (in-memory, or SQLite if --persist).
	var mem store.Store
	if flags.persist || cfg.Storage.Persist {
		dbPath, dbErr := store.DefaultDBPath()
		if dbErr != nil {
			return fmt.Errorf("sqlite path: %w", dbErr)
		}
		sq, sqErr := store.NewSQLite(dbPath, cfg.Storage.RingBufferSize)
		if sqErr != nil {
			return fmt.Errorf("sqlite: %w", sqErr)
		}
		defer sq.Close()
		// Run cleanup on startup.
		if n, cleanErr := sq.Cleanup(cfg.Storage.RetentionDays); cleanErr == nil && n > 0 {
			log.Info().Msgf("sqlite cleanup: deleted %d old requests", n)
		}
		mem = sq
		log.Info().Msgf("persisting requests to %s", dbPath)
	} else {
		mem = store.NewMemory(cfg.Storage.RingBufferSize)
	}
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

	srv, err := proxy.New(proxyCfg, mem, tracker, pricingDB)
	if err != nil {
		return fmt.Errorf("creating proxy: %w", err)
	}

	// Print startup banner.
	printBanner(cfg.Proxy.Port, cfg.Proxy.DashboardPort)

	// Start dashboard server (subscribes independently to the store).
	dashboardCh := mem.Subscribe()
	defer mem.Unsubscribe(dashboardCh)
	dashAddr := fmt.Sprintf(":%d", cfg.Proxy.DashboardPort)
	dash := dashboard.New(dashAddr, mem, dashboardCh, pricingDB)
	dashboardErr := make(chan error, 1)
	go func() {
		dashboardErr <- dash.Start(ctx)
	}()

	// Auto-open browser unless --no-browser.
	if !flags.noBrowser {
		go openBrowser(fmt.Sprintf("http://localhost:%d", cfg.Proxy.DashboardPort))
	}

	// Start proxy in background.
	proxyErr := make(chan error, 1)
	go func() {
		proxyErr <- srv.Start(ctx)
	}()

	// Build and start TUI.
	proxyAddr := fmt.Sprintf("localhost:%d", cfg.Proxy.Port)
	dashDisplayAddr := fmt.Sprintf("localhost:%d", cfg.Proxy.DashboardPort)
	app := tui.New(proxyAddr, dashDisplayAddr, tracker, srv.Events())

	// Wire alert thresholds (CLI flags take priority over config).
	alertCfg := tui.AlertConfig{
		CostThreshold:    cfg.Alerts.CostThreshold,
		LatencyThreshold: cfg.Alerts.LatencyThreshold,
		AlertOnError:     cfg.Alerts.AlertOnError,
	}
	if flags.alertCost > 0 {
		alertCfg.CostThreshold = flags.alertCost
	}
	if flags.alertLatency > 0 {
		alertCfg.LatencyThreshold = flags.alertLatency
	}
	if flags.alertError {
		alertCfg.AlertOnError = true
	}
	app.WithAlerts(alertCfg)

	if flags.quiet {
		// Quiet mode: no TUI — just block until signal.
		fmt.Fprintf(os.Stderr, "probe %s listening on :%d (quiet mode)\n", version, cfg.Proxy.Port)
		<-ctx.Done()
	} else {
		p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("tui: %w", err)
		}
		// TUI exited (user pressed q) — cancel context to stop proxy.
		cancel()
	}

	// Wait for proxy/dashboard to shut down or pick up errors.
	select {
	case err := <-proxyErr:
		if err != nil && err != context.Canceled {
			return fmt.Errorf("proxy: %w", err)
		}
	case err := <-dashboardErr:
		if err != nil && err != context.Canceled {
			log.Warn().Err(err).Msg("dashboard error")
		}
	}

	// Print session summary.
	stats := tracker.Stats()
	fmt.Printf("\nSession complete: %d requests | %s total\n",
		stats.RequestCount, formatCostSummary(stats.TotalCost))

	return nil
}

// openBrowser opens url in the system default browser.
func openBrowser(url string) {
	// Small delay so the server is ready before the browser connects.
	time.Sleep(300 * time.Millisecond)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
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
	return fmt.Sprintf("$%.8f", cost)
}

func buildVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("probe %s (built %s, %s)\n", version, buildDate, runtime.Version())
		},
	}
}

// buildInspect constructs the inspect subcommand, which provides CLI-level
// access to a captured request by its sequence number.
func buildInspect() *cobra.Command {
	var (
		showStream bool
		replayStr  bool
		speed      float64
		curlFlag   bool
		curlLang   string
	)

	cmd := &cobra.Command{
		Use:   "inspect <seq>",
		Short: "Inspect a captured request by sequence number",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInspect(args[0], showStream, replayStr, speed, curlFlag, curlLang)
		},
	}

	cmd.Flags().BoolVar(&showStream, "stream", false, "Show stream chunk timeline")
	cmd.Flags().BoolVar(&replayStr, "replay", false, "Replay stream at original timing")
	cmd.Flags().Float64Var(&speed, "speed", 1.0, "Replay speed multiplier")
	cmd.Flags().BoolVar(&curlFlag, "curl", false, "Print a curl command reproducing this request")
	cmd.Flags().StringVar(&curlLang, "lang", "curl", "Output language: curl, python, node")
	return cmd
}

func runInspect(seqStr string, showStream bool, replayFlag bool, speed float64, curlFlag bool, curlLang string) error {
	if !curlFlag {
		fmt.Println("inspect: connect to a running probe session to inspect requests.")
		fmt.Println("Use 'probe listen' and press Enter on a request to view details.")
		fmt.Println("Use --curl to generate a curl command for a captured request.")
		return nil
	}

	cfg, _ := config.Load()
	mem := store.NewMemory(cfg.Storage.RingBufferSize)
	var seq int
	if _, err := fmt.Sscanf(seqStr, "%d", &seq); err != nil {
		return fmt.Errorf("invalid sequence number %q", seqStr)
	}
	req := mem.GetBySeq(seq)
	if req == nil {
		return fmt.Errorf("request #%d not found in current session (persistence coming in Phase 6)", seq)
	}
	fmt.Println(export.ToCurl(req, curlLang))
	return nil
}

// ── replay command ────────────────────────────────────────────────────────────

func buildReplay() *cobra.Command {
	var (
		modelFlag    string
		providerFlag string
		tempFlag     float64
		setTemp      bool
		maxTokFlag   int
		setMaxTok    bool
		systemFlag   string
		exportFlag   string
	)

	cmd := &cobra.Command{
		Use:   "replay <seq>",
		Short: "Re-send a captured request, optionally with parameter changes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReplay(args[0], replayOpts{
				model:       modelFlag,
				provider:    store.ProviderName(providerFlag),
				hasTemp:     setTemp,
				temperature: tempFlag,
				hasMaxTok:   setMaxTok,
				maxTokens:   maxTokFlag,
				system:      systemFlag,
				exportFile:  exportFlag,
			})
		},
	}

	cmd.Flags().StringVar(&modelFlag, "model", "", "Override model")
	cmd.Flags().StringVar(&providerFlag, "provider", "", "Translate to provider (openai|anthropic)")
	cmd.Flags().Float64Var(&tempFlag, "temperature", 0, "Override temperature")
	cmd.Flags().BoolVar(&setTemp, "set-temperature", false, "Apply --temperature override")
	cmd.Flags().IntVar(&maxTokFlag, "max-tokens", 0, "Override max_tokens")
	cmd.Flags().BoolVar(&setMaxTok, "set-max-tokens", false, "Apply --max-tokens override")
	cmd.Flags().StringVar(&systemFlag, "system", "", "Replace system prompt")
	cmd.Flags().StringVar(&exportFlag, "export", "", "Export comparison to markdown file")

	return cmd
}

type replayOpts struct {
	model       string
	provider    store.ProviderName
	hasTemp     bool
	temperature float64
	hasMaxTok   bool
	maxTokens   int
	system      string
	exportFile  string
}

func runReplay(seqStr string, opts replayOpts) error {
	cfg, _ := config.Load()
	mem := store.NewMemory(cfg.Storage.RingBufferSize)
	fetchCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	livePricing, _ := cost.FetchLiteLLM(fetchCtx)
	cancel()
	pricingDB, _ := cost.NewDB(cfg.Pricing.Custom, livePricing)

	var seq int
	if _, err := fmt.Sscanf(seqStr, "%d", &seq); err != nil {
		return fmt.Errorf("invalid sequence number %q", seqStr)
	}

	orig := mem.GetBySeq(seq)
	if orig == nil {
		return fmt.Errorf("request #%d not found in current session (persistence coming in Phase 6)", seq)
	}

	engine := replay.New(mem, pricingDB)
	ro := replay.Options{
		Model:        opts.model,
		Provider:     opts.provider,
		SystemPrompt: opts.system,
	}
	if opts.hasTemp {
		t := opts.temperature
		ro.Temperature = &t
	}
	if opts.hasMaxTok {
		n := opts.maxTokens
		ro.MaxTokens = &n
	}

	fmt.Printf("Replaying request #%d...\n", seq)
	result, err := engine.Replay(context.Background(), orig, ro)
	if err != nil {
		return fmt.Errorf("replay failed: %w", err)
	}

	fmt.Printf("\nReplay complete (#%d):\n", result.Req.Seq)
	if len(result.ParameterDiffs) > 0 {
		fmt.Printf("  Changes: %s\n", strings.Join(result.ParameterDiffs, ", "))
	}
	fmt.Printf("  Model:    %s\n", result.Req.Model)
	fmt.Printf("  Status:   %d\n", result.Req.StatusCode)
	fmt.Printf("  Latency:  %dms\n", result.Req.Latency.Milliseconds())
	if result.Req.PricingKnown {
		fmt.Printf("  Cost:     $%.8f\n", result.Req.TotalCost)
	}
	fmt.Printf("  Tokens:   %d in / %d out\n", result.Req.InputTokens, result.Req.OutputTokens)

	if opts.exportFile != "" {
		cmp := replay.Compare(orig, result.Req)
		md := export.ComparisonMarkdown(cmp, version)
		if err := os.WriteFile(opts.exportFile, []byte(md), 0o644); err != nil {
			return fmt.Errorf("writing export: %w", err)
		}
		fmt.Printf("\nExported comparison to %s\n", opts.exportFile)
	}

	return nil
}

// ── compare command ───────────────────────────────────────────────────────────

func buildCompare() *cobra.Command {
	var (
		modelsFlag string
		exportFlag string
	)

	cmd := &cobra.Command{
		Use:   "compare <seqA> [seqB]",
		Short: "Compare two captured requests, or fan out to multiple models",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompare(args, modelsFlag, exportFlag)
		},
	}

	cmd.Flags().StringVar(&modelsFlag, "models", "", "Comma-separated models for multi-model comparison")
	cmd.Flags().StringVar(&exportFlag, "export", "", "Export comparison to markdown file")

	return cmd
}

func runCompare(args []string, modelsFlag string, exportFlag string) error {
	cfg, _ := config.Load()
	mem := store.NewMemory(cfg.Storage.RingBufferSize)
	fetchCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	livePricing, _ := cost.FetchLiteLLM(fetchCtx)
	cancel()
	pricingDB, _ := cost.NewDB(cfg.Pricing.Custom, livePricing)
	engine := replay.New(mem, pricingDB)

	var seqA int
	if _, err := fmt.Sscanf(args[0], "%d", &seqA); err != nil {
		return fmt.Errorf("invalid sequence number %q", args[0])
	}
	reqA := mem.GetBySeq(seqA)
	if reqA == nil {
		return fmt.Errorf("request #%d not found (persistence coming in Phase 6)", seqA)
	}

	// Multi-model compare.
	if modelsFlag != "" {
		models := strings.Split(modelsFlag, ",")
		fmt.Printf("Comparing %d models for request #%d...\n\n", len(models), seqA)
		var results []*replay.Result
		for _, m := range models {
			m = strings.TrimSpace(m)
			fmt.Printf("  Sending to %s...\n", m)
			r, err := engine.Replay(context.Background(), reqA, replay.Options{Model: m})
			if err != nil {
				fmt.Printf("  ⚠ %s failed: %v\n", m, err)
				continue
			}
			results = append(results, r)
		}
		mr := replay.BuildMultiTable(results)
		fmt.Println("\n" + replay.RenderMultiText(mr))
		if exportFlag != "" {
			md := export.MultiComparisonMarkdown(mr, version)
			if err := os.WriteFile(exportFlag, []byte(md), 0o644); err != nil {
				return fmt.Errorf("writing export: %w", err)
			}
			fmt.Printf("Exported to %s\n", exportFlag)
		}
		return nil
	}

	// Two-request diff.
	if len(args) < 2 {
		return fmt.Errorf("provide two sequence numbers or --models for multi-model compare")
	}
	var seqB int
	if _, err := fmt.Sscanf(args[1], "%d", &seqB); err != nil {
		return fmt.Errorf("invalid sequence number %q", args[1])
	}
	reqB := mem.GetBySeq(seqB)
	if reqB == nil {
		return fmt.Errorf("request #%d not found", seqB)
	}

	cmp := replay.Compare(reqA, reqB)
	fmt.Println(replay.RenderComparisonText(cmp))

	if exportFlag != "" {
		md := export.ComparisonMarkdown(cmp, version)
		if err := os.WriteFile(exportFlag, []byte(md), 0o644); err != nil {
			return fmt.Errorf("writing export: %w", err)
		}
		fmt.Printf("Exported to %s\n", exportFlag)
	}

	return nil
}

// ── export command ────────────────────────────────────────────────────────────

func buildExport() *cobra.Command {
	var (
		formatFlag  string
		outputFlag  string
		filterFlag  []string
		requestFlag int
		compactFlag bool
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export captured requests to HAR, JSON, NDJSON, or Markdown",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(formatFlag, outputFlag, filterFlag, requestFlag, compactFlag)
		},
	}

	cmd.Flags().StringVar(&formatFlag, "format", "json", "Export format: har, json, ndjson, markdown")
	cmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringArrayVar(&filterFlag, "filter", nil, "Filter: key=value (provider, model, status)")
	cmd.Flags().IntVar(&requestFlag, "request", 0, "Export a single request by seq (markdown only)")
	cmd.Flags().BoolVar(&compactFlag, "compact", false, "Compact JSON output (no indentation)")

	return cmd
}

func runExport(format, output string, filters []string, requestSeq int, compact bool) error {
	cfg, _ := config.Load()
	mem := store.NewMemory(cfg.Storage.RingBufferSize)

	requests := export.FilterRequests(mem.All(), filters)
	if len(requests) == 0 {
		return fmt.Errorf("no requests found in current session (start 'probe listen' first)")
	}

	var data []byte
	var err error

	switch format {
	case "har":
		har := export.ToHAR(requests, version)
		data, err = export.HARToJSON(har)
	case "json":
		data, err = export.ToJSON(requests, compact)
	case "ndjson":
		data, err = export.ToNDJSON(requests)
	case "markdown", "md":
		if requestSeq > 0 {
			req := mem.GetBySeq(requestSeq)
			if req == nil {
				return fmt.Errorf("request #%d not found", requestSeq)
			}
			data = []byte(export.RequestMarkdown(req, version))
		} else {
			data = []byte(export.SessionMarkdown(requests, version))
		}
	default:
		return fmt.Errorf("unknown format %q: use har, json, ndjson, or markdown", format)
	}

	if err != nil {
		return fmt.Errorf("generating export: %w", err)
	}

	if output == "" {
		fmt.Print(string(data))
		return nil
	}

	if err := os.WriteFile(output, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", output, err)
	}
	fmt.Printf("Exported %d requests to %s (%s format)\n", len(requests), output, format)
	return nil
}

// ── analyze command ───────────────────────────────────────────────────────────

func buildAnalyze() *cobra.Command {
	var (
		wastFlag  bool
		dupFlag   bool
	)

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze captured requests for waste, duplicates, and inefficiencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalyze(wastFlag, dupFlag)
		},
	}

	cmd.Flags().BoolVar(&wastFlag, "waste", false, "Report on token waste patterns")
	cmd.Flags().BoolVar(&dupFlag, "duplicates", false, "Report on duplicate requests")

	return cmd
}

func runAnalyze(wasteMode, dupMode bool) error {
	cfg, _ := config.Load()
	mem := store.NewMemory(cfg.Storage.RingBufferSize)
	requests := mem.All()

	if len(requests) == 0 {
		return fmt.Errorf("no requests found in current session (start 'probe listen' first)")
	}

	if !wasteMode && !dupMode {
		// Default: show both.
		wasteMode = true
		dupMode = true
	}

	if dupMode {
		groups := analyze.DetectDuplicates(requests)
		if len(groups) == 0 {
			fmt.Println("No duplicate requests detected.")
		} else {
			fmt.Printf("Found %d duplicate groups:\n\n", len(groups))
			var totalSavings float64
			for _, g := range groups {
				seqs := make([]string, len(g.Requests))
				for i, r := range g.Requests {
					seqs[i] = fmt.Sprintf("#%d", r.Seq)
				}
				fmt.Printf("  Requests %s are identical (model: %s)\n", strings.Join(seqs, ", "), g.Requests[0].Model)
				if g.PotentialSavings > 0 {
					fmt.Printf("  → Adding a cache would save $%.6f\n", g.PotentialSavings)
					totalSavings += g.PotentialSavings
				}
			}
			if totalSavings > 0 {
				fmt.Printf("\nTotal potential savings: $%.6f\n", totalSavings)
			}
		}
		fmt.Println()
	}

	if wasteMode {
		report := analyze.AnalyzeWaste(requests)
		if len(report.Suggestions) == 0 {
			fmt.Println("No token waste patterns detected.")
		} else {
			fmt.Printf("Token waste analysis (%d suggestions):\n\n", len(report.Suggestions))
			for _, s := range report.Suggestions {
				fmt.Printf("  • %s\n", s)
			}
		}
	}

	return nil
}

// ── history command ───────────────────────────────────────────────────────────

func buildHistory() *cobra.Command {
	var (
		costFlag    bool
		errorsFlag  bool
		modelFlag   string
		cleanupFlag bool
		limitFlag   int
	)

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Browse persisted request history from ~/.probe/history.db",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHistory(costFlag, errorsFlag, modelFlag, cleanupFlag, limitFlag)
		},
	}

	cmd.Flags().BoolVar(&costFlag, "cost", false, "Show cost summary")
	cmd.Flags().BoolVar(&errorsFlag, "errors", false, "Show only error requests")
	cmd.Flags().StringVar(&modelFlag, "model", "", "Filter by model name")
	cmd.Flags().BoolVar(&cleanupFlag, "cleanup", false, "Run retention cleanup and exit")
	cmd.Flags().IntVar(&limitFlag, "limit", 100, "Maximum number of requests to show")

	return cmd
}

func runHistory(costMode, errorsOnly bool, modelFilter string, cleanup bool, limit int) error {
	dbPath, err := store.DefaultDBPath()
	if err != nil {
		return fmt.Errorf("sqlite path: %w", err)
	}

	cfg, _ := config.Load()
	sq, err := store.NewSQLite(dbPath, cfg.Storage.RingBufferSize)
	if err != nil {
		return fmt.Errorf("opening history db: %w (run 'probe listen --persist' first)", err)
	}
	defer sq.Close()

	if cleanup {
		n, cleanErr := sq.Cleanup(cfg.Storage.RetentionDays)
		if cleanErr != nil {
			return fmt.Errorf("cleanup: %w", cleanErr)
		}
		fmt.Printf("Deleted %d requests older than %d days\n", n, cfg.Storage.RetentionDays)
		return nil
	}

	filter := func(r *store.Request) bool {
		if errorsOnly && r.StatusCode < 400 && r.Status != store.StatusError {
			return false
		}
		if modelFilter != "" && r.Model != modelFilter {
			return false
		}
		return true
	}

	requests, err := sq.History(limit, filter)
	if err != nil {
		return fmt.Errorf("reading history: %w", err)
	}

	if len(requests) == 0 {
		fmt.Println("No matching requests found.")
		return nil
	}

	if costMode {
		var total float64
		modelCosts := make(map[string]float64)
		for _, r := range requests {
			total += r.TotalCost
			modelCosts[r.Model] += r.TotalCost
		}
		fmt.Printf("%-30s %10s\n", "Model", "Cost")
		fmt.Printf("%-30s %10s\n", strings.Repeat("-", 30), strings.Repeat("-", 10))
		for m, c := range modelCosts {
			fmt.Printf("%-30s $%9.6f\n", m, c)
		}
		fmt.Printf("\nTotal: $%.6f across %d requests\n", total, len(requests))
		return nil
	}

	// Default: table output.
	fmt.Printf("%-4s %-12s %-30s %-8s %-10s %-8s\n", "#", "Time", "Model", "Status", "Latency", "Cost")
	fmt.Println(strings.Repeat("-", 80))
	for _, r := range requests {
		costStr := "n/a"
		if r.PricingKnown {
			costStr = fmt.Sprintf("$%.6f", r.TotalCost)
		}
		fmt.Printf("%-4d %-12s %-30s %-8s %-10s %-8s\n",
			r.Seq,
			r.StartedAt.Format("15:04:05"),
			r.Model,
			string(r.Status),
			fmt.Sprintf("%dms", r.Latency.Milliseconds()),
			costStr,
		)
	}
	fmt.Printf("\n%d requests shown\n", len(requests))
	return nil
}
