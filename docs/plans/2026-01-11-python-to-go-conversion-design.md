# Design: Conversion du Générateur DICOM MRI de Python vers Go

**Date**: 2026-01-11
**Objectif**: Transcription complète du programme Python `generate_dicom_mri.py` vers Go, en conservant toutes les fonctionnalités et en garantissant la compatibilité avec les plateformes médicales.

## Contexte

Le programme Python actuel génère des fichiers DICOM d'IRM valides pour tester des plateformes médicales. Il fonctionne bien mais doit être transcrit en Go, le langage standard de la société.

### Fonctionnalités à Préserver

- Génération de séries DICOM multi-fichiers (un fichier par image)
- Support de multiples études (studies)
- Métadonnées MRI réalistes (SIEMENS, GE, PHILIPS)
- Overlay texte "File X/Y" sur chaque image
- Génération déterministe avec seed
- Création de DICOMDIR avec hiérarchie PT*/ST*/SE*
- UIDs déterministes pour reproductibilité
- Noms de patients français réalistes
- Support tailles KB/MB/GB

### Contraintes

- Code Python original **préservé** (pas de suppression)
- Nouveau code Go dans répertoire `go/`
- Compatibilité totale avec output Python (même seed = même résultat)
- Aucune dépendance système externe (implémentation DICOMDIR pure Go)

## Architecture

### Structure des Fichiers

```
go/
├── cmd/
│   └── generate-dicom-mri/
│       └── main.go                 # Point d'entrée CLI
├── internal/
│   ├── dicom/
│   │   ├── metadata.go            # Génération métadonnées DICOM
│   │   ├── generator.go           # Génération images DICOM
│   │   └── dicomdir.go            # Création DICOMDIR
│   ├── image/
│   │   ├── pixel.go               # Génération pixel data
│   │   └── overlay.go             # Texte overlay sur images
│   └── util/
│       ├── size.go                # Parse taille (KB/MB/GB)
│       ├── uid.go                 # Génération UIDs déterministes
│       └── names.go               # Génération noms patients
├── go.mod
├── go.sum
├── README.md
└── integration_test.go
```

### Dépendances Go

**Bibliothèques externes**:
- `github.com/suyashkumar/dicom` - Manipulation DICOM (lecture/écriture)
- `github.com/golang/freetype` - Rendu texte sur images
- `golang.org/x/image/font` - Support polices

**Bibliothèque standard**:
- `flag` - Parsing arguments CLI
- `crypto/sha256` - Génération UIDs déterministes
- `math/rand` - Génération aléatoire reproductible
- `regexp` - Parsing taille (KB/MB/GB)
- `image` - Manipulation images
- `os`, `path/filepath`, `io` - Opérations fichiers

## Composants Détaillés

### 1. CLI et Point d'Entrée (`cmd/generate-dicom-mri/main.go`)

**Arguments CLI** (package `flag`):
- `--num-images` (int, requis) - Nombre d'images/coupes
- `--total-size` (string, requis) - Taille totale (ex: "1GB")
- `--output` (string, défaut: "dicom_series") - Dossier de sortie
- `--seed` (int, optionnel) - Seed pour reproductibilité
- `--num-studies` (int, défaut: 1) - Nombre d'études

**Validations**:
- `num-images > 0`
- `num-studies > 0`
- `num-studies <= num-images`
- Format taille valide

**Logique principale**:
1. Parse et valide arguments
2. Parse taille totale en bytes
3. Vérifie espace disque disponible
4. Calcule dimensions optimales (width, height)
5. Génère ou utilise seed fourni
6. Génère informations patient partagées
7. Pour chaque study:
   - Génère UIDs déterministes
   - Génère paramètres série
   - Pour chaque image:
     - Génère métadonnées
     - Génère pixel data avec overlay
     - Écrit fichier .dcm
8. Crée DICOMDIR et hiérarchie
9. Nettoie fichiers temporaires

### 2. Parsing de Taille (`internal/util/size.go`)

