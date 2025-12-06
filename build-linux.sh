#!/bin/bash
# Build script for Linux only

set -e

echo "=== Building Shotgun Code for Linux ==="
echo ""

# Check if wails is installed
if ! command -v wails &> /dev/null; then
    echo "❌ Wails not found. Install with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    exit 1
fi

cd backend

echo "Building Linux executable..."
wails build -clean

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Build successful!"
    echo ""
    
    # Show output file
    if [ -f "bin/ShotgunWB" ]; then
        ls -lh bin/ShotgunWB | awk '{print "Output: bin/ShotgunWB\nSize: " $5}'
    fi
else
    echo ""
    echo "✗ Build failed!"
    exit 1
fi

cd ..
