# helix-copilot

`helix-copilot` provides GitHub Copilot support for a patched Helix editor build with native ghost text / inline completion.

This repository is maintained under the `naipi11` GitHub account.

## Upstream Helix

This project is based on the upstream Helix editor:

- Upstream repository: <https://github.com/helix-editor/helix>
- Helix website: <https://helix-editor.com/>
- Helix license and runtime files remain from the upstream project where applicable.

The `helix/` directory contains the patched Helix source used to build `hx` with native inline-completion rendering and acceptance behavior.

## What is included

- `hx`: patched Helix editor binary with native Copilot ghost text support.
- `helix-copilot`: Go CLI and LSP bridge.
- `helix-copilot lsp`: starts a proxy between Helix and `@github/copilot-language-server`.
- `helix-copilot login`: runs the GitHub Copilot device login flow.
- `helix-copilot configure-helix`: safely merges Copilot language-server settings into Helix `languages.toml`.
- `helix-copilot model <name>`: stores the selected Copilot model.

## Requirements

- GitHub account with GitHub Copilot access.
- Node.js and npm available in `PATH`.
- Go 1.24+ if building `helix-copilot` from source.
- Rust toolchain if building patched `hx` from source.

The LSP bridge runs the official Copilot language server through:

```bash
npx --yes @github/copilot-language-server --stdio
```

## Install from release

After a release is published, download the archive for your platform from:

```text
https://github.com/naipi11/helix_copilot/releases
```

Each release archive is intended to contain:

```text
helix-copilot        # or helix-copilot.exe on Windows
hx                  # or hx.exe on Windows
runtime/            # Helix runtime files
```

### Linux / macOS install script

```bash
curl -fsSL https://raw.githubusercontent.com/naipi11/helix_copilot/main/scripts/install.sh | bash
```

Optional environment variables:

```bash
VERSION=v0.1.0 BIN_DIR="$HOME/.local/bin" bash scripts/install.sh
```

### Windows PowerShell install script

```powershell
iwr https://raw.githubusercontent.com/naipi11/helix_copilot/main/scripts/install.ps1 -UseBasicParsing | iex
```

Optional parameters when running locally:

```powershell
./scripts/install.ps1 -Version v0.1.0 -BinDir "$HOME\bin"
```

## Build from source

### Build the Go CLI

```bash
go build -o helix-copilot ./cmd/helix-copilot
```

Install it into your Go bin directory:

```bash
go install ./cmd/helix-copilot
```

### Build patched Helix

```bash
cd helix
cargo build --release --locked
```

The patched editor binary will be at:

```text
helix/target/release/hx
```

On Windows, the binary is `hx.exe`.

Make sure both `hx` and `helix-copilot` are in your `PATH`.

## GitHub Copilot login

Run:

```bash
helix-copilot login
```

The command starts the Copilot language server, requests a device login code, and prompts you to complete authorization in the browser.

## Configure Helix

Run:

```bash
helix-copilot configure-helix
```

By default this updates:

```text
~/.config/helix/languages.toml
```

To test the merge output first:

```bash
helix-copilot configure-helix --output ./languages.test.toml
```

The command merges instead of blindly overwriting:

- Adds or updates `[language-server.copilot]`.
- Adds `copilot` to existing `language-servers` arrays without duplicating it.
- Preserves other language settings such as `auto-format`, `indent`, debugger templates, and grammar entries.
- Adds a Python `pylsp + copilot` setup and disables noisy style-only diagnostics from `pylsp`.

## Usage

Open a supported source file with the patched `hx` binary. In insert mode, Copilot suggestions appear as ghost text when available.

Key behavior:

- `Tab`: accept visible ghost text; falls back to normal smart tab when no ghost text exists.
- `Esc`: reject ghost text and return to normal mode.
- `:model <name>`: in the patched Helix build, calls `helix-copilot model <name>` to store the selected model. Restart the language server or editor after changing models.

You can also set the model outside Helix:

```bash
helix-copilot model gpt-5.4-mini
```

## Package manager templates

This repository includes starter packaging files:

- `.goreleaser.yaml`
- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `packaging/scoop/helix-copilot.json`
- `packaging/homebrew/helix-copilot.rb`

The Scoop and Homebrew files are templates. Replace `TODO` checksums after publishing real release assets.

## Development checks

Run the Go tests for project-owned packages:

```bash
go test ./cmd/helix-copilot ./internal/config ./internal/helixconfig ./internal/login ./internal/lsp
```

Run the patched Helix check:

```bash
cd helix
cargo check -p helix-term
```

Avoid `go test ./...` at the repository root because it descends into vendored/upstream Helix tree-sitter grammar bindings that are not part of the Go CLI module.

## License

This repository contains original `helix-copilot` code plus patched Helix source. See the relevant source files and upstream Helix licensing for details.
