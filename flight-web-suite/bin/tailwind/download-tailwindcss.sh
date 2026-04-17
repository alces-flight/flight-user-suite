#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# file name
TAILWIND_BIN="tailwindcss"
TAILWIND_BIN_PATH="${SCRIPT_DIR}/${TAILWIND_BIN}"

echo "Downloading Tailwind CSS CLI to ${TAILWIND_BIN_PATH}..."

# OS detector
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  curl -sSL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 -o "$TAILWIND_BIN_PATH"
elif [[ "$OSTYPE" == "darwin"* ]]; then
  curl -sSL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64 -o "$TAILWIND_BIN_PATH"
else
  echo "OS not supported. Try to download manually"
  exit 1
fi

# Set permission to make the file executable
echo "Setting the file as executable..."
chmod +x "$TAILWIND_BIN_PATH"

echo "Tailwind CSS CLI is Installed!"
echo "Executable path: $TAILWIND_BIN_PATH"
