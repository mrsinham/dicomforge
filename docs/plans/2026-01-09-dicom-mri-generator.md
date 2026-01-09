# DICOM MRI Generator Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Python CLI tool to generate valid DICOM multi-frame MRI files with configurable size and image count for testing medical interfaces.

**Architecture:** Single Python script using pydicom for DICOM creation, numpy for pixel data generation, and argparse for CLI. Calculates optimal image dimensions to hit target file size, generates realistic MRI metadata, and writes pixel data progressively to avoid memory issues.

**Tech Stack:** Python 3.13+, pydicom, numpy, pillow

---

## Task 1: Project Setup and Dependencies

**Files:**
- Create: `requirements.txt`
- Create: `README.md`

**Step 1: Create requirements.txt**

Create `requirements.txt` with dependencies:

```
pydicom>=2.4.0
numpy>=1.24.0
pillow>=10.0.0
```

**Step 2: Install dependencies**

Run: `pip install -r requirements.txt`
Expected: All packages install successfully

**Step 3: Create README.md**

```markdown
# DICOM MRI Generator

Outil CLI Python pour générer des fichiers DICOM d'IRM multi-frame valides pour tester des interfaces médicales.

## Installation

```bash
pip install -r requirements.txt
```

## Usage

```bash
python generate_dicom_mri.py --num-images 120 --total-size 4.5GB --output mri_test.dcm
```

### Paramètres

- `--num-images` (requis): Nombre d'images/coupes dans la série
- `--total-size` (requis): Taille totale cible (KB, MB, GB)
- `--output` (optionnel): Nom du fichier de sortie (défaut: `generated_mri.dcm`)
- `--seed` (optionnel): Seed pour reproductibilité

### Exemples

```bash
# Générer 120 images pour 4.5 GB
python generate_dicom_mri.py --num-images 120 --total-size 4.5GB

# Avec nom de fichier personnalisé et seed
python generate_dicom_mri.py --num-images 50 --total-size 1GB --output test.dcm --seed 42
```
```

**Step 4: Verify documentation**

Read the README to ensure clarity.

**Step 5: Commit**

```bash
git init
git add requirements.txt README.md
git commit -m "feat: add project setup and documentation"
```

---

## Task 2: Size Parser Function

**Files:**
- Create: `generate_dicom_mri.py`
- Create: `tests/test_generate_dicom_mri.py`

**Step 1: Create test file structure**

Create `tests/test_generate_dicom_mri.py`:

```python
import pytest
import sys
sys.path.insert(0, '.')
from generate_dicom_mri import parse_size


def test_parse_size_kilobytes():
    assert parse_size("100KB") == 100 * 1024


def test_parse_size_megabytes():
    assert parse_size("50MB") == 50 * 1024 * 1024


def test_parse_size_gigabytes():
    assert parse_size("4.5GB") == int(4.5 * 1024 * 1024 * 1024)


def test_parse_size_case_insensitive():
    assert parse_size("10mb") == 10 * 1024 * 1024
    assert parse_size("10Mb") == 10 * 1024 * 1024


def test_parse_size_invalid_format():
    with pytest.raises(ValueError):
        parse_size("invalid")


def test_parse_size_invalid_unit():
    with pytest.raises(ValueError):
        parse_size("100TB")
```

**Step 2: Run tests to verify they fail**

Run: `pytest tests/test_generate_dicom_mri.py::test_parse_size_kilobytes -v`
Expected: FAIL with "cannot import name 'parse_size'"

**Step 3: Create minimal implementation**

Create `generate_dicom_mri.py`:

