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
