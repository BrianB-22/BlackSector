#!/usr/bin/env bash
# BlackSector Build Script
# Compiles the server binary with appropriate optimization flags

set -e

# Configuration
BINARY_NAME="blacksector"
CMD_PATH="./cmd/server"
OUTPUT_DIR="bin"
MODULE_PATH="github.com/BrianB-22/BlackSector"

# Version information (can be overridden by environment variables)
VERSION="${VERSION:-dev}"
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S_UTC')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Build mode (default: dev)
BUILD_MODE="${BUILD_MODE:-dev}"

# Target platform (default: current platform)
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Print usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build BlackSector server binary

OPTIONS:
    -m, --mode MODE         Build mode: dev (default) or prod
    -o, --output DIR        Output directory (default: bin)
    -v, --version VERSION   Version string (default: dev)
    -t, --target TARGET     Target platform (e.g., linux/amd64, darwin/arm64)
    -c, --clean             Clean build artifacts before building
    -h, --help              Show this help message

EXAMPLES:
    # Development build (current platform)
    $0

    # Production build
    $0 --mode prod

    # Cross-compile for Linux amd64
    $0 --mode prod --target linux/amd64

    # Clean and build
    $0 --clean --mode prod

ENVIRONMENT VARIABLES:
    BUILD_MODE              Build mode (dev or prod)
    VERSION                 Version string
    GOOS                    Target operating system
    GOARCH                  Target architecture

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--mode)
            BUILD_MODE="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -t|--target)
            IFS='/' read -r GOOS GOARCH <<< "$2"
            shift 2
            ;;
        -c|--clean)
            CLEAN=1
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Validate build mode
if [[ "$BUILD_MODE" != "dev" && "$BUILD_MODE" != "prod" ]]; then
    log_error "Invalid build mode: $BUILD_MODE (must be 'dev' or 'prod')"
    exit 1
fi

# Clean if requested
if [[ -n "$CLEAN" ]]; then
    log_info "Cleaning build artifacts..."
    rm -rf "$OUTPUT_DIR"
    go clean -cache
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Determine output binary name
OUTPUT_BINARY="$OUTPUT_DIR/$BINARY_NAME"
if [[ "$GOOS" == "windows" ]]; then
    OUTPUT_BINARY="${OUTPUT_BINARY}.exe"
fi
if [[ "$GOOS" != "$(go env GOOS)" || "$GOARCH" != "$(go env GOARCH)" ]]; then
    OUTPUT_BINARY="$OUTPUT_DIR/${BINARY_NAME}-${GOOS}-${GOARCH}"
    if [[ "$GOOS" == "windows" ]]; then
        OUTPUT_BINARY="${OUTPUT_BINARY}.exe"
    fi
fi

# Build ldflags for version information
LDFLAGS="-X 'main.Version=$VERSION'"
LDFLAGS="$LDFLAGS -X 'main.BuildTime=$BUILD_TIME'"
LDFLAGS="$LDFLAGS -X 'main.GitCommit=$GIT_COMMIT'"
LDFLAGS="$LDFLAGS -X 'main.GitBranch=$GIT_BRANCH'"

# Add production optimizations
if [[ "$BUILD_MODE" == "prod" ]]; then
    # Strip debug info and symbol table for smaller binary
    LDFLAGS="$LDFLAGS -s -w"
    BUILD_FLAGS="-trimpath"
    log_info "Building in PRODUCTION mode (optimized, stripped)"
else
    BUILD_FLAGS=""
    log_info "Building in DEVELOPMENT mode (debug symbols included)"
fi

# Print build configuration
log_info "Build Configuration:"
echo "  Binary:      $OUTPUT_BINARY"
echo "  Version:     $VERSION"
echo "  Commit:      $GIT_COMMIT"
echo "  Branch:      $GIT_BRANCH"
echo "  Build Time:  $BUILD_TIME"
echo "  Target:      $GOOS/$GOARCH"
echo "  Mode:        $BUILD_MODE"
echo ""

# Run go build
log_info "Compiling..."
GOOS="$GOOS" GOARCH="$GOARCH" go build \
    $BUILD_FLAGS \
    -ldflags "$LDFLAGS" \
    -o "$OUTPUT_BINARY" \
    "$CMD_PATH"

# Check if build succeeded
if [[ $? -eq 0 ]]; then
    # Get binary size
    BINARY_SIZE=$(du -h "$OUTPUT_BINARY" | cut -f1)
    
    log_info "Build successful!"
    echo "  Output:      $OUTPUT_BINARY"
    echo "  Size:        $BINARY_SIZE"
    
    # Make binary executable (not needed on Windows)
    if [[ "$GOOS" != "windows" ]]; then
        chmod +x "$OUTPUT_BINARY"
    fi
    
    # Show version info if it's a local build
    if [[ "$GOOS" == "$(go env GOOS)" && "$GOARCH" == "$(go env GOARCH)" ]]; then
        echo ""
        log_info "To run the server:"
        echo "  $OUTPUT_BINARY --config config/server.json"
    fi
else
    log_error "Build failed"
    exit 1
fi
