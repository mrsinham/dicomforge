# Plan de Tests d'Intégration - Générateur DICOM Go

**Date**: 2026-01-18
**Objectif**: Valider le fonctionnement end-to-end du générateur DICOM Go et sa compatibilité avec la version Python

---

## Vue d'Ensemble

### Stratégie de Test

Les tests d'intégration vérifient que tous les composants fonctionnent ensemble correctement :
- Génération complète de séries DICOM
- Organisation hiérarchique des fichiers
- Validité des fichiers DICOM générés
- Compatibilité avec viewers DICOM
- Reproductibilité avec seed
- Comparaison avec la version Python

### Structure des Tests

```
go/
├── tests/
│   ├── integration_test.go           # Tests principaux
│   ├── validation_test.go            # Validation DICOM
│   ├── compatibility_test.go         # Comparaison Python/Go
│   └── testdata/                     # Données de test
│       ├── expected_structure.json
│       └── reference_metadata.json
└── scripts/
    ├── compare_outputs.sh            # Script de comparaison
    └── validate_dicom.py             # Validation avec pydicom
```

---

## Test Suite 1: Génération Basique

### Test 1.1: Génération Simple
**Fichier**: `integration_test.go`

```go
func TestGenerateSeries_Basic(t *testing.T) {
    // Setup
    outputDir := t.TempDir()

    opts := dicom.GeneratorOptions{
        NumImages:  10,
        TotalSize:  "50MB",
        OutputDir:  outputDir,
        Seed:       42,
        NumStudies: 1,
    }

    // Execute
    files, err := dicom.GenerateDICOMSeries(opts)

    // Verify
    assert.NoError(t, err)
    assert.Equal(t, 10, len(files))

    // Check files exist
    for _, file := range files {
        assert.FileExists(t, file.Path)
    }
}
```

**Validation**:
- ✅ 10 fichiers créés
- ✅ Pas d'erreurs
- ✅ Tous les fichiers existent
- ✅ Taille totale ≈ 50MB (± 5%)

### Test 1.2: Calcul de Dimensions
```go
func TestCalculateDimensions_ValidSizes(t *testing.T) {
    tests := []struct {
        name       string
        totalBytes int64
        numImages  int
        wantWidth  int
        wantHeight int
    }{
        {"100MB_10images", 100 * 1024 * 1024, 10, 2304, 2304},
        {"1GB_50images", 1024 * 1024 * 1024, 50, 3200, 3200},
        {"10MB_5images", 10 * 1024 * 1024, 5, 1024, 1024},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            w, h, err := dicom.CalculateDimensions(tt.totalBytes, tt.numImages)
            assert.NoError(t, err)
            assert.Equal(t, tt.wantWidth, w)
            assert.Equal(t, tt.wantHeight, h)
        })
    }
}
```

### Test 1.3: Validation des Paramètres
```go
func TestGenerateSeries_InvalidParameters(t *testing.T) {
    tests := []struct {
        name    string
        opts    dicom.GeneratorOptions
        wantErr string
    }{
        {
            name: "negative_images",
            opts: dicom.GeneratorOptions{
                NumImages: -1,
                TotalSize: "100MB",
            },
            wantErr: "number of images must be > 0",
        },
        {
            name: "invalid_size",
            opts: dicom.GeneratorOptions{
                NumImages: 10,
                TotalSize: "invalid",
            },
            wantErr: "invalid format",
        },
        {
            name: "too_many_studies",
            opts: dicom.GeneratorOptions{
                NumImages:  5,
                TotalSize:  "10MB",
                NumStudies: 10,
            },
            wantErr: "num studies cannot exceed num images",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := dicom.GenerateDICOMSeries(tt.opts)
            assert.Error(t, err)
            assert.Contains(t, err.Error(), tt.wantErr)
        })
    }
}
```

---

## Test Suite 2: Structure et Organisation