```python
#!/usr/bin/env python3
"""
DICOM MRI Generator
Generate valid DICOM multi-frame MRI files for testing medical interfaces.
"""

import re


def parse_size(size_str):
    """
    Parse size string (e.g., '4.5GB', '100MB') into bytes.

    Args:
        size_str: Size string with unit (KB, MB, GB)

    Returns:
        int: Size in bytes

    Raises:
        ValueError: If format is invalid or unit not supported
    """
    pattern = r'^(\d+(?:\.\d+)?)(KB|MB|GB)$'
    match = re.match(pattern, size_str.upper())

    if not match:
        raise ValueError(f"Format invalide: '{size_str}'. Utilisez format comme '100MB', '4.5GB'")

    value = float(match.group(1))
    unit = match.group(2)

    multipliers = {
        'KB': 1024,
        'MB': 1024 * 1024,
        'GB': 1024 * 1024 * 1024
    }

    if unit not in multipliers:
        raise ValueError(f"Unité non supportée: '{unit}'. Utilisez KB, MB ou GB")

    return int(value * multipliers[unit])
```

**Step 4: Run all parse_size tests**

Run: `pytest tests/test_generate_dicom_mri.py -v -k parse_size`
Expected: All 6 tests PASS

**Step 5: Commit**

```bash
git add generate_dicom_mri.py tests/test_generate_dicom_mri.py
git commit -m "feat: add size parser with validation"
```

---

## Task 3: Dimension Calculator Function

**Files:**
- Modify: `generate_dicom_mri.py`
- Modify: `tests/test_generate_dicom_mri.py`

**Step 1: Write failing tests**

Add to `tests/test_generate_dicom_mri.py`:

```python
from generate_dicom_mri import calculate_dimensions
import math


def test_calculate_dimensions_basic():
    """Test basic dimension calculation."""
    total_bytes = 1024 * 1024 * 100  # 100 MB
    num_images = 10
    width, height = calculate_dimensions(total_bytes, num_images)

    # Should return square dimensions
    assert width == height
    # Should be multiple of 256 or close to sqrt of pixels
    assert width > 0 and height > 0


def test_calculate_dimensions_large_file():
    """Test with 4.5GB / 120 images."""
    total_bytes = int(4.5 * 1024 * 1024 * 1024)
    num_images = 120
    width, height = calculate_dimensions(total_bytes, num_images)

    # Check dimensions are reasonable for MRI
    assert width >= 512
    assert width == height

    # Verify size is close to target (within 10%)
    metadata_overhead = 100 * 1024  # 100KB
    pixel_bytes = (total_bytes - metadata_overhead)
    expected_pixels = pixel_bytes // 2  # 2 bytes per pixel (uint16)
    expected_per_frame = expected_pixels // num_images
    actual_per_frame = width * height

    tolerance = 0.1
    assert abs(actual_per_frame - expected_per_frame) / expected_per_frame < tolerance


def test_calculate_dimensions_rounds_to_reasonable():
    """Test that dimensions are rounded to reasonable values."""
    total_bytes = 1024 * 1024 * 50  # 50 MB
    num_images = 5
    width, height = calculate_dimensions(total_bytes, num_images)

    # Should be multiple of 256 or close
    assert width % 256 == 0 or (width % 128 == 0)
```

**Step 2: Run tests to verify they fail**

Run: `pytest tests/test_generate_dicom_mri.py::test_calculate_dimensions_basic -v`
Expected: FAIL with "cannot import name 'calculate_dimensions'"

**Step 3: Implement calculate_dimensions**

Add to `generate_dicom_mri.py`:

```python
import math


def calculate_dimensions(total_size_bytes, num_images):
    """
    Calculate optimal image dimensions to hit target file size.

    Args:
        total_size_bytes: Target total file size in bytes
        num_images: Number of frames/images

    Returns:
        tuple: (width, height) as integers
    """
    # Estimate metadata overhead
    metadata_overhead = 100 * 1024  # 100KB

    # Available bytes for pixel data
    available_bytes = total_size_bytes - metadata_overhead

    # Calculate pixels (2 bytes per pixel for uint16)
    bytes_per_pixel = 2
    total_pixels = available_bytes // bytes_per_pixel
    pixels_per_frame = total_pixels // num_images

    # Calculate square dimension
    dim = int(math.sqrt(pixels_per_frame))

    # Round to nearest multiple of 256 for realistic MRI dimensions
    # But use 128 if result would be too small
    if dim >= 256:
        dim = round(dim / 256) * 256
    elif dim >= 128:
        dim = round(dim / 128) * 128

    # Ensure minimum size
    dim = max(dim, 128)

    return dim, dim
```