**Fonction**: `ParseSize(sizeStr string) (int64, error)`

**Implémentation**:
```go
// Pattern regex: (\d+(?:\.\d+)?)(KB|MB|GB)
// Exemple: "1.5GB" → 1610612736 bytes
```

**Multiplicateurs**:
- KB: 1024
- MB: 1024 * 1024
- GB: 1024 * 1024 * 1024

**Gestion erreurs**:
- Format invalide → error
- Unité non supportée → error

### 3. Génération UIDs Déterministes (`internal/util/uid.go`)

**Fonction**: `GenerateDeterministicUID(seed string) string`

**Algorithme** (identique au Python):
1. Prefix DICOM: `1.2.826.0.1.3680043.8.498`
2. Hash SHA256 du seed string
3. Convertir hash en numérique (premiers 30 chars hex)
4. Découper en segments de 10 chiffres
5. Supprimer leading zeros (sauf "0")
6. Limiter à 3 segments après prefix
7. Assurer longueur totale ≤ 64 caractères

**Garantie**: Même seed → Même UID (compatibilité Python/Go)

### 4. Génération Noms Patients (`internal/util/names.go`)

**Fonction**: `GeneratePatientName(sex string) string`

**Données** (identiques au Python):
- Prénoms masculins (38 prénoms français)
- Prénoms féminins (36 prénoms français)
- Noms de famille (57 noms français)

**Format sortie**: `LASTNAME^FIRSTNAME` (format DICOM)

**Sélection**: Utilise `math/rand` avec seed pour reproductibilité

### 5. Génération Métadonnées DICOM (`internal/dicom/metadata.go`)

**Structure**: `MetadataOptions`
```go
type MetadataOptions struct {
    NumImages           int
    Width, Height       int
    InstanceNumber      int

    // Shared across series
    StudyUID            string
    SeriesUID           string
    PatientID           string
    PatientName         string
    PatientBirthDate    string
    PatientSex          string
    StudyDate           string
    StudyTime           string
    StudyID             string
    StudyDescription    string
    AccessionNumber     string
    SeriesNumber        int

    // MRI parameters (shared across series)
    PixelSpacing              float64
    SliceThickness            float64
    SpacingBetweenSlices      float64
    EchoTime                  float64
    RepetitionTime            float64
    FlipAngle                 float64
    SequenceName              string
    Manufacturer              string
    Model                     string
    FieldStrength             float64
}
```

**Fonction**: `GenerateMetadata(opts MetadataOptions) (*dicom.Dataset, error)`

**Tags DICOM créés** (via `dicom.NewElement`):

*Patient Information*:
- (0010,0010) Patient Name
- (0010,0020) Patient ID
- (0010,0030) Patient Birth Date
- (0010,0040) Patient Sex

*Study Information*:
- (0020,000D) Study Instance UID
- (0008,0020) Study Date
- (0008,0030) Study Time
- (0020,0010) Study ID
- (0008,1030) Study Description
- (0008,0050) Accession Number

*Series Information*:
- (0020,000E) Series Instance UID
- (0020,0011) Series Number
- (0008,103E) Series Description
- (0008,0060) Modality = "MR"

*Instance Information*:
- (0020,0013) Instance Number
- (0008,0018) SOP Instance UID
- (0008,0016) SOP Class UID = "1.2.840.10008.5.1.4.1.1.4" (MR Image Storage)

*MRI Parameters*:
- (0018,0050) Slice Thickness
- (0018,0088) Spacing Between Slices
- (0028,0030) Pixel Spacing
- (0018,0081) Echo Time
- (0018,0080) Repetition Time
- (0018,1314) Flip Angle
- (0018,0024) Sequence Name
- (0008,0070) Manufacturer
- (0008,1090) Manufacturer Model Name
- (0018,0087) Magnetic Field Strength
- (0018,0084) Imaging Frequency (calculé: fieldStrength * 42.58 MHz)

