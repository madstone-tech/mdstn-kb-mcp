#!/bin/bash
# Generate shell completions for kbvault
# This script uses the kbvault binary to generate completion scripts

set -e

# Ensure the binary is built
if [ ! -f "./bin/kbvault" ]; then
    echo "Building kbvault binary..."
    go build -o ./bin/kbvault ./cmd/kbvault
fi

# Create completions directory
mkdir -p completions

echo "Generating shell completions..."

# Generate bash completion
./bin/kbvault completion bash > completions/kbvault.bash
echo "✓ Generated bash completion"

# Generate zsh completion
./bin/kbvault completion zsh > completions/_kbvault
echo "✓ Generated zsh completion"

# Generate fish completion
./bin/kbvault completion fish > completions/kbvault.fish
echo "✓ Generated fish completion"

echo "All completions generated successfully!"
