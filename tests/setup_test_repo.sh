#!/bin/bash

TEST_REPO_DIR_NAME="sjiek_test_repo"
# Get the directory where this script is located, then go one level up for the project root.
PROJECT_ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)
TEST_REPO_PATH="$PROJECT_ROOT_DIR/$TEST_REPO_DIR_NAME"

# Default output directories sjiek might create
SJIEK_DEFAULT_OUTPUT_DIR_HOME="$HOME/llm_context_diffs"
SJIEK_FALLBACK_OUTPUT_DIR_RELATIVE="$PROJECT_ROOT_DIR/sjiek_diffs" # Relative to where sjiek might be run if home fails

echo "--- sjiek test environment setup ---"

# Clean up previous test repo if it exists
if [ -d "$TEST_REPO_PATH" ]; then
    echo "Removing previous test repository: $TEST_REPO_PATH"
    rm -rf "$TEST_REPO_PATH"
fi

# Clean up potential default output directories (use with caution if you use these names)
# For testing, this ensures a clean slate.
if [ -d "$SJIEK_DEFAULT_OUTPUT_DIR_HOME" ]; then
    echo "Removing previous default output directory from home: $SJIEK_DEFAULT_OUTPUT_DIR_HOME"
    rm -rf "$SJIEK_DEFAULT_OUTPUT_DIR_HOME"
fi
if [ -d "$SJIEK_FALLBACK_OUTPUT_DIR_RELATIVE" ]; then
    echo "Removing previous fallback output directory: $SJIEK_FALLBACK_OUTPUT_DIR_RELATIVE"
    rm -rf "$SJIEK_FALLBACK_OUTPUT_DIR_RELATIVE"
fi


echo "Creating new test repository at: $TEST_REPO_PATH"
mkdir -p "$TEST_REPO_PATH"
cd "$TEST_REPO_PATH" || { echo "Failed to cd into $TEST_REPO_PATH"; exit 1; }

# Initialize git repo
git init -b main
git config user.email "test@example.com" # Set local config
git config user.name "Test User"
echo "Git repository initialized."

# Create and commit initial file
echo "Initial content for file1.txt" > file1.txt
git add file1.txt
git commit -m "Initial commit with file1.txt"
echo "Committed file1.txt"

echo "Second line for file1.txt" >> file1.txt
git commit -am "Second commit for file1.txt (modified file1.txt)"
echo "Committed modification to file1.txt"

# Create some changes for diffing
echo ">>> Making changes for testing diffs..."

# 1. Unstaged change to an existing file
echo "This is an unstaged modification in file1.txt" >> file1.txt
echo "file1.txt: unstaged modification added."

# 2. New file, staged
echo "Content for staged_file.txt" > staged_file.txt
git add staged_file.txt
echo "staged_file.txt: created and staged."

# 3. New file, unstaged
echo "Content for unstaged_file.txt" > unstaged_file.txt
echo "unstaged_file.txt: created, unstaged."

# 4. File to be modified and staged
echo "Original content for modified_staged.txt" > modified_staged.txt
git add modified_staged.txt
git commit -m "Add modified_staged.txt"
echo "This is a staged modification in modified_staged.txt" >> modified_staged.txt
git add modified_staged.txt
echo "modified_staged.txt: created, committed, then modified and staged."

echo ""
echo "--- Test Repository Status ---"
git status
echo "------------------------------"

echo ""
echo "Test repository setup complete in '$TEST_REPO_PATH'."
echo "You can now 'cd $TEST_REPO_PATH' and run your 'sjiek' Go binary."
echo "Example: If 'sjiek' binary is in '$PROJECT_ROOT_DIR', run:"
echo "  cd $TEST_REPO_PATH"
echo "  $PROJECT_ROOT_DIR/sjiek -i"
echo "  $PROJECT_ROOT_DIR/sjiek -n all_changes.txt --diff-type all -c"
echo "-----------------------------------"
