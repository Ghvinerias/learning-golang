package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"slick-autobuild/internal/config"
	"slick-autobuild/internal/planner"
)

var (
	flagConfig      = flag.String("config", "build.yaml", "Path to config file")
	flagConcurrency = flag.Int("concurrency", 0, "Max concurrent builds (default: CPU cores)")
	flagJSON        = flag.Bool("json", false, "JSON logging output")
	flagNoCache     = flag.Bool("no-cache", false, "Disable build cache")
	flagOnly        = flag.String("only", "", "Comma separated project paths to include")
	flagDryRun      = flag.Bool("dry-run", false, "Plan only; do not execute builds")
	flagVersion     = flag.Bool("version", false, "Print version and exit")
)

const version = "0.0.1-dev"

func main() {
	flag.Parse()

	if *flagVersion {
		fmt.Println(version)
		return
	}

	args := flag.Args()
	cmd := "build"
	if len(args) > 0 {
		cmd = args[0]
	}

	// Basic switch – only plan/build for MVP scaffold
	switch cmd {
	case "plan":
		if err := runPlan(); err != nil {
			fatal(err)
		}
	case "build":
		if err := runBuild(); err != nil {
			fatal(err)
		}
	default:
		fatal(fmt.Errorf("unknown command: %s", cmd))
	}
}

func runPlan() error {
	cfg, err := config.Load(*flagConfig)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	selected := parseOnly()
	plan := planner.Expand(cfg, selected)
	printPlan(plan)
	return nil
}

func runBuild() error {
	if *flagDryRun {
		return runPlan()
	}
	cfg, err := config.Load(*flagConfig)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	selected := parseOnly()
	plan := planner.Expand(cfg, selected)
	conc := *flagConcurrency
	if conc <= 0 {
		conc = runtime.NumCPU()
	}
	fmt.Printf("Executing %d build task(s) with concurrency=%d\n", len(plan.Tasks), conc)
	// Placeholder: simulate builds
	ctx := context.Background()
	sem := make(chan struct{}, conc)
	for _, t := range plan.Tasks {
		sem <- struct{}{}
		go func(task planner.Task) {
			defer func() { <-sem }()
			start := time.Now()
			fmt.Printf("[BUILD] %s (%s %s)\n", task.Path, task.Kind, task.Version)
			// Simulate build runtime
			time.Sleep(200 * time.Millisecond)
			fmt.Printf("[OK] %s in %v\n", task.Path, time.Since(start))
		}(t)
	}
	// drain
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}
	fmt.Println("All tasks completed (simulated).")
	_ = ctx
	return nil
}

func parseOnly() map[string]struct{} {
	m := map[string]struct{}{}
	if *flagOnly == "" {
		return m
	}
	for _, p := range strings.Split(*flagOnly, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			m[p] = struct{}{}
		}
	}
	return m
}

func printPlan(p planner.Plan) {
	fmt.Printf("Plan: %d task(s)\n", len(p.Tasks))
	for _, t := range p.Tasks {
		fmt.Printf(" - %s | kind=%s version=%s\n", t.Path, t.Kind, t.Version)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
