#!/bin/bash
# Build script for macOS only

set -e

echo "=== Building Shotgun Code for macOS ==="
echo ""

# Check if wails is installed
if ! command -v wails &> /dev/null; then
    echo "❌ Wails not found. Install with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    exit 1
fi

cd backend

echo "Building macOS application (universal binary)..."
wails build -clean

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Build successful!"
    echo ""
    
    # Show output file
    if [ -d "bin/ShotgunWB.app" ]; then
        du -sh bin/ShotgunWB.app | awk '{print "Output: bin/ShotgunWB.app\nSize: " $1}'
    fi
else
    echo ""
    echo "✗ Build failed!"
    exit 1
fi

cd ..
