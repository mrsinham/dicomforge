package image

import "math/rand/v2"

// GenerateSingleImage generates random pixel data for a single MRI image.
//
// Returns a slice of uint16 values in 12-bit range (0-4095) typical for MRI.
// The seed parameter ensures reproducible generation.
// Returns nil if dimensions are invalid (zero, negative, or would overflow).
func GenerateSingleImage(width, height int, seed int64) []uint16 {
	if width <= 0 || height <= 0 {
		return nil
	}

	// Check for potential overflow on 32-bit systems
	if width > 0 && height > 0 {
		maxSize := int(^uint(0) >> 1) // max int value
		if width > maxSize/height {
			return nil // would overflow
		}
	}

	// Seed the random number generator for reproducibility
	rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed)))

	// Generate random pixels in 12-bit range (0-4095)
	size := width * height
	pixels := make([]uint16, size)

	for i := 0; i < size; i++ {
		pixels[i] = uint16(rng.IntN(4096))
	}

	return pixels
}
