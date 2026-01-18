#!/bin/bash
# Compare DICOM generation between Python and Go implementations

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
GO_DIR="$PROJECT_ROOT/go"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "======================================"
echo "Python vs Go DICOM Comparison"
echo "======================================"
echo ""

# Configuration
SEED=${SEED:-42}
NUM_IMAGES=${NUM_IMAGES:-5}
SIZE=${SIZE:-10MB}
PYTHON_OUTPUT="$PROJECT_ROOT/test-python-seed$SEED"
GO_OUTPUT="$PROJECT_ROOT/test-go-seed$SEED"

echo "Configuration:"
echo "  Seed: $SEED"
echo "  Images: $NUM_IMAGES"
echo "  Size: $SIZE"
echo "  Python output: $PYTHON_OUTPUT"
echo "  Go output: $GO_OUTPUT"
echo ""

# Clean up previous runs
if [ -d "$PYTHON_OUTPUT" ]; then
    echo "Cleaning up previous Python output..."
    rm -rf "$PYTHON_OUTPUT"
fi

if [ -d "$GO_OUTPUT" ]; then
    echo "Cleaning up previous Go output..."
    rm -rf "$GO_OUTPUT"
fi

echo ""
echo "======================================"
echo "1. Generating with Python"
echo "======================================"

if [ ! -f "$PROJECT_ROOT/generate_dicom_mri.py" ]; then
    echo -e "${RED}Error: Python script not found at $PROJECT_ROOT/generate_dicom_mri.py${NC}"
    exit 1
fi

cd "$PROJECT_ROOT"

# Check if Python and pydicom are available
if ! command -v python3 &> /dev/null; then
    echo -e "${RED}Error: python3 not found${NC}"
    exit 1
fi

if ! python3 -c "import pydicom" 2>/dev/null; then
    echo -e "${YELLOW}Warning: pydicom not installed. Install with: pip install pydicom${NC}"
    echo "Skipping Python generation..."
    SKIP_PYTHON=1
fi

if [ -z "$SKIP_PYTHON" ]; then
    python3 generate_dicom_mri.py \
        --num-images $NUM_IMAGES \
        --total-size $SIZE \
        --output "$PYTHON_OUTPUT" \
        --seed $SEED \
        --num-studies 1

    if [ $? -ne 0 ]; then
        echo -e "${RED}Python generation failed${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ Python generation complete${NC}"
else
    echo -e "${YELLOW}⊘ Python generation skipped${NC}"
fi

echo ""
echo "======================================"
echo "2. Generating with Go"
echo "======================================"

cd "$GO_DIR"

# Check if Go binary exists
GO_BINARY="$GO_DIR/bin/generate-dicom-mri"

if [ ! -f "$GO_BINARY" ]; then
    echo "Go binary not found. Building..."
    if ! go build -o "$GO_BINARY" ./cmd/generate-dicom-mri 2>&1; then
        echo -e "${RED}Error: Failed to build Go binary${NC}"
        echo "Try running: cd go && go build -o bin/generate-dicom-mri ./cmd/generate-dicom-mri"
        exit 1
    fi
    echo -e "${GREEN}✓ Go binary built${NC}"
fi

"$GO_BINARY" \
    --num-images $NUM_IMAGES \
    --total-size $SIZE \
    --output "$GO_OUTPUT" \
    --seed $SEED \
    --num-studies 1

if [ $? -ne 0 ]; then
    echo -e "${RED}Go generation failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Go generation complete${NC}"

echo ""
echo "======================================"
echo "3. Comparing Outputs"
echo "======================================"

# Count files
if [ -z "$SKIP_PYTHON" ]; then
    PYTHON_COUNT=$(find "$PYTHON_OUTPUT" -name 'IM*' -type f | wc -l)
    echo "Python files: $PYTHON_COUNT"
fi

GO_COUNT=$(find "$GO_OUTPUT" -name 'IM*' -type f | wc -l)
echo "Go files: $GO_COUNT"

if [ -z "$SKIP_PYTHON" ]; then
    if [ "$PYTHON_COUNT" -eq "$GO_COUNT" ]; then
        echo -e "${GREEN}✓ File counts match${NC}"
    else
        echo -e "${RED}✗ File counts differ${NC}"
    fi
fi

echo ""
echo "Directory structures:"
echo ""
echo "Python:"
if [ -z "$SKIP_PYTHON" ]; then
    find "$PYTHON_OUTPUT" -type d | sort | head -10
else
    echo "  (skipped)"
fi

echo ""
echo "Go:"
find "$GO_OUTPUT" -type d | sort | head -10

echo ""
echo "======================================"
echo "4. Extracting Metadata"
echo "======================================"

