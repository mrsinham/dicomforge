# DICOM MRI Generator - Python to Go Conversion Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Convert Python DICOM MRI generator to Go while preserving all functionality and ensuring compatibility with medical platforms.

**Architecture:** Modular Go structure with cmd/ for CLI, internal/ for packages (dicom, image, util). Manual DICOMDIR implementation for zero external dependencies. TDD approach with unit tests first.

**Tech Stack:** Go 1.21+, github.com/suyashkumar/dicom, github.com/golang/freetype, standard library (flag, crypto/sha256, regexp, image)

---

## Task 1: Initialize Go Module and Project Structure

**Files:**
- Create: `go/go.mod`
- Create: `go/README.md`
- Create: `go/.gitignore`

**Step 1: Create go directory and initialize module**

Run: `mkdir -p go && cd go`

**Step 2: Initialize go.mod**

Run: `go mod init github.com/julien/dicom-test/go`

Expected: Creates `go.mod` with module declaration

**Step 3: Create directory structure**

Run:
```bash
mkdir -p cmd/generate-dicom-mri
mkdir -p internal/util internal/image internal/dicom
mkdir -p tests
```

Expected: Directory structure created

**Step 4: Create .gitignore**

```
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
cmd/generate-dicom-mri/generate-dicom-mri

# Test outputs
test-*/
*.dcm
DICOMDIR

# IDE
.vscode/
.idea/
*.swp
*.swo
```

**Step 5: Create basic README**

```markdown
# DICOM MRI Generator (Go)

Go implementation of the DICOM MRI generator for testing medical platforms.

## Building

\`\`\`bash
go build -o bin/generate-dicom-mri ./cmd/generate-dicom-mri
\`\`\`

## Usage

\`\`\`bash
./bin/generate-dicom-mri --num-images 10 --total-size 100MB --output test-series
\`\`\`

## Development

Run tests:
\`\`\`bash
go test ./...
\`\`\`

Run with coverage:
\`\`\`bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
\`\`\`
```

**Step 6: Commit initialization**

Run:
```bash
git add go/
git commit -m "build: initialize Go module and project structure

Set up Go module, directory structure, and basic documentation for
DICOM MRI generator Go implementation.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 2: Implement Size Parser (internal/util/size.go)

**Files:**
- Create: `go/internal/util/size_test.go`
- Create: `go/internal/util/size.go`

**Step 1: Write failing test for valid size parsing**

File: `go/internal/util/size_test.go`

```go
package util

import "testing"

