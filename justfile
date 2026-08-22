# Per-repo task manifest. Run `just` (or `just --list`) to see every verb.
#
# Recipes take trailing arguments directly: `just evalkit-export <dir>`, where
# the retired form was `ward exec evalkit-export -- <dir>`.
#
# One line of comment per recipe on purpose: just reads only the LAST comment
# line above a recipe, so a wrapped description silently truncates to its tail.
# That is agentic-os#1048's finding, kept here rather than rediscovered.
#
# Retiring `ward exec` per coilysiren/inbox#366. `.ward/ward.yaml` survives
# carrying catalog metadata only, because check_catalog_block pins that exact
# path upstream. Tracked at agentic-os#1081.

set positional-arguments

# Default target: list every available recipe.
default:
    @just --list --unsorted

# Build acompose and report isolated host convergence stages.
smoke *ARGS:
    @sh scripts/smoke.sh "$@"

# Run the smoke with captured acompose convergence transcripts.
smoke-verbose *ARGS:
    @sh scripts/smoke.sh --verbose "$@"

# Run Go and palette tests, then the pre-commit suite.
test *ARGS:
    @sh scripts/test.sh "$@"

# Compile every package.
build *ARGS:
    @go build ./... "$@"

# Run Go static analysis.
lint *ARGS:
    @go vet ./... "$@"

# Format Go source files.
fmt *ARGS:
    @gofmt -w . "$@"

# Install the agent-compose binary into GOBIN.
install *ARGS:
    @go install ./cmd/agent-compose "$@"

# Reconcile go.mod and go.sum with the source imports.
tidy *ARGS:
    @go mod tidy "$@"

# Run the complete pre-commit suite explicitly.
pre-commit *ARGS:
    @pre-commit run --all-files "$@"

# Re-extract the vendored Claude Code verb and theme-token lists from the installed binary.
harness-refresh *ARGS:
    @go test ./internal/nativeui -run TestVendoredHarnessData -update -v "$@"

# Generate canonical palette data and build the local explorer.
palette-build *ARGS:
    @sh scripts/palette-web.sh build "$@"

# Type-check, build, and verify the local palette explorer.
palette-test *ARGS:
    @sh scripts/palette-web.sh test "$@"

# Generate canonical palette data and serve the local explorer.
palette-serve *ARGS:
    @sh scripts/palette-web.sh dev "$@"

# Reconcile the local explorer package lock with package.json.
palette-tidy *ARGS:
    @npm --prefix web/personality-palette install --package-lock-only --ignore-scripts --no-audit --no-fund "$@"

# Print the capability by test-type matrix the current roster implies.
evalkit-matrix *ARGS:
    @sh scripts/eval-matrix.sh "$@"

# Compose one bundle per role and write its delivery as the eval system prompt.
evalkit-prompts *ARGS:
    @sh scripts/eval-prompts.sh "$@"

# Send one live request through Agent Proxy to prove the eval transport works.
evalkit-smoke *ARGS:
    @sh scripts/eval-smoke.sh "$@"

# Lint, format-check, type-check, and test the evalkit Python package.
evalkit-check *ARGS:
    @sh scripts/evalkit-check.sh "$@"

# Reconcile the evalkit virtualenv with pyproject.toml.
evalkit-sync *ARGS:
    @uv sync "$@"

# Run the board through Inspect against Agent Proxy, unscored, five epochs.
evalkit-run *ARGS:
    @sh scripts/eval-run.sh "$@"

# Open the Inspect log viewer on the most recent board run.
evalkit-view *ARGS:
    @uv run inspect view --log-dir .evalkit/logs "$@"

# Project a committed run into a display payload. One way, and refuses anything unsafe for a public surface.
evalkit-export *ARGS:
    @uv run aos-eval export "$@"

# Read an Inspect eval log and build the dataset the annotator grades.
evalkit-filter *ARGS:
    @uv run python -m evalkit.filter "$@"

# Cluster annotation critiques into a ranked failure taxonomy.
evalkit-taxonomy *ARGS:
    @uv run aos-eval taxonomy "$@"

# Annotate the eval dataset by hand, one keystroke per challenge.
evalkit-annotate *ARGS:
    @sh scripts/eval-annotate.sh "$@"

# Cross-compile version-stamped release binaries into dist/.
release-build *ARGS:
    @sh scripts/release-build.sh "$@"

# Render the brew formula and scoop manifest from dist/ binaries.
release-package *ARGS:
    @sh scripts/render-packaging.sh "$@"

# Render v2 fixture package metadata and inspect its public contract.
release-package-test *ARGS:
    @sh scripts/render-packaging-test.sh "$@"

# Decide whether one validated revision should publish a product release.
release-impact *ARGS:
    @sh scripts/release-impact.sh "$@"

# Exercise documentation, result, product, hold, and manual release fixtures.
release-impact-test *ARGS:
    @sh scripts/release-impact-test.sh "$@"
