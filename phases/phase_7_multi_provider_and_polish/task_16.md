# Task 16: README, Terminal GIFs & Benchmarks

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Write the README, create terminal GIF demos, and run benchmarks to prove probe adds minimal latency overhead.

## Requirements

- README sections: What is Probe, Quick Start (30-second setup), Installation, Proxy mode vs Base URL mode, All CLI commands, Configuration reference, FAQ
- Terminal GIF: `probe listen` → requests flowing in → press Enter on a request → detail view → press `s` for stream view
- Benchmark: measure proxy overhead with `go test -bench=BenchmarkProxy` — prove < 2ms added latency for non-streaming
- Benchmark: measure streaming overhead — prove < 1ms per chunk added latency
- Benchmark results included in README as a table
- `make bench` target runs benchmarks and prints results

## Files

- `README.md`
- `docs/`
- `Makefile`
- `internal/proxy/proxy_bench_test.go`
