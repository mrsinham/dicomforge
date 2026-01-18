# Compatibility Scripts

Scripts for comparing Python and Go DICOM implementations and validating cross-compatibility.

## Scripts

### extract_metadata.py

Extracts metadata from DICOM files for comparison.

**Usage:**
```bash
python3 extract_metadata.py <dicom_directory> [output_json]
```

**Example:**
```bash
# Output to stdout
python3 extract_metadata.py test-output

# Save to file
python3 extract_metadata.py test-output metadata.json
```

**Output:** JSON file with:
- File count
- Patient information (ID, Name, DOB, Sex)
- Study information (UID, ID, Description)
- Series information (UID, Number, Modality)
- Instance information (UID, Number)
- Image dimensions (Rows, Columns, BitsAllocated)
- MRI parameters (Manufacturer, Model, TE, TR, etc.)

### validate_dicom.py

Validates DICOM files using pydicom.

**Usage:**
```bash
python3 validate_dicom.py <dicom_directory>
```

**Example:**
```bash
python3 validate_dicom.py test-go-output
```

**Validates:**
- Required tags present (PatientName, StudyUID, etc.)
- Tag values correct (Modality=MR, BitsAllocated=16, etc.)
- SOP Class UID for MR Image Storage
- Pixel data exists and correct size
- UID format (contains dots, max 64 chars)
- Patient name format (contains ^)

**Exit codes:**
- 0: All files valid
- 1: Validation errors found

### compare_python_go.sh

Comprehensive comparison between Python and Go implementations.

**Usage:**
```bash
# Use defaults (seed=42, 5 images, 10MB)
./compare_python_go.sh

# Custom parameters
SEED=99 NUM_IMAGES=10 SIZE=20MB ./compare_python_go.sh
```

**What it does:**
1. Generates DICOM series with Python (if available)
2. Generates DICOM series with Go
3. Compares file counts and directory structure
4. Extracts metadata from both outputs
5. Validates both with pydicom
6. Compares metadata (Patient ID, Study UID, etc.)

**Requirements:**
- Python 3 with pydicom installed
- Go binary built at `go/bin/generate-dicom-mri`
- Python generator script at root

**Environment variables:**
- `SEED`: Random seed (default: 42)
- `NUM_IMAGES`: Number of images (default: 5)
- `SIZE`: Total size (default: 10MB)

## Requirements

### Python

```bash
pip install pydicom
```

### Go

Build the binary first:
```bash
cd /home/user/dicom-test/go
go build -o bin/generate-dicom-mri ./cmd/generate-dicom-mri
```

## Usage Examples

### Quick Validation

Validate Go-generated files:
```bash
# Generate files
cd /home/user/dicom-test/go
./bin/generate-dicom-mri --num-images 5 --total-size 10MB --output ../test-go --seed 42

# Validate
./scripts/validate_dicom.py ../test-go
```

### Extract and Inspect Metadata

```bash
# Extract metadata
./scripts/extract_metadata.py ../test-go metadata.json

# View metadata
cat metadata.json | python3 -m json.tool

# Check specific fields
cat metadata.json | python3 -c "import json, sys; data=json.load(sys.stdin); print(f\"Files: {data['file_count']}\"); print(f\"Patient: {data['files'][0]['patient_name']}\")"
```

### Full Comparison

```bash
# Compare with default parameters
./scripts/compare_python_go.sh

# Compare with custom parameters
SEED=42 NUM_IMAGES=10 SIZE=50MB ./scripts/compare_python_go.sh

# Compare outputs are saved to:
# - test-python-seed42/
# - test-go-seed42/
# - metadata-python.json
# - metadata-go.json
```

### Compare Specific Aspects

```bash
# After running compare_python_go.sh, inspect differences:

# Compare patient IDs
python3 -c "
import json
py = json.load(open('metadata-python.json'))
go = json.load(open('metadata-go.json'))
print(f\"Python PatientID: {py['files'][0]['patient_id']}\")
print(f\"Go PatientID: {go['files'][0]['patient_id']}\")
print(f\"Match: {py['files'][0]['patient_id'] == go['files'][0]['patient_id']}\")
"

# Compare dimensions
python3 -c "
import json
py = json.load(open('metadata-python.json'))
go = json.load(open('metadata-go.json'))
py_dims = f\"{py['files'][0]['rows']}x{py['files'][0]['columns']}\"
go_dims = f\"{go['files'][0]['rows']}x{go['files'][0]['columns']}\"
print(f\"Python: {py_dims}\")
print(f\"Go: {go_dims}\")
print(f\"Match: {py_dims == go_dims}\")
"
```

## Integration with Tests

The compatibility tests (`tests/compatibility_test.go`) use these scripts:

```bash
# Run compatibility tests
cd /home/user/dicom-test/go
go test ./tests -v -run TestCompatibility

# Run specific test
go test ./tests -v -run TestCompatibility_PythonValidation

# Skip if Python not available
go test ./tests -v -run TestCompatibility  # Will skip if pydicom not found
```

## Expected Differences

When comparing Python vs Go outputs with the same seed:

### ✅ Should Match
- File count
- Modality (MR)
- Image dimensions (calculated from size)
- DICOM structure (PT*/ST*/SE*)
- Tag presence and types

### ⚠️ May Differ
- **Patient ID**: Different RNG implementations
- **Patient Name**: Different RNG implementations
- **Manufacturer/Model**: Different RNG sequences
- **Pixel values**: Different RNG (Python: numpy, Go: math/rand)

### 📝 Known Limitations
- UID generation uses same algorithm but different output dir handling
- Text overlay rendering differs (Python: TrueType fonts, Go: basic rectangles)
- DICOMDIR format: Python uses FileSet, Go uses manual implementation

## Troubleshooting

### "pydicom not installed"

```bash
pip install pydicom
# or
pip3 install pydicom
```

### "Python generator script not found"

Ensure you're in the correct directory:
```bash
cd /home/user/dicom-test
ls generate_dicom_mri.py  # Should exist
```

### "Go binary not found"

Build it first:
```bash
cd /home/user/dicom-test/go
go build -o bin/generate-dicom-mri ./cmd/generate-dicom-mri
```

### Permission denied

Make scripts executable:
```bash
chmod +x scripts/*.sh scripts/*.py
```

## Output Cleanup

After testing:
```bash
# Remove test outputs
rm -rf test-python-seed* test-go-seed*
rm -f metadata-python.json metadata-go.json

# Or use the suggestion from compare_python_go.sh output
```