*Image Parameters*:
- (0028,0002) Samples Per Pixel = 1
- (0028,0004) Photometric Interpretation = "MONOCHROME2"
- (0028,0010) Rows
- (0028,0011) Columns
- (0028,0100) Bits Allocated = 16
- (0028,0101) Bits Stored = 16
- (0028,0102) High Bit = 15
- (0028,0103) Pixel Representation = 0 (unsigned)

*Position et Orientation*:
- (0020,0032) Image Position Patient = [0, 0, slicePosition]
- (0020,0037) Image Orientation Patient = [1, 0, 0, 0, 1, 0] (axial)
- (0020,1041) Slice Location

*Display Parameters*:
- (0028,1050) Window Center = "2048"
- (0028,1051) Window Width = "4096"
- (0028,1055) Window Center & Width Explanation = "Full Range"
- (0028,1052) Rescale Intercept = "0"
- (0028,1053) Rescale Slope = "1"
- (0028,1054) Rescale Type = "US"

*File Meta*:
- (0002,0010) Transfer Syntax UID = ExplicitVRLittleEndian
- (0002,0002) Media Storage SOP Class UID
- (0002,0003) Media Storage SOP Instance UID
- (0002,0012) Implementation Class UID

**Character Set**: ISO_IR 192 (UTF-8) pour accents français

### 6. Génération Pixel Data (`internal/image/pixel.go`)

**Fonction**: `GenerateSingleImage(width, height int, seed int64) []uint16`

**Implémentation**:
- Utilise `math/rand` avec seed
- Génère pixels aléatoires range 12-bit (0-4095, typique MRI)
- Retourne slice `[]uint16` de taille `width * height`

**Fonction**: `AddTextOverlay(pixels []uint16, width, height, imageNum, totalImages int, font *truetype.Font) error`

**Implémentation**:
1. Convertir `[]uint16` vers `image.Gray16`
2. Scale 0-4095 → 0-65535 pour meilleur contraste
3. Convertir vers RGB pour drawing
4. Utiliser `freetype.Context` pour dessiner texte
5. Texte: `fmt.Sprintf("File %d/%d", imageNum, totalImages)`
6. Position: centré horizontalement, 5% du haut
7. Style: texte blanc avec outline noir épais (3-5px)
8. Convertir retour vers grayscale
9. Scale retour 0-4095
10. Clip pour garantir range 12-bit

**Police**: Charger une seule fois au démarrage, réutiliser
- Taille: `width / 16` (proportionnel à image)
- Paths: DejaVuSans-Bold.ttf (Linux standard paths)

### 7. Création DICOMDIR (`internal/dicom/dicomdir.go`)

**Implémentation manuelle** (pas de bibliothèque disponible pour création)

#### Structure de Données

```go
type DirectoryNode struct {
    RecordType string              // "PATIENT", "STUDY", "SERIES", "IMAGE"
    Tags       map[uint32]interface{} // Tag → Valeur
    Children   []*DirectoryNode
    Next       *DirectoryNode
    Offset     uint32              // Calculé en passe 2
}
```

#### Fonction Principale

**Fonction**: `CreateDICOMDIR(outputDir string, files []DICOMFileInfo) error`

**Étapes**:
1. Lire tous les fichiers .dcm pour extraire métadonnées
2. Construire arbre logique (Patient→Study→Series→Image)
3. Organiser hiérarchie physique fichiers
4. Calculer offsets Directory Records
5. Construire fichier DICOMDIR
6. Écrire à la racine output directory

#### Construction Arbre Logique (Passe 1)

**Fonction**: `buildDirectoryTree(files []DICOMFileInfo) *DirectoryNode`

**Algorithme**:
1. Grouper par PatientID → Patients
2. Pour chaque Patient, grouper par StudyUID → Studies
3. Pour chaque Study, grouper par SeriesUID → Series
4. Pour chaque Series, lister Images (triées par InstanceNumber)
5. Construire chaînage Next (siblings) et Children (hiérarchie)

#### Directory Records Tags