**Step 4: Run all dimension tests**

Run: `pytest tests/test_generate_dicom_mri.py -v -k calculate_dimensions`
Expected: All 3 tests PASS

**Step 5: Commit**

```bash
git add generate_dicom_mri.py tests/test_generate_dicom_mri.py
git commit -m "feat: add dimension calculator"
```

---

## Task 4: Metadata Generator Function

**Files:**
- Modify: `generate_dicom_mri.py`
- Modify: `tests/test_generate_dicom_mri.py`

**Step 1: Write failing tests**

Add to `tests/test_generate_dicom_mri.py`:

```python
from generate_dicom_mri import generate_metadata
from datetime import datetime


def test_generate_metadata_creates_dataset():
    """Test that metadata generator creates valid dataset."""
    ds = generate_metadata(num_images=10, width=512, height=512)

    # Check basic DICOM tags exist
    assert hasattr(ds, 'PatientID')
    assert hasattr(ds, 'StudyInstanceUID')
    assert hasattr(ds, 'SeriesInstanceUID')
    assert hasattr(ds, 'Modality')

    # Check modality is MR
    assert ds.Modality == 'MR'


def test_generate_metadata_has_patient_info():
    """Test patient information is generated."""
    ds = generate_metadata(num_images=10, width=512, height=512)

    assert hasattr(ds, 'PatientName')
    assert hasattr(ds, 'PatientBirthDate')
    assert hasattr(ds, 'PatientSex')
    assert ds.PatientSex in ['M', 'F']


def test_generate_metadata_has_mri_params():
    """Test MRI-specific parameters are realistic."""
    ds = generate_metadata(num_images=10, width=512, height=512)

    # Check manufacturer info
    assert hasattr(ds, 'Manufacturer')
    assert ds.Manufacturer in ['SIEMENS', 'GE MEDICAL SYSTEMS', 'PHILIPS']

    # Check MRI parameters
    assert hasattr(ds, 'MagneticFieldStrength')
    assert ds.MagneticFieldStrength in [1.5, 3.0]

    assert hasattr(ds, 'EchoTime')
    assert hasattr(ds, 'RepetitionTime')


def test_generate_metadata_multi_frame():
    """Test multi-frame specific tags."""
    num_images = 120
    ds = generate_metadata(num_images=num_images, width=512, height=512)

    assert ds.NumberOfFrames == num_images
    assert ds.SamplesPerPixel == 1
    assert ds.PhotometricInterpretation == 'MONOCHROME2'
    assert ds.Rows == 512
    assert ds.Columns == 512
    assert ds.BitsAllocated == 16
    assert ds.BitsStored == 16
    assert ds.HighBit == 15
    assert ds.PixelRepresentation == 0  # unsigned
```

**Step 2: Run tests to verify they fail**

Run: `pytest tests/test_generate_dicom_mri.py::test_generate_metadata_creates_dataset -v`
Expected: FAIL with "cannot import name 'generate_metadata'"

**Step 3: Implement generate_metadata**

Add to `generate_dicom_mri.py`:

