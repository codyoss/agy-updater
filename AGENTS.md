# Instructions for AI Coding Agents

To keep this project maintainable and well-documented, all AI agents modifying this codebase must adhere to the following rules:

1. **Documentation Freshness**:
   - Whenever any code files are added, modified, renamed, or deleted, the corresponding documentation in [docs/architecture.md](file:///home/codyoss/code/agy-updater/docs/architecture.md) MUST be updated immediately in the same turn.
   - Maintain the detailed, per-file granularity format. Do not let documentation become stale.

2. **Standard Library Constraints**:
   - Keep the project minimal and free of external dependencies.
   - Do not add any new dependencies to `go.mod` unless explicitly instructed by the user. Rely entirely on the standard library.

3. **Atomic Operations**:
   - Always perform directory moves and setup staging folders atomically using staging directories (e.g. `.new`) and `os.Rename`.
   - Never write directly to target `/opt` locations without verifying tarball contents first.