**PATIENT Record**:
- (0004,1400) Offset of Next Directory Record (uint32)
- (0004,1410) Record In-use Flag (uint16) = 0xFFFF
- (0004,1420) Offset of Referenced Lower-Level (uint32)
- (0004,1430) Directory Record Type = "PATIENT"
- (0010,0010) Patient Name
- (0010,0020) Patient ID

**STUDY Record**:
- Tags structurels +
- (0004,1430) = "STUDY"
- (0020,000D) Study Instance UID
- (0008,0020) Study Date
- (0008,0030) Study Time
- (0008,0050) Accession Number
- (0008,1030) Study Description

**SERIES Record**:
- Tags structurels +
- (0004,1430) = "SERIES"
- (0020,000E) Series Instance UID
- (0008,0060) Modality
- (0020,0011) Series Number

**IMAGE Record**:
- Tags structurels +
- (0004,1430) = "IMAGE"
- (0004,1500) Referenced File ID (chemin relatif)
- (0004,1510) Referenced SOP Class UID
- (0004,1511) Referenced SOP Instance UID
- (0020,0013) Instance Number

#### Calcul Offsets (Passe 2)

**Fonction**: `calculateOffsets(root *DirectoryNode) uint32`

**Algorithme**:
1. Commencer offset après File Meta + Directory Record Sequence header
2. Parcourir arbre depth-first
3. Pour chaque node:
   - Calculer taille encodée des tags
   - Assigner offset courant
   - Accumuler pour prochain record
4. Mettre à jour tags (0004,1400) et (0004,1420) avec offsets calculés

**Taille d'un Record**: Somme des tailles de tous ses éléments DICOM encodés

#### Organisation Hiérarchie Fichiers

**Fonction**: `OrganizeFileHierarchy(outputDir string, tree *DirectoryNode) error`

**Convention nommage**:
- Patient: `PT` + 6 chiffres zero-padded (PT000000, PT000001, ...)
- Study: `ST` + 6 chiffres (ST000000, ST000001, ...)
- Series: `SE` + 6 chiffres (SE000000, SE000001, ...)
- Image: `IM` + 6 chiffres (IM000001, IM000002, ...)

**Structure résultante**:
```
output_dir/
├── DICOMDIR
├── PT000000/
│   ├── ST000000/
│   │   └── SE000000/
│   │       ├── IM000001
│   │       ├── IM000002
│   │       └── ...
│   └── ST000001/
│       └── SE000000/
│           └── ...
└── PT000001/
    └── ...
```

**Opération**: Copier ou déplacer fichiers IMG*.dcm vers hiérarchie

#### File Meta du DICOMDIR

**Tags spéciaux**:
- (0002,0001) File Meta Information Version = [0, 1]
- (0002,0002) Media Storage SOP Class UID = "1.2.840.10008.1.3.10" (Media Storage Directory Storage)
- (0002,0003) Media Storage SOP Instance UID = généré
- (0002,0010) Transfer Syntax UID = ExplicitVRLittleEndian
- (0002,0012) Implementation Class UID = généré
- (0004,1130) File-set ID = nom output directory
- (0004,1141) File-set Descriptor File ID (optionnel)
- (0004,1200) Offset of First Directory Record
- (0004,1202) Offset of Last Directory Record
- (0004,1212) File-set Consistency Flag = 0 (unknown)

### 8. Générateur Principal (`internal/dicom/generator.go`)

**Fonction**: `GenerateSeries(opts GeneratorOptions) error`

**Structure**: `GeneratorOptions`
```go
type GeneratorOptions struct {
    NumImages    int
    TotalSize    string
    OutputDir    string
    Seed         int64
    NumStudies   int
}
```

**Orchestration**:
1. Parse taille totale
2. Calcule dimensions (width, height)
3. Crée output directory
4. Génère seed si non fourni (hash du outputDir)
5. Seed RNG global
6. Génère info patient partagée
7. Load police pour overlay
8. Pour chaque study:
   - Génère UIDs déterministes
   - Génère paramètres série
   - Pour chaque image:
     - Crée métadonnées
     - Génère pixel data
     - Ajoute overlay
     - Écrit fichier
     - Progress indicator
