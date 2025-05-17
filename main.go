package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/kballard/go-shellquote"
)

type Config struct {
	OutputDir       string
	Filename        string
	DiffType        string
	Timestamp       bool
	CopyToClipboard bool
	Interactive     bool
	Intro           bool // To skip Harmonica intro
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cfg := Config{}

	flag.StringVar(&cfg.OutputDir, "o", "", "Output directory (default: ~/llm_context_diffs or ./sjiek_diffs)")
	flag.StringVar(&cfg.Filename, "n", "", "Filename (default: current_diff.txt)")
	flag.StringVar(&cfg.DiffType, "diff-type", "all", "Diff type: all, staged, unstaged")
	flag.BoolVar(&cfg.Timestamp, "t", false, "Add timestamp to filename")
	flag.BoolVar(&cfg.CopyToClipboard, "c", false, "Copy diff to clipboard")
	flag.BoolVar(&cfg.Interactive, "i", false, "Run in interactive mode using gum")
	flag.BoolVar(&cfg.Intro, "intro", false, "Show the startup intro animation") // Corrected flag
	flag.Parse()

	// --- Handle Intro ---
	introWasShown := false
	if cfg.Intro && commandExists("gum") {
		title := "🫧 sjiek 🫧"
		slogan := "🫦 chew on this 🫦"
		shellCommandForSpinner := fmt.Sprintf("sleep 0.7; echo \"%s\"", slogan)
		gumSpinnerCmd := exec.Command("gum", "spin", "--spinner", "pulse", "--title", title, "--show-output", "--", "sh", "-c", shellCommandForSpinner)
		gumSpinnerCmd.Stdin = os.Stdin
		gumSpinnerCmd.Stdout = os.Stdout
		gumSpinnerCmd.Stderr = os.Stderr
		if err := gumSpinnerCmd.Run(); err != nil {
			fmt.Printf("%s %s\n", title, slogan)
		}
		fmt.Println()
		introWasShown = true
	} else if cfg.Intro { // --intro but no gum
		fmt.Println("🫧 sjiek 🫧 - 🫦 chew on this 🫦")
		fmt.Println()
		// introWasShown = true; // Not strictly needed here if gumStyle isn't shown for non-interactive
	}

	applyDefaults(&cfg) // Apply defaults regardless of mode first
	if err := validateConfig(&cfg); err != nil {
		log.Fatalf("Initial configuration error: %v", err)
	}

	// --- Determine Mode and Execute ---
	if cfg.Interactive { // Explicit -i flag means interactive mode
		if !introWasShown && commandExists("gum") { // Show gumStyle welcome if intro didn't run
			gumStyle("🫦 chew on this 🫦", "--padding", "1")
		} else if !introWasShown { // No intro, no gum, but interactive mode requested
			fmt.Println("sjiek: interactive mode") // Simple text title
		}

		err := runInteractiveMode(&cfg) // This will prompt user to override defaults
		if err != nil {
			if strings.Contains(err.Error(), "user cancelled") || strings.Contains(err.Error(), "user cancelled or TUI failed") {
				log.Println("Operation cancelled by user.")
				os.Exit(0)
			}
			log.Fatalf("Interactive mode failed: %v", err)
		}
		// Defaults were already applied. runInteractiveMode modifies cfg based on user input.
		// Re-validate if necessary, though runInteractiveMode should ensure valid choices.
		if err := validateConfig(&cfg); err != nil { // Good to re-validate after user input
			log.Fatalf("Configuration error after interactive mode: %v", err)
		}
		// Now process with the interactively gathered config
		err = processRequest(cfg)
		if err != nil {
			log.Fatalf("Error processing request after interactive mode: %v", err)
		}
	} else {
		// Non-interactive mode (default or specific flags given, but not -i)
		// applyDefaults() has already set up cfg.
		// If specific flags were given (e.g., -n, -o, -c, -t, --diff-type), they would have overridden the defaults.
		// If no flags were given (and -i was not given), it runs with pure defaults.
		if flag.NFlag() == 0 && !cfg.Intro { // Truly no flags at all (and no intro shown)
			// This is the "sjiek" command run for pure speed with defaults.
			// Optionally print a very minimal message or nothing.
			// For speed, perhaps nothing is best.
			// Or a very quick confirmation of what it's about to do if defaults are not obvious.
			// Let's assume for now that processRequest will print success/failure.
		} else if flag.NFlag() > 0 && !cfg.Intro && !cfg.Interactive {
			// Flags were given, but not -i and not --intro.
			// This is also non-interactive, using the specified flags.
			// No special message needed here either.
		}

		err := processRequest(cfg)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
	}
}

