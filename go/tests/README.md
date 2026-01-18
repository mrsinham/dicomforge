# Integration Tests for Go DICOM Generator

This directory contains comprehensive integration tests for the DICOM MRI generator.

## Test Statistics

- **Total Test Files**: 6
- **Total Test Functions**: 35+
- **Test Categories**: Integration, Validation, Errors, Performance, Reproducibility, Utilities

## Test Files

### 1. integration_test.go
Core integration tests for end-to-end workflows.

#### TestGenerateSeries_Basic
Tests basic DICOM series generation with 5 images:
- Verifies correct number of files generated
- Checks that all files exist
- Validates UIDs and patient IDs are set

#### TestOrganizeFiles_DICOMDIRStructure
Tests DICOMDIR organization and file hierarchy:
- Verifies DICOMDIR file exists
- Checks PT000000/ST000000/SE000000/ hierarchy
- Validates image files are moved correctly
- Ensures temporary IMG*.dcm files are cleaned up

#### TestValidation_RequiredTags
Validates DICOM file contents:
- Checks all required DICOM tags are present
- Verifies tag values (Modality=MR, BitsAllocated=16, etc.)
- Parses generated files with suyashkumar/dicom library

#### TestMultiStudy
Tests multi-study generation:
- Generates 15 images across 3 studies
- Verifies correct directory structure for each study
- Validates image distribution

#### TestReproducibility_SameSeed
Tests reproducibility:
- Generates two series with same seed
- Verifies PatientID is identical
- Checks deterministic behavior

#### TestCalculateDimensions
Tests dimension calculation algorithm:
- Validates dimensions for various size/image combinations
- Checks dimensions are within expected ranges

### 2. validation_test.go
Advanced DICOM validation tests.

#### TestValidation_MRIParameters
Validates MRI-specific DICOM tags:
- Manufacturer, Model, Field Strength
- EchoTime, RepetitionTime, FlipAngle
- PixelSpacing, SliceThickness, SequenceName
- Verifies values are from expected lists

#### TestValidation_PixelData
Tests pixel data integrity:
- Verifies pixel data is not encapsulated
- Checks frame count
- Validates pixel data size matches dimensions
- Ensures pixel data is not all zeros

#### TestValidation_ImagePosition
Tests spatial information tags:
- ImagePositionPatient exists and is valid
- ImageOrientationPatient for axial orientation
- SliceLocation varies across slices

#### TestValidation_PatientInfo
Validates patient information consistency:
- All images in series have same patient info
- Patient name format (LASTNAME^FIRSTNAME)
- Patient sex is M or F
- Birth date format (YYYYMMDD)

#### TestValidation_UIDUniqueness
Tests UID uniqueness:
- SOP Instance UIDs are unique across all images
- Instance Numbers are unique and sequential
- Validates 10+ images have distinct identifiers

### 3. errors_test.go
Error handling and edge case tests.

#### TestErrors_InvalidNumImages
Tests error handling for invalid image counts:
- Zero images → error
- Negative images → error
- One image → valid
- Normal count → valid

#### TestErrors_InvalidTotalSize
Tests size string parsing errors:
- Invalid format → error
- Empty string → error
- Negative size → error
- Zero size → error
- Valid KB/MB/GB → success

#### TestErrors_TooSmallSize
Tests handling of size too small for metadata (< 100KB).

#### TestErrors_InvalidNumStudies
Tests various study count scenarios:
- Zero studies
- Negative studies
- More studies than images

#### TestEdgeCase_SingleImage
Tests generation with exactly 1 image.

#### TestEdgeCase_LargeNumberOfImages
Tests with 100 images (skipped in short mode).

#### TestEdgeCase_VerySmallImages
Tests minimal size (500KB) generation.

#### TestEdgeCase_ManyStudies
Tests 50 images across 10 studies.

#### TestCalculateDimensions_EdgeCases
Tests dimension calculation edge cases:
- Zero bytes
- Negative bytes
- Very small sizes
- Large sizes (4GB)

### 4. performance_test.go
Performance benchmarks and tests.

