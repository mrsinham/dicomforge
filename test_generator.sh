#!/bin/bash
# Script de test pour le générateur DICOM

set -e

echo "========================================="
echo "Test du générateur DICOM MRI"
echo "========================================="
echo ""

# Test 1: Petit fichier
echo "Test 1: Génération d'un petit fichier (10MB, 5 images)..."
python generate_dicom_mri.py --num-images 5 --total-size 10MB --output test_small.dcm --seed 42

if [ -f "test_small.dcm" ]; then
    echo "✓ Fichier créé: test_small.dcm"
    size=$(ls -lh test_small.dcm | awk '{print $5}')
    echo "  Taille: $size"

    # Vérifier avec pydicom
    python -c "import pydicom; ds = pydicom.dcmread('test_small.dcm'); print(f'  Modalité: {ds.Modality}, Frames: {ds.NumberOfFrames}, Dimensions: {ds.Rows}x{ds.Columns}')"
    echo "✓ Fichier DICOM valide"
    rm test_small.dcm
else
    echo "✗ Erreur: fichier non créé"
    exit 1
fi

echo ""

# Test 2: Tests unitaires
echo "Test 2: Exécution des tests unitaires..."
pytest tests/test_generate_dicom_mri.py -v --tb=short

echo ""

# Test 3: Tests d'intégration
echo "Test 3: Exécution des tests d'intégration..."
pytest tests/test_integration.py -v --tb=short

echo ""
echo "========================================="
echo "✓ Tous les tests ont réussi!"
echo "========================================="
echo ""
echo "Pour tester un grand fichier (4.5GB), exécutez:"
echo "  python generate_dicom_mri.py --num-images 120 --total-size 4.5GB --output test_large.dcm"