### Test 2.1: Hiérarchie DICOMDIR
```go
func TestOrganizeFiles_DICOMDIRStructure(t *testing.T) {
    // Setup: generate files
    outputDir := t.TempDir()
    files := generateTestFiles(t, outputDir, 10, 1)

    // Execute
    err := dicom.OrganizeFilesIntoDICOMDIR(outputDir, files)
    assert.NoError(t, err)

    // Verify structure
    // 1. DICOMDIR exists
    dicomdirPath := filepath.Join(outputDir, "DICOMDIR")
    assert.FileExists(t, dicomdirPath)

    // 2. PT000000 directory exists
    patientDir := filepath.Join(outputDir, "PT000000")
    assert.DirExists(t, patientDir)

    // 3. ST000000 directory exists
    studyDir := filepath.Join(patientDir, "ST000000")
    assert.DirExists(t, studyDir)

    // 4. SE000000 directory exists
    seriesDir := filepath.Join(studyDir, "SE000000")
    assert.DirExists(t, seriesDir)

    // 5. Image files exist (IM000001, IM000002, ...)
    for i := 1; i <= 10; i++ {
        imageFile := filepath.Join(seriesDir, fmt.Sprintf("IM%06d", i))
        assert.FileExists(t, imageFile)
    }

    // 6. No IMG*.dcm files in root
    matches, _ := filepath.Glob(filepath.Join(outputDir, "IMG*.dcm"))
    assert.Empty(t, matches, "Temporary IMG*.dcm files should be cleaned up")
}
```

### Test 2.2: Multi-Études
```go
func TestOrganizeFiles_MultipleStudies(t *testing.T) {
    outputDir := t.TempDir()

    opts := dicom.GeneratorOptions{
        NumImages:  30,
        TotalSize:  "100MB",
        OutputDir:  outputDir,
        Seed:       42,
        NumStudies: 3,
    }

    files, err := dicom.GenerateDICOMSeries(opts)
    assert.NoError(t, err)

    err = dicom.OrganizeFilesIntoDICOMDIR(outputDir, files)
    assert.NoError(t, err)

    // Verify 3 study directories
    patientDir := filepath.Join(outputDir, "PT000000")

    for i := 0; i < 3; i++ {
        studyDir := filepath.Join(patientDir, fmt.Sprintf("ST%06d", i))
        assert.DirExists(t, studyDir)

        // Each study should have SE000000
        seriesDir := filepath.Join(studyDir, "SE000000")
        assert.DirExists(t, seriesDir)
    }

    // Count total images across all studies
    totalImages := 0
    filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
        if !info.IsDir() && strings.HasPrefix(info.Name(), "IM") {
            totalImages++
        }
        return nil
    })
    assert.Equal(t, 30, totalImages)
}
```

### Test 2.3: Répartition des Images
```go
func TestGenerateSeries_ImageDistribution(t *testing.T) {
    outputDir := t.TempDir()

    opts := dicom.GeneratorOptions{
        NumImages:  25,  // 25 images, 3 studies = 9, 8, 8
        TotalSize:  "50MB",
        OutputDir:  outputDir,
        Seed:       42,
        NumStudies: 3,
    }

    files, err := dicom.GenerateDICOMSeries(opts)
    assert.NoError(t, err)

    // Group by study
    studyCounts := make(map[string]int)
    for _, file := range files {
        studyCounts[file.StudyUID]++
    }

    assert.Equal(t, 3, len(studyCounts))

    // First study should have 9 images (25 / 3 = 8 remainder 1)
    // Check distribution is reasonable
    for _, count := range studyCounts {
        assert.True(t, count >= 8 && count <= 9)
    }
}
```

---

## Test Suite 3: Validation DICOM

### Test 3.1: Tags Requis
**Fichier**: `validation_test.go`