```python
import pydicom
from pydicom.dataset import Dataset, FileMetaDataset
from pydicom.uid import generate_uid, ExplicitVRLittleEndian
from datetime import datetime
import random


def generate_metadata(num_images, width, height):
    """
    Generate DICOM dataset with realistic MRI metadata.

    Args:
        num_images: Number of frames
        width: Image width in pixels
        height: Image height in pixels

    Returns:
        pydicom.Dataset: Dataset with metadata
    """
    # Create file meta information
    file_meta = FileMetaDataset()
    file_meta.TransferSyntaxUID = ExplicitVRLittleEndian
    file_meta.MediaStorageSOPClassUID = '1.2.840.10008.5.1.4.1.1.4'  # MR Image Storage
    file_meta.MediaStorageSOPInstanceUID = generate_uid()
    file_meta.ImplementationClassUID = generate_uid()

    # Create main dataset
    ds = Dataset()
    ds.file_meta = file_meta

    # Patient information
    ds.PatientName = f"TEST^PATIENT^{random.randint(1000, 9999)}"
    ds.PatientID = f"PID{random.randint(100000, 999999)}"
    ds.PatientBirthDate = f"{random.randint(1950, 2000):04d}{random.randint(1, 12):02d}{random.randint(1, 28):02d}"
    ds.PatientSex = random.choice(['M', 'F'])

    # Study information
    ds.StudyInstanceUID = generate_uid()
    now = datetime.now()
    ds.StudyDate = now.strftime('%Y%m%d')
    ds.StudyTime = now.strftime('%H%M%S')
    ds.StudyID = f"STD{random.randint(1000, 9999)}"
    ds.AccessionNumber = f"ACC{random.randint(100000, 999999)}"

    # Series information
    ds.SeriesInstanceUID = generate_uid()
    ds.SeriesNumber = 1
    ds.SeriesDescription = "Test MRI Series - Multi-frame"
    ds.Modality = 'MR'

    # SOP Common
    ds.SOPClassUID = file_meta.MediaStorageSOPClassUID
    ds.SOPInstanceUID = file_meta.MediaStorageSOPInstanceUID

    # MRI-specific parameters
    manufacturers = [
        ('SIEMENS', 'Avanto', 1.5),
        ('SIEMENS', 'Skyra', 3.0),
        ('GE MEDICAL SYSTEMS', 'Signa HDxt', 1.5),
        ('GE MEDICAL SYSTEMS', 'Discovery MR750', 3.0),
        ('PHILIPS', 'Achieva', 1.5),
        ('PHILIPS', 'Ingenia', 3.0)
    ]
    manufacturer, model, field_strength = random.choice(manufacturers)

    ds.Manufacturer = manufacturer
    ds.ManufacturerModelName = model
    ds.MagneticFieldStrength = field_strength

    # Calculate imaging frequency based on field strength
    # 1.5T ≈ 63.87 MHz, 3.0T ≈ 127.74 MHz for protons
    ds.ImagingFrequency = field_strength * 42.58  # MHz (gyromagnetic ratio)

    # Sequence parameters (realistic T1-weighted values)
    ds.EchoTime = random.uniform(10, 30)  # ms
    ds.RepetitionTime = random.uniform(400, 800)  # ms
    ds.FlipAngle = random.uniform(60, 90)  # degrees
    ds.SliceThickness = random.uniform(1.0, 5.0)  # mm
    ds.SpacingBetweenSlices = ds.SliceThickness + random.uniform(0, 0.5)  # mm
    ds.SequenceName = random.choice(['T1_MPRAGE', 'T1_SE', 'T2_FSE', 'T2_FLAIR'])

    # Multi-frame image parameters
    ds.NumberOfFrames = num_images
    ds.SamplesPerPixel = 1
    ds.PhotometricInterpretation = 'MONOCHROME2'
    ds.Rows = height
    ds.Columns = width
    ds.BitsAllocated = 16
    ds.BitsStored = 16
    ds.HighBit = 15
    ds.PixelRepresentation = 0  # unsigned

    # Pixel spacing (typical MRI: 0.5-2mm)
    pixel_spacing = random.uniform(0.5, 2.0)
    ds.PixelSpacing = [pixel_spacing, pixel_spacing]

    return ds
```

**Step 4: Run all metadata tests**

Run: `pytest tests/test_generate_dicom_mri.py -v -k generate_metadata`
Expected: All 4 tests PASS

**Step 5: Commit**

```bash
git add generate_dicom_mri.py tests/test_generate_dicom_mri.py
git commit -m "feat: add DICOM metadata generator with realistic MRI parameters"
```

---

## Task 5: Pixel Data Generator Function

