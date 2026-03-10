#!/bin/bash
# Connect to BlackSector server with test client

echo "Building test client..."
go build -o testclient ./cmd/testclient

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Connecting to BlackSector server..."
echo ""

./testclient -token test_token_12345 -host localhost:2222
