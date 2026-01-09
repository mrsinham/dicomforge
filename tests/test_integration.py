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
