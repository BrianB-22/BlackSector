#!/bin/bash
# BlackSector Server Startup Script

echo "Building BlackSector server..."
go build -o blacksector-server ./cmd/server

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Starting BlackSector server on port 2222..."
echo "Connect with: ssh localhost -p 2222"
echo "Press Ctrl+C to stop the server"
echo ""

./blacksector-server