```go
func TestValidation_RequiredTags(t *testing.T) {
    outputDir := t.TempDir()
    files := generateAndOrganizeFiles(t, outputDir, 5, 1)

    // Parse first DICOM file
    firstImage := filepath.Join(outputDir, "PT000000", "ST000000", "SE000000", "IM000001")
    ds, err := dicom.ParseFile(firstImage, nil)
    assert.NoError(t, err)

    // Verify required tags exist
    requiredTags := []tag.Tag{
        tag.PatientName,
        tag.PatientID,
        tag.PatientBirthDate,
        tag.PatientSex,
        tag.StudyInstanceUID,
        tag.SeriesInstanceUID,
        tag.SOPInstanceUID,
        tag.Modality,
        tag.Rows,
        tag.Columns,
        tag.BitsAllocated,
        tag.PhotometricInterpretation,
    }

    for _, t := range requiredTags {
        elem, err := ds.FindElementByTag(t)
        assert.NoError(t, err, "Tag %v should exist", t)
        assert.NotNil(t, elem)
    }
}
```

### Test 3.2: Valeurs de Tags
```go
func TestValidation_TagValues(t *testing.T) {
    outputDir := t.TempDir()
    files := generateAndOrganizeFiles(t, outputDir, 5, 1)

    firstImage := filepath.Join(outputDir, "PT000000", "ST000000", "SE000000", "IM000001")
    ds, err := dicom.ParseFile(firstImage, nil)
    assert.NoError(t, err)

    // Check Modality = "MR"
    modality, err := ds.FindElementByTag(tag.Modality)
    assert.NoError(t, err)
    assert.Equal(t, "MR", modality.Value.String())

    // Check SOP Class UID for MR Image Storage
    sopClassUID, err := ds.FindElementByTag(tag.SOPClassUID)
    assert.NoError(t, err)
    assert.Equal(t, "1.2.840.10008.5.1.4.1.1.4", sopClassUID.Value.String())

    // Check BitsAllocated = 16
    bitsAllocated, err := ds.FindElementByTag(tag.BitsAllocated)
    assert.NoError(t, err)
    assert.Equal(t, 16, bitsAllocated.Value.GetValue().(int))

    // Check Photometric Interpretation
    pi, err := ds.FindElementByTag(tag.PhotometricInterpretation)
    assert.NoError(t, err)
    assert.Equal(t, "MONOCHROME2", pi.Value.String())
}
```

### Test 3.3: Paramètres MRI
```go
func TestValidation_MRIParameters(t *testing.T) {
    outputDir := t.TempDir()
    files := generateAndOrganizeFiles(t, outputDir, 5, 1)

    firstImage := filepath.Join(outputDir, "PT000000", "ST000000", "SE000000", "IM000001")
    ds, err := dicom.ParseFile(firstImage, nil)
    assert.NoError(t, err)

    // Check MRI-specific tags exist
    mriTags := []tag.Tag{
        tag.Manufacturer,
        tag.ManufacturerModelName,
        tag.MagneticFieldStrength,
        tag.EchoTime,
        tag.RepetitionTime,
        tag.FlipAngle,
        tag.PixelSpacing,
        tag.SliceThickness,
    }

    for _, t := range mriTags {
        elem, err := ds.FindElementByTag(t)
        assert.NoError(t, err, "MRI tag %v should exist", t)
        assert.NotNil(t, elem)
    }

    // Check manufacturer is one of expected values
    mfr, _ := ds.FindElementByTag(tag.Manufacturer)
    mfrValue := mfr.Value.String()
    validMfrs := []string{"SIEMENS", "GE MEDICAL SYSTEMS", "PHILIPS"}
    assert.Contains(t, validMfrs, mfrValue)
}
```

### Test 3.4: Données Pixel
```go
func TestValidation_PixelData(t *testing.T) {
    outputDir := t.TempDir()
    files := generateAndOrganizeFiles(t, outputDir, 3, 1)

    firstImage := filepath.Join(outputDir, "PT000000", "ST000000", "SE000000", "IM000001")
    ds, err := dicom.ParseFile(firstImage, nil)
    assert.NoError(t, err)

    // Get pixel data
    pixelDataElem, err := ds.FindElementByTag(tag.PixelData)
    assert.NoError(t, err)

    pixelInfo := pixelDataElem.Value.(dicom.PixelDataInfo)
    assert.False(t, pixelInfo.IsEncapsulated)
    assert.Equal(t, 1, len(pixelInfo.Frames))

    frame := pixelInfo.Frames[0]
    assert.False(t, frame.Encapsulated)

    // Check pixel data exists and has correct size
    rows, _ := ds.FindElementByTag(tag.Rows)
    cols, _ := ds.FindElementByTag(tag.Columns)

    rowsVal := rows.Value.GetValue().(int)
    colsVal := cols.Value.GetValue().(int)

    expectedSize := rowsVal * colsVal * 2 // 2 bytes per pixel (16-bit)
    assert.Equal(t, expectedSize, len(frame.NativeData.Data))
}
```