#### Benchmarks
- **BenchmarkGenerateSeries_Small**: 5 images, 10MB
- **BenchmarkGenerateSeries_Medium**: 20 images, 50MB
- **BenchmarkGenerateSeries_Large**: 50 images, 200MB (skip in short mode)
- **BenchmarkCalculateDimensions**: Dimension calculation speed
- **BenchmarkOrganizeFiles**: DICOMDIR organization speed

#### TestPerformance_MemoryUsage
Measures memory allocation for 50 images (200MB):
- Should use < 1GB RAM
- Logs allocated and total memory

#### TestPerformance_GenerationSpeed
Tests generation speed for different sizes:
- Small (5 images, 10MB): < 2 seconds
- Medium (20 images, 50MB): < 5 seconds
- Large (50 images, 200MB): < 15 seconds

### 5. reproducibility_test.go
Tests for deterministic generation.

#### TestReproducibility_DifferentSeed
Verifies different seeds produce different results:
- Seed 42 vs Seed 99
- PatientID should differ

#### TestReproducibility_AutoSeedFromDir
Tests auto-seed from directory name:
- Same directory name → consistent seed
- Different paths with same name

#### TestReproducibility_MultipleSeries
Generates 3 series with same seed:
- All should have identical PatientID
- Validates consistency

#### TestReproducibility_UIDGeneration
Tests UID generation determinism:
- Same seed → same UID
- Multiple test seeds
- Verifies UID format and length

#### TestReproducibility_PatientNames
Tests patient name generation:
- Format validation (LASTNAME^FIRSTNAME)
- French name characteristics
- Non-empty parts

#### TestReproducibility_PixelData
Tests pixel data reproducibility:
- Same seed → same file size
- Compares two generations

#### TestReproducibility_StudyUIDs
Tests study UID consistency:
- Validates UID format
- Checks UID length limits

### 6. utilities_test.go
Tests for utility functions.

#### TestUtil_ParseSize
Tests size string parsing with 20+ cases:
- Bytes, KB, MB, GB formats
- Upper/lowercase variations
- Decimal values (1.5MB, 2.5GB)
- With/without spaces
- Invalid formats

#### TestUtil_GeneratePatientName
Tests patient name generation:
- Male and female names
- Format validation
- Variability (generates multiple unique names)
- Shows examples

#### TestUtil_GenerateDeterministicUID
Tests UID generation:
- Short and long seeds
- Special characters in seeds
- UID format validation (digits and dots only)
- No leading zeros in components
- Length limits (max 64 chars)

#### TestUtil_UIDDeterminism
Verifies same seed → same UID consistently.

#### TestUtil_UIDUniqueness
Verifies different seeds → different UIDs.

#### TestUtil_PatientNameFormat
Validates DICOM name format:
- Exactly one '^' separator
- Valid character set
- Handles French accented characters

#### TestUtil_SizeEdgeCases
Tests edge cases in size parsing:
- 1B, 1KB, 1MB, 1GB
- Fractional sizes (0.5KB, 0.1MB)
- Rounding tolerance

## Running the Tests

### Prerequisites

1. **Build environment**: Requires Go 1.21+
2. **Network connection**: First run needs to download dependencies
3. **Disk space**: Tests create temporary DICOM files (cleaned up automatically)

### Run All Tests

```bash
cd /home/user/dicom-test/go
go test ./tests -v

# Or use the helper script
./tests/run_tests.sh -v
```

### Run Specific Test File

```bash
# Run only integration tests
go test ./tests -v -run "^Test.*" integration_test.go

# Run only validation tests
go test ./tests -v -run "^TestValidation.*"

# Run only error handling tests
go test ./tests -v -run "^TestErrors.*"

# Run only reproducibility tests
go test ./tests -v -run "^TestReproducibility.*"

# Run only utility tests
go test ./tests -v -run "^TestUtil.*"
```

### Run Specific Test Function

```bash
# Run only basic generation test
go test ./tests -v -run TestGenerateSeries_Basic

# Run only DICOMDIR structure test
go test ./tests -v -run TestOrganizeFiles_DICOMDIRStructure

# Run only MRI parameters validation
go test ./tests -v -run TestValidation_MRIParameters

# Run only patient info validation
go test ./tests -v -run TestValidation_PatientInfo
```

### Run with Coverage

