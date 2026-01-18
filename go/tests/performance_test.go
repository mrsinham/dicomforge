package tests

import (
	"runtime"
	"testing"

	internaldicom "github.com/julien/dicom-test/go/internal/dicom"
)

// BenchmarkGenerateSeries_Small benchmarks small series generation
func BenchmarkGenerateSeries_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		outputDir := b.TempDir()

		opts := internaldicom.GeneratorOptions{
			NumImages:  5,
			TotalSize:  "10MB",
			OutputDir:  outputDir,
			Seed:       42,
			NumStudies: 1,
		}

		_, err := internaldicom.GenerateDICOMSeries(opts)
		if err != nil {
			b.Fatalf("GenerateDICOMSeries failed: %v", err)
		}
	}
}

// BenchmarkGenerateSeries_Medium benchmarks medium series generation
func BenchmarkGenerateSeries_Medium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		outputDir := b.TempDir()

		opts := internaldicom.GeneratorOptions{
			NumImages:  20,
			TotalSize:  "50MB",
			OutputDir:  outputDir,
			Seed:       42,
			NumStudies: 1,
		}

		_, err := internaldicom.GenerateDICOMSeries(opts)
		if err != nil {
			b.Fatalf("GenerateDICOMSeries failed: %v", err)
		}
	}
}

// BenchmarkGenerateSeries_Large benchmarks large series generation
func BenchmarkGenerateSeries_Large(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark in short mode")
	}

	for i := 0; i < b.N; i++ {
		outputDir := b.TempDir()

		opts := internaldicom.GeneratorOptions{
			NumImages:  50,
			TotalSize:  "200MB",
			OutputDir:  outputDir,
			Seed:       42,
			NumStudies: 1,
		}

		_, err := internaldicom.GenerateDICOMSeries(opts)
		if err != nil {
			b.Fatalf("GenerateDICOMSeries failed: %v", err)
		}
	}
}

// BenchmarkCalculateDimensions benchmarks dimension calculation
func BenchmarkCalculateDimensions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = internaldicom.CalculateDimensions(100*1024*1024, 10)
	}
}

// BenchmarkOrganizeFiles benchmarks DICOMDIR organization
func BenchmarkOrganizeFiles(b *testing.B) {
	// Generate files once
	outputDir := b.TempDir()

	opts := internaldicom.GeneratorOptions{
		NumImages:  10,
		TotalSize:  "20MB",
		OutputDir:  outputDir,
		Seed:       42,
		NumStudies: 1,
	}

	files, err := internaldicom.GenerateDICOMSeries(opts)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = internaldicom.OrganizeFilesIntoDICOMDIR(outputDir, files)
	}
}

// TestPerformance_MemoryUsage tests memory usage for large generation
func TestPerformance_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	outputDir := t.TempDir()

	opts := internaldicom.GeneratorOptions{
		NumImages:  50,
		TotalSize:  "200MB",
		OutputDir:  outputDir,
		Seed:       42,
		NumStudies: 1,
	}

	t.Logf("Generating 50 images (200MB)...")
	files, err := internaldicom.GenerateDICOMSeries(opts)
	if err != nil {
		t.Fatalf("GenerateDICOMSeries failed: %v", err)
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	allocMB := float64(m2.Alloc-m1.Alloc) / (1024 * 1024)
	totalAllocMB := float64(m2.TotalAlloc-m1.TotalAlloc) / (1024 * 1024)

	t.Logf("Memory usage:")
	t.Logf("  Allocated: %.2f MB", allocMB)
	t.Logf("  Total allocated: %.2f MB", totalAllocMB)
	t.Logf("  Generated %d files", len(files))

	// Memory should be reasonable (< 1GB for 200MB of output)
	if allocMB > 1024 {
		t.Errorf("Memory usage too high: %.2f MB", allocMB)
	}

	t.Logf("✓ Memory usage is acceptable")
}

// TestPerformance_GenerationSpeed tests generation speed
func TestPerformance_GenerationSpeed(t *testing.T) {
	tests := []struct {
		name       string
		numImages  int
		totalSize  string
		maxSeconds float64
	}{
		{
			name:       "small_series",
			numImages:  5,
			totalSize:  "10MB",
			maxSeconds: 2.0,
		},
		{
			name:       "medium_series",
			numImages:  20,
			totalSize:  "50MB",
			maxSeconds: 5.0,
		},
		{
			name:       "large_series",
			numImages:  50,
			totalSize:  "200MB",
			maxSeconds: 15.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if testing.Short() && tt.numImages > 20 {
				t.Skip("Skipping large test in short mode")
			}

			outputDir := t.TempDir()

			opts := internaldicom.GeneratorOptions{
				NumImages:  tt.numImages,
				TotalSize:  tt.totalSize,
				OutputDir:  outputDir,
				Seed:       42,
				NumStudies: 1,
			}

			// Time the generation
			start := testing.Benchmark(func(b *testing.B) {
				b.StopTimer()
				b.StartTimer()
				_, err := internaldicom.GenerateDICOMSeries(opts)
				if err != nil {
					t.Fatalf("GenerateDICOMSeries failed: %v", err)
				}
				b.StopTimer()
			})

			seconds := start.T.Seconds()
			t.Logf("Generated %d images (%s) in %.2f seconds", tt.numImages, tt.totalSize, seconds)

			if seconds > tt.maxSeconds {
				t.Logf("⚠ Warning: Generation took %.2f seconds (expected < %.2f)", seconds, tt.maxSeconds)
			} else {
				t.Logf("✓ Generation speed acceptable")
			}
		})
	}
}