---

## Test Suite 4: Reproductibilité

### Test 4.1: Seed Identique
```go
func TestReproducibility_SameSeed(t *testing.T) {
    seed := int64(42)

    // Generate first series
    outputDir1 := t.TempDir()
    opts1 := dicom.GeneratorOptions{
        NumImages:  5,
        TotalSize:  "10MB",
        OutputDir:  outputDir1,
        Seed:       seed,
        NumStudies: 1,
    }
    files1, err := dicom.GenerateDICOMSeries(opts1)
    assert.NoError(t, err)

    // Generate second series with same seed
    outputDir2 := t.TempDir()
    opts2 := opts1
    opts2.OutputDir = outputDir2
    files2, err := dicom.GenerateDICOMSeries(opts2)
    assert.NoError(t, err)

    // Compare UIDs (should be identical)
    assert.Equal(t, len(files1), len(files2))

    for i := 0; i < len(files1); i++ {
        assert.Equal(t, files1[i].PatientID, files2[i].PatientID)
        assert.Equal(t, files1[i].StudyUID, files2[i].StudyUID)
        assert.Equal(t, files1[i].SeriesUID, files2[i].SeriesUID)
        // Note: SOPInstanceUID depends on output dir, so will differ
    }
}
```

### Test 4.2: Seed Différent
```go
func TestReproducibility_DifferentSeed(t *testing.T) {
    outputDir1 := t.TempDir()
    opts1 := dicom.GeneratorOptions{
        NumImages:  5,
        TotalSize:  "10MB",
        OutputDir:  outputDir1,
        Seed:       42,
        NumStudies: 1,
    }
    files1, _ := dicom.GenerateDICOMSeries(opts1)

    outputDir2 := t.TempDir()
    opts2 := opts1
    opts2.OutputDir = outputDir2
    opts2.Seed = 99
    files2, _ := dicom.GenerateDICOMSeries(opts2)

    // UIDs should be different
    assert.NotEqual(t, files1[0].PatientID, files2[0].PatientID)
    assert.NotEqual(t, files1[0].StudyUID, files2[0].StudyUID)
    assert.NotEqual(t, files1[0].SeriesUID, files2[0].SeriesUID)
}
```

### Test 4.3: Auto-Seed depuis OutputDir
```go
func TestReproducibility_AutoSeedFromDir(t *testing.T) {
    outputDir := "test-series-123"

    // Generate twice with same dir name (no explicit seed)
    opts := dicom.GeneratorOptions{
        NumImages:  3,
        TotalSize:  "5MB",
        OutputDir:  filepath.Join(t.TempDir(), outputDir),
        NumStudies: 1,
    }
    files1, _ := dicom.GenerateDICOMSeries(opts)

    opts.OutputDir = filepath.Join(t.TempDir(), outputDir)
    files2, _ := dicom.GenerateDICOMSeries(opts)

    // Should produce same patient/study IDs
    assert.Equal(t, files1[0].PatientID, files2[0].PatientID)
    assert.Equal(t, files1[0].StudyUID, files2[0].StudyUID)
}
```

---

## Test Suite 5: Compatibilité Python

### Test 5.1: Comparaison UIDs
**Fichier**: `compatibility_test.go`

