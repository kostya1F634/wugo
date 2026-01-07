# Repository Guidelines

## Project Structure & Module Organization
- `cmd/wugo/main.go` is the CLI entry point.
- `internal/app` orchestrates CLI parsing and application flow.
- `internal/image` handles URL/local image processing and naming.
- `internal/wallpaper` contains KDE Plasma integrations and file URI helpers.
- `go.mod` and `go.sum` define Go module metadata and dependencies.
- `Makefile` provides build/setup/link shortcuts.
- `bin/` is created by `make bin` to store the compiled binary.
- `README.md` contains user-facing usage and installation notes.

## Build, Test, and Development Commands
- `make setup`: run `go mod tidy` and `go mod download` to sync dependencies.
- `make build`: compile `./bin/wugo` from `./cmd/wugo`.
- `make bin`: create `bin/` and run setup + build.
- `make link`: symlink `bin/wugo` to `/usr/local/bin/wugo` (requires sudo).
- `go run ./cmd/wugo <image>`: run the CLI locally without building.
- `go test ./...`: run all unit tests.

## Coding Style & Naming Conventions
- Use standard Go formatting; run `go fmt ./...` or `gofmt -w cmd/wugo/*.go internal/**/*.go`.
- Indentation is gofmt tabs; keep lines readable.
- Name locals in `camelCase`; use `PascalCase` for exported identifiers if added.
- Prefer small helper functions over deep nesting (see `internal/image`).

## Testing Guidelines
- Tests live next to the code (`internal/*`), using `*_test.go`.
- Use table-driven tests for pure helpers (URL parsing, extensions, file URI building).
- For HTTP logic, use `httptest` to avoid real network calls.

## Commit & Pull Request Guidelines
- Commit messages are short, imperative, and title-cased (e.g., “Update README.md”).
- PRs should describe behavior changes, list manual test steps, and link issues if any.
- If you add flags or change CLI behavior, update `README.md` examples.

## Platform & Configuration Notes
- Targets KDE Plasma. Uses DBus (`org.kde.plasmashell`) and writes `~/.config/kscreenlockerrc`.
- Default wallpaper directory is `~/wallpapers`; override with `-d <dir>` or skip moving with `-nm`.
- Only use trusted image URLs and paths.
