# Extended DICOM Features Design

## Context

DICOMFORGE is a CLI tool for generating test DICOM files. Currently it only supports MR modality with basic metadata. This design extends it to support a medical platform that needs to:

- Display DICOM files for medical experts
- Parse and categorize incoming files
- Handle edge cases and malformed data gracefully

## Features Overview

### Priority High

1. **Additional Modalities** — CT, CR/DX, US, MG
2. **Advanced Categorization Tags** — Institution, physicians, body part, protocol, priority
3. **Multi-Series Studies** — 1-N series per study with spatial coherence
4. **Valid Edge Cases** — Special characters, long names, missing optional tags

### Priority Medium

5. **Malformed Files Mode** — Controlled generation of invalid files for robustness testing
6. **MPR-Compatible Data** — Coherent geometry for multi-planar reconstruction
7. **Multiple Window Presets** — W/L presets per modality

### Priority Low (Future)

8. **Presentation States** — DICOM annotations and saved view states
9. **Multi-Frame Support** — US clips, cardiac cine sequences
10. **Pixel Realism** — Modality-specific visual patterns

---

## Feature 1: Additional Modalities

### New CLI Option

```bash
--modality <MR|CT|CR|DX|US|MG>  # Default: MR (current behavior)
```

### CT (Computed Tomography)

**SOP Class UID:** `1.2.840.10008.5.1.4.1.1.2` (CT Image Storage)

**Specific Tags:**
| Tag | VR | Description | Generated Values |
|-----|-----|-------------|------------------|
| KVP | DS | Tube voltage | 80, 100, 120, 140 |
| XRayTubeCurrent | IS | Tube current (mA) | 100-400 |
| ConvolutionKernel | SH | Reconstruction filter | SOFT, STANDARD, BONE, LUNG |
| RescaleIntercept | DS | HU offset | -1024 |
| RescaleSlope | DS | HU scale | 1 |
| SliceThickness | DS | Slice thickness | 0.5-3.0 mm |

**Pixel Data:**
- 16-bit signed (PixelRepresentation = 1)
- Values in Hounsfield Units range: -1024 to +3071
- Window presets: BRAIN (40/80), BONE (400/2000), LUNG (-600/1500)

### CR/DX (Computed/Digital Radiography)

**SOP Class UIDs:**
- CR: `1.2.840.10008.5.1.4.1.1.1` (Computed Radiography)
- DX: `1.2.840.10008.5.1.4.1.1.1.1` (Digital X-Ray)

**Specific Tags:**
| Tag | VR | Description | Generated Values |
|-----|-----|-------------|------------------|
| ViewPosition | CS | View orientation | AP, PA, LAT, LL, RL |
| BodyPartExamined | CS | Anatomy | CHEST, HAND, KNEE, SPINE, SKULL |
| ImagerPixelSpacing | DS | Detector pixel size | 0.1-0.2 mm |
| DistanceSourceToDetector | DS | SID | 1000-1800 mm |
| DistanceSourceToPatient | DS | SOD | 800-1500 mm |
| Exposure | IS | Exposure (mAs) | 1-50 |

**Pixel Data:**
- Single image (not a series of slices)
- Higher resolution: 2000-4000 pixels
- PhotometricInterpretation: MONOCHROME1 or MONOCHROME2

### US (Ultrasound)

**SOP Class UID:** `1.2.840.10008.5.1.4.1.1.6.1` (Ultrasound Image Storage)

**Specific Tags:**
| Tag | VR | Description | Generated Values |
|-----|-----|-------------|------------------|
| SequenceOfUltrasoundRegions | SQ | Calibration regions | Single region with pixel spacing |
| UltrasoundColorDataPresent | US | Color Doppler flag | 0 (grayscale for now) |
| FrameTime | DS | Time between frames | 33 ms (30 fps) |
| NumberOfFrames | IS | Frame count | 1 (single frame initially) |
| TransducerType | CS | Probe type | LINEAR, CONVEX, PHASED |
| TransducerFrequency | DS | Probe frequency | 2-15 MHz |

**Pixel Data:**
- Typical resolution: 640x480 to 1024x768
- Speckle-like texture (future enhancement)
- May include color (RGB for Doppler - future)

### MG (Mammography)

**SOP Class UID:** `1.2.840.10008.5.1.4.1.1.1.2` (Digital Mammography)

**Specific Tags:**
| Tag | VR | Description | Generated Values |
|-----|-----|-------------|------------------|
| ImageLaterality | CS | Breast side | L, R |
| ViewPosition | CS | Standard views | CC, MLO, ML, LM |
| BodyPartExamined | CS | Fixed value | BREAST |
| AnodeTargetMaterial | CS | X-ray target | MOLYBDENUM, RHODIUM, TUNGSTEN |
| FilterMaterial | CS | Filter type | MOLYBDENUM, RHODIUM, SILVER |
| CompressionForce | DS | Compression (N) | 80-200 |
| OrganDose | DS | Glandular dose (mGy) | 1-3 |