```bash
go test ./tests -v -cover

# Detailed coverage report
go test ./tests -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Short Tests Only

```bash
# Skips long-running tests (100+ images, performance tests)
go test ./tests -v -short
```

### Run Benchmarks

```bash
# Run all benchmarks
go test ./tests -bench=. -benchmem

# Run specific benchmark
go test ./tests -bench=BenchmarkGenerateSeries_Small

# Run benchmarks with timing
go test ./tests -bench=. -benchtime=10s
```

### Run Parallel Tests

```bash
# Run tests in parallel (default Go behavior)
go test ./tests -v -parallel=4
```

### Run Tests with Timeout

```bash
# Set timeout for long-running tests
go test ./tests -v -timeout 10m
```

## Expected Output

Successful test run:
```
=== RUN   TestGenerateSeries_Basic
    integration_test.go:25: Generating DICOM series in: /tmp/TestGenerateSeries_Basic...
    integration_test.go:45: Generated file 1: /tmp/.../IMG0001.dcm
    integration_test.go:45: Generated file 2: /tmp/.../IMG0002.dcm
    ...
    integration_test.go:61: ✓ Basic generation test passed
--- PASS: TestGenerateSeries_Basic (0.50s)

=== RUN   TestOrganizeFiles_DICOMDIRStructure
    integration_test.go:79: Generated 5 files, organizing into DICOMDIR...
    integration_test.go:94: ✓ DICOMDIR exists: /tmp/.../DICOMDIR
    integration_test.go:102: ✓ Patient directory exists: /tmp/.../PT000000
    ...
    integration_test.go:133: ✓ DICOMDIR structure test passed
--- PASS: TestOrganizeFiles_DICOMDIRStructure (0.60s)

...

PASS
ok      github.com/julien/dicom-test/go/tests   3.456s
```

## Test Data

All tests use `t.TempDir()` which:
- Creates unique temporary directories for each test
- Automatically cleans up after test completion
- Prevents conflicts between parallel tests

## Troubleshooting

### Network Errors

If you see errors about downloading dependencies:
```
go: downloading github.com/suyashkumar/dicom v1.1.0
dial tcp: lookup proxy.golang.org: no such host
```

**Solution**: Ensure you have internet connection for first run, or use vendored dependencies:
```bash
go mod vendor
go test ./tests -mod=vendor -v
```

### Module Errors

If you see:
```
go: updates to go.mod needed; to update it:
    go mod tidy
```

**Solution**: Run from the go directory and update modules:
```bash
cd /home/user/dicom-test/go
go mod tidy
go test ./tests -v
```

### Permission Errors

If tests fail creating files:
```
permission denied: /tmp/...
```

**Solution**: Ensure /tmp is writable or set TMPDIR:
```bash
export TMPDIR=/path/to/writable/dir
go test ./tests -v
```

## Adding New Tests

To add a new integration test:

1. Create a new test function in `integration_test.go`:
```go
func TestMyNewFeature(t *testing.T) {
    outputDir := t.TempDir()

    // Test setup
    opts := internaldicom.GeneratorOptions{
        NumImages: 5,
        TotalSize: "10MB",
        OutputDir: outputDir,
        Seed: 42,
        NumStudies: 1,
    }

    // Execute
    files, err := internaldicom.GenerateDICOMSeries(opts)
    if err != nil {
        t.Fatalf("Failed: %v", err)
    }

    // Verify
    // ... your assertions here ...

    t.Logf("✓ My test passed")
}
```

2. Run your new test:
```bash
go test ./tests -v -run TestMyNewFeature
```

## CI/CD Integration

For automated testing in CI pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run integration tests
  run: |
    cd go
    go test ./tests -v -timeout 5m
```

## Performance

Expected test execution times (approximate):
- TestGenerateSeries_Basic: 0.3-0.5s
- TestOrganizeFiles_DICOMDIRStructure: 0.4-0.7s
- TestValidation_RequiredTags: 0.3-0.5s
- TestMultiStudy: 0.8-1.2s
- TestReproducibility_SameSeed: 0.5-0.8s
- TestCalculateDimensions: <0.1s

**Total**: ~3-5 seconds for all tests

## Next Steps

See `docs/plans/2026-01-18-go-integration-tests.md` for the complete testing plan including:
- Additional test suites (validation, compatibility, performance)
- Python comparison scripts
- Benchmark tests
- Extended validation with pydicom