**Files:**
- Modify: `generate_dicom_mri.py`
- Modify: `tests/test_generate_dicom_mri.py`

**Step 1: Write failing tests**

Add to `tests/test_generate_dicom_mri.py`:

```python
from generate_dicom_mri import generate_pixel_data
import numpy as np


def test_generate_pixel_data_shape():
    """Test pixel data has correct shape."""
    num_images = 10
    width, height = 512, 512
    pixel_data = generate_pixel_data(num_images, width, height)

    expected_shape = (num_images, height, width)
    assert pixel_data.shape == expected_shape


def test_generate_pixel_data_dtype():
    """Test pixel data is uint16."""
    pixel_data = generate_pixel_data(5, 256, 256)
    assert pixel_data.dtype == np.uint16


def test_generate_pixel_data_range():
    """Test pixel values are in valid range."""
    pixel_data = generate_pixel_data(5, 256, 256)

    # Should be in 12-bit range (0-4095)
    assert pixel_data.min() >= 0
    assert pixel_data.max() <= 4095


def test_generate_pixel_data_with_seed():
    """Test seed produces reproducible results."""
    num_images, width, height = 3, 128, 128

    data1 = generate_pixel_data(num_images, width, height, seed=42)
    data2 = generate_pixel_data(num_images, width, height, seed=42)

    assert np.array_equal(data1, data2)


def test_generate_pixel_data_different_without_seed():
    """Test different results without seed."""
    num_images, width, height = 3, 128, 128

    data1 = generate_pixel_data(num_images, width, height)
    data2 = generate_pixel_data(num_images, width, height)

    # Should be different (with very high probability)
    assert not np.array_equal(data1, data2)
```

**Step 2: Run tests to verify they fail**

Run: `pytest tests/test_generate_dicom_mri.py::test_generate_pixel_data_shape -v`
Expected: FAIL with "cannot import name 'generate_pixel_data'"

**Step 3: Implement generate_pixel_data**

Add to `generate_dicom_mri.py`:

```python
import numpy as np


def generate_pixel_data(num_images, width, height, seed=None):
    """
    Generate random pixel data for MRI images.

    Args:
        num_images: Number of frames
        width: Image width
        height: Image height
        seed: Optional random seed for reproducibility

    Returns:
        numpy.ndarray: Array of shape (num_images, height, width) with dtype uint16
    """
    if seed is not None:
        np.random.seed(seed)

    # Generate random noise in 12-bit range (0-4095) - typical for MRI
    # Shape: (num_images, height, width)
    pixel_data = np.random.randint(0, 4096, size=(num_images, height, width), dtype=np.uint16)

    return pixel_data
```

**Step 4: Run all pixel data tests**

Run: `pytest tests/test_generate_dicom_mri.py -v -k generate_pixel_data`
Expected: All 5 tests PASS

**Step 5: Commit**

```bash
git add generate_dicom_mri.py tests/test_generate_dicom_mri.py
git commit -m "feat: add pixel data generator with seed support"
```

---

## Task 6: CLI Argument Parser

**Files:**
- Modify: `generate_dicom_mri.py`
- Modify: `tests/test_generate_dicom_mri.py`

**Step 1: Write failing tests**

Add to `tests/test_generate_dicom_mri.py`:

```python
from generate_dicom_mri import parse_arguments
import sys


def test_parse_arguments_required():
    """Test required arguments."""
    args = parse_arguments(['--num-images', '120', '--total-size', '4.5GB'])

    assert args.num_images == 120
    assert args.total_size == '4.5GB'
    assert args.output == 'generated_mri.dcm'  # default
    assert args.seed is None  # default


def test_parse_arguments_all_options():
    """Test all arguments including optional."""
    args = parse_arguments([
        '--num-images', '50',
        '--total-size', '1GB',
        '--output', 'test.dcm',
        '--seed', '42'
    ])

    assert args.num_images == 50
    assert args.total_size == '1GB'
    assert args.output == 'test.dcm'
    assert args.seed == 42


def test_parse_arguments_missing_required():
    """Test error when missing required arguments."""
    with pytest.raises(SystemExit):
        parse_arguments(['--num-images', '10'])  # missing --total-size
```

