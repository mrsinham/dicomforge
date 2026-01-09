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
import numpy as np
import argparse
import sys
import os
from pydicom.fileset import FileSet


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

    # DICOM limit: pixel data must be < 2^32 bytes (4,294,967,296)
    # The length field is 32-bit unsigned, so max is 2^32 - 1
    # Use a safe margin of 10MB below the limit
    MAX_PIXEL_DATA_SIZE = (2**32 - 1) - (10 * 1024 * 1024)  # ~4.28 GB with safety margin

    # If requested size exceeds DICOM limit, cap it
    if available_bytes > MAX_PIXEL_DATA_SIZE:
        available_bytes = MAX_PIXEL_DATA_SIZE
        print(f"Attention: Taille limitée à 4 GB (limite DICOM pour pixel data)")

    # Calculate pixels (2 bytes per pixel for uint16)
    bytes_per_pixel = 2
    total_pixels = available_bytes // bytes_per_pixel
    pixels_per_frame = total_pixels // num_images

    # Calculate square dimension
    dim = int(math.sqrt(pixels_per_frame))

    # Round DOWN to nearest multiple of 256 for realistic MRI dimensions
    # Important: must round down to ensure we don't exceed size limit
    if dim >= 256:
        dim = (dim // 256) * 256  # Floor division to round down
    elif dim >= 128:
        dim = (dim // 128) * 128

    # Ensure minimum size
    dim = max(dim, 128)

    return dim, dim


def generate_metadata(num_images, width, height, instance_number=None, study_uid=None, series_uid=None):
    """
    Generate DICOM dataset with realistic MRI metadata.

    Args:
        num_images: Number of frames (used for series info, but each file has 1 frame)
        width: Image width in pixels
        height: Image height in pixels
        instance_number: Instance number for this image (1-based)
        study_uid: Shared Study Instance UID (if None, generates new)
        series_uid: Shared Series Instance UID (if None, generates new)

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
    ds.StudyInstanceUID = study_uid if study_uid else generate_uid()
    now = datetime.now()
    ds.StudyDate = now.strftime('%Y%m%d')
    ds.StudyTime = now.strftime('%H%M%S')
    ds.StudyID = f"STD{random.randint(1000, 9999)}"
    ds.AccessionNumber = f"ACC{random.randint(100000, 999999)}"

    # Series information
    ds.SeriesInstanceUID = series_uid if series_uid else generate_uid()
    ds.SeriesNumber = 1
    ds.SeriesDescription = f"Test MRI Series - {num_images} images"
    ds.Modality = 'MR'

    # Instance number (position in series)
    if instance_number is not None:
        ds.InstanceNumber = instance_number

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

    # Image parameters (single frame per file)
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


def generate_pixel_data(num_images, width, height, seed=None):
    """
    Generate random pixel data for MRI images.

    Args:
        num_images: Number of frames to generate
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


def generate_single_image(width, height, seed=None):
    """
    Generate random pixel data for a single MRI image.

    Args:
        width: Image width
        height: Image height
        seed: Optional random seed for reproducibility

    Returns:
        numpy.ndarray: Array of shape (height, width) with dtype uint16
    """
    if seed is not None:
        np.random.seed(seed)

    # Generate random noise in 12-bit range (0-4095) - typical for MRI
    pixel_data = np.random.randint(0, 4096, size=(height, width), dtype=np.uint16)

    return pixel_data


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
        default='dicom_series',
        help='Nom du dossier de sortie (défaut: dicom_series)'
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

        print(f"Résolution: {width}x{height} pixels par image")
        print(f"Taille estimée: {format_bytes(estimated_size)} ({args.num_images} images)")

        # Create output directory
        output_dir = args.output
        if os.path.exists(output_dir):
            print(f"Attention: Le dossier {output_dir} existe déjà")
        else:
            os.makedirs(output_dir)
            print(f"Création du dossier: {output_dir}")

        # Generate shared UIDs for the series
        study_uid = generate_uid()
        series_uid = generate_uid()

        print(f"Génération de {args.num_images} fichiers DICOM...")

        # Set seed for reproducibility if specified
        if args.seed is not None:
            np.random.seed(args.seed)
            random.seed(args.seed)

        total_size = 0

        # Generate each DICOM file
        for i in range(1, args.num_images + 1):
            # Generate metadata for this instance
            ds = generate_metadata(
                num_images=args.num_images,
                width=width,
                height=height,
                instance_number=i,
                study_uid=study_uid,
                series_uid=series_uid
            )

            # Generate pixel data for this single image
            pixel_data = generate_single_image(width, height)

            # Add pixel data to dataset
            ds.PixelData = pixel_data.tobytes()

            # Write DICOM file
            filename = f"IMG{i:04d}.dcm"
            filepath = os.path.join(output_dir, filename)
            ds.save_as(filepath, write_like_original=False)

            total_size += os.path.getsize(filepath)

            # Progress indicator
            if i % 10 == 0 or i == args.num_images:
                progress = (i / args.num_images) * 100
                print(f"  Progression: {i}/{args.num_images} ({progress:.0f}%)")

        print(f"\n✓ {args.num_images} fichiers DICOM créés dans: {output_dir}/")
        print(f"  Taille totale: {format_bytes(total_size)}")

        # Create DICOMDIR file
        print("\nCréation du fichier DICOMDIR...")
        try:
            fs = FileSet()

            # Add all DICOM files to the fileset
            for i in range(1, args.num_images + 1):
                filename = f"IMG{i:04d}.dcm"
                filepath = os.path.join(output_dir, filename)
                fs.add(filepath)

            # Write DICOMDIR
            dicomdir_path = os.path.join(output_dir, "DICOMDIR")
            fs.write(dicomdir_path)

            print(f"✓ DICOMDIR créé: {dicomdir_path}")
            print(f"\nLa série DICOM est prête à être importée!")
            print(f"Importez le dossier complet: {output_dir}/")

        except Exception as e:
            print(f"Attention: Erreur lors de la création du DICOMDIR: {e}")
            print(f"Les fichiers DICOM sont valides, mais le DICOMDIR n'a pas pu être créé.")

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