```go
func TestCompatibility_UIDGeneration(t *testing.T) {
    // Test UID generation algorithm matches Python
    testCases := []struct {
        seed     string
        expected string // UID from Python version
    }{
        {
            seed:     "test_study_1",
            expected: "1.2.826.0.1.3680043.8.498.xxx", // Fill from Python
        },
        {
            seed:     "test_study_2",
            expected: "1.2.826.0.1.3680043.8.498.yyy",
        },
    }

    for _, tc := range testCases {
        uid := util.GenerateDeterministicUID(tc.seed)

        // Check format matches
        assert.True(t, strings.HasPrefix(uid, "1.2.826.0.1.3680043.8.498."))
        assert.LessOrEqual(t, len(uid), 64)

        // TODO: Verify exact match with Python after cross-validation
        // assert.Equal(t, tc.expected, uid)
    }
}
```

### Test 5.2: Nom de Patient
```go
func TestCompatibility_PatientNames(t *testing.T) {
    // Set same seed as Python test
    rand.Seed(42)

    // Generate patient name
    name := util.GeneratePatientName("M")

    // Check format
    assert.Contains(t, name, "^")
    parts := strings.Split(name, "^")
    assert.Equal(t, 2, len(parts))

    // TODO: Compare with Python output for same seed
    // expectedFromPython := "DUPONT^Jean"
    // assert.Equal(t, expectedFromPython, name)
}
```

### Test 5.3: Structure de Métadonnées
```go
func TestCompatibility_MetadataStructure(t *testing.T) {
    // Generate with known seed
    outputDir := t.TempDir()
    opts := dicom.GeneratorOptions{
        NumImages:  3,
        TotalSize:  "5MB",
        OutputDir:  outputDir,
        Seed:       42,
        NumStudies: 1,
    }

    files, _ := dicom.GenerateDICOMSeries(opts)
    dicom.OrganizeFilesIntoDICOMDIR(outputDir, files)

    // Parse first file
    firstImage := filepath.Join(outputDir, "PT000000", "ST000000", "SE000000", "IM000001")
    ds, _ := dicom.ParseFile(firstImage, nil)

    // Extract key metadata
    metadata := extractKeyMetadata(ds)

    // TODO: Compare with Python-generated file
    // expectedMetadata := loadExpectedMetadata("testdata/python_reference.json")
    // compareMetadata(metadata, expectedMetadata)
}
```

---

## Test Suite 6: Performance

### Test 6.1: Temps de Génération
**Fichier**: `performance_test.go`

```go
func BenchmarkGenerateSeries_Small(b *testing.B) {
    for i := 0; i < b.N; i++ {
        outputDir := b.TempDir()
        opts := dicom.GeneratorOptions{
            NumImages:  10,
            TotalSize:  "50MB",
            OutputDir:  outputDir,
            Seed:       42,
            NumStudies: 1,
        }

        _, err := dicom.GenerateDICOMSeries(opts)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkGenerateSeries_Medium(b *testing.B) {
    for i := 0; i < b.N; i++ {
        outputDir := b.TempDir()
        opts := dicom.GeneratorOptions{
            NumImages:  50,
            TotalSize:  "500MB",
            OutputDir:  outputDir,
            Seed:       42,
            NumStudies: 1,
        }

        _, err := dicom.GenerateDICOMSeries(opts)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Test 6.2: Utilisation Mémoire
```go
func TestPerformance_MemoryUsage(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping memory test in short mode")
    }

    var m1, m2 runtime.MemStats
    runtime.ReadMemStats(&m1)

    outputDir := t.TempDir()
    opts := dicom.GeneratorOptions{
        NumImages:  100,
        TotalSize:  "1GB",
        OutputDir:  outputDir,
        Seed:       42,
        NumStudies: 1,
    }

    _, err := dicom.GenerateDICOMSeries(opts)
    assert.NoError(t, err)

    runtime.ReadMemStats(&m2)

    memUsed := m2.TotalAlloc - m1.TotalAlloc
    t.Logf("Memory used: %d MB", memUsed/(1024*1024))

    // Should not use excessive memory
    assert.Less(t, memUsed, uint64(2*1024*1024*1024)) // < 2GB
}
```

---

## Scripts de Support

### Script 1: Comparaison Python/Go
**Fichier**: `scripts/compare_outputs.sh`

```bash
#!/bin/bash
set -e