# Extract metadata using Python script
if [ -z "$SKIP_PYTHON" ]; then
    echo "Extracting Python metadata..."
    python3 "$SCRIPT_DIR/extract_metadata.py" "$PYTHON_OUTPUT" "$PROJECT_ROOT/metadata-python.json"
fi

echo "Extracting Go metadata..."
python3 "$SCRIPT_DIR/extract_metadata.py" "$GO_OUTPUT" "$PROJECT_ROOT/metadata-go.json"

echo ""
echo "======================================"
echo "5. Validating DICOM Files"
echo "======================================"

echo "Validating Go-generated files..."
if python3 "$SCRIPT_DIR/validate_dicom.py" "$GO_OUTPUT"; then
    echo -e "${GREEN}✓ All Go files are valid DICOM${NC}"
else
    echo -e "${RED}✗ Some Go files have validation errors${NC}"
fi

if [ -z "$SKIP_PYTHON" ]; then
    echo ""
    echo "Validating Python-generated files..."
    if python3 "$SCRIPT_DIR/validate_dicom.py" "$PYTHON_OUTPUT"; then
        echo -e "${GREEN}✓ All Python files are valid DICOM${NC}"
    else
        echo -e "${RED}✗ Some Python files have validation errors${NC}"
    fi
fi

echo ""
echo "======================================"
echo "6. Comparing Metadata"
echo "======================================"

if [ -z "$SKIP_PYTHON" ]; then
    echo "Comparing patient IDs, study UIDs, and other metadata..."

    # Use Python to compare JSON files
    python3 << 'EOFPYTHON'
import json
import sys

try:
    with open('metadata-python.json') as f:
        py_data = json.load(f)
    with open('metadata-go.json') as f:
        go_data = json.load(f)

    print(f"\nPython: {py_data['file_count']} files")
    print(f"Go:     {go_data['file_count']} files")

    if py_data['file_count'] != go_data['file_count']:
        print("❌ File counts differ!")
        sys.exit(1)

    # Compare first file metadata
    py_first = py_data['files'][0] if py_data['files'] else {}
    go_first = go_data['files'][0] if go_data['files'] else {}

    print("\n--- First File Comparison ---")

    # Patient ID
    py_pid = py_first.get('patient_id', 'N/A')
    go_pid = go_first.get('patient_id', 'N/A')
    match = "✓" if py_pid == go_pid else "✗"
    print(f"{match} Patient ID:")
    print(f"  Python: {py_pid}")
    print(f"  Go:     {go_pid}")

    # Patient Name
    py_name = py_first.get('patient_name', 'N/A')
    go_name = go_first.get('patient_name', 'N/A')
    match = "✓" if py_name == go_name else "✗"
    print(f"{match} Patient Name:")
    print(f"  Python: {py_name}")
    print(f"  Go:     {go_name}")

    # Modality
    py_mod = py_first.get('modality', 'N/A')
    go_mod = go_first.get('modality', 'N/A')
    match = "✓" if py_mod == go_mod else "✗"
    print(f"{match} Modality: {py_mod} vs {go_mod}")

    # Dimensions
    py_dims = f"{py_first.get('rows', 'N/A')}x{py_first.get('columns', 'N/A')}"
    go_dims = f"{go_first.get('rows', 'N/A')}x{go_first.get('columns', 'N/A')}"
    match = "✓" if py_dims == go_dims else "✗"
    print(f"{match} Dimensions: {py_dims} vs {go_dims}")

    # Manufacturer (might differ due to RNG)
    py_mfr = py_first.get('manufacturer', 'N/A')
    go_mfr = go_first.get('manufacturer', 'N/A')
    print(f"Manufacturer:")
    print(f"  Python: {py_mfr}")
    print(f"  Go:     {go_mfr}")

    print("\n✓ Metadata comparison complete")

except Exception as e:
    print(f"Error comparing metadata: {e}")
    sys.exit(1)
EOFPYTHON
else
    echo -e "${YELLOW}⊘ Python comparison skipped (no Python output)${NC}"
fi

echo ""
echo "======================================"
echo "Summary"
echo "======================================"

echo ""
echo "Generated outputs:"
if [ -z "$SKIP_PYTHON" ]; then
    echo "  Python: $PYTHON_OUTPUT"
fi
echo "  Go:     $GO_OUTPUT"

echo ""
echo "Metadata files:"
if [ -z "$SKIP_PYTHON" ]; then
    echo "  Python: $PROJECT_ROOT/metadata-python.json"
fi
echo "  Go:     $PROJECT_ROOT/metadata-go.json"

echo ""
echo -e "${GREEN}✓ Comparison complete!${NC}"
echo ""
echo "To clean up:"
if [ -z "$SKIP_PYTHON" ]; then
    echo "  rm -rf $PYTHON_OUTPUT $GO_OUTPUT metadata-python.json metadata-go.json"
else
    echo "  rm -rf $GO_OUTPUT metadata-go.json"
fi