func applyDefaults(cfg *Config) {
	if cfg.OutputDir == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			cfg.OutputDir = filepath.Join(homeDir, "llm_context_diffs")
		} else {
			log.Printf("Warning: could not get user home directory: %v. Defaulting OutputDir to ./sjiek_diffs\n", err)
			cfg.OutputDir = "./sjiek_diffs"
		}
	}
	if cfg.Filename == "" {
		cfg.Filename = "current_diff.txt"
	}
	if cfg.DiffType == "" {
		cfg.DiffType = "all"
	}
}

func validateConfig(cfg *Config) error {
	validDiffTypes := map[string]bool{"all": true, "staged": true, "unstaged": true}
	if !validDiffTypes[cfg.DiffType] {
		return fmt.Errorf("invalid diff-type: '%s'. Must be one of 'all', 'staged', 'unstaged'", cfg.DiffType)
	}
	if cfg.Filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	return nil
}

func commandExists(cmdName string) bool {
	_, err := exec.LookPath(cmdName)
	return err == nil
}

// gumStyle is used for the welcome message in interactive mode
func gumStyle(text string, args ...string) {
	if !commandExists("gum") {
		fmt.Println(text)
		return
	}
	cmdArgs := append([]string{"style"}, args...)
	cmdArgs = append(cmdArgs, text)
	cmd := exec.Command("gum", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot get user home directory: %w", err)
		}
		return filepath.Join(homeDir, path[2:]), nil
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("could not determine absolute path for '%s': %w", path, err)
	}
	return absPath, nil
}

func confirmGumCommand(prompt string) (bool, error) {
	if !commandExists("gum") {
		return false, fmt.Errorf("gum command not found. Please install gum")
	}
	cmd := exec.Command("gum", "confirm", prompt)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err == nil {
		return true, nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 1 {
			return false, nil
		}
		if exitErr.ExitCode() == 130 {
			return false, fmt.Errorf("user cancelled (SIGINT)")
		}
		return false, fmt.Errorf("gum confirm failed/cancelled (exit code %d)", exitErr.ExitCode())
	}
	return false, fmt.Errorf("gum confirm execution failed: %v", err)
}

func runGumCommand(args ...string) (string, error) {
	if !commandExists("gum") {
		return "", fmt.Errorf("gum command not found. Please install gum")
	}
	if !commandExists("sh") {
		return "", fmt.Errorf("sh (shell) command not found, required for gum workaround")
	}

	tmpFile, err := os.CreateTemp("", "sjiek-gum-selection-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpFileName := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFileName)
		return "", fmt.Errorf("failed to close temp file before writing: %w", err)
	}
	defer func() {
		if err := os.Remove(tmpFileName); err != nil {
			log.Printf("Warning: failed to remove temporary file %s: %v\n", tmpFileName, err)
		}
	}()

	quotedGumArgs := make([]string, len(args))
	for i, arg := range args {
		quotedGumArgs[i] = shellquote.Join(arg)
	}
	gumCmdPart := "gum " + strings.Join(quotedGumArgs, " ")
	shellCmdString := fmt.Sprintf("%s > %s", gumCmdPart, shellquote.Join(tmpFileName))

	cmd := exec.Command("sh", "-c", shellCmdString)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	runErr := cmd.Run()

	selectionBytes, readErr := os.ReadFile(tmpFileName)
	selectedItem := ""
	if readErr == nil {
		selectedItem = strings.TrimSpace(string(selectionBytes))
	}

	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 130 {
				return "", fmt.Errorf("user cancelled (SIGINT to shell)")
			}
			if selectedItem == "" && (exitErr.ExitCode() == 1 || exitErr.ExitCode() == 0) {
				return "", fmt.Errorf("user cancelled or TUI failed (shell exit %d, no selection)", exitErr.ExitCode())
			}
			return selectedItem, fmt.Errorf("shell command for gum failed (exit %d)", exitErr.ExitCode())
		}
		return selectedItem, fmt.Errorf("shell command execution for gum failed: %v", runErr)
	}

	if selectedItem == "" {
		return "", fmt.Errorf("user cancelled (no selection written by gum)")
	}
	return selectedItem, nil
}