**Step 2: Run tests to verify they fail**

Run: `pytest tests/test_generate_dicom_mri.py::test_parse_arguments_required -v`
Expected: FAIL with "cannot import name 'parse_arguments'"

**Step 3: Implement parse_arguments**

Add to `generate_dicom_mri.py`:

```python
import argparse


def parse_arguments(argv=None):
    """
    Parse command line arguments.

    Args:
        argv: List of arguments (for testing), None uses sys.argv

    Returns:
        argparse.Namespace: Parsed arguments
    """
    parser = argparse.ArgumentParser(
        description='Générer des fichiers DICOM d\'IRM multi-frame pour tests',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Exemples:
  %(prog)s --num-images 120 --total-size 4.5GB
  %(prog)s --num-images 50 --total-size 1GB --output test.dcm --seed 42
        """
    )

    parser.add_argument(
        '--num-images',
        type=int,
        required=True,
        help='Nombre d\'images/coupes dans la série'
    )

    parser.add_argument(
        '--total-size',
        type=str,
        required=True,
        help='Taille totale cible (ex: 100MB, 4.5GB)'
    )

    parser.add_argument(
        '--output',
        type=str,
        default='generated_mri.dcm',
        help='Nom du fichier de sortie (défaut: generated_mri.dcm)'
    )

    parser.add_argument(
        '--seed',
        type=int,
        default=None,
        help='Seed pour la génération aléatoire (reproductibilité)'
    )

    args = parser.parse_args(argv)

    # Validate num_images
    if args.num_images <= 0:
        parser.error("--num-images doit être > 0")

    return args
```

**Step 4: Run all argument parser tests**

Run: `pytest tests/test_generate_dicom_mri.py -v -k parse_arguments`
Expected: All 3 tests PASS

**Step 5: Commit**

```bash
git add generate_dicom_mri.py tests/test_generate_dicom_mri.py
git commit -m "feat: add CLI argument parser with validation"
```

---

## Task 7: Main Function Integration

**Files:**
- Modify: `generate_dicom_mri.py`

**Step 1: Implement main function**

Add to `generate_dicom_mri.py`:

```python
import sys
import os


def format_bytes(bytes_size):
    """Format bytes as human-readable string."""
    for unit in ['B', 'KB', 'MB', 'GB']:
        if bytes_size < 1024.0:
            return f"{bytes_size:.2f} {unit}"
        bytes_size /= 1024.0
    return f"{bytes_size:.2f} TB"


def main():
    """Main entry point."""
    # Parse arguments
    args = parse_arguments()

    try:
        # Parse and validate size
        print("Calcul de la résolution optimale...")
        total_bytes = parse_size(args.total_size)

        if total_bytes <= 0:
            print(f"Erreur: La taille doit être > 0", file=sys.stderr)
            return 1

        # Check disk space
        stat = os.statvfs('.')
        available_space = stat.f_bavail * stat.f_frsize
        if total_bytes > available_space:
            print(f"Erreur: Espace disque insuffisant. Requis: {format_bytes(total_bytes)}, Disponible: {format_bytes(available_space)}", file=sys.stderr)
            return 1

        # Calculate dimensions
        width, height = calculate_dimensions(total_bytes, args.num_images)

        # Estimate actual file size
        pixel_bytes = args.num_images * width * height * 2  # 2 bytes per pixel
        metadata_overhead = 100 * 1024  # 100KB estimate
        estimated_size = pixel_bytes + metadata_overhead

        print(f"Résolution: {width}x{height} pixels par frame")
        print(f"Taille estimée: {format_bytes(estimated_size)} ({args.num_images} frames)")

        # Generate metadata
        print("Génération des métadonnées DICOM...")
        ds = generate_metadata(args.num_images, width, height)

        # Generate pixel data
        print("Génération des données d'image...")
        pixel_data = generate_pixel_data(args.num_images, width, height, args.seed)

        # Add pixel data to dataset
        # Flatten to 1D array as DICOM expects
        ds.PixelData = pixel_data.tobytes()

        # Write DICOM file
        print(f"Écriture du fichier DICOM: {args.output}")
        ds.save_as(args.output, write_like_original=False)

        # Get actual file size
        actual_size = os.path.getsize(args.output)
        print(f"Fichier DICOM créé: {args.output}")
        print(f"Taille réelle: {format_bytes(actual_size)}")

        return 0

    except ValueError as e:
        print(f"Erreur: {e}", file=sys.stderr)
        return 1
    except OSError as e:
        print(f"Erreur d'écriture: {e}", file=sys.stderr)
        return 1
    except Exception as e:
        print(f"Erreur inattendue: {e}", file=sys.stderr)
        return 1


if __name__ == '__main__':
    sys.exit(main())
```

