# DICOM MRI Generator - Design Document

**Date:** 2026-01-09
**Status:** Validated

## Overview

Script Python CLI pour générer des fichiers DICOM d'IRM multi-frame valides avec taille et nombre d'images paramétrables, destiné au test d'interfaces médicales.

## Requirements

- Générer un fichier DICOM multi-frame unique (série IRM)
- Paramètres: nombre d'images et taille totale du fichier
- Images avec bruit aléatoire (pixels aléatoires)
- Métadonnées DICOM complètes et réalistes
- Utilisation de pydicom

## Architecture

### Interface CLI

```bash
python generate_dicom_mri.py --num-images 120 --total-size 4.5GB --output mri_test.dcm
```

**Paramètres:**
- `--num-images` (requis): Nombre d'images/coupes dans la série
- `--total-size` (requis): Taille totale cible (accepte KB, MB, GB)
- `--output` (optionnel): Nom du fichier de sortie (défaut: `generated_mri.dcm`)
- `--seed` (optionnel): Seed pour la génération aléatoire (reproductibilité)

### Dépendances

- `pydicom`: Création et manipulation de fichiers DICOM
- `numpy`: Génération des données d'images
- `pillow`: Manipulation d'images si nécessaire
- Bibliothèque standard Python: argparse, pathlib, etc.

### Calcul automatique de résolution

1. Parser la taille cible en octets (ex: "4.5GB" → 4,831,838,208 octets)
2. Soustraire l'overhead des métadonnées DICOM (~100KB)
3. Calculer la taille disponible par image
4. Déterminer une résolution carrée en 16-bit
5. Ajuster pour obtenir des dimensions réalistes (multiples de 256 proches)

**Exemple pour 4.5GB / 120 images:**
- ~37.5 MB par frame de données pures
- ~19M pixels par frame (37.5MB / 2 bytes)
- Résolution: ~4360x4360 pixels

## Métadonnées DICOM

### Tags obligatoires (Patient/Study/Series)

- Patient ID, Name (générés aléatoirement)
- Patient Birth Date, Sex
- Study Instance UID, Study Date, Study Time
- Series Instance UID, Series Number, Series Description
- Modality: "MR" (Magnetic Resonance)

### Tags spécifiques IRM (réalistes)

- Manufacturer: "SIEMENS", "GE MEDICAL SYSTEMS", ou "PHILIPS"
- Manufacturer Model Name: "Avanto", "Signa HDxt", etc.
- Magnetic Field Strength: 1.5T ou 3.0T
- Imaging Frequency: 63.87 MHz (pour 1.5T) ou équivalent
- Echo Time (TE), Repetition Time (TR): valeurs typiques pour T1/T2
- Flip Angle, Slice Thickness, Spacing Between Slices
- Sequence Name: "T1_MPRAGE", "T2_FSE", etc.

### Tags multi-frame

- Number of Frames: paramètre `--num-images`
- Frame Increment Pointer
- Pixel Data avec toutes les frames concaténées

### Génération des UIDs

Utilisation de `pydicom.uid.generate_uid()` pour garantir l'unicité des UIDs DICOM.

## Génération des données d'image

### Format des pixels

- Type: 16-bit unsigned integers (uint16) - standard pour l'IRM
- Plage de valeurs: 0-4095 (12-bit dynamique typique)
- Photometric Interpretation: "MONOCHROME2" (noir = faible signal, blanc = fort signal)

### Génération du bruit aléatoire

```python
# Pour chaque frame
frame_data = np.random.randint(0, 4096, size=(height, width), dtype=np.uint16)
```

Distribution uniforme. Si seed fourni, utiliser `np.random.seed(seed)` pour reproductibilité.

### Algorithme de calcul de taille

1. Parser taille cible (ex: "4.5GB" → octets)
2. Estimer overhead DICOM (~100KB)
3. Taille disponible pour pixels = taille_cible - overhead
4. Taille par pixel = 2 octets (uint16)
5. Total pixels nécessaires = taille_disponible / 2
6. Pixels par frame = total_pixels / num_images
7. Calculer dimensions: `dim = int(sqrt(pixels_par_frame))`
8. Ajuster pour dimensions réalistes (arrondir à multiples de 256)

### Validation

Afficher résolution calculée et taille estimée finale avant génération.

## Performance et gestion des erreurs

### Performance

- Génération frame par frame et écriture progressive (éviter de charger 4.5GB en mémoire)
- Buffer pour accumuler frames puis écrire par blocs
- Barre de progression ou pourcentage pendant génération
- Temps estimé: ~30-60 secondes pour 4.5GB

### Gestion des erreurs

- Validation des paramètres (taille > 0, num_images > 0)
- Vérification de l'espace disque disponible
- Gestion des erreurs d'écriture (permissions, disque plein)
- Messages d'erreur clairs en français

### Output du script

```
Calculating optimal resolution...
Resolution: 4360x4360 pixels per frame
Expected file size: ~4.52 GB (120 frames)
Generating DICOM file...
Progress: [████████████████████] 100% (120/120 frames)
DICOM file created: mri_test.dcm
Actual size: 4.51 GB
```

## Structure du code

### Fonctions principales

1. `parse_arguments()`: Parse les arguments CLI
2. `parse_size(size_str)`: Convertit "4.5GB" en octets
3. `calculate_dimensions(total_size, num_images)`: Calcule résolution optimale
4. `generate_metadata(num_images, width, height)`: Crée dataset DICOM avec métadonnées
5. `generate_pixel_data(num_images, width, height, seed)`: Génère frames de bruit
6. `main()`: Orchestre le processus complet

### Documentation

- Docstrings pour chaque fonction
- README.md avec exemples d'utilisation
- requirements.txt pour les dépendances

## Files à créer

```
dicom-test/
├── generate_dicom_mri.py    # Script principal
├── requirements.txt          # Dépendances Python
├── README.md                 # Documentation utilisateur
└── docs/
    └── plans/
        └── 2026-01-09-dicom-mri-generator-design.md  # Ce document
```

## Validation

Design validé le 2026-01-09. Prêt pour implémentation.
