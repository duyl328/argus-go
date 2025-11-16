# Repository Guidelines

## Project Structure & Module Organization
- Root Go service is in `./` with `main.go`, `go.mod`, and `config.yaml`.
- Core backend lives in `internal/` (API, handlers, services, repositories, workflows, utilities).
- Reusable libraries are in `pkg/` (logging, SSE, system utilities, image helpers).
- HTTP request examples live in `tests/` (`*.http`) and complement Go tests.
- Supporting assets and tools: `tools/` (external binaries), `scripts/` (build), `db/`, `logs/`, `examples/`, `doc/`.

## Build, Test, and Development Commands
- Install deps: `go mod download`.
- Run in dev: `go run main.go` (uses `config.yaml`).
- Build local binary: `go build -o rear.exe .`.
- Cross-platform build: `bash scripts/build.sh` (from `rear/`, requires Bash and Go).
- Run tests: `go test ./...` (unit/integration), plus manual API checks via `.http` files in `tests/`.

## Coding Style & Naming Conventions
- Language: Go 1.24, format all code with `gofmt`/editor-on-save (`go fmt ./...` before committing).
- Package names: short, lowercase (`filesystem`, `logger`, `sse`); files use `snake_case.go`.
- Types and exported functions: `PascalCase`; locals and unexported symbols: `camelCase`.
- Keep handlers thin, push business logic into `internal/service` and persistence into `internal/repositories`.

## Testing Guidelines
- Add/maintain Go tests under the same package with `*_test.go` suffix (`go test ./pkg/...` for focused runs).
- Mirror new behavior with table-driven tests where practical.
- Prefer deterministic tests; avoid relying on real external tools or network when possible.
- Use `.http` files in `tests/` to document and validate API flows when endpoints change.

## Commit & Pull Request Guidelines
- Follow conventional commits: `feat(scope): summary`, `fix(scope): summary`, `chore(scope): summary`, `docs(scope): summary`, etc. Scopes like `sse`, `filesystem`, `repo`, `tests`.
- Use present-tense, concise messages; keep one logical change per commit.
- PRs should include: brief description, motivation/context, checklist of changes, how to test (commands and sample requests), and screenshots or log snippets when relevant.

## Security & Configuration Tips
- Never commit real credentials or private file paths; use `config.yaml.example` as the template and keep local `config.yaml` environment-specific.
- Treat `tools/` binaries as build/runtime dependencies; do not modify or repackage them without confirming licenses and platform impact.
- When introducing new external tools or ports, document them in `README.md` and `doc/` and ensure they are configurable via `config.yaml`.

## Agent-Specific Instructions
- Keep changes minimal and scoped; prefer extending existing modules under `internal/` or `pkg/` instead of creating new top-level directories.
- Do not introduce new Go dependencies unless necessary; if added, group them logically in `go.mod` and explain the rationale in the PR description.
- When modifying behavior, update or add tests in the nearest relevant package and, if API-visible, adjust `.http` examples in `tests/`.
