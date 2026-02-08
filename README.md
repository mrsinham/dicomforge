# DICOM Generator (dicomforge)

[![CI](https://github.com/mrsinham/dicomforge/actions/workflows/ci.yml/badge.svg)](https://github.com/mrsinham/dicomforge/actions/workflows/ci.yml)
[![Release](https://github.com/mrsinham/dicomforge/actions/workflows/release.yml/badge.svg)](https://github.com/mrsinham/dicomforge/actions/workflows/release.yml)

A CLI tool to generate valid DICOM series for testing medical imaging platforms. Supports multiple modalities: MR, CT, CR, DX, US, and MG.

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
# Generate 10 MR DICOM images totaling 100MB
./dicomforge --num-images 10 --total-size 100MB

# Generate a full MRI series (120 slices, 1GB)
./dicomforge --num-images 120 --total-size 1GB --output mri_series

# Generate a CT series (100 slices)
./dicomforge --num-images 100 --total-size 200MB --modality CT --output ct_series

# Generate chest X-ray images (DX)
./dicomforge --num-images 2 --total-size 50MB --modality DX --body-part CHEST

# Generate ultrasound images
./dicomforge --num-images 20 --total-size 30MB --modality US

# Generate mammography images
./dicomforge --num-images 4 --total-size 100MB --modality MG
```

> **[See Complete Examples Guide](docs/EXAMPLES.md)** - Detailed examples for all features: multi-series, custom tags, edge cases, clinical trial simulations, and more.

## Interactive Wizard

For a guided experience, use the interactive wizard:

```bash
# Launch the wizard
dicomforge wizard

# Or use the --interactive flag
dicomforge --interactive
```

**Key features:**
- **Guided configuration** - Step-by-step prompts for all options
- **Help panel** - Context-sensitive help for each field
- **Live preview** - See the resulting DICOM structure before generating
- **Config save/load** - Save your configuration to YAML for later use

**Wizard flow:**
1. Global settings (modality, total images, output directory)
2. Patient configuration (name, ID, birth date, sex)
3. Study and series setup per patient
4. Preview the configuration
5. Generate or save config for later

```bash
# Load a saved configuration
dicomforge --config myconfig.yaml

# Edit an existing config with the wizard
dicomforge wizard --from myconfig.yaml
```

> **[See Examples Guide](docs/EXAMPLES.md#interactive-wizard)** for detailed wizard usage and example config files.

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
| `--modality` | Imaging modality: `MR`, `CT`, `CR`, `DX`, `US`, `MG` | `MR` |
| `--num-studies` | Number of studies to generate | `1` |
| `--num-patients` | Number of patients (studies distributed among them) | `1` |
| `--workers` | Number of parallel workers | CPU core count |
| `--edge-cases` | Percentage of patients with edge case variations (0-100) | `0` |
| `--edge-case-types` | Comma-separated edge case types to enable | all types |
| `--corrupt` | Vendor corruption types (comma-separated, or `all`) | disabled |
| `--help` | Show help message | - |

### Modality Support

| Modality | Description | SOP Class |
|----------|-------------|-----------|
| `MR` | Magnetic Resonance Imaging | MR Image Storage |
| `CT` | Computed Tomography | CT Image Storage |
| `CR` | Computed Radiography | Computed Radiography Image Storage |
| `DX` | Digital X-Ray | Digital X-Ray Image Storage |
| `US` | Ultrasound | Ultrasound Image Storage |
| `MG` | Mammography | Digital Mammography X-Ray Image Storage |

**MR-specific features:** Realistic parameters (EchoTime, RepetitionTime, FlipAngle), scanner models from Siemens, GE, and Philips (1.5T and 3.0T).

**CT-specific features:** Hounsfield units (RescaleIntercept=-1024), KVP, XRayTubeCurrent, ConvolutionKernel, scanner models with detector rows (64-320 rows).

**CR/DX-specific features:** ViewPosition, ImagerPixelSpacing, DistanceSourceToDetector, Exposure parameters.

**US-specific features:** TransducerType (LINEAR, CONVEX, PHASED), TransducerFrequency, 8-bit grayscale images.

**MG-specific features:** ImageLaterality (L/R), ViewPosition (CC, MLO), AnodeTargetMaterial, CompressionForce, high-resolution 14-bit images.

### Edge Case Types

When using `--edge-cases`, you can specify which types to enable with `--edge-case-types`:

| Type | Description |
|------|-------------|
| `special-chars` | Names with accents, hyphens, apostrophes (Müller-Schmidt, O'Connor, François) |
| `long-names` | Names at DICOM's 64-character limit |
| `old-dates` | Birth dates from 1900-1950, or partial dates (YYYY, YYYYMM) |
| `varied-ids` | Patient IDs with dashes, letters, spaces, or at max length |
| `missing-tags` | Omit optional DICOM tags (BodyPartExamined, StudyDescription, etc.) |

### Vendor Corruption (Robustness Testing)

The `--corrupt` flag injects vendor-specific private DICOM tags and malformed elements into **all** generated files, reproducing real-world scanner quirks that crash fragile DICOM readers. This is based on real corrupted files observed from Siemens scanners in production.

```bash
# Inject all corruption types
dicomforge --num-images 10 --total-size 10MB --corrupt all

# Inject only Siemens CSA private tags
dicomforge --num-images 10 --total-size 10MB --corrupt siemens-csa

# Combine multiple types
dicomforge --num-images 10 --total-size 10MB --corrupt siemens-csa,ge-private
```

| Type | Description |
|------|-------------|
| `siemens-csa` | Siemens CSA private tags: creator `(0029,0010)`, CSA Image Header `(0029,1010)` and Series Header `(0029,1020)` with realistic "SV10" binary format, crash-trigger private SQ `(0029,1102)` |
| `ge-private` | GE GEMS private tags: creators `(0009,0010)` + `(0043,0010)`, software version `(0009,10E3)`, multi-valued diffusion params `(0043,1039)` |
| `philips-private` | Philips private tags: creators `(2001,0010)` + `(2005,0010)`, nested private sequence `(2005,100E)` with scale/intercept data |
| `malformed-lengths` | Reproduces real dcmdump warnings: `(0070,0253)` FL with length not multiple of 4, `(7FE0,0010)` PixelData OW with odd byte count |
| `all` | Shorthand for all corruption types |

> **Note:** Unlike `--edge-cases` (percentage-based, per-patient), corruption applies to **all** generated files when enabled. The `--corrupt` and `--edge-cases` flags can be used together.

> **[See Examples Guide](docs/EXAMPLES.md#vendor-corruption-for-robustness-testing)** for detailed corruption examples and use cases.

### Examples

```bash
# Basic usage: 120 MR images, 1GB total
./dicomforge --num-images 120 --total-size 1GB

# Generate CT scan (100 slices)
./dicomforge --num-images 100 --total-size 200MB --modality CT

# Custom output directory with fixed seed for reproducibility
./dicomforge --num-images 50 --total-size 500MB --output patient_001 --seed 42

# Generate multiple studies (useful for testing study management)
./dicomforge --num-images 30 --total-size 500MB --num-studies 3

# Generate multiple patients with studies distributed among them
./dicomforge --num-images 60 --total-size 1GB --num-studies 6 --num-patients 2

# CT with specific body part
./dicomforge --num-images 100 --total-size 300MB --modality CT --body-part CHEST

# Limit parallelism (useful on resource-constrained systems)
./dicomforge --num-images 100 --total-size 1GB --workers 4

# Large dataset for stress testing
./dicomforge --num-images 500 --total-size 4GB --output stress_test

# Generate edge cases for robustness testing (25% of patients)
./dicomforge --num-images 100 --total-size 1GB --num-patients 20 \
  --edge-cases 25 --edge-case-types "special-chars,long-names"

# Generate all edge case types for comprehensive testing
./dicomforge --num-images 50 --total-size 500MB --num-studies 10 --num-patients 10 --edge-cases 50

# Generate corrupted DICOM files for robustness testing
./dicomforge --num-images 10 --total-size 10MB --corrupt all

# Generate with Siemens crash-trigger private tags
./dicomforge --num-images 10 --total-size 10MB --corrupt siemens-csa

# Combine corruption and edge cases
./dicomforge --num-images 20 --total-size 20MB --corrupt siemens-csa,ge-private --edge-cases 50
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
- **Multiple modalities**: Supports MR, CT, CR, DX, US, and MG with modality-specific parameters
- **DICOMDIR support**: Automatic directory index file creation
- **PT/ST/SE hierarchy**: Standard patient/study/series folder structure
- **Visual overlay**: Each image shows "File X/Y" text for easy verification
- **Parallel generation**: Worker pool for fast generation (~4.5x speedup)
- **Realistic metadata**: Simulated parameters from major vendors (Siemens, GE, Philips, Canon)
- **Realistic patient names**: Generated patient names (80% English, 20% French)
- **Edge case generation**: Special characters, long names, old dates, varied IDs for robustness testing
- **Vendor corruption**: Inject Siemens CSA, GE GEMS, Philips private tags and malformed elements for parser robustness testing
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
├── cmd/dicomforge/            # CLI entry point
├── internal/
│   ├── dicom/                 # DICOM generation and DICOMDIR
│   │   ├── corruption/        # Vendor-specific corruption tags
│   │   ├── edgecases/         # Edge case generation
│   │   └── modalities/        # Modality-specific generators (MR, CT, CR, DX, US, MG)
│   ├── image/                 # Pixel data generation
│   └── util/                  # Utilities (UID generation, size parsing)
├── tests/                     # Integration tests
│   └── e2e/                   # End-to-end tests (Gherkin/Cucumber)
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