9. Collecte infos fichiers générés
10. Crée DICOMDIR
11. Organise hiérarchie
12. Nettoie fichiers temporaires

**Calcul dimensions** (fonction `CalculateDimensions`):
```go
func CalculateDimensions(totalBytes int64, numImages int) (width, height int, err error)
```

**Algorithme**:
1. Soustraire metadata overhead (100 KB)
2. Available bytes pour pixel data
3. Limiter à max DICOM (2^32 - 10MB ≈ 4.28 GB)
4. Calculer pixels totaux: availableBytes / 2 (uint16)
5. Pixels par frame: totalPixels / numImages
6. Dimension: sqrt(pixelsPerFrame)
7. Arrondir DOWN au multiple de 256 (ou 128 si < 256)
8. Minimum: 128x128

**Important**: Arrondir DOWN pour ne jamais dépasser taille cible

## Tests

### Tests Unitaires

**`internal/util/size_test.go`**:
- `TestParseSize_Valid`: "100MB", "1.5GB", "500KB"
- `TestParseSize_Invalid`: "100", "1.5TB", "abc"
- `TestParseSize_EdgeCases`: "0MB", "0.5KB"

**`internal/util/uid_test.go`**:
- `TestGenerateDeterministicUID_Consistency`: Même seed → Même UID
- `TestGenerateDeterministicUID_Different`: Seeds différents → UIDs différents
- `TestGenerateDeterministicUID_Length`: ≤ 64 chars
- `TestGenerateDeterministicUID_NoLeadingZeros`: Segments valides

**`internal/util/names_test.go`**:
- `TestGeneratePatientName_Format`: Format "LASTNAME^FIRSTNAME"
- `TestGeneratePatientName_Deterministic`: Même seed → Même nom
- `TestGeneratePatientName_Sex`: Vérifie prénoms selon sexe

**`internal/dicom/metadata_test.go`**:
- `TestGenerateMetadata_RequiredTags`: Tous tags requis présents
- `TestGenerateMetadata_MRIParameters`: Valeurs dans ranges valides
- `TestGenerateMetadata_InstanceNumber`: Correct assignment
- `TestGenerateMetadata_Dimensions`: Width/Height corrects

**`internal/dicom/dicomdir_test.go`**:
- `TestBuildDirectoryTree_Structure`: Hiérarchie correcte
- `TestCalculateOffsets_Validity`: Offsets non-nuls et croissants
- `TestDirectoryRecord_Tags`: Tags requis pour chaque type
- `TestOrganizeFileHierarchy_Naming`: Conventions PT*/ST*/SE*

**`internal/image/pixel_test.go`**:
- `TestGenerateSingleImage_Range`: Pixels dans 0-4095
- `TestGenerateSingleImage_Size`: Dimensions correctes
- `TestGenerateSingleImage_Deterministic`: Même seed → Même image

**`internal/image/overlay_test.go`**:
- `TestAddTextOverlay_Presence`: Texte détectable dans image
- `TestAddTextOverlay_Range`: Pixels restent dans 0-4095

### Tests d'Intégration

**`integration_test.go`** (racine `go/`):

**Test 1**: Génération série simple
```go
func TestGenerateSeries_Basic(t *testing.T) {
    // Générer 10 images, 100MB, seed fixe
    // Vérifier:
    // - 10 fichiers .dcm créés
    // - DICOMDIR existe
    // - Hiérarchie PT*/ST*/SE* correcte
    // - Taille totale ≈ 100MB
}
```

**Test 2**: Multiples études
```go
func TestGenerateSeries_MultiStudy(t *testing.T) {
    // Générer 30 images, 3 studies
    // Vérifier:
    // - 3 dossiers ST* sous PT000000
    // - DICOMDIR contient 3 STUDY records
    // - Répartition correcte (10 images par study)
}
```

