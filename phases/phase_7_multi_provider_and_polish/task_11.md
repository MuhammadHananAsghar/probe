# Task 11: Cross-Compile & goreleaser CI/CD

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Set up goreleaser to produce cross-compiled binaries for all target platforms and automate GitHub releases.

## Requirements

- `.goreleaser.yml` config: build targets `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
- Strip debug symbols (`-s -w`); compress with UPX for smaller binaries
- GitHub Actions workflow: `.github/workflows/release.yml` triggers on `git tag v*`
- Release artifacts: `probe_linux_amd64.tar.gz`, `probe_darwin_arm64.tar.gz`, `probe_windows_amd64.zip`, checksums file
- Embed version string and build date via `-ldflags "-X main.version=..."`
- `probe version` prints version, build date, and Go version

## Files

- `.goreleaser.yml`
- `.github/workflows/release.yml`
- `cmd/probe/main.go`
