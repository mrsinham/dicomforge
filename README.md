# DICOM MRI Generator

[![CI](https://github.com/mrsinham/dicomforge/actions/workflows/ci.yml/badge.svg)](https://github.com/mrsinham/dicomforge/actions/workflows/ci.yml)
[![Release](https://github.com/mrsinham/dicomforge/actions/workflows/release.yml/badge.svg)](https://github.com/mrsinham/dicomforge/actions/workflows/release.yml)

A CLI tool to generate valid DICOM MRI series for testing medical imaging platforms.

**Generates multiple DICOM files** (one per image) in a directory, using the standard format expected by medical platforms and PACS systems.

## Installation

### Download pre-built binaries

Download the latest release for your platform from [GitHub Releases](https://github.com/mrsinham/dicomforge/releases):

- `dicomforge-linux-amd64` - Linux (x86_64)
- `dicomforge-linux-arm64` - Linux (ARM64)
- `dicomforge-darwin-amd64` - macOS (Intel)
- `dicomforge-darwin-arm64` - macOS (Apple Silicon)
- `dicomforge-windows-amd64.exe` - Windows

```bash
# Example for Linux x86_64
curl -LO https://github.com/mrsinham/dicomforge/releases/latest/download/dicomforge-linux-amd64
chmod +x dicomforge-linux-amd64
sudo mv dicomforge-linux-amd64 /usr/local/bin/dicomforge
```

### Homebrew (macOS/Linux)

```bash
# Add the tap
brew tap mrsinham/tap

# Install dicomforge
brew install dicomforge
```

### Nix

```bash
# Run without installing
nix run github:mrsinham/dicomforge -- --num-images 10 --total-size 100MB

# Install to profile
nix profile install github:mrsinham/dicomforge

# Development shell (Go 1.24 + tools)
nix develop github:mrsinham/dicomforge
```

### Go install

```bash
go install github.com/mrsinham/dicomforge/cmd/dicomforge@latest
```

### Build from source

```bash
git clone https://github.com/mrsinham/dicomforge.git
cd dicomforge
go build -o dicomforge ./cmd/dicomforge/
```

## Quick Start

```bash
# Generate 10 DICOM images totaling 100MB
./dicomforge --num-images 10 --total-size 100MB

# Generate a full MRI series (120 slices, 1GB)
./dicomforge --num-images 120 --total-size 1GB --output mri_series
```

## Usage

```bash
./dicomforge --num-images <N> --total-size <SIZE> [options]
```

### Required Arguments

| Argument | Description |
|----------|-------------|
| `--num-images` | Number of images/slices to generate |
| `--total-size` | Total target size (e.g., `100MB`, `1GB`, `4.5GB`) |

### Optional Arguments

| Argument | Description | Default |
|----------|-------------|---------|
| `--output` | Output directory name | `dicom_series` |
| `--seed` | Random seed for reproducibility | auto-generated |
| `--num-studies` | Number of studies to generate | `1` |
| `--num-patients` | Number of patients (studies distributed among them) | `1` |
| `--workers` | Number of parallel workers | CPU core count |
| `--edge-cases` | Percentage of patients with edge case variations (0-100) | `0` |
| `--edge-case-types` | Comma-separated edge case types to enable | all types |
| `--help` | Show help message | - |

### Edge Case Types

When using `--edge-cases`, you can specify which types to enable with `--edge-case-types`:

| Type | Description |
|------|-------------|
| `special-chars` | Names with accents, hyphens, apostrophes (Müller-Schmidt, O'Connor, François) |
| `long-names` | Names at DICOM's 64-character limit |
| `old-dates` | Birth dates from 1900-1950, or partial dates (YYYY, YYYYMM) |
| `varied-ids` | Patient IDs with dashes, letters, spaces, or at max length |
| `missing-tags` | Omit optional DICOM tags (BodyPartExamined, StudyDescription, etc.) |

### Examples

```bash
# Basic usage: 120 images, 1GB total
./dicomforge --num-images 120 --total-size 1GB

# Custom output directory with fixed seed for reproducibility
./dicomforge --num-images 50 --total-size 500MB --output patient_001 --seed 42

# Generate multiple studies (useful for testing study management)
./dicomforge --num-images 30 --total-size 500MB --num-studies 3

# Generate multiple patients with studies distributed among them
./dicomforge --num-images 60 --total-size 1GB --num-studies 6 --num-patients 2

# Limit parallelism (useful on resource-constrained systems)
./dicomforge --num-images 100 --total-size 1GB --workers 4

# Large dataset for stress testing
./dicomforge --num-images 500 --total-size 4GB --output stress_test

# Generate edge cases for robustness testing (25% of patients)
./dicomforge --num-images 100 --total-size 1GB --num-patients 20 \
  --edge-cases 25 --edge-case-types "special-chars,long-names"

# Generate all edge case types for comprehensive testing
./dicomforge --num-images 50 --total-size 500MB --num-studies 10 --num-patients 10 --edge-cases 50
```

## Output Structure

The generator creates a standard DICOMDIR structure:

```
output_directory/
├── DICOMDIR                      # Directory index file
└── PT000000/                     # Patient directory
    └── ST000000/                 # Study directory
        └── SE000000/             # Series directory
            ├── IM000001          # Image 1
            ├── IM000002          # Image 2
            └── ...
```

This hierarchy follows the DICOM standard and is compatible with:
- PACS systems (Orthanc, dcm4chee, etc.)
- DICOM viewers (Horos, OsiriX, RadiAnt, etc.)
- Medical imaging platforms

## Features

- **Standard DICOM format**: Generates valid DICOM files readable by any compliant software
- **DICOMDIR support**: Automatic directory index file creation
- **PT/ST/SE hierarchy**: Standard patient/study/series folder structure
- **Visual overlay**: Each image shows "File X/Y" text for easy verification
- **Parallel generation**: Worker pool for fast generation (~4.5x speedup)
- **Realistic metadata**: Simulated MRI parameters from major vendors (Siemens, GE, Philips)
- **Realistic patient names**: Generated patient names (80% English, 20% French)
- **Edge case generation**: Special characters, long names, old dates, varied IDs for robustness testing
- **Reproducible output**: Same seed produces identical files
- **Window/Level tags**: Proper display settings for DICOM viewers

## Performance

Benchmarks on a 24-core CPU:

| Images | Total Size | Sequential | Parallel (24 workers) |
|--------|------------|------------|----------------------|
| 50     | 100MB      | ~3.1s      | ~0.7s                |
| 120    | 1GB        | ~15s       | ~3s                  |
| 500    | 4GB        | ~60s       | ~12s                 |

## Reproducibility

The generator supports deterministic output:

```bash
# These two commands produce identical files
./dicomforge --num-images 10 --total-size 100MB --output test --seed 42
./dicomforge --num-images 10 --total-size 100MB --output test --seed 42
```

When no seed is provided, a deterministic seed is generated from the output directory name, ensuring that regenerating with the same output directory produces the same patient/study IDs.

## Testing

```bash
# Run unit tests
go test ./internal/...

# Run integration tests
go test ./tests/...

# Run all tests
go test ./...

# Run with verbose output
go test -v ./...
```

## Project Structure

```
.
├── cmd/dicomforge/    # CLI entry point
├── internal/
│   ├── dicom/                 # DICOM generation and DICOMDIR
│   ├── image/                 # Pixel data generation
│   └── util/                  # Utilities (UID generation, size parsing)
├── tests/                     # Integration tests
├── scripts/                   # Validation scripts
├── python/                    # Legacy Python version
│   ├── generate_dicom_mri.py
│   ├── requirements.txt
│   └── tests/
└── go.mod
```

## Legacy Python Version

The original Python implementation is preserved in the `python/` directory:

```bash
cd python
pip install -r requirements.txt
python generate_dicom_mri.py --num-images 10 --total-size 100MB
```

Note: The Go version is recommended for production use due to better performance and parallel generation support.

## Use Cases

- **Platform testing**: Generate test data for medical imaging platforms
- **PACS integration**: Test DICOM import/export functionality
- **Viewer development**: Create sample data for DICOM viewer development
- **Load testing**: Generate large datasets for performance testing
- **CI/CD pipelines**: Reproducible test data generation

## License

MIT License - See LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.