**Test 3**: Compatibilité Python
```go
func TestGenerateSeries_PythonCompatibility(t *testing.T) {
    // Même seed que Python (ex: 42)
    // Vérifier:
    // - Mêmes StudyUID, SeriesUID
    // - Même PatientID, PatientName
    // - UIDs identiques entre Python et Go
}
```

**Test 4**: Validation DICOM
```go
func TestGenerateSeries_DICOMValidity(t *testing.T) {
    // Générer série
    // Parser avec suyashkumar/dicom
    // Vérifier:
    // - Tags requis présents
    // - Valeurs dans ranges valides
    // - Transfer syntax correct
}
```

### Validation Croisée Python/Go

**Script de comparaison** (`scripts/compare_outputs.sh`):
```bash
#!/bin/bash

SEED=42
NUM_IMAGES=10
SIZE="100MB"

# Générer avec Python
python generate_dicom_mri.py \
    --num-images $NUM_IMAGES \
    --total-size $SIZE \
    --output test-py \
    --seed $SEED

# Générer avec Go
go run go/cmd/generate-dicom-mri/main.go \
    --num-images $NUM_IMAGES \
    --total-size $SIZE \
    --output test-go \
    --seed $SEED

# Extraire et comparer métadonnées clés
python scripts/extract_metadata.py test-py > metadata-py.txt
python scripts/extract_metadata.py test-go > metadata-go.txt

diff metadata-py.txt metadata-go.txt
```

**Métadonnées à comparer**:
- StudyInstanceUID
- SeriesInstanceUID
- PatientID
- PatientName
- SOPInstanceUID (pour chaque image)

### Tests de Performance

**Benchmarks Go** (`*_bench_test.go`):
```go
func BenchmarkGenerateSingleImage(b *testing.B) {
    // Mesurer temps génération 1 image 512x512
}

func BenchmarkGenerateMetadata(b *testing.B) {
    // Mesurer temps création métadonnées
}

func BenchmarkCreateDICOMDIR(b *testing.B) {
    // Mesurer temps création DICOMDIR (100 images)
}
```

**Objectifs performance** (comparaison avec Python):
- 10 images (100MB): < 5 secondes
- 50 images (500MB): 5-15 secondes
- 120 images (1GB): 15-30 secondes

**Attendu**: Go devrait être 2-5x plus rapide que Python pour génération pixel data et I/O.

## Plan d'Implémentation

### Phase 1: Fondations (2-3 jours)

1. Setup projet Go
   - Initialiser `go.mod`
   - Installer dépendances
   - Créer structure dossiers

2. Implémentation `internal/util`
   - `size.go` + tests
   - `uid.go` + tests
   - `names.go` + tests

3. Validation tests unitaires util

### Phase 2: Génération Images (3-4 jours)

4. Implémentation `internal/image`
   - `pixel.go` + tests
   - `overlay.go` + tests
   - Chargement polices

5. Implémentation `internal/dicom/metadata.go`
   - Construction tags DICOM
   - Tests validation tags

6. Tests d'intégration génération simple (sans DICOMDIR)

### Phase 3: DICOMDIR (4-5 jours)

7. Implémentation `internal/dicom/dicomdir.go`
   - Construction arbre logique
   - Calcul offsets
   - Création Directory Records
   - Tests unitaires

8. Implémentation organisation hiérarchie fichiers
   - Création PT*/ST*/SE*
   - Copie/déplacement fichiers
   - Tests

9. Écriture fichier DICOMDIR complet
   - File Meta
   - Directory Record Sequence
   - Validation avec parseur

### Phase 4: CLI et Intégration (2-3 jours)

10. Implémentation `cmd/generate-dicom-mri/main.go`
    - Parsing arguments
    - Orchestration complète
    - Progress indicators
    - Gestion erreurs

11. Implémentation `internal/dicom/generator.go`
    - Calcul dimensions
    - Boucle génération études
    - Intégration tous composants

12. Tests d'intégration complets

### Phase 5: Validation et Documentation (2-3 jours)