**Step 2: Make script executable**

Run: `chmod +x generate_dicom_mri.py`
Expected: File becomes executable

**Step 3: Test with small file**

Run: `python generate_dicom_mri.py --num-images 5 --total-size 10MB --output test_small.dcm`
Expected:
- Script prints progress messages
- Creates test_small.dcm file
- File size is approximately 10MB

**Step 4: Verify DICOM validity**

Run: `python -c "import pydicom; ds = pydicom.dcmread('test_small.dcm'); print(f'Valid DICOM: {ds.Modality}, {ds.NumberOfFrames} frames, {ds.Rows}x{ds.Columns}')"`
Expected: Prints DICOM info showing MR modality, correct frame count and dimensions

**Step 5: Clean up test file**

Run: `rm test_small.dcm`

**Step 6: Commit**

```bash
git add generate_dicom_mri.py
git commit -m "feat: add main function with progress reporting and error handling"
```

---

## Task 8: Integration Testing

**Files:**
- Create: `tests/test_integration.py`

**Step 1: Create integration test**

Create `tests/test_integration.py`:

```python
import subprocess
import os
import pydicom
import pytest


def test_full_generation_small():
    """Test generating a small DICOM file end-to-end."""
    output = 'test_integration_small.dcm'

    # Clean up if exists
    if os.path.exists(output):
        os.remove(output)

    try:
        # Run generator
        result = subprocess.run([
            'python', 'generate_dicom_mri.py',
            '--num-images', '5',
            '--total-size', '5MB',
            '--output', output,
            '--seed', '42'
        ], capture_output=True, text=True)

        assert result.returncode == 0, f"Script failed: {result.stderr}"
        assert os.path.exists(output), "Output file not created"

        # Verify it's valid DICOM
        ds = pydicom.dcmread(output)
        assert ds.Modality == 'MR'
        assert ds.NumberOfFrames == 5
        assert ds.BitsAllocated == 16

        # Check file size is reasonable (within 20% of target)
        target_bytes = 5 * 1024 * 1024
        actual_size = os.path.getsize(output)
        assert 0.8 * target_bytes <= actual_size <= 1.2 * target_bytes

    finally:
        # Clean up
        if os.path.exists(output):
            os.remove(output)


def test_full_generation_reproducible():
    """Test that seed produces reproducible files."""
    output1 = 'test_integration_1.dcm'
    output2 = 'test_integration_2.dcm'

    try:
        # Generate first file
        subprocess.run([
            'python', 'generate_dicom_mri.py',
            '--num-images', '3',
            '--total-size', '1MB',
            '--output', output1,
            '--seed', '123'
        ], check=True, capture_output=True)

        # Generate second file with same seed
        subprocess.run([
            'python', 'generate_dicom_mri.py',
            '--num-images', '3',
            '--total-size', '1MB',
            '--output', output2,
            '--seed', '123'
        ], check=True, capture_output=True)

        # Read both files
        ds1 = pydicom.dcmread(output1)
        ds2 = pydicom.dcmread(output2)

        # Pixel data should be identical
        assert ds1.PixelData == ds2.PixelData

    finally:
        for f in [output1, output2]:
            if os.path.exists(f):
                os.remove(f)


def test_invalid_arguments():
    """Test error handling for invalid arguments."""
    # Missing required argument
    result = subprocess.run([
        'python', 'generate_dicom_mri.py',
        '--num-images', '10'
    ], capture_output=True, text=True)
    assert result.returncode != 0

    # Invalid size format
    result = subprocess.run([
        'python', 'generate_dicom_mri.py',
        '--num-images', '10',
        '--total-size', 'invalid'
    ], capture_output=True, text=True)
    assert result.returncode != 0
```