SEED=42
NUM_IMAGES=10
SIZE="100MB"

echo "=== Generating with Python ==="
cd /home/user/dicom-test
python generate_dicom_mri.py \
    --num-images $NUM_IMAGES \
    --total-size $SIZE \
    --output test-py \
    --seed $SEED

echo ""
echo "=== Generating with Go ==="
cd /home/user/dicom-test/go
./bin/generate-dicom-mri \
    --num-images $NUM_IMAGES \
    --total-size $SIZE \
    --output ../test-go \
    --seed $SEED

echo ""
echo "=== Comparing structures ==="
cd /home/user/dicom-test

# Compare directory structure
echo "Python structure:"
find test-py -type d | sort
echo ""
echo "Go structure:"
find test-go -type d | sort

# Count files
echo ""
echo "Python file count: $(find test-py -name 'IM*' | wc -l)"
echo "Go file count: $(find test-go -name 'IM*' | wc -l)"

# Extract and compare metadata
echo ""
echo "=== Extracting metadata ==="
python scripts/extract_metadata.py test-py > metadata-py.json
python scripts/extract_metadata.py test-go > metadata-go.json

echo "Comparing UIDs and patient info..."
python scripts/compare_metadata.py metadata-py.json metadata-go.json
```

### Script 2: Validation DICOM
**Fichier**: `scripts/validate_dicom.py`

```python
#!/usr/bin/env python3
"""Validate DICOM files generated by Go implementation."""

import sys
import os
import pydicom
from pathlib import Path

def validate_dicom_file(filepath):
    """Validate a single DICOM file."""
    errors = []

    try:
        ds = pydicom.dcmread(filepath)

        # Check required tags
        required_tags = [
            'PatientName', 'PatientID', 'StudyInstanceUID',
            'SeriesInstanceUID', 'SOPInstanceUID', 'Modality',
            'Rows', 'Columns', 'BitsAllocated'
        ]

        for tag in required_tags:
            if not hasattr(ds, tag):
                errors.append(f"Missing tag: {tag}")

        # Check values
        if hasattr(ds, 'Modality') and ds.Modality != 'MR':
            errors.append(f"Invalid Modality: {ds.Modality} (expected MR)")

        if hasattr(ds, 'BitsAllocated') and ds.BitsAllocated != 16:
            errors.append(f"Invalid BitsAllocated: {ds.BitsAllocated}")

        # Check pixel data
        if not hasattr(ds, 'PixelData'):
            errors.append("Missing PixelData")

    except Exception as e:
        errors.append(f"Parse error: {str(e)}")

    return errors

def main():
    if len(sys.argv) < 2:
        print("Usage: validate_dicom.py <dicom_directory>")
        sys.exit(1)

    dicom_dir = Path(sys.argv[1])

    # Find all DICOM files
    dicom_files = list(dicom_dir.rglob('IM*'))

    print(f"Validating {len(dicom_files)} DICOM files...")

    total_errors = 0
    for filepath in dicom_files:
        errors = validate_dicom_file(filepath)
        if errors:
            print(f"\n{filepath}:")
            for error in errors:
                print(f"  ❌ {error}")
            total_errors += len(errors)

    if total_errors == 0:
        print(f"\n✓ All {len(dicom_files)} files are valid!")
        return 0
    else:
        print(f"\n❌ Found {total_errors} errors in DICOM files")
        return 1

if __name__ == '__main__':
    sys.exit(main())
```

### Script 3: Extraction de Métadonnées
**Fichier**: `scripts/extract_metadata.py`

```python
#!/usr/bin/env python3
"""Extract metadata from DICOM files for comparison."""

import sys
import json
import pydicom
from pathlib import Path