13. Validation croisée Python/Go
    - Script comparaison
    - Tests reproductibilité
    - Correction discordances

14. Benchmarks performance
    - Mesures temps exécution
    - Comparaison Python/Go
    - Optimisations si nécessaire

15. Documentation
    - README Go avec exemples
    - Commentaires GoDoc
    - Guide migration Python→Go

**Total estimé**: 13-18 jours

## Risques et Mitigations

### Risque 1: Complexité DICOMDIR

**Probabilité**: Haute
**Impact**: Moyen

**Description**: Implémentation manuelle DICOMDIR complexe, risque bugs offsets

**Mitigation**:
- Tests exhaustifs avec multiples configurations
- Validation avec parseurs DICOM existants
- Comparaison byte-par-byte avec output pydicom
- Documentation détaillée spec DICOM Part 10

### Risque 2: Incompatibilité UIDs

**Probabilité**: Moyenne
**Impact**: Haut

**Description**: Algorithme Go génère UIDs différents de Python pour même seed

**Mitigation**:
- Tests unitaires comparaison directe Python/Go
- Même algorithme hash (SHA256)
- Validation chaque étape conversion
- Script validation automatique

### Risque 3: Différences Pixel Data

**Probabilité**: Moyenne
**Impact**: Faible

**Description**: RNG Go vs Python peuvent donner résultats différents même avec seed

**Mitigation**:
- Accepter différences pixel data (pas critique pour tests)
- Focus compatibilité métadonnées (StudyUID, SeriesUID, PatientID)
- Ou: implémenter même algorithme RNG si nécessaire

### Risque 4: Performance DICOMDIR

**Probabilité**: Faible
**Impact**: Faible

**Description**: Création DICOMDIR manuelle peut être lente pour grandes séries

**Mitigation**:
- Optimiser calcul offsets (single pass si possible)
- Utiliser buffering I/O efficace
- Benchmarks et profiling
- Parallélisation si nécessaire (goroutines)

### Risque 5: Polices Manquantes

**Probabilité**: Moyenne
**Impact**: Faible

**Description**: Police TrueType absente sur système, overlay texte échoue

**Mitigation**:
- Fallback vers police default
- Embarquer police dans binaire (embed)
- Documentation prérequis
- Tests avec/sans police

## Critères de Succès

### Fonctionnel

- [ ] Tous arguments CLI Python supportés
- [ ] Génération DICOMDIR avec hiérarchie PT*/ST*/SE*
- [ ] Support multiples études
- [ ] Overlay texte sur images
- [ ] Reproductibilité avec seed
- [ ] UIDs déterministes identiques à Python
- [ ] Noms patients français réalistes

### Qualité

- [ ] Tous tests unitaires passent (>90% coverage)
- [ ] Tests d'intégration passent
- [ ] Validation croisée Python/Go réussie
- [ ] DICOMDIR validé par parseur externe
- [ ] Code documenté (GoDoc)
- [ ] Pas de dépendance système externe (dcmtk)

### Performance

- [ ] 10 images (100MB): < 5 secondes
- [ ] 120 images (1GB): < 30 secondes
- [ ] Performance ≥ Python (idéalement 2-5x plus rapide)

### Compatibilité

- [ ] Fichiers DICOM importables dans plateformes médicales
- [ ] DICOMDIR reconnu par viewers DICOM (Weasis, etc.)
- [ ] Même seed → Mêmes UIDs/IDs que Python
- [ ] Format strictement conforme DICOM Part 10

## Références

**Standards DICOM**:
- DICOM Part 3: Information Object Definitions
- DICOM Part 5: Data Structures and Encoding
- DICOM Part 10: Media Storage and File Format for Media Interchange

**Bibliothèques**:
- [suyashkumar/dicom](https://github.com/suyashkumar/dicom) - Bibliothèque DICOM Go
- [pydicom](https://pydicom.github.io/) - Référence Python

**Code source**:
- `generate_dicom_mri.py` - Version Python existante (référence)
- `tests/test_*.py` - Tests Python à reproduire en Go