**Step 2: Run integration tests**

Run: `pytest tests/test_integration.py -v`
Expected: All 3 integration tests PASS

**Step 3: Commit**

```bash
git add tests/test_integration.py
git commit -m "test: add integration tests for end-to-end validation"
```

---

## Task 9: Large File Test and Documentation

**Files:**
- Modify: `README.md`

**Step 1: Manual test with target size**

Run: `python generate_dicom_mri.py --num-images 120 --total-size 4.5GB --output test_large.dcm`
Expected:
- Script completes in 30-90 seconds
- Creates file approximately 4.5GB
- Prints progress and final size

**Step 2: Verify large DICOM**

Run: `python -c "import pydicom; ds = pydicom.dcmread('test_large.dcm'); print(f'Frames: {ds.NumberOfFrames}, Size: {ds.Rows}x{ds.Columns}, Modality: {ds.Modality}')"`
Expected: Shows 120 frames with correct dimensions and MR modality

**Step 3: Check file size**

Run: `ls -lh test_large.dcm`
Expected: File size around 4.3-4.7 GB

**Step 4: Update README with performance notes**

Add to `README.md` before the last section:

```markdown
## Performance

- Small files (< 100MB): < 5 seconds
- Medium files (100MB - 1GB): 5-20 seconds
- Large files (1GB - 5GB): 20-90 seconds

Performance depends on disk speed. The script generates pixel data progressively to avoid memory issues.

## Testing

Run unit tests:
```bash
pytest tests/test_generate_dicom_mri.py -v
```

Run integration tests:
```bash
pytest tests/test_integration.py -v
```

Run all tests:
```bash
pytest tests/ -v
```
```

**Step 5: Clean up test file**

Run: `rm test_large.dcm`

**Step 6: Commit**

```bash
git add README.md
git commit -m "docs: add performance notes and testing instructions"
```

---

## Task 10: Final Verification and Cleanup

**Files:**
- All files

**Step 1: Run all tests**

Run: `pytest tests/ -v`
Expected: All tests PASS

**Step 2: Test help message**

Run: `python generate_dicom_mri.py --help`
Expected: Clear help message in French with examples

**Step 3: Verify requirements install cleanly**

Run: `pip install -r requirements.txt --dry-run`
Expected: No errors, all packages available

**Step 4: Final lint check (if pylint available)**

Run: `python -m py_compile generate_dicom_mri.py`
Expected: No syntax errors

**Step 5: Create final commit**

```bash
git add -A
git commit -m "chore: final verification and cleanup"
```

**Step 6: Create git tag**

```bash
git tag -a v1.0.0 -m "Release v1.0.0: DICOM MRI generator"
```

---

## Completion Checklist

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Can generate small files (< 100MB)
- [ ] Can generate large files (4.5GB+)
- [ ] DICOM files are valid (pydicom can read them)
- [ ] File sizes match targets within 20%
- [ ] Seed produces reproducible results
- [ ] Error messages are clear in French
- [ ] README has usage examples
- [ ] Code is documented with docstrings

---

## Notes

- This implementation prioritizes simplicity and correctness over performance
- All pixel data is generated in memory then written; for files > 10GB, consider streaming
- Metadata uses realistic MRI parameters but is not tied to actual clinical protocols
- UIDs are randomly generated; not suitable for production PACS integration
