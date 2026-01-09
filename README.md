# DICOM MRI Generator

Outil CLI Python pour générer des séries DICOM d'IRM valides pour tester des interfaces médicales.

**Génère plusieurs fichiers DICOM** (un par image) dans un dossier, format standard attendu par les plateformes médicales.

## Installation

```bash
pip install -r requirements.txt
```

## Usage

```bash
python generate_dicom_mri.py --num-images 120 --total-size 1GB --output mri_series
```

Cela créera un dossier `mri_series/` contenant 120 fichiers DICOM individuels + fichier DICOMDIR:
```
mri_series/
├── DICOMDIR       # Fichier d'index de la série
├── IMG0001.dcm
├── IMG0002.dcm
├── ...
└── IMG0120.dcm
```

### Paramètres

- `--num-images` (requis): Nombre d'images/coupes dans la série
- `--total-size` (requis): Taille totale cible pour tous les fichiers (KB, MB, GB)
- `--output` (optionnel): Nom du dossier de sortie (défaut: `dicom_series`)
- `--seed` (optionnel): Seed pour reproductibilité

### Exemples

```bash
# Générer 120 images pour 1 GB total
python generate_dicom_mri.py --num-images 120 --total-size 1GB

# Avec nom de dossier personnalisé et seed
python generate_dicom_mri.py --num-images 50 --total-size 500MB --output my_mri --seed 42

# Série IRM complète pour test plateforme médicale
python generate_dicom_mri.py --num-images 120 --total-size 1GB --output test_patient_001
```

## Utilisation avec Plateforme Médicale

Après génération, importez **l'intégralité du dossier** dans votre plateforme:

1. Générez la série: `python generate_dicom_mri.py --num-images 120 --total-size 1GB --output patient_test`
2. Dans votre plateforme médicale, sélectionnez **tout le dossier** `patient_test/`
3. La plateforme reconnaîtra automatiquement les 120 images comme une série IRM unique

## Caractéristiques

- ✅ Génère des fichiers DICOM individuels (format standard)
- ✅ **Fichier DICOMDIR** automatiquement créé (index de la série)
- ✅ Tous les fichiers partagent le même Study UID et Series UID
- ✅ Chaque fichier a un InstanceNumber unique (ordre de la série)
- ✅ Métadonnées MRI réalistes (SIEMENS, GE, PHILIPS)
- ✅ Compatible avec plateformes médicales PACS
- ✅ Reproductible avec seed

## Performance

- 10 images (100MB total): < 5 seconds
- 50 images (500MB total): 5-15 seconds
- 120 images (1GB total): 15-30 seconds

Performance depends on disk speed and number of files. Each file is generated and written individually.

## Testing

Run unit tests:
```bash
pytest tests/test_generate_dicom_mri.py -v
```

Run integration tests:
```bash
pytest tests/test_integration.py -v
```

Run all tests:
```bash
pytest tests/ -v
```
