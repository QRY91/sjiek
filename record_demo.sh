#!/bin/bash

set -e

# --- Configuration ---
TEMPLATE_TAPE_FILE="sjiek_demo.tape.template"
FINAL_TAPE_FILE="sjiek_demo.tape" # VHS will use this
OUTPUT_GIF_NAME="sjiek_demo.gif"
ASSETS_DIR="./assets"
VHS_OUTPUT_DIR="."

PROJECT_ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
cd "$PROJECT_ROOT_DIR"

echo "--- sjiek demo recording script ---"

# --- Dynamic Values (configurable here) ---
# These are the default values sjiek uses, which we need to "erase" in the demo
DEFAULT_OUTPUT_DIR_IN_SJIEK="$HOME/llm_context_diffs" # Or get this from sjiek --help if possible
DEFAULT_FILENAME_IN_SJIEK="current_diff.txt"

# Interaction choices for the demo
DEMO_OUTPUT_DIR_ACTION="default" # "default" or "type_new"
DEMO_NEW_OUTPUT_DIR="./output_diffs_demo" # Only used if DEMO_OUTPUT_DIR_ACTION is "type_new"

# --- Helper function to generate backspace commands ---
generate_backspaces() {
    local text_to_erase="$1"
    local num_chars=${#text_to_erase}
    local backspace_commands=""
    for (( i=0; i<num_chars; i++ )); do
        backspace_commands+="Backspace@50ms Sleep 50ms\n"
    done
    # Add a small pause after backspacing
    if [ "$num_chars" -gt 0 ]; then
        backspace_commands+="Sleep 200ms\n"
    fi
    echo -e "$backspace_commands" # -e to interpret \n
}

# --- Generate the final .tape file from the template ---
echo "Generating '$FINAL_TAPE_FILE' from '$TEMPLATE_TAPE_FILE'..."

# Calculate backspaces for filename
FILENAME_BACKSPACES_CMDS=$(generate_backspaces "$DEFAULT_FILENAME_IN_SJIEK")

# Handle output directory interaction
OUTPUT_DIR_INTERACTION_CMDS=""
if [ "$DEMO_OUTPUT_DIR_ACTION" == "type_new" ]; then
    OUTPUT_DIR_BACKSPACES_CMDS=$(generate_backspaces "$DEFAULT_OUTPUT_DIR_IN_SJIEK")
    # Using Ctrl+A Backspace might be more reliable if gum supports it well in VHS
    # OUTPUT_DIR_INTERACTION_CMDS="Ctrl+A Sleep 100ms\nBackspace Sleep 100ms\nSleep 200ms\n"
    OUTPUT_DIR_INTERACTION_CMDS+="$OUTPUT_DIR_BACKSPACES_CMDS"
    OUTPUT_DIR_INTERACTION_CMDS+="Type \"$DEMO_NEW_OUTPUT_DIR\" Sleep 500ms\n"
elif [ "$DEMO_OUTPUT_DIR_ACTION" == "default" ]; then
    OUTPUT_DIR_INTERACTION_CMDS="" # Just press Enter
fi

# Use sed for replacement. Using a different delimiter for sed like | or #
# in case paths/defaults contain /
# Create a temporary file for sed output to avoid issues with in-place editing of the same file being read
sed \
    -e "s|%%FILENAME_BACKSPACES%%|$FILENAME_BACKSPACES_CMDS|g" \
    -e "s|%%OUTPUT_DIR_INTERACTION%%|$OUTPUT_DIR_INTERACTION_CMDS|g" \
    "$TEMPLATE_TAPE_FILE" > "$FINAL_TAPE_FILE"

echo "'$FINAL_TAPE_FILE' generated."

# --- The rest of the script is similar to before ---
if ! command -v vhs &> /dev/null; then
    echo "ERROR: 'vhs' command not found." exit 1
fi
if [ ! -f "$FINAL_TAPE_FILE" ]; then
    echo "ERROR: Generated VHS tape file '$FINAL_TAPE_FILE' not found." exit 1
fi

echo "INFO: Ensure 'sjiek' is built and accessible by commands in '$FINAL_TAPE_FILE'."
sleep 1

echo "Running VHS with tape file: $FINAL_TAPE_FILE..."
vhs "$FINAL_TAPE_FILE"

EXPECTED_VHS_OUTPUT_PATH="$VHS_OUTPUT_DIR/$OUTPUT_GIF_NAME"
if [ ! -f "$EXPECTED_VHS_OUTPUT_PATH" ]; then
    echo "ERROR: VHS did not create expected output: $EXPECTED_VHS_OUTPUT_PATH" exit 1
fi
echo "VHS recording complete: $EXPECTED_VHS_OUTPUT_PATH"

if [ "$VHS_OUTPUT_DIR/$OUTPUT_GIF_NAME" != "$ASSETS_DIR/$OUTPUT_GIF_NAME" ]; then
    echo "Moving GIF to $ASSETS_DIR/ ..."
    mkdir -p "$ASSETS_DIR"
    mv "$EXPECTED_VHS_OUTPUT_PATH" "$ASSETS_DIR/$OUTPUT_GIF_NAME"
    echo "GIF moved to $ASSETS_DIR/$OUTPUT_GIF_NAME"
fi

echo ""
echo "--- Demo Recording Process Finished ---"
echo "Review '$ASSETS_DIR/$OUTPUT_GIF_NAME'."
echo "To change demo interactions, edit this script ('$0') or '$TEMPLATE_TAPE_FILE'."