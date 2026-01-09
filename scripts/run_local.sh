#!/bin/bash

# Flight Event Throttler - Local Development Run Script

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Flight Event Throttler - Local Setup${NC}"
echo -e "${GREEN}========================================${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go 1.22.5 or higher"
    exit 1
fi

echo -e "${YELLOW}Go version:${NC}"
go version

# Set default environment variables if not already set
export PORT=${PORT:-8080}
export LOG_LEVEL=${LOG_LEVEL:-INFO}
export BUFFER_TYPE=${BUFFER_TYPE:-ring}
export BUFFER_SIZE=${BUFFER_SIZE:-10000}
export RATE_LIMIT_RPS=${RATE_LIMIT_RPS:-100}

echo -e "\n${YELLOW}Environment Configuration:${NC}"
echo "  PORT: $PORT"
echo "  LOG_LEVEL: $LOG_LEVEL"
echo "  BUFFER_TYPE: $BUFFER_TYPE"
echo "  BUFFER_SIZE: $BUFFER_SIZE"
echo "  RATE_LIMIT_RPS: $RATE_LIMIT_RPS"

# Download dependencies
echo -e "\n${YELLOW}Downloading dependencies...${NC}"
go mod download

# Build the application
echo -e "\n${YELLOW}Building application...${NC}"
go build -o bin/flight-event-throttler ./cmd/server

if [ $? -eq 0 ]; then
    echo -e "${GREEN}Build successful!${NC}"
else
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi

# Run the application
echo -e "\n${YELLOW}Starting Flight Event Throttler...${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop${NC}\n"

./bin/flight-event-throttler