**Pixel Data:**
- High resolution: 3000-5000 pixels
- 12-16 bits stored
- PhotometricInterpretation: MONOCHROME1 (typically)

### Implementation Notes

- Create `internal/dicom/modalities/` package with per-modality tag generators
- Each modality implements a `MetadataGenerator` interface
- Factory function selects generator based on `--modality` flag
- Existing MR implementation moves to `modalities/mr.go`

---

## Feature 2: Advanced Categorization Tags

### New CLI Options

```bash
--institution <name>           # e.g., "CHU Bordeaux" (random if not specified)
--department <name>            # e.g., "Radiologie" (random if not specified)
--body-part <part>             # e.g., "HEAD" (random per modality if not specified)
--priority <HIGH|ROUTINE|LOW>  # Default: ROUTINE
--varied-metadata              # Generate varied institutions/physicians across studies
```

### New Tags to Generate

**Origin Identification:**
| Tag | VR | Description | Example Values |
|-----|-----|-------------|----------------|
| InstitutionName | LO | Hospital/clinic name | "CHU Bordeaux", "Hopital Saint-Louis", "Clinique du Parc" |
| InstitutionAddress | ST | Full address | "Place Amelie Raba-Leon, 33000 Bordeaux" |
| StationName | SH | Machine identifier | "IRM_NEURO_01", "CT_URG_02" |
| InstitutionalDepartmentName | LO | Department | "Radiologie", "Urgences", "Cardiologie" |

**Physicians and Workflow:**
| Tag | VR | Description | Example Values |
|-----|-----|-------------|----------------|
| ReferringPhysicianName | PN | Ordering physician | Realistic French/English names |
| PerformingPhysicianName | PN | Radiologist/technician | Realistic names |
| OperatorsName | PN | Technician | Realistic names |
| RequestingPhysician | PN | Requester (if different) | Realistic names |

**Clinical Context:**
| Tag | VR | Description | Example Values |
|-----|-----|-------------|----------------|
| BodyPartExamined | CS | Anatomy (DICOM defined terms) | HEAD, CHEST, ABDOMEN, LSPINE, KNEE |
| ProtocolName | LO | Acquisition protocol | "BRAIN_ROUTINE", "THORAX_TRAUMA" |
| RequestedProcedureDescription | LO | Exam requested | "IRM cerebrale sans injection" |
| ReasonForStudy | LO | Clinical indication | "Cephalees persistantes" |

**Priority and Status:**
| Tag | VR | Description | Values |
|-----|-----|-------------|--------|
| Priority | CS | Exam priority | HIGH, ROUTINE, LOW |
| InstanceAvailability | CS | Data availability | ONLINE |

### Implementation Notes

- Add name generator for French physician names (reuse pattern from patient names)
- Create lookup tables for realistic institutions, departments, protocols
- BodyPartExamined should correlate with modality and StudyDescription
- Protocol names should be consistent with modality and body part

---

## Feature 3: Multi-Series Studies

### New CLI Options

```bash
--series-per-study <min>-<max>   # e.g., "3-5" (default: "1-1" = current behavior)
--series-per-study <n>           # Fixed number, e.g., "4"
```

### Series Configuration by Modality

**MR Brain (3-6 series):**
| Series | SequenceName | SeriesDescription | Contrast |
|--------|--------------|-------------------|----------|
| 1 | T1_SE | "T1 SAG" | No |
| 2 | T2_FSE | "T2 AX" | No |
| 3 | T2_FLAIR | "FLAIR AX" | No |
| 4 | T1_MPRAGE | "T1 SAG +C" | Yes (Gadolinium) |
| 5 | DWI | "DWI AX" | No |
| 6 | T2_STAR | "T2* GRE" | No |

**CT Thorax/Abdomen (2-4 series):**
| Series | SeriesDescription | ContrastBolusAgent |
|--------|-------------------|-------------------|
| 1 | "Sans contraste" | None |
| 2 | "Arteriel" | "IOMERON 400" |
| 3 | "Portal" | "IOMERON 400" |
| 4 | "Tardif" | "IOMERON 400" |

**MR Knee (4-5 series):**
| Series | SequenceName | ImageOrientationPatient | SeriesDescription |
|--------|--------------|------------------------|-------------------|
| 1 | T1_SE | Sagittal | "T1 SAG" |
| 2 | T2_FSE | Sagittal | "T2 SAG FAT-SAT" |
| 3 | PD_FSE | Coronal | "PD COR" |
| 4 | T2_FSE | Axial | "T2 AX" |

