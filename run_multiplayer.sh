#!/bin/zsh

# Default count to 2 if not provided
COUNT=${1:-2}
GODOT_PATH="/Applications/Godot.app/Contents/MacOS/Godot"
PROJECT_DIR="$(pwd)/frontend"

echo "Launching $COUNT multiplayer instances..."

for i in $(seq 1 $COUNT); do
    echo "Launching Player $i..."
    # Run in background
    "$GODOT_PATH" --path "$PROJECT_DIR" -- --profile=$i --side-by-side=$COUNT &
done

echo "All instances launched."