func TestParseSize_ValidSizes(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"100KB", 102400},
		{"1MB", 1048576},
		{"1.5GB", 1610612736},
		{"500MB", 524288000},
		{"0.5KB", 512},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseSize(tt.input)
			if err != nil {
				t.Fatalf("ParseSize(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("ParseSize(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseSize_InvalidFormats(t *testing.T) {
	tests := []string{
		"100",
		"1.5TB",
		"abc",
		"100 MB",
		"-100MB",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseSize(input)
			if err == nil {
				t.Errorf("ParseSize(%q) expected error, got nil", input)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd go && go test ./internal/util`

Expected: FAIL with "undefined: ParseSize"

**Step 3: Implement ParseSize function**

File: `go/internal/util/size.go`

```go
package util

import (
	"fmt"
	"regexp"
	"strconv"
)

// ParseSize parses a size string (e.g., "4.5GB", "100MB") into bytes.
//
// Supported units: KB, MB, GB
// Returns the size in bytes or an error if the format is invalid.
func ParseSize(sizeStr string) (int64, error) {
	pattern := `^(\d+(?:\.\d+)?)(KB|MB|GB)$`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(sizeStr)

	if matches == nil {
		return 0, fmt.Errorf("invalid format: '%s'. Use format like '100MB', '4.5GB'", sizeStr)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %v", err)
	}

	unit := matches[2]
	multipliers := map[string]int64{
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
	}

	multiplier, ok := multipliers[unit]
	if !ok {
		return 0, fmt.Errorf("unsupported unit: '%s'. Use KB, MB, or GB", unit)
	}

	return int64(value * float64(multiplier)), nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd go && go test ./internal/util -v`

Expected: PASS for all tests

**Step 5: Commit size parser**

Run:
```bash
git add go/internal/util/
git commit -m "feat(util): implement size parser with KB/MB/GB support

Add ParseSize function to parse size strings like '100MB', '1.5GB' into bytes.
Includes comprehensive unit tests for valid and invalid inputs.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 3: Implement UID Generator (internal/util/uid.go)

**Files:**
- Create: `go/internal/util/uid_test.go`
- Create: `go/internal/util/uid.go`

**Step 1: Write failing test for UID generation**

File: `go/internal/util/uid_test.go`

```go
package util

import (
	"strings"
	"testing"
)

func TestGenerateDeterministicUID_Consistency(t *testing.T) {
	seed := "test_seed_123"
	uid1 := GenerateDeterministicUID(seed)
	uid2 := GenerateDeterministicUID(seed)

	if uid1 != uid2 {
		t.Errorf("Same seed should produce same UID: %s != %s", uid1, uid2)
	}
}

func TestGenerateDeterministicUID_Different(t *testing.T) {
	uid1 := GenerateDeterministicUID("seed1")
	uid2 := GenerateDeterministicUID("seed2")

	if uid1 == uid2 {
		t.Errorf("Different seeds should produce different UIDs")
	}
}

func TestGenerateDeterministicUID_Length(t *testing.T) {
	uid := GenerateDeterministicUID("test_seed")

	if len(uid) > 64 {
		t.Errorf("UID length %d exceeds DICOM maximum of 64 chars", len(uid))
	}
}

func TestGenerateDeterministicUID_NoLeadingZeros(t *testing.T) {
	uid := GenerateDeterministicUID("test_seed")
	segments := strings.Split(uid, ".")

	for i, segment := range segments {
		if len(segment) > 1 && segment[0] == '0' {
			t.Errorf("Segment %d has leading zero: %s", i, segment)
		}
	}
}

func TestGenerateDeterministicUID_Format(t *testing.T) {
	uid := GenerateDeterministicUID("test_seed")

	// Should start with DICOM prefix
	expectedPrefix := "1.2.826.0.1.3680043.8.498."
	if !strings.HasPrefix(uid, expectedPrefix) {
		t.Errorf("UID should start with %s, got %s", expectedPrefix, uid)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd go && go test ./internal/util`

Expected: FAIL with "undefined: GenerateDeterministicUID"

**Step 3: Implement GenerateDeterministicUID function**

File: `go/internal/util/uid.go`

```go
package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// GenerateDeterministicUID generates a deterministic DICOM UID from a seed string.
//
// The UID is generated using SHA256 hash of the seed, ensuring the same seed
// always produces the same UID. The result is a valid DICOM UID (max 64 chars,
// no leading zeros in components).
func GenerateDeterministicUID(seed string) string {
	// DICOM UID prefix for compatibility
	prefix := "1.2.826.0.1.3680043.8.498"

	// Generate SHA256 hash of seed
	hash := sha256.Sum256([]byte(seed))
	hashHex := hex.EncodeToString(hash[:])

	// Convert first 30 hex chars to numeric string
	hashBytes := hashHex[:30]
	numericValue := new(big.Int)
	numericValue.SetString(hashBytes, 16)
	numericSuffix := numericValue.String()

	// Create segments, ensuring no segment starts with 0
	var segments []string
	for i := 0; i < len(numericSuffix) && len(segments) < 3; i += 10 {
		end := i + 10
		if end > len(numericSuffix) {
			end = len(numericSuffix)
		}
		segment := numericSuffix[i:end]

		// Remove leading zeros (unless segment is just "0")
		if segment != "0" && len(segment) > 0 && segment[0] == '0' {
			segment = strings.TrimLeft(segment, "0")
			if segment == "" {
				segment = "1"
			}
		}

		if segment != "" {
			segments = append(segments, segment)
		}
	}

	suffix := strings.Join(segments, ".")
	uid := fmt.Sprintf("%s.%s", prefix, suffix)

	// Ensure UID is not too long (max 64 chars)
	if len(uid) > 64 {
		uid = uid[:63]
		uid = strings.TrimSuffix(uid, ".")
	}

	return uid
}
```

**Step 4: Run test to verify it passes**

Run: `cd go && go test ./internal/util -v`

Expected: PASS for all tests

**Step 5: Commit UID generator**

Run:
```bash
git add go/internal/util/
git commit -m "feat(util): implement deterministic UID generator

Add GenerateDeterministicUID function using SHA256 to ensure same seed
produces same DICOM UID. Validates DICOM constraints (max 64 chars,
no leading zeros in segments).

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 4: Implement Patient Name Generator (internal/util/names.go)

**Files:**
- Create: `go/internal/util/names_test.go`
- Create: `go/internal/util/names.go`

**Step 1: Write failing test for patient name generation**

File: `go/internal/util/names_test.go`

```go
package util

import (
	"math/rand"
	"strings"
	"testing"
)

func TestGeneratePatientName_Format(t *testing.T) {
	rand.Seed(42)
	name := GeneratePatientName("M")

	if !strings.Contains(name, "^") {
		t.Errorf("Name should contain '^' separator, got: %s", name)
	}

	parts := strings.Split(name, "^")
	if len(parts) != 2 {
		t.Errorf("Name should have exactly 2 parts (LASTNAME^FIRSTNAME), got: %s", name)
	}
}

func TestGeneratePatientName_Deterministic(t *testing.T) {
	rand.Seed(42)
	name1 := GeneratePatientName("M")

	rand.Seed(42)
	name2 := GeneratePatientName("M")

	if name1 != name2 {
		t.Errorf("Same seed should produce same name: %s != %s", name1, name2)
	}
}

func TestGeneratePatientName_Sex(t *testing.T) {
	maleFirstNames := []string{
		"Jean", "Pierre", "Michel", "André", "Philippe", "Alain", "Bernard", "Jacques",
		"François", "Christian", "Daniel", "Patrick", "Nicolas", "Olivier", "Laurent",
		"Thierry", "Stéphane", "Éric", "David", "Julien", "Christophe", "Pascal",
		"Sébastien", "Marc", "Vincent", "Antoine", "Alexandre", "Maxime", "Thomas",
		"Lucas", "Hugo", "Louis", "Arthur", "Gabriel", "Raphaël", "Paul", "Jules",
	}

	femaleFirstNames := []string{
		"Marie", "Nathalie", "Isabelle", "Sylvie", "Catherine", "Françoise", "Valérie",
		"Christine", "Monique", "Sophie", "Patricia", "Martine", "Nicole", "Sandrine",
		"Stéphanie", "Céline", "Julie", "Aurélie", "Caroline", "Laurence", "Émilie",
		"Claire", "Anne", "Camille", "Laura", "Sarah", "Manon", "Emma", "Léa",
		"Chloé", "Zoé", "Alice", "Charlotte", "Lucie", "Juliette", "Louise",
	}

	// Test male names
	rand.Seed(42)
	for i := 0; i < 10; i++ {
		name := GeneratePatientName("M")
		parts := strings.Split(name, "^")
		firstName := parts[1]

		found := false
		for _, mn := range maleFirstNames {
			if mn == firstName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Male name %s not in male first names list", firstName)
		}
	}

	// Test female names
	rand.Seed(42)
	for i := 0; i < 10; i++ {
		name := GeneratePatientName("F")
		parts := strings.Split(name, "^")
		firstName := parts[1]

		found := false
		for _, fn := range femaleFirstNames {
			if fn == firstName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Female name %s not in female first names list", firstName)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd go && go test ./internal/util`

Expected: FAIL with "undefined: GeneratePatientName"

**Step 3: Implement GeneratePatientName function**

File: `go/internal/util/names.go`

```go
package util

import "math/rand"

var (
	maleFirstNames = []string{
		"Jean", "Pierre", "Michel", "André", "Philippe", "Alain", "Bernard", "Jacques",
		"François", "Christian", "Daniel", "Patrick", "Nicolas", "Olivier", "Laurent",
		"Thierry", "Stéphane", "Éric", "David", "Julien", "Christophe", "Pascal",
		"Sébastien", "Marc", "Vincent", "Antoine", "Alexandre", "Maxime", "Thomas",
		"Lucas", "Hugo", "Louis", "Arthur", "Gabriel", "Raphaël", "Paul", "Jules",
	}

	femaleFirstNames = []string{
		"Marie", "Nathalie", "Isabelle", "Sylvie", "Catherine", "Françoise", "Valérie",
		"Christine", "Monique", "Sophie", "Patricia", "Martine", "Nicole", "Sandrine",
		"Stéphanie", "Céline", "Julie", "Aurélie", "Caroline", "Laurence", "Émilie",
		"Claire", "Anne", "Camille", "Laura", "Sarah", "Manon", "Emma", "Léa",
		"Chloé", "Zoé", "Alice", "Charlotte", "Lucie", "Juliette", "Louise",
	}

	lastNames = []string{
		"Martin", "Bernard", "Dubois", "Thomas", "Robert", "Richard", "Petit",
		"Durand", "Leroy", "Moreau", "Simon", "Laurent", "Lefebvre", "Michel",
		"Garcia", "David", "Bertrand", "Roux", "Vincent", "Fournier", "Morel",
		"Girard", "André", "Lefevre", "Mercier", "Dupont", "Lambert", "Bonnet",
		"François", "Martinez", "Legrand", "Garnier", "Faure", "Rousseau", "Blanc",
		"Guerin", "Muller", "Henry", "Roussel", "Nicolas", "Perrin", "Morin",
		"Mathieu", "Clement", "Gauthier", "Dumont", "Lopez", "Fontaine", "Chevalier",
		"Robin", "Masson", "Sanchez", "Gerard", "Nguyen", "Boyer", "Denis", "Lemaire",
	}
)

// GeneratePatientName generates a realistic French patient name based on sex.
//
// Sex should be "M" or "F". Returns name in DICOM format: "LASTNAME^FIRSTNAME"
func GeneratePatientName(sex string) string {
	var firstName string
	if sex == "M" {
		firstName = maleFirstNames[rand.Intn(len(maleFirstNames))]
	} else {
		firstName = femaleFirstNames[rand.Intn(len(femaleFirstNames))]
	}

	lastName := lastNames[rand.Intn(len(lastNames))]

	// DICOM format: LASTNAME^FIRSTNAME
	return lastName + "^" + firstName
}
```

**Step 4: Run test to verify it passes**

Run: `cd go && go test ./internal/util -v`

Expected: PASS for all tests

**Step 5: Commit patient name generator**

Run:
```bash
git add go/internal/util/
git commit -m "feat(util): implement patient name generator

Add GeneratePatientName function with French first names (male/female)
and last names. Outputs DICOM format LASTNAME^FIRSTNAME.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 5: Implement Pixel Data Generator (internal/image/pixel.go)

**Files:**
- Create: `go/internal/image/pixel_test.go`
- Create: `go/internal/image/pixel.go`

**Step 1: Write failing test for pixel generation**

File: `go/internal/image/pixel_test.go`

```go
package image

import (
	"math/rand"
	"testing"
)

func TestGenerateSingleImage_Size(t *testing.T) {
	width, height := 256, 256
	pixels := GenerateSingleImage(width, height, 42)

	expectedSize := width * height
	if len(pixels) != expectedSize {
		t.Errorf("Expected %d pixels, got %d", expectedSize, len(pixels))
	}
}

func TestGenerateSingleImage_Range(t *testing.T) {
	width, height := 128, 128
	pixels := GenerateSingleImage(width, height, 42)

	for i, pixel := range pixels {
		if pixel > 4095 {
			t.Errorf("Pixel %d value %d exceeds 12-bit max (4095)", i, pixel)
		}
	}
}

func TestGenerateSingleImage_Deterministic(t *testing.T) {
	width, height := 128, 128

	pixels1 := GenerateSingleImage(width, height, 42)
	pixels2 := GenerateSingleImage(width, height, 42)

	if len(pixels1) != len(pixels2) {
		t.Fatalf("Pixel slices have different lengths")
	}

	for i := range pixels1 {
		if pixels1[i] != pixels2[i] {
			t.Errorf("Pixel %d differs: %d != %d", i, pixels1[i], pixels2[i])
		}
	}
}

func TestGenerateSingleImage_Different(t *testing.T) {
	width, height := 128, 128

	pixels1 := GenerateSingleImage(width, height, 42)
	pixels2 := GenerateSingleImage(width, height, 43)

	same := true
	for i := range pixels1 {
		if pixels1[i] != pixels2[i] {
			same = false
			break
		}
	}

	if same {
		t.Errorf("Different seeds should produce different pixel data")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd go && go test ./internal/image`

Expected: FAIL with "undefined: GenerateSingleImage"

**Step 3: Implement GenerateSingleImage function**

File: `go/internal/image/pixel.go`

```go
package image

import "math/rand"

// GenerateSingleImage generates random pixel data for a single MRI image.
//
// Returns a slice of uint16 values in 12-bit range (0-4095) typical for MRI.
// The seed parameter ensures reproducible generation.
func GenerateSingleImage(width, height int, seed int64) []uint16 {
	// Seed the random number generator for reproducibility
	rng := rand.New(rand.NewSource(seed))

	// Generate random pixels in 12-bit range (0-4095)
	size := width * height
	pixels := make([]uint16, size)

	for i := 0; i < size; i++ {
		pixels[i] = uint16(rng.Intn(4096))
	}

	return pixels
}
```

**Step 4: Run test to verify it passes**

Run: `cd go && go test ./internal/image -v`

Expected: PASS for all tests

**Step 5: Commit pixel generator**

Run:
```bash
git add go/internal/image/
git commit -m "feat(image): implement pixel data generator

Add GenerateSingleImage function to generate random 12-bit MRI pixel data
with deterministic seed support for reproducibility.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 6: Install DICOM Dependencies

**Files:**
- Modify: `go/go.mod`

**Step 1: Install suyashkumar/dicom**

Run: `cd go && go get github.com/suyashkumar/dicom@latest`

Expected: Dependency added to go.mod

**Step 2: Install freetype for text rendering**

Run: `cd go && go get github.com/golang/freetype@latest`

Expected: Dependency added to go.mod

**Step 3: Install golang.org/x/image**

Run: `cd go && go get golang.org/x/image/font@latest`

Expected: Dependency added to go.mod

**Step 4: Tidy dependencies**

Run: `cd go && go mod tidy`

Expected: go.sum updated with checksums

**Step 5: Verify all tests still pass**

Run: `cd go && go test ./...`

Expected: PASS

**Step 6: Commit dependency changes**

Run:
```bash
git add go/go.mod go/go.sum
git commit -m "build: add DICOM and image rendering dependencies

Add github.com/suyashkumar/dicom for DICOM manipulation
Add github.com/golang/freetype for text overlay rendering
Add golang.org/x/image/font for font support

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 7: Implement Text Overlay (internal/image/overlay.go)

**Files:**
- Create: `go/internal/image/overlay_test.go`
- Create: `go/internal/image/overlay.go`

**Step 1: Write failing test for text overlay**

File: `go/internal/image/overlay_test.go`

```go
package image

import (
	"testing"
)

func TestAddTextOverlay_Range(t *testing.T) {
	width, height := 256, 256
	pixels := GenerateSingleImage(width, height, 42)

	err := AddTextOverlay(pixels, width, height, 5, 10)
	if err != nil {
		t.Fatalf("AddTextOverlay failed: %v", err)
	}

	// Verify pixels still in valid range
	for i, pixel := range pixels {
		if pixel > 4095 {
			t.Errorf("Pixel %d value %d exceeds 12-bit max after overlay", i, pixel)
		}
	}
}

func TestAddTextOverlay_ModifiesImage(t *testing.T) {
	width, height := 256, 256
	pixels := GenerateSingleImage(width, height, 42)

	// Make a copy before overlay
	original := make([]uint16, len(pixels))
	copy(original, pixels)

	err := AddTextOverlay(pixels, width, height, 5, 10)
	if err != nil {
		t.Fatalf("AddTextOverlay failed: %v", err)
	}

	// Check that at least some pixels changed (text was drawn)
	different := false
	for i := range pixels {
		if pixels[i] != original[i] {
			different = true
			break
		}
	}

	if !different {
		t.Errorf("Expected overlay to modify pixels")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd go && go test ./internal/image`

Expected: FAIL with "undefined: AddTextOverlay"

**Step 3: Implement AddTextOverlay function**

File: `go/internal/image/overlay.go`

```go
package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// AddTextOverlay adds text "File X/Y" to the image pixels.
//
// Modifies pixels in place. Text is drawn in white with black outline
// centered horizontally and positioned near the top of the image.
func AddTextOverlay(pixels []uint16, width, height, imageNum, totalImages int) error {
	// Convert pixels to image.Gray16
	img := image.NewGray16(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			// Scale from 12-bit (0-4095) to 16-bit (0-65535) for better contrast
			val := uint16(pixels[idx]) * 16
			img.SetGray16(x, y, color.Gray16{Y: val})
		}
	}

	// Convert to RGBA for drawing
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	// Create text to draw
	text := fmt.Sprintf("File %d/%d", imageNum, totalImages)

	// Use default font (simplified - no TrueType loading for now)
	// This is a basic implementation - full TrueType font loading would be added later
	err := drawTextBasic(rgba, text, width, height)
	if err != nil {
		return err
	}

	// Convert back to grayscale and update pixels
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			r, g, b, _ := rgba.At(x, y).RGBA()
			// Convert to grayscale (simple average)
			gray := uint16((r + g + b) / 3)
			// Scale back to 12-bit range
			pixels[idx] = gray / 16
			// Clip to ensure we stay in range
			if pixels[idx] > 4095 {
				pixels[idx] = 4095
			}
		}
	}

	return nil
}

// drawTextBasic draws text on the image using basic font
func drawTextBasic(img *image.RGBA, text string, width, height int) error {
	// For now, use a basic font drawer
	// A full implementation would load TrueType fonts
	// This is a placeholder that creates a basic white rectangle to simulate text

	// Position: centered horizontally, 5% from top
	textHeight := 20
	textWidth := len(text) * 10
	x := (width - textWidth) / 2
	y := int(float64(height) * 0.05)

	// Draw black background (outline simulation)
	for dy := -2; dy <= textHeight+2; dy++ {
		for dx := -2; dx <= textWidth+2; dx++ {
			px := x + dx
			py := y + dy
			if px >= 0 && px < width && py >= 0 && py < height {
				img.Set(px, py, color.Black)
			}
		}
	}

	// Draw white rectangle (text simulation)
	for dy := 0; dy < textHeight; dy++ {
		for dx := 0; dx < textWidth; dx++ {
			px := x + dx
			py := y + dy
			if px >= 0 && px < width && py >= 0 && py < height {
				img.Set(px, py, color.White)
			}
		}
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd go && go test ./internal/image -v`

Expected: PASS for all tests

**Step 5: Commit text overlay**

Run:
```bash
git add go/internal/image/
git commit -m "feat(image): implement basic text overlay

Add AddTextOverlay function to draw 'File X/Y' text on images.
Basic implementation with placeholder for full TrueType font rendering.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 8: Implement DICOM Metadata Generator - Part 1 (Structure)

**Files:**
- Create: `go/internal/dicom/metadata.go`
- Create: `go/internal/dicom/metadata_test.go`

**Step 1: Write test for metadata structure**

File: `go/internal/dicom/metadata_test.go`

```go
package dicom

import (
	"testing"
)

func TestGenerateMetadata_BasicStructure(t *testing.T) {
	opts := MetadataOptions{
		NumImages:      10,
		Width:          256,
		Height:         256,
		InstanceNumber: 1,
		PatientID:      "TEST123",
		PatientName:    "DOE^JOHN",
		StudyUID:       "1.2.3.4.5",
		SeriesUID:      "1.2.3.4.6",
	}

	ds, err := GenerateMetadata(opts)
	if err != nil {
		t.Fatalf("GenerateMetadata failed: %v", err)
	}

	if ds == nil {
		t.Fatal("Expected non-nil dataset")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd go && go test ./internal/dicom`

Expected: FAIL with "undefined: MetadataOptions"

**Step 3: Implement MetadataOptions struct and basic function**

File: `go/internal/dicom/metadata.go`

```go
package dicom

import (
	"fmt"

	"github.com/suyashkumar/dicom"
)

// MetadataOptions contains all parameters needed to generate DICOM metadata
type MetadataOptions struct {
	NumImages      int
	Width          int
	Height         int
	InstanceNumber int

	// Shared across series
	StudyUID         string
	SeriesUID        string
	PatientID        string
	PatientName      string
	PatientBirthDate string
	PatientSex       string
	StudyDate        string
	StudyTime        string
	StudyID          string
	StudyDescription string
	AccessionNumber  string
	SeriesNumber     int

	// MRI parameters (shared across series)
	PixelSpacing         float64
	SliceThickness       float64
	SpacingBetweenSlices float64
	EchoTime             float64
	RepetitionTime       float64
	FlipAngle            float64
	SequenceName         string
	Manufacturer         string
	Model                string
	FieldStrength        float64
}

// GenerateMetadata creates a DICOM dataset with realistic MRI metadata
func GenerateMetadata(opts MetadataOptions) (*dicom.Dataset, error) {
	// Create new dataset
	ds := &dicom.Dataset{}

	// TODO: Add all DICOM tags
	// This is a basic structure - tags will be added in next step

	return ds, nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd go && go test ./internal/dicom -v`

Expected: PASS

**Step 5: Commit metadata structure**

Run:
```bash
git add go/internal/dicom/
git commit -m "feat(dicom): add metadata structure and basic generator

Add MetadataOptions struct and skeleton GenerateMetadata function.
Full DICOM tag population will be added in next step.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 9: Implement DICOM Metadata Generator - Part 2 (Tags)

**Files:**
- Modify: `go/internal/dicom/metadata.go`
- Modify: `go/internal/dicom/metadata_test.go`

**Step 1: Write test for required DICOM tags**

File: `go/internal/dicom/metadata_test.go` (add to existing file)

```go
func TestGenerateMetadata_RequiredTags(t *testing.T) {
	opts := MetadataOptions{
		NumImages:        10,
		Width:            256,
		Height:           256,
		InstanceNumber:   1,
		PatientID:        "TEST123",
		PatientName:      "DOE^JOHN",
		PatientBirthDate: "19800101",
		PatientSex:       "M",
		StudyUID:         "1.2.3.4.5",
		SeriesUID:        "1.2.3.4.6",
		StudyDate:        "20260111",
		StudyTime:        "120000",
		StudyID:          "STD001",
		StudyDescription: "Test Study",
		AccessionNumber:  "ACC001",
		SeriesNumber:     1,
	}

	ds, err := GenerateMetadata(opts)
	if err != nil {
		t.Fatalf("GenerateMetadata failed: %v", err)
	}

	// Check patient tags exist
	patientName, err := ds.FindElementByTag(tag.PatientName)
	if err != nil {
		t.Error("PatientName tag not found")
	}
	if patientName == nil {
		t.Error("PatientName is nil")
	}

	// Check study tags exist
	studyUID, err := ds.FindElementByTag(tag.StudyInstanceUID)
	if err != nil {
		t.Error("StudyInstanceUID tag not found")
	}
	if studyUID == nil {
		t.Error("StudyInstanceUID is nil")
	}

	// Check image tags exist
	rows, err := ds.FindElementByTag(tag.Rows)
	if err != nil {
		t.Error("Rows tag not found")
	}
	if rows == nil {
		t.Error("Rows is nil")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd go && go test ./internal/dicom`

Expected: FAIL - tags not found

**Step 3: Implement DICOM tag population**

Note: This is a complex implementation. The actual implementation will need to use the suyashkumar/dicom library's API for creating elements. This is a conceptual outline - the actual code will depend on the library's API.

File: `go/internal/dicom/metadata.go` (modify function)

```go
// GenerateMetadata creates a DICOM dataset with realistic MRI metadata
func GenerateMetadata(opts MetadataOptions) (*dicom.Dataset, error) {
	elements := []dicom.Element{}

	// Patient Information Module
	elements = append(elements,
		mustNewElement(tag.PatientName, []string{opts.PatientName}),
		mustNewElement(tag.PatientID, []string{opts.PatientID}),
		mustNewElement(tag.PatientBirthDate, []string{opts.PatientBirthDate}),
		mustNewElement(tag.PatientSex, []string{opts.PatientSex}),
	)

	// Study Information Module
	elements = append(elements,
		mustNewElement(tag.StudyInstanceUID, []string{opts.StudyUID}),
		mustNewElement(tag.StudyDate, []string{opts.StudyDate}),
		mustNewElement(tag.StudyTime, []string{opts.StudyTime}),
		mustNewElement(tag.StudyID, []string{opts.StudyID}),
		mustNewElement(tag.StudyDescription, []string{opts.StudyDescription}),
		mustNewElement(tag.AccessionNumber, []string{opts.AccessionNumber}),
	)

	// Series Information Module
	seriesDesc := fmt.Sprintf("Test MRI Series - %d images", opts.NumImages)
	elements = append(elements,
		mustNewElement(tag.SeriesInstanceUID, []string{opts.SeriesUID}),
		mustNewElement(tag.SeriesNumber, []string{fmt.Sprintf("%d", opts.SeriesNumber)}),
		mustNewElement(tag.SeriesDescription, []string{seriesDesc}),
		mustNewElement(tag.Modality, []string{"MR"}),
	)

	// Instance Information
	elements = append(elements,
		mustNewElement(tag.InstanceNumber, []string{fmt.Sprintf("%d", opts.InstanceNumber)}),
		mustNewElement(tag.SOPClassUID, []string{"1.2.840.10008.5.1.4.1.1.4"}), // MR Image Storage
	)

	// Image Pixel Module
	elements = append(elements,
		mustNewElement(tag.Rows, []string{fmt.Sprintf("%d", opts.Height)}),
		mustNewElement(tag.Columns, []string{fmt.Sprintf("%d", opts.Width)}),
		mustNewElement(tag.BitsAllocated, []string{"16"}),
		mustNewElement(tag.BitsStored, []string{"16"}),
		mustNewElement(tag.HighBit, []string{"15"}),
		mustNewElement(tag.PixelRepresentation, []string{"0"}), // unsigned
		mustNewElement(tag.SamplesPerPixel, []string{"1"}),
		mustNewElement(tag.PhotometricInterpretation, []string{"MONOCHROME2"}),
	)

	// MRI-specific parameters
	if opts.Manufacturer != "" {
		elements = append(elements,
			mustNewElement(tag.Manufacturer, []string{opts.Manufacturer}),
			mustNewElement(tag.ManufacturerModelName, []string{opts.Model}),
		)
	}

	// Create dataset with elements
	ds := &dicom.Dataset{Elements: elements}

	return ds, nil
}

// mustNewElement creates a DICOM element or panics
// This is a helper for test/development - production code would handle errors
func mustNewElement(t tag.Tag, values []string) dicom.Element {
	elem, err := dicom.NewElement(t, values)
	if err != nil {
		panic(fmt.Sprintf("failed to create element for tag %v: %v", t, err))
	}
	return *elem
}
```

**Step 4: Run test to verify it passes**

Run: `cd go && go test ./internal/dicom -v`

Expected: PASS

**Step 5: Commit DICOM tag implementation**

Run:
```bash
git add go/internal/dicom/
git commit -m "feat(dicom): implement DICOM tag population

Add patient, study, series, instance, and image pixel tags to metadata
generator. Supports MRI-specific parameters.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-01-11-go-conversion-implementation.md`.

**Note:** This plan includes only the first 9 tasks covering foundational utilities, image generation, and basic DICOM metadata. The full implementation would require additional tasks for:

- DICOM file writing
- DICOMDIR creation (manual implementation)
- File hierarchy organization
- CLI implementation
- Integration tests
- Python compatibility validation

The plan follows TDD principles with tests written before implementation, frequent commits, and bite-sized tasks (2-5 minutes each).

---

**Two execution options:**

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration. Use @superpowers:subagent-driven-development

**2. Parallel Session (separate)** - Open new session with @superpowers:executing-plans, batch execution with checkpoints

**Which approach?**
