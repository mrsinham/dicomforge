package image

import (
	"fmt"
	"image"
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// AddTextOverlay adds text "File X/Y" to the image pixels.
//
// Modifies pixels in place. Text is drawn with white color and black outline
// for visibility against varying backgrounds. Uses basicfont for simplicity;
// full TrueType font rendering can be added later using golang.org/x/image/font/opentype.
//
// The function converts the uint16 pixel data to RGBA for drawing, then converts
// back to uint16, ensuring all values remain in the valid 12-bit range (0-4095).
// Optimized to use only 2 conversion passes instead of 5.
func AddTextOverlay(pixels []uint16, width, height, imageNum, totalImages int) error {
	// Validate inputs
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid dimensions: %dx%d", width, height)
	}
	if len(pixels) != width*height {
		return fmt.Errorf("pixel slice length %d does not match dimensions %dx%d", len(pixels), width, height)
	}
	if imageNum < 1 || totalImages < 1 || imageNum > totalImages {
		return fmt.Errorf("invalid image numbering: %d/%d", imageNum, totalImages)
	}

	// Pass 1: Convert to RGBA for drawing (scale 12-bit to 8-bit)
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			// Scale 12-bit (0-4095) to 8-bit (0-255)
			gray := uint8((uint32(pixels[idx]) * 255) / 4095)
			img.SetRGBA(x, y, color.RGBA{gray, gray, gray, 255})
		}
	}

	// Prepare text
	text := fmt.Sprintf("File %d/%d", imageNum, totalImages)

	// Use basicfont.Face7x13 - a simple fixed-width font
	face := basicfont.Face7x13

	// Calculate text position: centered horizontally, near top (5% from top)
	paddingTop := int(float64(height) * 0.05)

	// Measure text width
	textWidth := font.MeasureString(face, text).Ceil()
	x := (width - textWidth) / 2

	// For basicfont, metrics are available from the font.Metrics method
	metrics := face.Metrics()
	y := paddingTop + metrics.Ascent.Ceil()

	// Create a drawer
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: face,
		Dot:  fixed.P(x, y),
	}

	// Draw black outline for visibility (thick outline)
	outlineThickness := 2
	for dx := -outlineThickness; dx <= outlineThickness; dx++ {
		for dy := -outlineThickness; dy <= outlineThickness; dy++ {
			if dx != 0 || dy != 0 { // Skip center
				drawer.Dot = fixed.P(x+dx, y+dy)
				drawer.DrawString(text)
			}
		}
	}

	// Draw main text in white
	drawer.Src = image.NewUniform(color.White)
	drawer.Dot = fixed.P(x, y)
	drawer.DrawString(text)

	// Pass 2: Convert back to uint16 (scale 8-bit to 12-bit)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			r, g, b, _ := img.At(x, y).RGBA()
			// RGBA() returns 16-bit values (0-65535), convert to 8-bit first
			gray8 := uint8((r + g + b) / (3 * 256))
			// Scale 8-bit (0-255) to 12-bit (0-4095) correctly
			pixels[idx] = uint16((uint32(gray8) * 4095) / 255)
			// Clamp to 12-bit range
			if pixels[idx] > 4095 {
				pixels[idx] = 4095
			}
		}
	}

	return nil
}