func runInteractiveMode(cfg *Config) error {
	var err error
	diffTypeDisplay := map[string]string{
		"all":      "all (staged & unstaged)",
		"staged":   "staged (for commit)",
		"unstaged": "unstaged (working changes)",
	}
	currentSelectedDisplay := diffTypeDisplay[cfg.DiffType]
	if currentSelectedDisplay == "" {
		currentSelectedDisplay = diffTypeDisplay["all"]
	}

	selectedDiffDisplay, err := runGumCommand("choose", diffTypeDisplay["all"], diffTypeDisplay["staged"], diffTypeDisplay["unstaged"], "--header", "Select Diff Type:", "--selected", currentSelectedDisplay)
	if err != nil {
		return err
	}
	found := false
	for key, display := range diffTypeDisplay {
		if display == selectedDiffDisplay {
			cfg.DiffType = key
			found = true
			break
		}
	}
	if !found && selectedDiffDisplay != "" {
		return fmt.Errorf("unexpected selection for Diff Type: '%s'", selectedDiffDisplay)
	} else if !found && selectedDiffDisplay == "" {
		return fmt.Errorf("no selection made for Diff Type (cancelled)")
	}

	tempOutputDir, err := runGumCommand("input", "--value", cfg.OutputDir, "--placeholder", "e.g., ~/llm_context_diffs/", "--header", "Output Directory:")
	if err != nil {
		return err
	}
	if tempOutputDir != "" {
		cfg.OutputDir = tempOutputDir
	}

	tempFilename, err := runGumCommand("input", "--value", cfg.Filename, "--placeholder", "e.g., current_diff.txt", "--header", "Filename:")
	if err != nil {
		return err
	}
	if tempFilename != "" {
		cfg.Filename = tempFilename
	}

	cfg.Timestamp, err = confirmGumCommand("Add a timestamp to the filename?")
	if err != nil {
		return err
	}

	cfg.CopyToClipboard, err = confirmGumCommand("Copy diff to clipboard?")
	if err != nil {
		return err
	}
	return nil
}

func processRequest(cfg Config) error {
	// fmt.Printf("Processing with config: %+v\n", cfg)

	gitRepoCheckCmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	output, err := gitRepoCheckCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git repository status: %w. Is this a git repository?", err)
	}
	if strings.TrimSpace(string(output)) != "true" {
		currentDir, _ := os.Getwd()
		return fmt.Errorf("directory '%s' is not a git repository", currentDir)
	}

	var gitArgs []string
	diffDescription := ""
	switch cfg.DiffType {
	case "all":
		gitArgs = []string{"diff", "HEAD"}
		diffDescription = "all uncommitted changes"
	case "staged":
		gitArgs = []string{"diff", "--staged"}
		diffDescription = "staged changes"
	case "unstaged":
		gitArgs = []string{"diff"}
		diffDescription = "unstaged changes"
	default:
		return fmt.Errorf("internal error: invalid diff-type '%s'", cfg.DiffType)
	}

	diffCmd := exec.Command("git", gitArgs...)
	diffOutputBytes, err := diffCmd.CombinedOutput()
	diffContent := string(diffOutputBytes)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 && strings.Contains(diffContent, "diff --git") {
			} else if exitErr.ExitCode() != 0 {
				return fmt.Errorf("git diff command failed with exit code %d: %s", exitErr.ExitCode(), diffContent)
			}
		} else {
			return fmt.Errorf("git diff command execution failed: %w. Output: %s", err, diffContent)
		}
	}

	trimmedDiffContent := strings.TrimSpace(diffContent)
	if !strings.HasPrefix(trimmedDiffContent, "diff --git") {
		diffContent = ""
	}

	actualFilename := cfg.Filename
	if cfg.Timestamp {
		timestamp := time.Now().Format("20060102_150405")
		ext := filepath.Ext(cfg.Filename)
		name := strings.TrimSuffix(cfg.Filename, ext)
		if name == "" {
			name = "diff"
		}
		actualFilename = fmt.Sprintf("%s_%s%s", name, timestamp, ext)
	}
	if filepath.Ext(actualFilename) == "" {
		actualFilename += ".txt"
	}

	expandedOutputDir, err := expandPath(cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("could not expand output directory path '%s': %w", cfg.OutputDir, err)
	}
	if err := os.MkdirAll(expandedOutputDir, 0755); err != nil {
		return fmt.Errorf("could not create output directory '%s': %w", expandedOutputDir, err)
	}
	outputFilepath := filepath.Join(expandedOutputDir, actualFilename)

	if diffContent == "" {
		fmt.Printf("No changes for diff type '%s'. Empty file saved to: %s\n", cfg.DiffType, outputFilepath)
	}
	if err := os.WriteFile(outputFilepath, []byte(diffContent), 0644); err != nil {
		return fmt.Errorf("could not write diff to file '%s': %w", outputFilepath, err)
	}
	if diffContent != "" {
		fmt.Printf("💾 diff for '%s' saved successfully to: %s\n", diffDescription, outputFilepath)
	}

	if cfg.CopyToClipboard {
		if diffContent == "" {
			if err := clipboard.WriteAll(""); err != nil {
				fmt.Printf("Warning: Could not copy empty string to clipboard: %v\n", err)
			} else {
				fmt.Println("📋 empty string copied to clipboard!")
			}
		} else {
			if err := clipboard.WriteAll(diffContent); err != nil {
				fmt.Printf("Warning: Could not copy diff to clipboard: %v\n", err)
				fmt.Println("  Ensure a clipboard utility is installed (e.g., xclip/xsel on Linux, pbcopy on macOS, wl-clipboard for Wayland).")
			} else {
				fmt.Println("📋 diff copied to clipboard!")
			}
		}
	}
	return nil
}