def extract_metadata(dicom_dir):
    """Extract key metadata from all DICOM files."""
    dicom_dir = Path(dicom_dir)
    dicom_files = sorted(dicom_dir.rglob('IM*'))

    metadata = {
        'file_count': len(dicom_files),
        'files': []
    }

    for filepath in dicom_files:
        ds = pydicom.dcmread(filepath)

        file_meta = {
            'filename': str(filepath.relative_to(dicom_dir)),
            'patient_id': str(ds.PatientID),
            'patient_name': str(ds.PatientName),
            'study_uid': str(ds.StudyInstanceUID),
            'series_uid': str(ds.SeriesInstanceUID),
            'sop_instance_uid': str(ds.SOPInstanceUID),
            'instance_number': int(ds.InstanceNumber),
            'modality': str(ds.Modality),
            'rows': int(ds.Rows),
            'columns': int(ds.Columns),
        }

        # Add MRI parameters if present
        if hasattr(ds, 'Manufacturer'):
            file_meta['manufacturer'] = str(ds.Manufacturer)
        if hasattr(ds, 'EchoTime'):
            file_meta['echo_time'] = float(ds.EchoTime)

        metadata['files'].append(file_meta)

    return metadata

def main():
    if len(sys.argv) < 2:
        print("Usage: extract_metadata.py <dicom_directory>")
        sys.exit(1)

    metadata = extract_metadata(sys.argv[1])
    print(json.dumps(metadata, indent=2))

if __name__ == '__main__':
    main()
```

---

## Plan d'Exécution

### Phase 1: Tests Basiques (1-2 heures)
1. Implémenter Test Suite 1 (génération basique)
2. Implémenter Test Suite 2 (structure)
3. Exécuter et corriger erreurs

### Phase 2: Validation DICOM (2-3 heures)
4. Implémenter Test Suite 3 (validation tags)
5. Créer script de validation Python
6. Valider tous les fichiers générés

### Phase 3: Reproductibilité (1 heure)
7. Implémenter Test Suite 4 (reproductibilité)
8. Tester avec différents seeds
9. Vérifier auto-seed

### Phase 4: Compatibilité Python (2-3 heures)
10. Créer scripts de comparaison
11. Générer séries avec Python et Go (même seed)
12. Comparer UIDs et métadonnées
13. Documenter différences

### Phase 5: Performance (1 heure)
14. Implémenter benchmarks
15. Mesurer temps et mémoire
16. Comparer avec Python

### Phase 6: Documentation (1 heure)
17. Documenter résultats
18. Créer rapport de compatibilité
19. Mettre à jour README

**Total estimé**: 8-11 heures

---

## Critères de Succès

### Tests Unitaires
- [ ] ✅ Tous les tests passent
- [ ] ✅ Couverture > 80%

### Validation DICOM
- [ ] ✅ Tous les tags requis présents
- [ ] ✅ Valeurs dans ranges valides
- [ ] ✅ Fichiers lisibles par pydicom
- [ ] ✅ Structure PT*/ST*/SE* correcte

### Reproductibilité
- [ ] ✅ Même seed → Mêmes UIDs
- [ ] ✅ Même outputDir → Mêmes IDs
- [ ] ✅ Seeds différents → UIDs différents

### Compatibilité Python
- [ ] ✅ Même algorithme UID (vérification manuelle)
- [ ] ✅ Même format de noms patients
- [ ] ✅ Même structure de métadonnées
- [ ] ✅ Fichiers compatibles avec viewers DICOM

### Performance
- [ ] ✅ 10 images (100MB): < 5 secondes
- [ ] ✅ 120 images (1GB): < 30 secondes
- [ ] ✅ Mémoire < 2GB pour 1GB de données

---

## Prochaines Étapes

1. **Créer les fichiers de test** dans `go/tests/`
2. **Implémenter les tests** suite par suite
3. **Créer les scripts de support** dans `go/scripts/`
4. **Exécuter la validation complète**
5. **Documenter les résultats** dans `TESTING_REPORT.md`
6. **Corriger les bugs découverts**
7. **Optimiser les performances** si nécessaire

## Notes

- Les tests d'intégration nécessitent que le binaire soit compilé
- Les tests de compatibilité nécessitent Python + pydicom installé
- Les benchmarks doivent être exécutés sur une machine non chargée
- Documenter toutes les différences avec Python pour décider si acceptables
