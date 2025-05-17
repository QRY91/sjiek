# sjiek_demo.tape
# Demo of sjiek interactive usage

# Output GIF file
Output sjiek_demo.gif

# Terminal Setup (adjust as needed)
Set Theme "Dracula" # Or "MaterialDarker", "Dracula", "Nord", etc.
Set FontFamily "JetBrains Mono" # Or your preferred mono font
Set FontSize 18
Set Width 1200
Set Height 750
Set TypingSpeed 100ms # Speed of simulated typing
Set LoopOffset "20%"  # Where the GIF loops from (e.g., 20% from the end)

# --- Setup a temporary git repository for the demo ---
Hide # Hide these setup commands from the GIF
Type "mkdir -p /tmp/sjiek_vhs_demo && cd /tmp/sjiek_vhs_demo" Enter Sleep 500ms
Type "rm -rf ./* .git" Enter Sleep 500ms # Clean up if run before
Type "git init -b main" Enter Sleep 500ms
# Set dummy git user for the demo repo to avoid global config prompts if not set
Type "git config user.name 'sjiek demo'" Enter Sleep 500ms
Type "git config user.email 'demo@example.com'" Enter Sleep 500ms
# Create some files and changes
Type "echo 'Hello World from sjiek!' > file1.txt" Enter Sleep 500ms
Type "git add file1.txt && git commit -m 'Initial commit'" Enter Sleep 1s
Type "echo 'An unstaged change to file1.' >> file1.txt" Enter Sleep 500ms
Type "echo 'A new staged file.' > staged_file.txt" Enter Sleep 500ms
Type "git add staged_file.txt" Enter Sleep 500ms
Type "clear" Enter Sleep 500ms # Clear setup clutter
Show # Start showing commands for the GIF

# --- Demonstrate sjiek in interactive mode ---
# IMPORTANT: Ensure 'sjiek' (your compiled binary) is in the PATH
# or provide the absolute path to it.
# For this demo, let's assume it's in PATH.
# If not, you might need to do: Type "/path/to/your/sjiek" Enter

# Start sjiek (interactive mode)
Type "sjiek -i" Sleep 500ms Enter
Sleep 1s # Wait for the welcome message and first prompt

# 1. Diff Type: (select "unstaged" as before)
Down@200ms Sleep 500ms 
Down@200ms Sleep 500ms 
Enter@200ms Sleep 1s    

# 2. Output Directory: Type "./output_diffs" to save locally within the demo repo
#    Clear default (assuming it's long enough, e.g., 30 chars for /home/user/llm_context_diffs)
# Ctrl+A Sleep 100ms Backspace Sleep 100ms # Select all and delete is often better than many backspaces
# Or many Backspaces:
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
Backspace@50ms Sleep 5ms # Repeat ~30 times
# ... (enough backspaces) ...
Sleep 200ms
Type "./output_diffs" Sleep 500ms # Type the relative path
Enter@200ms Sleep 1s

# 3. Filename: Type "demo_context.txt" (clear default as before)
#    (Assuming default "current_diff.txt" - 16 chars)
# Ctrl+A Sleep 100ms Backspace Sleep 100ms
# Or many Backspaces:
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
Backspace@50ms Sleep 5ms # Repeat 16 times
# ... (enough backspaces) ...
Sleep 200ms
Type "demo_context.txt" Sleep 500ms
Enter@200ms Sleep 1s

# 4. Timestamp: Confirm "Yes"
Type "y" Sleep 500ms
Enter@200ms Sleep 1s

# 5. Copy to Clipboard: Confirm "No"
Right@200ms Sleep 500ms 
Enter@200ms Sleep 1s    

Sleep 2s # Wait for sjiek to finish

# Now, the file should be in ./output_diffs/demo_context_YYYYMMDD_HHMMSS.txt
# relative to /tmp/sjiek_vhs_demo/
Type "ls -l ./output_diffs/" Sleep 500ms Enter # List the output directory
Sleep 1.5s
Type "cat ./output_diffs/demo_context_*.txt" Sleep 500ms Enter # Cat the file using wildcard for timestamp
Sleep 4s