### Spatial Coherence

All series in a study must share:
- `FrameOfReferenceUID` — Same coordinate system
- `PatientPosition` — Same patient position (HFS, FFS, etc.)
- Consistent `ImagePositionPatient` origin (may differ by orientation)

Series may vary:
- `ImageOrientationPatient` — Axial [1,0,0,0,1,0], Sagittal [0,1,0,0,0,-1], Coronal [1,0,0,0,0,-1]
- `SliceThickness`, `PixelSpacing` — Protocol-dependent
- Image count per series

### Implementation Notes

- Distribute `--num-images` across series (configurable: equal vs varied)
- Generate `FrameOfReferenceUID` per study, shared by all series
- Create series template configurations per modality
- Add `ContrastBolusAgent`, `ContrastBolusStartTime` tags for contrast series

---

## Feature 4: Valid Edge Cases

### New CLI Option

```bash
--edge-cases <percentage>   # e.g., "10" = 10% of files have edge case variations
--edge-case-types <types>   # Comma-separated: "special-chars,long-names,missing-tags,old-dates"
```

### Edge Case Types

**Special Characters (`special-chars`):**
- Patient names with accents: "François Müller", "José García", "Øystein Næss"
- Asian characters: "田中太郎" (Tanaka Taro)
- Special chars: "O'Brien", "Mary-Jane", "Jean-Pierre"
- Mixed: "Dr. José García-Müller III"

**Long Names (`long-names`):**
- PatientName at max length (64 chars for LO)
- StudyDescription at max length
- InstitutionName at max length
- Test truncation handling

**Missing Optional Tags (`missing-tags`):**
Files randomly missing (but remaining valid):
- BodyPartExamined
- StudyDescription / SeriesDescription
- ReferringPhysicianName
- PatientBirthDate
- InstitutionName
- ProtocolName

**Date Variations (`old-dates`):**
- Very old birth dates: 1900-01-01
- Partial dates: "1990" (year only), "199006" (year-month)
- Future study dates (for validation testing)

**PatientID Variations (`varied-ids`):**
- With dashes: "123-456-789"
- With letters: "ABC123456"
- With spaces: "123 456 789"
- Very long IDs

### Implementation Notes

- Create `internal/dicom/edgecases/` package
- Edge case applicators wrap normal tag generation
- Percentage applies per-file, randomly selected
- Multiple edge case types can combine in same file

---

## Feature 5: Malformed Files Mode

### New CLI Option

```bash
--quirks-mode <types>       # Comma-separated: "bad-vr,truncated,encoding,out-of-range"
--quirks-percentage <n>     # Percentage of files affected (default: 10)
```

### Quirk Types

**Bad Value Representation (`bad-vr`):**
- Store integer in DS (Decimal String) field without decimal
- Store text in IS (Integer String) field
- Wrong VR for standard tags (e.g., DA stored as LO)

**Truncated Data (`truncated`):**
- Pixel data shorter than declared dimensions
- Missing last bytes of file
- Incomplete sequences (SQ without terminator)

**Encoding Issues (`encoding`):**
- Wrong character set declaration
- UTF-8 content with ISO-8859-1 declared
- Invalid characters for declared encoding
- Missing Specific Character Set tag with non-ASCII content

**Out of Range (`out-of-range`):**
- Negative values in unsigned fields
- BitsStored > BitsAllocated
- InstanceNumber = 0 or negative
- Invalid date formats: "20251301" (month 13)

**Missing Required Tags (`missing-required`):**
- Missing SOPClassUID
- Missing SOPInstanceUID
- Missing PatientID
- Missing Rows/Columns with pixel data present

### Implementation Notes

- Malformed files should be clearly marked (filename suffix? metadata?)
- Generate manifest listing which files have which quirks
- These files may not load in strict DICOM parsers - that's intentional
- Separate output directory option: `--quirks-output-dir`

---

## Feature 6: MPR-Compatible Data

### Requirements for MPR Reconstruction

For multi-planar reconstruction to work correctly:

**Geometric Consistency:**
- `PixelSpacing` identical across all images in series
- `SliceThickness` consistent and accurate
- `SpacingBetweenSlices` equals actual distance (no gaps or overlaps)
- `ImagePositionPatient` forms regular grid

**Isotropic or Near-Isotropic:**
- Ideal: `PixelSpacing[0] == PixelSpacing[1] == SliceThickness`
- Acceptable: ratio < 3:1 between any dimensions
- Example: 0.5mm x 0.5mm x 1.0mm

**Frame of Reference:**
- Single `FrameOfReferenceUID` for entire volume
- Consistent `ImageOrientationPatient` (no rotation between slices)
- Sequential `ImagePositionPatient` Z-values

### New CLI Option

