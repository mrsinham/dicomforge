#!/usr/bin/env python3
"""
DICOM MRI Generator
Generate valid DICOM multi-frame MRI files for testing medical interfaces.
"""

import re
import math
import pydicom
from pydicom.dataset import Dataset, FileMetaDataset
from pydicom.uid import generate_uid, ExplicitVRLittleEndian
from datetime import datetime
import random


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
