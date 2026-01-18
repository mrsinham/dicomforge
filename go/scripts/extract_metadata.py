#!/usr/bin/env python3
"""
Extract metadata from DICOM files for comparison between Python and Go implementations.
"""

import sys
import json
from pathlib import Path

try:
    import pydicom
except ImportError:
    print("Error: pydicom is not installed", file=sys.stderr)
    print("Install with: pip install pydicom", file=sys.stderr)
    sys.exit(1)


def extract_metadata(dicom_dir):
    """Extract key metadata from all DICOM files in a directory."""
    dicom_dir = Path(dicom_dir)

    # Find all DICOM files (including those without .dcm extension)
    dicom_files = []
    for pattern in ['**/*.dcm', '**/IM*', '**/IMG*.dcm']:
        dicom_files.extend(sorted(dicom_dir.glob(pattern)))

    # Remove duplicates and sort
    dicom_files = sorted(set(dicom_files))

    if not dicom_files:
        print(f"No DICOM files found in {dicom_dir}", file=sys.stderr)
        return None

    metadata = {
        'source_directory': str(dicom_dir),
        'file_count': len(dicom_files),
        'files': []
    }

    for filepath in dicom_files:
        try:
            ds = pydicom.dcmread(str(filepath))

            file_meta = {
                'filename': str(filepath.relative_to(dicom_dir)),
                'file_size': filepath.stat().st_size,
            }

            # Patient information
            if hasattr(ds, 'PatientID'):
                file_meta['patient_id'] = str(ds.PatientID)
            if hasattr(ds, 'PatientName'):
                file_meta['patient_name'] = str(ds.PatientName)
            if hasattr(ds, 'PatientBirthDate'):
                file_meta['patient_birth_date'] = str(ds.PatientBirthDate)
            if hasattr(ds, 'PatientSex'):
                file_meta['patient_sex'] = str(ds.PatientSex)

            # Study information
            if hasattr(ds, 'StudyInstanceUID'):
                file_meta['study_uid'] = str(ds.StudyInstanceUID)
            if hasattr(ds, 'StudyID'):
                file_meta['study_id'] = str(ds.StudyID)
            if hasattr(ds, 'StudyDescription'):
                file_meta['study_description'] = str(ds.StudyDescription)
            if hasattr(ds, 'StudyDate'):
                file_meta['study_date'] = str(ds.StudyDate)

            # Series information
            if hasattr(ds, 'SeriesInstanceUID'):
                file_meta['series_uid'] = str(ds.SeriesInstanceUID)
            if hasattr(ds, 'SeriesNumber'):
                file_meta['series_number'] = int(ds.SeriesNumber)
            if hasattr(ds, 'Modality'):
                file_meta['modality'] = str(ds.Modality)

            # Instance information
            if hasattr(ds, 'SOPInstanceUID'):
                file_meta['sop_instance_uid'] = str(ds.SOPInstanceUID)
            if hasattr(ds, 'InstanceNumber'):
                file_meta['instance_number'] = int(ds.InstanceNumber)

            # Image dimensions
            if hasattr(ds, 'Rows'):
                file_meta['rows'] = int(ds.Rows)
            if hasattr(ds, 'Columns'):
                file_meta['columns'] = int(ds.Columns)
            if hasattr(ds, 'BitsAllocated'):
                file_meta['bits_allocated'] = int(ds.BitsAllocated)

            # MRI parameters
            if hasattr(ds, 'Manufacturer'):
                file_meta['manufacturer'] = str(ds.Manufacturer)
            if hasattr(ds, 'ManufacturerModelName'):
                file_meta['model'] = str(ds.ManufacturerModelName)
            if hasattr(ds, 'MagneticFieldStrength'):
                file_meta['field_strength'] = float(ds.MagneticFieldStrength)
            if hasattr(ds, 'EchoTime'):
                file_meta['echo_time'] = float(ds.EchoTime)
            if hasattr(ds, 'RepetitionTime'):
                file_meta['repetition_time'] = float(ds.RepetitionTime)

            metadata['files'].append(file_meta)

        except Exception as e:
            print(f"Warning: Could not read {filepath}: {e}", file=sys.stderr)
            continue

    return metadata


def main():
    if len(sys.argv) < 2:
        print("Usage: extract_metadata.py <dicom_directory> [output_json]")
        print("\nExtracts metadata from all DICOM files in a directory")
        print("\nExample:")
        print("  python3 extract_metadata.py test-output")
        print("  python3 extract_metadata.py test-output metadata.json")
        sys.exit(1)

    dicom_dir = sys.argv[1]

    if not Path(dicom_dir).exists():
        print(f"Error: Directory '{dicom_dir}' does not exist", file=sys.stderr)
        sys.exit(1)

    print(f"Extracting metadata from: {dicom_dir}", file=sys.stderr)
    metadata = extract_metadata(dicom_dir)

    if metadata is None:
        sys.exit(1)

    print(f"Found {metadata['file_count']} DICOM files", file=sys.stderr)

    # Output to file or stdout
    if len(sys.argv) >= 3:
        output_file = sys.argv[2]
        with open(output_file, 'w') as f:
            json.dump(metadata, f, indent=2)
        print(f"Metadata written to: {output_file}", file=sys.stderr)
    else:
        print(json.dumps(metadata, indent=2))

    return 0


if __name__ == '__main__':
    sys.exit(main())
