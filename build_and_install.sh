#!/bin/bash

# build_and_install.sh
# Builds the sjiek binary and installs it to ~/.local/bin/

# Exit immediately if a command exits with a non-zero status.
set -e

echo "--- Building sjiek ---"
# The -ldflags="-s -w" are optional:
# -s: Omit the symbol table and debug information.
# -w: Omit the DWARF symbol table.
# These flags reduce the binary size but make debugging harder if you were to use delve.
# For a release, they are often used. For development, you might omit them.
go build -ldflags="-s -w" -o sjiek .

if [ ! -f ./sjiek ]; then
    echo "Build failed! sjiek binary not found."
    exit 1
fi
echo "Build successful: ./sjiek created."

echo "--- Making sjiek executable ---"
chmod +x ./sjiek

# --- Installation ---
INSTALL_DIR="$HOME/.local/bin"

# Create the installation directory if it doesn't exist
if [ ! -d "$INSTALL_DIR" ]; then
    echo "Creating installation directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"
fi

echo "--- Moving sjiek to $INSTALL_DIR ---"
mv ./sjiek "$INSTALL_DIR/"

# Check if $INSTALL_DIR is in PATH and provide a message
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "WARNING: Installation directory $INSTALL_DIR is not in your PATH."
    echo "You may need to add it to your shell's configuration file (e.g., ~/.bashrc, ~/.zshrc):"
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    echo "Then, restart your shell or source the configuration file."
else
    echo "Installation successful! 'sjiek' is now available in your PATH."
fi

echo ""
echo "To run sjiek, type: sjiek"