```bash
--mpr-compatible            # Ensure volumetric data suitable for MPR
--voxel-size <x,y,z>        # e.g., "0.5,0.5,1.0" mm (optional override)
```

### Implementation Notes

- Enforce geometric constraints when flag is set
- Calculate image count from volume size / voxel dimensions
- Validate that `--num-images` is compatible with volumetric constraints
- Add `SliceLocation` tag with accurate Z position

---

## Feature 7: Multiple Window Presets

### VOI LUT (Window/Level) Tags

Currently generated: single `WindowCenter` and `WindowWidth` per image.

**Enhancement:** Multiple presets per image via `VOILUTSequence` or repeated W/C, W/W values.

### Presets by Modality

**CT:**
| Preset Name | WindowCenter | WindowWidth |
|-------------|--------------|-------------|
| BRAIN | 40 | 80 |
| SUBDURAL | 75 | 215 |
| BONE | 400 | 2000 |
| LUNG | -600 | 1500 |
| MEDIASTINUM | 40 | 400 |
| ABDOMEN | 40 | 350 |
| LIVER | 60 | 150 |

**MR:**
| Preset Name | WindowCenter | WindowWidth |
|-------------|--------------|-------------|
| DEFAULT | 500 | 1000 |
| BRIGHT | 300 | 600 |
| CONTRAST | 600 | 1200 |

**MG:**
| Preset Name | WindowCenter | WindowWidth |
|-------------|--------------|-------------|
| DEFAULT | 2048 | 4096 |
| DENSE | 3000 | 2000 |
| FATTY | 1500 | 3000 |

### Tags to Generate

| Tag | VR | Description |
|-----|-----|-------------|
| WindowCenter | DS | Can be multi-valued: "40\\75\\400" |
| WindowWidth | DS | Can be multi-valued: "80\\215\\2000" |
| WindowCenterWidthExplanation | LO | Preset names: "BRAIN\\SUBDURAL\\BONE" |

### Implementation Notes

- Generate appropriate presets based on modality and body part
- CT always includes at least SOFT TISSUE and BONE presets
- Use multi-valued strings (backslash-separated per DICOM standard)

---

## Feature 8: Presentation States (Future)

### Overview

DICOM Presentation State objects store display settings separately from images:
- Window/level settings
- Geometric transformations (rotation, flip, zoom)
- Annotations (text, arrows, measurements)
- Shutters (masking regions)

**SOP Class UID:** `1.2.840.10008.5.1.4.1.1.11.1` (Grayscale Softcopy Presentation State)

### Implementation Scope (Future)

- Generate basic GSPS files referencing generated images
- Include simple annotations (patient name overlay, acquisition info)
- Store in same DICOMDIR structure
- Flag: `--generate-presentation-states`

---

## CLI Summary

### New Flags

```bash
# Modality selection
--modality <MR|CT|CR|DX|US|MG>

# Categorization
--institution <name>
--department <name>
--body-part <part>
--priority <HIGH|ROUTINE|LOW>
--varied-metadata

# Multi-series
--series-per-study <n|min-max>

# Edge cases
--edge-cases <percentage>
--edge-case-types <type1,type2,...>

# Malformed files
--quirks-mode <type1,type2,...>
--quirks-percentage <n>

# Geometry
--mpr-compatible
--voxel-size <x,y,z>
```

### Example Commands

```bash
# Basic CT generation
dicomforge --num-images 100 --total-size 500MB --modality CT

# Multi-series MR brain study
dicomforge --num-images 150 --total-size 300MB \
  --modality MR --body-part HEAD --series-per-study 3-5

# Stress test with edge cases
dicomforge --num-images 500 --total-size 2GB \
  --varied-metadata --edge-cases 15 \
  --edge-case-types special-chars,missing-tags

# Robustness testing
dicomforge --num-images 50 --total-size 100MB \
  --quirks-mode bad-vr,encoding --quirks-percentage 20

# MPR-compatible volume
dicomforge --num-images 200 --total-size 400MB \
  --modality CT --mpr-compatible --voxel-size 0.5,0.5,1.0
```

---

## Implementation Order

Recommended implementation sequence:

1. **Feature 2: Categorization Tags** — Low complexity, high value, no architecture change
2. **Feature 4: Valid Edge Cases** — Builds on feature 2, tests parsing robustness
3. **Feature 1: Additional Modalities** — Requires refactoring metadata generation
4. **Feature 3: Multi-Series Studies** — Requires modality support, significant logic
5. **Feature 7: Window Presets** — Simple addition, modality-dependent
6. **Feature 6: MPR-Compatible Data** — Geometry constraints, validation
7. **Feature 5: Malformed Files** — Separate mode, can be done independently
8. **Feature 8: Presentation States** — Future, separate SOP class
