package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/julien/dicom-test/go/internal/dicom"
)

func main() {
	// Define command-line flags
	numImages := flag.Int("num-images", 0, "Number of images/slices to generate (required)")
	totalSize := flag.String("total-size", "", "Total size (e.g., '100MB', '1GB') (required)")
	outputDir := flag.String("output", "dicom_series", "Output directory")
	seed := flag.Int64("seed", 0, "Seed for reproducibility (optional, auto-generated if not specified)")
	numStudies := flag.Int("num-studies", 1, "Number of studies to generate")
	workers := flag.Int("workers", 0, fmt.Sprintf("Number of parallel workers (default: %d = CPU cores)", runtime.NumCPU()))
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	// Show help
	if *help {
		printHelp()
		os.Exit(0)
	}

	// Validate required arguments
	if *numImages <= 0 {
		fmt.Fprintf(os.Stderr, "Error: --num-images must be > 0\n")
		printUsage()
		os.Exit(1)
	}

	if *totalSize == "" {
		fmt.Fprintf(os.Stderr, "Error: --total-size is required\n")
		printUsage()
		os.Exit(1)
	}

	if *numStudies <= 0 {
		fmt.Fprintf(os.Stderr, "Error: --num-studies must be > 0\n")
		printUsage()
		os.Exit(1)
	}

	if *numStudies > *numImages {
		fmt.Fprintf(os.Stderr, "Error: --num-studies cannot be greater than --num-images\n")
		os.Exit(1)
	}

	// Create generator options
	opts := dicom.GeneratorOptions{
		NumImages:  *numImages,
		TotalSize:  *totalSize,
		OutputDir:  *outputDir,
		Seed:       *seed,
		NumStudies: *numStudies,
		Workers:    *workers,
	}

	// Generate DICOM series
	fmt.Println("DICOM MRI Generator (Go)")
	fmt.Println("========================")
	fmt.Println()

	generatedFiles, err := dicom.GenerateDICOMSeries(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating DICOM series: %v\n", err)
		os.Exit(1)
	}

	// Organize into DICOMDIR structure
	if err := dicom.OrganizeFilesIntoDICOMDIR(*outputDir, generatedFiles); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating DICOMDIR: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Generation complete!")
	fmt.Printf("  Import directory: %s\n", *outputDir)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "\nUsage:")
	fmt.Fprintln(os.Stderr, "  generate-dicom-mri --num-images <N> --total-size <SIZE> [options]")
	fmt.Fprintln(os.Stderr, "\nRequired:")
	flag.PrintDefaults()
}

func printHelp() {
	fmt.Println("DICOM MRI Generator (Go)")
	fmt.Println("========================")
	fmt.Println()
	fmt.Println("Generate valid DICOM multi-file MRI series for testing medical platforms.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  generate-dicom-mri --num-images <N> --total-size <SIZE> [options]")
	fmt.Println()
	fmt.Println("Required arguments:")
	fmt.Println("  --num-images <N>      Number of DICOM images/slices to generate")
	fmt.Println("  --total-size <SIZE>   Total size (e.g., '100MB', '1GB', '4.5GB')")
	fmt.Println()
	fmt.Println("Optional arguments:")
	fmt.Println("  --output <DIR>        Output directory (default: 'dicom_series')")
	fmt.Println("  --seed <N>            Seed for reproducibility (auto-generated if not specified)")
	fmt.Println("  --num-studies <N>     Number of studies to generate (default: 1)")
	fmt.Printf("  --workers <N>         Number of parallel workers (default: %d = CPU cores)\n", runtime.NumCPU())
	fmt.Println("  --help                Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Generate 10 images, 100MB total")
	fmt.Println("  generate-dicom-mri --num-images 10 --total-size 100MB")
	fmt.Println()
	fmt.Println("  # Generate 120 images, 4.5GB, with specific seed")
	fmt.Println("  generate-dicom-mri --num-images 120 --total-size 4.5GB --seed 42")
	fmt.Println()
	fmt.Println("  # Generate 30 images across 3 studies")
	fmt.Println("  generate-dicom-mri --num-images 30 --total-size 500MB --num-studies 3")
	fmt.Println()
	fmt.Println("  # Generate with 4 parallel workers (for limited resources)")
	fmt.Println("  generate-dicom-mri --num-images 100 --total-size 1GB --workers 4")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  The program creates a DICOM series with:")
	fmt.Println("  - DICOMDIR index file")
	fmt.Println("  - PT000000/ST000000/SE000000/ hierarchy (patient/study/series)")
	fmt.Println("  - Realistic MRI metadata (manufacturer, scanner, parameters)")
	fmt.Println("  - French patient names")
	fmt.Println("  - Text overlay showing 'File X/Y' on each image")
	fmt.Println()
	fmt.Println("Reproducibility:")
	fmt.Println("  Using the same seed ensures identical UIDs and patient info across runs.")
	fmt.Println("  Same output directory name also generates consistent IDs.")
}
