# Sjiek 🪄

> _"Sjiek: Context for AI to chew on"_

Sjiek (Dutch/Flemish for "gum," and a homonym for "chic/fancy") is a command-line tool designed to help you quickly generate `git diff` outputs of your current code changes. It saves these diffs to a file and can copy them to your clipboard, making it easy to provide context to Large Language Models (LLMs) or share your work-in-progress.

It features an interactive mode powered by [Charm Gum](https://github.com/charmbracelet/gum) for a user-friendly experience, but can also be controlled directly via command-line flags for scripting and automation.

## Features

*   Generate git diffs for:
    *   All uncommitted changes (staged and unstaged).
    *   Only staged changes (what's ready for commit).
    *   Only unstaged changes (current work not yet staged).
*   Save diff output to a configurable directory and filename.
*   Optionally add a timestamp to filenames for easy versioning of diffs.
*   Optionally copy the diff content directly to the system clipboard.
*   User-friendly interactive mode using `gum` for selecting options on-the-fly.
*   Direct command-line flag operation for quick, non-interactive use.
*   Built as a single, portable binary with Go.

## Requirements

*   **Git:** Sjiek operates on Git repositories and uses `git` commands.
*   **Gum (for interactive mode):** [Charm Gum](https://github.com/charmbracelet/gum) must be installed and accessible in your system's PATH for the interactive (`-i` or default no-flag) mode to function with its Terminal User Interface (TUI).
*   **Shell (`sh`):** Required by the interactive mode's `gum` integration for `gum choose` and `gum input`. This is standard on Linux and macOS.
*   **(Optional) Clipboard utility:** For the clipboard copy feature (`-c`) to work effectively, your operating system needs a clipboard utility that `github.com/atotto/clipboard` (the Go library used) can interface with.
    *   **Linux/X11:** `xclip` or `xsel` (e.g., `sudo apt install xclip`)
    *   **Linux/Wayland:** `wl-clipboard` (e.g., `sudo apt install wl-clipboard`)
    *   **macOS:** `pbcopy`/`pbpaste` (comes pre-installed)
    *   **Windows:** Should work out-of-the-box.

## Installation

### Using `go install` (Recommended)

Once the repository is public on GitHub (e.g., `github.com/yourusername/sjiek`):
```bash
go install github.com/yourusername/sjiek@latest
```
This command will download the source code, compile it, and place the `sjiek` binary in your `$GOPATH/bin` directory (or `$HOME/go/bin` if GOPATH is not set). Ensure this directory is included in your system's PATH environment variable.

### Building from Source

1.  **Clone the repository:**
    (Replace `yourusername` with your actual GitHub username)
    ```bash
    git clone https://github.com/yourusername/sjiek.git
    cd sjiek
    ```
2.  **Build the binary:**
    ```bash
    go build -o sjiek .
    ```
    This creates an executable file named `sjiek` in the current directory.
3.  **Move the binary to a directory in your PATH:**
    For example, to `~/.local/bin` (a common place for user-installed binaries on Linux):
    ```bash
    mkdir -p ~/.local/bin
    mv sjiek ~/.local/bin/
    ```
    *(Ensure `~/.local/bin` is in your PATH. You might need to add `export PATH="$HOME/.local/bin:$PATH"` to your shell's configuration file, like `~/.bashrc` or `~/.zshrc`, and then restart your shell or run `source ~/.bashrc`.)*

## Usage

Sjiek can be run interactively or directly via command-line flags. If no flags are provided, it defaults to interactive mode (requires `gum`).

```
sjiek [flags]
```

### Flags

*   `-o <directory>`: Specifies the output directory for the diff file.
    *   Default: `~/llm_context_diffs` (or `./sjiek_diffs` if the home directory is not accessible).
*   `-n <filename>`: Sets the filename for the diff file.
    *   Default: `current_diff.txt`.
*   `--diff-type <type>`: Determines the type of git diff to generate.
    *   Options:
        *   `all` (default): Shows all uncommitted changes (both staged and unstaged). Equivalent to `git diff HEAD`.
        *   `staged`: Shows only staged changes (what will be included in the next commit). Equivalent to `git diff --staged`.
        *   `unstaged`: Shows only unstaged changes (changes in the working directory not yet staged). Equivalent to `git diff`.
*   `-t`: Appends a timestamp (e.g., `_YYYYMMDD_HHMMSS`) to the filename.
*   `-c`: Copies the generated diff content to the system clipboard.
*   `-i`: Explicitly forces Sjiek to run in interactive mode. This is useful if you've set other flags but still want to be prompted for any remaining options.
*   `--help`: Displays the help message with all available flags.

### Interactive Mode

Running `sjiek` without any operational flags (or by using the `-i` flag) will launch the interactive mode:

```bash
sjiek
# or
sjiek -i
```

In this mode, you will be guided by `gum` prompts to:
1.  Select the desired diff type.
2.  Enter the output directory (or press Enter to use the default).
3.  Enter the filename (or press Enter for the default).
4.  Confirm whether to add a timestamp to the filename.
5.  Confirm whether to copy the diff to the clipboard.

### Examples

1.  **Run interactively for guided setup:**
    ```bash
    sjiek
    ```

2.  **Generate a diff of all changes, name it `latest_updates.diff`, and copy to clipboard:**
    ```bash
    sjiek -n latest_updates.diff -c
    ```

3.  **Get only staged changes, save to a timestamped file in a custom directory `~/project_diffs/`:**
    ```bash
    sjiek --diff-type staged -o "~/project_diffs/" -t
    ```

4.  **Get unstaged changes with a specific filename, without copying or timestamping:**
    ```bash
    sjiek --diff-type unstaged -n feature_x_WIP.patch
    ```

## Development

This project is built with Go.

*   Clone the repository.
*   Ensure Go (version 1.20 or newer is recommended) is installed.
*   Fetch dependencies: `go mod tidy`
*   Build: `go build -o sjiek .`
*   Run tests (if any are added in the future): `go test ./...`

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.