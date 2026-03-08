# now-cli

## Project

- Go CLI tool, binary name is `now`
- Repo: `opennow-labs/now-cli` (project name ≠ binary name)

## Build & Release

- Build: `make build` (includes Swift helper on macOS)
- Release: `goreleaser release --clean`, then `gh release create` to upload artifacts
- `GITHUB_TOKEN` used by goreleaser has limited permissions; use `gh` CLI for creating GitHub releases
- **Naming convention**: release archives use the binary name `now`, not the repo name `now-cli`
  - Correct: `now_darwin_amd64.tar.gz`, `now_linux_arm64.tar.gz`, `now_windows_amd64.zip`
  - Wrong: `now-cli_darwin_amd64.tar.gz`
  - This is controlled by `project_name: now` in `.goreleaser.yml`
