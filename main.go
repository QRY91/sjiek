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
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Keep for now, helps pinpoint Fatalf origins
	cfg := Config{}

	flag.StringVar(&cfg.OutputDir, "o", "", "Output directory (default: ~/llm_context_diffs or ./sjiek_diffs)")
	flag.StringVar(&cfg.Filename, "n", "", "Filename (default: current_diff.txt)")
	flag.StringVar(&cfg.DiffType, "diff-type", "all", "Diff type: all, staged, unstaged")
	flag.BoolVar(&cfg.Timestamp, "t", false, "Add timestamp to filename")
	flag.BoolVar(&cfg.CopyToClipboard, "c", false, "Copy diff to clipboard")
	flag.BoolVar(&cfg.Interactive, "i", false, "Run in interactive mode using gum")
	flag.Parse()

	applyDefaults(&cfg)
	if err := validateConfig(&cfg); err != nil {
		log.Fatalf("Initial configuration error: %v", err)
	}

	runAsInteractive := cfg.Interactive
	if flag.NFlag() == 0 {
		fmt.Println("No flags provided, running in interactive mode.")
		runAsInteractive = true
	}

	if runAsInteractive {
		fmt.Println("Sjiek Interactive Mode")
		gumStyle("sjiek: context for AI to chew on", "--border", "normal", "--margin", "1", "--padding", "1", "--border-foreground", "212")
		err := runInteractiveMode(&cfg)
		if err != nil {
			// "user cancelled" is a normal exit path for interactive mode
			if strings.Contains(err.Error(), "user cancelled") || strings.Contains(err.Error(), "user cancelled or TUI failed") {
				log.Println("Operation cancelled by user.")
				os.Exit(0)
			}
			log.Fatalf("Interactive mode failed: %v", err)
		}
		applyDefaults(&cfg) // Re-apply if interactive mode cleared some fields by returning empty
		if err := validateConfig(&cfg); err != nil {
			log.Fatalf("Configuration error after interactive mode: %v", err)
		}
		// fmt.Println("Interactive configuration complete.") // Can be a bit noisy
	}

	err := processRequest(cfg)
	if err != nil {
		log.Fatalf("Error: %v", err)
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

func gumStyle(text string, args ...string) {
	if !commandExists("gum") {
		fmt.Println(text) // Fallback if gum not found
		return
	}
	cmdArgs := append([]string{"style"}, args...)
	cmdArgs = append(cmdArgs, text)
	cmd := exec.Command("gum", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run() // Errors in gum style are not critical for sjiek's core functionality
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
		// If Abs fails, it might be because the path is already effectively absolute
		// or due to other issues. Returning the original path might be problematic.
		// Let's try to be a bit more robust or at least indicate the failure.
		return "", fmt.Errorf("could not determine absolute path for '%s': %w", path, err)
	}
	return absPath, nil
}

func confirmGumCommand(prompt string) (bool, error) {
	if !commandExists("gum") {
		return false, fmt.Errorf("gum command not found. Please install gum")
	}
	// log.Printf("Attempting to run gum confirm with prompt: '%s'\n", prompt)
	cmd := exec.Command("gum", "confirm", prompt)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout // TUI draws here
	cmd.Stderr = os.Stderr // TUI errors here

	err := cmd.Run()
	if err == nil {
		// log.Println("gum confirm returned true (exit 0)")
		return true, nil
	}

	// log.Printf("gum confirm error: %v\n", err)
	if exitErr, ok := err.(*exec.ExitError); ok {
		// log.Printf("gum confirm ExitError. ExitCode: %d.\n", exitErr.ExitCode())
		if exitErr.ExitCode() == 1 { // No
			// log.Println("gum confirm returned false (exit 1)")
			return false, nil
		}
		if exitErr.ExitCode() == 130 { // SIGINT (Ctrl+C)
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
	// log.Printf("Attempting to run gum (via sh) with args: %v\n", args)

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
		// log.Printf("Removing temporary file: %s\n", tmpFileName)
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
	// log.Printf("Executing shell command: sh -c \"%s\"\n", shellCmdString)

	cmd := exec.Command("sh", "-c", shellCmdString)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	runErr := cmd.Run()

	selectionBytes, readErr := os.ReadFile(tmpFileName)
	selectedItem := ""
	if readErr == nil {
		selectedItem = strings.TrimSpace(string(selectionBytes))
	} else {
		// log.Printf("Warning: failed to read temp file %s: %v.", tmpFileName, readErr)
		// This is often okay if runErr is also set (e.g., user cancelled gum)
	}
	// log.Printf("gum command selectedItem from temp file: '%s'\n", selectedItem)

	if runErr != nil {
		// log.Printf("Shell command for gum returned error: %v\n", runErr)
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			// log.Printf("Shell command ExitError. ExitCode: %d.\n", exitErr.ExitCode())
			if exitErr.ExitCode() == 130 {
				return "", fmt.Errorf("user cancelled (SIGINT to shell)")
			}
			// If selectedItem is empty, it's a strong indicator of cancellation or TUI not even starting
			if selectedItem == "" && (exitErr.ExitCode() == 1 || exitErr.ExitCode() == 0) {
				return "", fmt.Errorf("user cancelled or TUI failed (shell exit %d, no selection)", exitErr.ExitCode())
			}
			return selectedItem, fmt.Errorf("shell command for gum failed (exit %d)", exitErr.ExitCode())
		}
		return selectedItem, fmt.Errorf("shell command execution for gum failed: %v", runErr)
	}

	if selectedItem == "" {
		// log.Println("Shell command successful, but no selection made. Assuming cancellation.")
		return "", fmt.Errorf("user cancelled (no selection written by gum)")
	}
	return selectedItem, nil
}

func runInteractiveMode(cfg *Config) error {
	var err error
	// log.Println("Entering runInteractiveMode")

	diffTypeDisplay := map[string]string{
		"all":      "all (staged & unstaged)",
		"staged":   "staged (for commit)",
		"unstaged": "unstaged (working changes)",
	}
	currentSelectedDisplay := diffTypeDisplay[cfg.DiffType]
	if currentSelectedDisplay == "" { // Should be caught by applyDefaults
		currentSelectedDisplay = diffTypeDisplay["all"]
	}

	// log.Println("Prompting for Diff Type...")
	selectedDiffDisplay, err := runGumCommand("choose", diffTypeDisplay["all"], diffTypeDisplay["staged"], diffTypeDisplay["unstaged"], "--header", "Select Diff Type:", "--selected", currentSelectedDisplay)
	if err != nil {
		// log.Printf("Error from gum choose (Diff Type): %v\n", err)
		return err
	}
	// log.Printf("Selected Diff Type display: '%s'\n", selectedDiffDisplay)
	found := false
	for key, display := range diffTypeDisplay {
		if display == selectedDiffDisplay {
			cfg.DiffType = key
			found = true
			break
		}
	}
	if !found && selectedDiffDisplay != "" {
		// This case means gum returned something not in our map.
		// It's an unexpected state, possibly an error or a new gum option.
		return fmt.Errorf("unexpected selection for Diff Type: '%s'", selectedDiffDisplay)
	} else if !found && selectedDiffDisplay == "" {
		// This should ideally be caught by runGumCommand's error handling if selection is empty
		return fmt.Errorf("no selection made for Diff Type (cancelled)")
	}

	// log.Println("Prompting for Output Directory...")
	tempOutputDir, err := runGumCommand("input", "--value", cfg.OutputDir, "--placeholder", "e.g., ~/llm_context_diffs/", "--header", "Output Directory:")
	if err != nil {
		// log.Printf("Error from gum input (Output Directory): %v\n", err)
		return err
	}
	if tempOutputDir != "" { // User provided input
		cfg.OutputDir = tempOutputDir
	} // If empty, user pressed Enter, keep existing cfg.OutputDir (which is the default)
	// log.Printf("Selected Output Directory: '%s'\n", cfg.OutputDir)

	// log.Println("Prompting for Filename...")
	tempFilename, err := runGumCommand("input", "--value", cfg.Filename, "--placeholder", "e.g., current_diff.txt", "--header", "Filename:")
	if err != nil {
		// log.Printf("Error from gum input (Filename): %v\n", err)
		return err
	}
	if tempFilename != "" { // User provided input
		cfg.Filename = tempFilename
	} // If empty, user pressed Enter, keep existing cfg.Filename
	// log.Printf("Selected Filename: '%s'\n", cfg.Filename)

	// log.Println("Prompting for Timestamp...")
	cfg.Timestamp, err = confirmGumCommand("Add a timestamp to the filename?")
	if err != nil {
		// log.Printf("Error from gum confirm (Timestamp): %v\n", err)
		return err
	}
	// log.Printf("Timestamp selected: %t\n", cfg.Timestamp)

	// log.Println("Prompting for Copy to Clipboard...")
	cfg.CopyToClipboard, err = confirmGumCommand("Copy diff to clipboard?")
	if err != nil {
		// log.Printf("Error from gum confirm (Copy to Clipboard): %v\n", err)
		return err
	}
	// log.Printf("Copy to Clipboard selected: %t\n", cfg.CopyToClipboard)
	// log.Println("Exiting runInteractiveMode successfully")
	return nil
}

func processRequest(cfg Config) error {
	// fmt.Printf("Processing with config: %+v\n", cfg) // Keep for now, or make verbose

	gitRepoCheckCmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	output, err := gitRepoCheckCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git repository status: %w. Is this a git repository?", err)
	}
	if strings.TrimSpace(string(output)) != "true" {
		currentDir, _ := os.Getwd()
		return fmt.Errorf("directory '%s' is not a git repository", currentDir)
	}
	// log.Println("Currently in a git repository.")

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

	// log.Printf("Generating diff for %s using: git %s\n", diffDescription, strings.Join(gitArgs, " "))
	diffCmd := exec.Command("git", gitArgs...)
	diffOutputBytes, err := diffCmd.CombinedOutput()
	diffContent := string(diffOutputBytes)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Normal for git diff to exit 1 if there are changes.
			// We only consider it an error if it's not 1 OR if it's 1 but there's no diff output.
			if exitErr.ExitCode() == 1 && strings.Contains(diffContent, "diff --git") {
				// log.Println("Differences found (git diff exit code 1).")
			} else if exitErr.ExitCode() != 0 { // Any other non-zero exit code is an error
				return fmt.Errorf("git diff command failed with exit code %d: %s", exitErr.ExitCode(), diffContent)
			}
		} else { // Not an ExitError (e.g., command not found)
			return fmt.Errorf("git diff command execution failed: %w. Output: %s", err, diffContent)
		}
	}

	trimmedDiffContent := strings.TrimSpace(diffContent)
	if !strings.HasPrefix(trimmedDiffContent, "diff --git") {
		// log.Println("No actual diff content found.")
		diffContent = "" // Standardize to empty string for no diff
	}

	actualFilename := cfg.Filename
	if cfg.Timestamp {
		timestamp := time.Now().Format("20060102_150405")
		ext := filepath.Ext(cfg.Filename)
		name := strings.TrimSuffix(cfg.Filename, ext)
		if name == "" {
			name = "diff" // Default base name if original was just ".txt" or empty
		}
		actualFilename = fmt.Sprintf("%s_%s%s", name, timestamp, ext)
	}
	if filepath.Ext(actualFilename) == "" { // Ensure .txt if no extension provided
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
		// log.Println("No diff content to save. Creating empty file.")
		fmt.Printf("No changes for diff type '%s'. Empty file saved to: %s\n", cfg.DiffType, outputFilepath)
	}
	if err := os.WriteFile(outputFilepath, []byte(diffContent), 0644); err != nil {
		return fmt.Errorf("could not write diff to file '%s': %w", outputFilepath, err)
	}
	if diffContent != "" { // Only print success if there was actual content
		fmt.Printf("Diff for '%s' saved successfully to: %s\n", diffDescription, outputFilepath)
	}

	if cfg.CopyToClipboard {
		if diffContent == "" {
			// log.Println("No diff content to copy (diff was empty). Attempting to copy empty string.")
			if err := clipboard.WriteAll(""); err != nil {
				fmt.Printf("Warning: Could not copy empty string to clipboard: %v\n", err)
			} else {
				fmt.Println("Empty string copied to clipboard.")
			}
		} else {
			if err := clipboard.WriteAll(diffContent); err != nil {
				fmt.Printf("Warning: Could not copy diff to clipboard: %v\n", err)
				fmt.Println("  Ensure a clipboard utility is installed (e.g., xclip/xsel on Linux, pbcopy on macOS, wl-clipboard for Wayland).")
			} else {
				fmt.Println("Diff copied to clipboard.")
			}
		}
	}
	return nil
}
