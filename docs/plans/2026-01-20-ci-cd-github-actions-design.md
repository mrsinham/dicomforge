# Design CI/CD GitHub Actions - Dicomforge

**Date** : 2026-01-20
**Statut** : Approuvé
**Auteur** : Collaboration humain/IA

## Résumé

Mise en place d'un pipeline CI/CD complet pour le projet Dicomforge :
- Tests automatiques avec linting et analyse de sécurité
- Releases multi-plateformes déclenchées par tags
- Support de Go 1.22, 1.23 et 1.24

## Objectifs

1. Valider la qualité du code sur chaque PR/push
2. Détecter les vulnérabilités de sécurité
3. Générer automatiquement des binaires pour 5 plateformes lors des releases
4. Maintenir la compatibilité avec les 3 dernières versions de Go

## Architecture

### Structure des fichiers

```
.github/
└── workflows/
    ├── ci.yml        # Intégration continue (PR/push)
    └── release.yml   # Build et release (tags)
```

### Workflow 1 : CI (`ci.yml`)

**Déclencheurs** :
- Push sur `main`
- Pull requests vers `main`

**Jobs parallèles** :

#### Job `lint`
- **Runner** : ubuntu-latest
- **Go** : 1.24
- **Durée estimée** : ~1 minute
- **Outils** :
  - `golangci-lint` : linting complet
  - `staticcheck` : analyse statique avancée
  - `go vet` : détection d'erreurs communes

#### Job `security`
- **Runner** : ubuntu-latest
- **Go** : 1.24
- **Durée estimée** : ~30 secondes
- **Outils** :
  - `gosec` : analyse de vulnérabilités (injections, crypto faible, etc.)
- **Artifacts** : rapport uploadé en cas d'échec

#### Job `test`
- **Runner** : ubuntu-latest
- **Go** : matrice [1.22, 1.23, 1.24]
- **Durée estimée** : ~2 minutes
- **Commandes** :
  ```bash
  go test ./... -v -race -coverprofile=coverage.out
  ```
- **Options** :
  - `-race` : détection des data races
  - `-coverprofile` : génération du rapport de couverture
  - Upload Codecov (optionnel, si token configuré)
- **Configuration** :
  - `fail-fast: false` : voir tous les échecs de la matrice

### Workflow 2 : Release (`release.yml`)

**Déclencheur** :
```yaml
on:
  push:
    tags:
      - 'v*'
```

**Jobs séquentiels** :

#### Job `test`
- Validation rapide pré-release
- Go 1.24 uniquement
- Bloque la release si échec

#### Job `build`
- **Dépend de** : `test`
- **Matrice de compilation** :

| GOOS | GOARCH | Fichier de sortie |
|------|--------|-------------------|
| linux | amd64 | `dicomforge-linux-amd64` |
| linux | arm64 | `dicomforge-linux-arm64` |
| darwin | amd64 | `dicomforge-darwin-amd64` |
| darwin | arm64 | `dicomforge-darwin-arm64` |
| windows | amd64 | `dicomforge-windows-amd64.exe` |

- **Flags de compilation** :
  ```bash
  CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=${TAG}" \
    -trimpath \
    -o dicomforge-${GOOS}-${GOARCH} \
    ./cmd/dicomforge/
  ```
  - `-s -w` : suppression des symboles de debug (binaire plus petit)
  - `-X main.version` : injection de la version
  - `-trimpath` : builds reproductibles
  - `CGO_ENABLED=0` : binaires statiques

#### Job `release`
- **Dépend de** : `build`
- **Action** : `softprops/action-gh-release`
- **Fonctionnalités** :
  - Création automatique de la GitHub Release
  - Attachement des 5 binaires
  - Génération du fichier `checksums.txt` (SHA256)
  - Release notes depuis les commits

### Fichiers générés par release

```
dicomforge-linux-amd64
dicomforge-linux-arm64
dicomforge-darwin-amd64
dicomforge-darwin-arm64
dicomforge-windows-amd64.exe
checksums.txt
```

## Configuration requise

### Secrets GitHub

| Secret | Requis | Description |
|--------|--------|-------------|
| `GITHUB_TOKEN` | Auto | Fourni automatiquement par GitHub Actions |
| `CODECOV_TOKEN` | Optionnel | Pour les badges de couverture |

### Branch protection (recommandé)

Dans Settings → Branches → main :
- [x] Require status checks to pass before merging
  - `lint`
  - `security`
  - `test`
- [x] Require branches to be up to date before merging

## Utilisation

### Développement quotidien

```bash
# Le CI se déclenche automatiquement
git push origin feature-branch
# → ci.yml : lint + security + tests (Go 1.22, 1.23, 1.24)
```

### Créer une release

```bash
# 1. S'assurer d'être sur main à jour
git checkout main
git pull

# 2. Créer et pousser le tag
git tag v1.0.0
git push origin v1.0.0

# 3. La release est créée automatiquement
# → release.yml : test → build (5 plateformes) → GitHub Release
```

### Badges pour le README

```markdown
![CI](https://github.com/mrsinham/dicomforge/actions/workflows/ci.yml/badge.svg)
![Release](https://github.com/mrsinham/dicomforge/actions/workflows/release.yml/badge.svg)
```

## Diagramme de flux

```
┌─────────────────────────────────────────────────────────────┐
│                     Push / Pull Request                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │           ci.yml              │
              └───────────────────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
    ┌──────────┐        ┌──────────┐        ┌──────────┐
    │   lint   │        │ security │        │   test   │
    │          │        │          │        │  matrix  │
    │ golangci │        │  gosec   │        │ 1.22     │
    │ static   │        │          │        │ 1.23     │
    │ go vet   │        │          │        │ 1.24     │
    └──────────┘        └──────────┘        └──────────┘
          │                   │                   │
          └───────────────────┴───────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │  PR Mergeable ?  │
                    └──────────────────┘


┌─────────────────────────────────────────────────────────────┐
│                      Push tag v*                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │         release.yml           │
              └───────────────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │      test        │
                    └──────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │      build       │
                    │   5 plateformes  │
                    └──────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │     release      │
                    │  GitHub Release  │
                    │  + checksums     │
                    └──────────────────┘
```

## Prochaines étapes

1. Implémenter `ci.yml`
2. Implémenter `release.yml`
3. Ajouter la variable `version` dans `main.go` si absente
4. Configurer la branch protection
5. Ajouter les badges au README
6. Tester avec un tag `v0.0.1-test`

## Notes techniques

- Les binaires sont compilés avec `CGO_ENABLED=0` pour être entièrement statiques
- La version est injectée via `-ldflags` au moment du build
- Le cache des modules Go est géré automatiquement par `actions/setup-go`
- Codecov est optionnel et ne bloque pas le CI si non configuré
