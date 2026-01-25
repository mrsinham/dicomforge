# dicomforge Examples

This guide provides detailed examples for all features of dicomforge. Each section explains a specific capability with practical use cases.

## Table of Contents

- [Basic Usage](#basic-usage)
- [Modalities](#modalities)
- [Multi-Studies and Multi-Patients](#multi-studies-and-multi-patients)
- [Multi-Series per Study](#multi-series-per-study)
- [Custom Study Descriptions](#custom-study-descriptions)
- [Custom DICOM Tags](#custom-dicom-tags)
- [Categorization Options](#categorization-options)
- [Edge Cases for Robustness Testing](#edge-cases-for-robustness-testing)
- [Reproducibility](#reproducibility)
- [Performance Tuning](#performance-tuning)
- [Real-World Scenarios](#real-world-scenarios)

---

## Basic Usage

Generate a simple DICOM series with the two required arguments:

```bash
# Generate 10 MR images totaling 100MB
dicomforge --num-images 10 --total-size 100MB
```

**What it does:**
- Creates 10 DICOM image files
- Each file is approximately 10MB (100MB / 10 images)
- Uses MR (Magnetic Resonance) as the default modality
- Output goes to `dicom_series/` directory
- Creates DICOMDIR index and PT/ST/SE hierarchy

```bash
# Specify output directory
dicomforge --num-images 50 --total-size 500MB --output my_test_data
```

**Size formats supported:**
- `KB` - Kilobytes (e.g., `500KB`)
- `MB` - Megabytes (e.g., `100MB`)
- `GB` - Gigabytes (e.g., `4.5GB`)

---

## Modalities

dicomforge supports 6 imaging modalities, each with specific DICOM tags and realistic parameters.

### MR - Magnetic Resonance Imaging

```bash
dicomforge --num-images 120 --total-size 1GB --modality MR --output mri_brain
```

**MR-specific features:**
- Scanner models from Siemens, GE, Philips (1.5T and 3.0T)
- Realistic parameters: EchoTime, RepetitionTime, FlipAngle
- SOP Class: MR Image Storage

### CT - Computed Tomography

```bash
dicomforge --num-images 200 --total-size 400MB --modality CT --output ct_chest
```

**CT-specific features:**
- Hounsfield units (RescaleIntercept=-1024, RescaleType=HU)
- KVP, XRayTubeCurrent, ConvolutionKernel
- Scanner models with 64-320 detector rows
- SOP Class: CT Image Storage

### CR - Computed Radiography

```bash
dicomforge --num-images 5 --total-size 50MB --modality CR --output cr_series
```

**CR-specific features:**
- ViewPosition, ImagerPixelSpacing
- DistanceSourceToDetector, Exposure parameters
- SOP Class: Computed Radiography Image Storage

### DX - Digital X-Ray

```bash
dicomforge --num-images 2 --total-size 30MB --modality DX --body-part CHEST --output chest_xray
```

**DX-specific features:**
- Similar to CR but for digital detectors
- ViewPosition (AP, PA, LATERAL)
- SOP Class: Digital X-Ray Image Storage for Presentation

### US - Ultrasound

```bash
dicomforge --num-images 30 --total-size 50MB --modality US --output ultrasound
```

**US-specific features:**
- TransducerType (LINEAR, CONVEX, PHASED)
- TransducerFrequency
- 8-bit grayscale images
- SOP Class: Ultrasound Image Storage

### MG - Mammography

```bash
dicomforge --num-images 4 --total-size 200MB --modality MG --output mammography
```

**MG-specific features:**
- ImageLaterality (L/R)
- ViewPosition (CC, MLO)
- AnodeTargetMaterial, CompressionForce
- High-resolution 14-bit images
- SOP Class: Digital Mammography X-Ray Image Storage for Presentation

---

## Multi-Studies and Multi-Patients

Generate multiple studies for workflow testing, or distribute studies across multiple patients.

### Multiple Studies (Single Patient)

```bash
# Generate 3 studies for one patient (e.g., follow-up exams)
dicomforge --num-images 30 --total-size 300MB --num-studies 3 --output followup
```

**Use case:** Testing study management, comparing exams over time.

**Output structure:**
```
followup/
├── DICOMDIR
└── PT000000/           # 1 patient
    ├── ST000000/       # Study 1 (10 images)
    ├── ST000001/       # Study 2 (10 images)
    └── ST000002/       # Study 3 (10 images)
```

### Multiple Patients with Studies

```bash
# Generate 6 studies distributed across 2 patients (3 studies each)
dicomforge --num-images 60 --total-size 600MB --num-studies 6 --num-patients 2 --output multi_patient
```

**Use case:** Testing patient list views, patient merging, multi-patient workflows.

**Output structure:**
```
multi_patient/
├── DICOMDIR
├── PT000000/           # Patient 1
│   ├── ST000000/       # 3 studies
│   ├── ST000001/
│   └── ST000002/
└── PT000001/           # Patient 2
    ├── ST000000/       # 3 studies
    ├── ST000001/
    └── ST000002/
```

---

## Multi-Series per Study

Generate multiple series within a single study, useful for multi-sequence MR or multi-phase CT.

### Fixed Number of Series

```bash
# Generate a study with exactly 3 series
dicomforge --num-images 30 --total-size 300MB --series-per-study 3 --output multi_series
```

**Output structure:**
```
multi_series/
└── PT000000/
    └── ST000000/
        ├── SE000000/   # Series 1 (10 images)
        ├── SE000001/   # Series 2 (10 images)
        └── SE000002/   # Series 3 (10 images)
```

### Random Range of Series

```bash
# Generate 2 to 5 series per study (randomly chosen)
dicomforge --num-images 50 --total-size 500MB --series-per-study 2-5 --output variable_series
```

**Use case:** Testing dynamic layouts that handle varying series counts.

### MR Brain Protocol (Multiple Sequences)

```bash
# Simulate a typical brain MRI with T1, T2, FLAIR sequences
dicomforge --num-images 90 --total-size 500MB \
  --modality MR --body-part HEAD \
  --series-per-study 3 \
  --output brain_mri
```

**What it generates:**
- Each series gets a realistic SeriesDescription (e.g., "T1 SAG", "T2 AX", "FLAIR COR")
- Appropriate orientation for each sequence
- Shared FrameOfReferenceUID across all series (for fusion/overlay)

### CT Multi-Phase (Contrast Phases)

```bash
# Simulate CT with pre-contrast, arterial, and venous phases
dicomforge --num-images 300 --total-size 600MB \
  --modality CT \
  --series-per-study 3 \
  --output ct_multiphase
```

---

## Custom Study Descriptions

Name each study explicitly for clinical trial simulations or longitudinal studies.

```bash
# Longitudinal study: baseline and follow-ups at 3 and 6 months
dicomforge --num-images 30 --total-size 300MB \
  --num-studies 3 \
  --study-descriptions "IRM_T0,IRM_M3,IRM_M6" \
  --output longitudinal
```

**Requirements:**
- Number of descriptions must match `--num-studies`
- Comma-separated, no spaces around commas

**Use case:** Clinical trial data simulation, testing study matching by description.

```bash
# Named protocols
dicomforge --num-images 20 --total-size 200MB \
  --num-studies 2 \
  --study-descriptions "PRE-OP PLANNING,POST-OP CONTROL" \
  --output surgical
```

---

## Custom DICOM Tags

Override any DICOM tag value using the `--tag` flag. Repeatable for multiple tags.

### Single Tag

```bash
# Set institution name
dicomforge --num-images 10 --total-size 100MB \
  --tag "InstitutionName=CHU Bordeaux" \
  --output bordeaux_data
```

### Multiple Tags

```bash
# Set multiple tags
dicomforge --num-images 10 --total-size 100MB \
  --tag "InstitutionName=Memorial Hospital" \
  --tag "StationName=MR-Scanner-01" \
  --tag "ReferringPhysicianName=Dr. Smith" \
  --output custom_tags
```

### Common Tags to Override

| Tag | Description | Example |
|-----|-------------|---------|
| `InstitutionName` | Hospital/clinic name | `--tag "InstitutionName=Mayo Clinic"` |
| `StationName` | Scanner/workstation name | `--tag "StationName=CT-Room-3"` |
| `ReferringPhysicianName` | Ordering physician | `--tag "ReferringPhysicianName=Dr. Jones"` |
| `StudyDescription` | Study description (all studies) | `--tag "StudyDescription=BRAIN MRI"` |
| `PerformingPhysicianName` | Technologist | `--tag "PerformingPhysicianName=Tech Johnson"` |
| `OperatorsName` | Operator name | `--tag "OperatorsName=JSmith"` |

**Note:** Use `--study-descriptions` for per-study descriptions instead of `--tag StudyDescription`.

---

## Categorization Options

Fine-tune metadata for specific testing scenarios.

### Institution and Department

```bash
dicomforge --num-images 20 --total-size 200MB \
  --institution "University Hospital" \
  --department "Radiology" \
  --output categorized
```

### Body Part

```bash
# Specify anatomical region (affects generated metadata)
dicomforge --num-images 50 --total-size 500MB \
  --modality MR \
  --body-part HEAD \
  --output brain_scan

# Other body parts: CHEST, ABDOMEN, PELVIS, KNEE, SPINE, LSPINE, etc.
```

### Priority

```bash
# Set exam priority (affects RequestedProcedurePriority tag)
dicomforge --num-images 10 --total-size 100MB \
  --priority HIGH \
  --output urgent_exam

# Options: HIGH, ROUTINE, LOW
```

### Varied Metadata

```bash
# Generate different institutions/physicians across studies
dicomforge --num-images 30 --total-size 300MB \
  --num-studies 3 \
  --varied-metadata \
  --output varied
```

**Use case:** Testing grouping/filtering by institution or physician.

---

## Edge Cases for Robustness Testing

Generate unusual but valid DICOM data to test system robustness.

### Enable Edge Cases

```bash
# Apply edge cases to 50% of patients
dicomforge --num-images 100 --total-size 1GB \
  --num-patients 20 \
  --edge-cases 50 \
  --output edge_test
```

### Specific Edge Case Types

```bash
# Only special characters and long names
dicomforge --num-images 50 --total-size 500MB \
  --num-patients 10 \
  --edge-cases 100 \
  --edge-case-types "special-chars,long-names" \
  --output names_test
```

### Available Edge Case Types

| Type | Description | Example Values |
|------|-------------|----------------|
| `special-chars` | Names with accents, hyphens, apostrophes | `Müller-Schmidt`, `O'Connor`, `François` |
| `long-names` | Names at DICOM's 64-character limit | `ALEXANDROPOULOSWILLIAMSON^CHRISTOPHERJOHN...` |
| `old-dates` | Very old birth dates (1900-1950) or partial dates | `19250315`, `1940`, `194506` |
| `varied-ids` | Patient IDs with dashes, letters, spaces | `123-456-789`, `A1B2C3D4`, `PAT 12345 67` |
| `missing-tags` | Omit optional DICOM tags | Missing StudyDescription, BodyPartExamined |

### All Edge Cases (Comprehensive Testing)

```bash
dicomforge --num-images 100 --total-size 1GB \
  --num-patients 50 \
  --edge-cases 100 \
  --edge-case-types "special-chars,long-names,old-dates,varied-ids,missing-tags" \
  --output comprehensive_edge
```

---

## Reproducibility

Use seeds for deterministic output - identical data across runs.

### Fixed Seed

```bash
# Generate identical data every time
dicomforge --num-images 20 --total-size 200MB --seed 42 --output reproducible

# Running again produces exactly the same files
dicomforge --num-images 20 --total-size 200MB --seed 42 --output reproducible
```

**What's reproducible with the same seed:**
- Patient names and IDs
- Study and Series UIDs
- All DICOM metadata
- Image content

### Automatic Seed from Directory Name

Without `--seed`, dicomforge generates a deterministic seed from the output directory name:

```bash
# These produce consistent IDs (based on "test_data" name)
dicomforge --num-images 10 --total-size 100MB --output test_data
dicomforge --num-images 10 --total-size 100MB --output test_data  # Same patient/study IDs
```

---

## Performance Tuning

### Worker Control

```bash
# Use all CPU cores (default)
dicomforge --num-images 500 --total-size 4GB --output fast

# Limit to 4 workers (for resource-constrained systems)
dicomforge --num-images 500 --total-size 4GB --workers 4 --output limited

# Single-threaded (for debugging or minimal resource usage)
dicomforge --num-images 50 --total-size 500MB --workers 1 --output sequential
```

**Performance guidelines:**
- Default (all cores) is fastest for large datasets
- Reduce workers if system becomes unresponsive
- SSD storage significantly improves write performance

---

## Real-World Scenarios

### Scenario 1: PACS Migration Testing

Generate diverse data to test PACS import:

```bash
dicomforge --num-images 500 --total-size 5GB \
  --num-studies 50 \
  --num-patients 25 \
  --modality CT \
  --varied-metadata \
  --edge-cases 20 \
  --output pacs_migration_test
```

### Scenario 2: Clinical Trial Simulation

Simulate a longitudinal study with 10 patients, 3 visits each:

```bash
dicomforge --num-images 300 --total-size 3GB \
  --num-studies 30 \
  --num-patients 10 \
  --study-descriptions "BASELINE,WEEK_4,WEEK_12" \
  --modality MR \
  --body-part HEAD \
  --series-per-study 4 \
  --institution "Clinical Trial Site A" \
  --seed 2024 \
  --output clinical_trial
```

**Note:** With 30 studies and 3 descriptions, each patient gets all 3 timepoints.

### Scenario 3: Viewer Development

Generate multi-series MR for testing series navigation:

```bash
dicomforge --num-images 200 --total-size 1GB \
  --modality MR \
  --body-part HEAD \
  --series-per-study 5 \
  --output viewer_test
```

### Scenario 4: Load Testing

Stress test with large dataset:

```bash
dicomforge --num-images 2000 --total-size 20GB \
  --num-studies 100 \
  --num-patients 50 \
  --modality CT \
  --output stress_test
```

### Scenario 5: Multi-Modality Worklist

Generate different modalities separately, then combine:

```bash
# Create separate directories
dicomforge --num-images 50 --total-size 500MB --modality MR --output worklist/mr
dicomforge --num-images 100 --total-size 200MB --modality CT --output worklist/ct
dicomforge --num-images 10 --total-size 100MB --modality DX --output worklist/xray
dicomforge --num-images 20 --total-size 50MB --modality US --output worklist/us
```

### Scenario 6: Mammography Screening

Generate bilateral mammography exams:

```bash
dicomforge --num-images 4 --total-size 400MB \
  --modality MG \
  --institution "Breast Imaging Center" \
  --output mammo_screening
```

---

## Quick Reference

### Required Arguments

| Argument | Description |
|----------|-------------|
| `--num-images N` | Number of DICOM images to generate |
| `--total-size SIZE` | Total size (e.g., `100MB`, `1GB`) |

### All Options

| Option | Default | Description |
|--------|---------|-------------|
| `--output DIR` | `dicom_series` | Output directory |
| `--modality MOD` | `MR` | Modality: MR, CT, CR, DX, US, MG |
| `--seed N` | auto | Random seed for reproducibility |
| `--num-studies N` | `1` | Number of studies |
| `--num-patients N` | `1` | Number of patients |
| `--series-per-study N` | `1` | Series per study (or range: `2-5`) |
| `--study-descriptions LIST` | auto | Comma-separated study names |
| `--tag NAME=VALUE` | - | Custom DICOM tag (repeatable) |
| `--institution NAME` | random | Institution name |
| `--department NAME` | random | Department name |
| `--body-part PART` | random | Body part examined |
| `--priority LEVEL` | `ROUTINE` | Priority: HIGH, ROUTINE, LOW |
| `--varied-metadata` | `false` | Vary institutions/physicians |
| `--edge-cases N` | `0` | Percentage with edge cases (0-100) |
| `--edge-case-types LIST` | all | Comma-separated edge case types |
| `--workers N` | CPU cores | Parallel workers |
| `--help` | - | Show help |
| `--version` | - | Show version |
