package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard"
	"github.com/mrsinham/dicomforge/internal/dicom"
	"github.com/mrsinham/dicomforge/internal/dicom/corruption"
	"github.com/mrsinham/dicomforge/internal/dicom/edgecases"
	"github.com/mrsinham/dicomforge/internal/dicom/modalities"
	"github.com/mrsinham/dicomforge/internal/util"
)

// version is set at build time via -ldflags
var version = "dev"

func main() {
	// Check for wizard subcommand (before flag.Parse)
	if len(os.Args) > 1 && os.Args[1] == "wizard" {
		// Extract --from flag if present
		var fromConfig string
		for i, arg := range os.Args[2:] {
			if arg == "--from" && i+3 < len(os.Args) {
				fromConfig = os.Args[i+3]
			}
		}
		if err := wizard.Run(fromConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Define command-line flags
	numImages := flag.Int("num-images", 0, "Number of images/slices to generate (required)")
	totalSize := flag.String("total-size", "", "Total size (e.g., '100MB', '1GB') (required)")
	outputDir := flag.String("output", "dicom_series", "Output directory")
	seed := flag.Int64("seed", 0, "Seed for reproducibility (optional, auto-generated if not specified)")
	numStudies := flag.Int("num-studies", 1, "Number of studies to generate")
	studyDescriptions := flag.String("study-descriptions", "", "Comma-separated study descriptions (must match --num-studies count)")
	numPatients := flag.Int("num-patients", 1, "Number of patients (studies are distributed among patients)")
	workers := flag.Int("workers", 0, fmt.Sprintf("Number of parallel workers (default: %d = CPU cores)", runtime.NumCPU()))

	// Modality selection
	modality := flag.String("modality", "MR", "Imaging modality: MR, CT, CR, DX, US, MG (default: MR)")

	// Multi-series support
	seriesPerStudy := flag.String("series-per-study", "1", "Number of series per study (e.g., '3' or '2-5' for random range)")

	// Categorization options
	institution := flag.String("institution", "", "Institution name (random if not specified)")
	department := flag.String("department", "", "Department name (random if not specified)")
	bodyPart := flag.String("body-part", "", "Body part examined (random per modality if not specified)")
	priority := flag.String("priority", "ROUTINE", "Exam priority: HIGH, ROUTINE, LOW")
	variedMetadata := flag.Bool("varied-metadata", false, "Generate varied institutions/physicians across studies")

	// Custom tag options
	var tagFlags []string
	flag.Func("tag", "Set DICOM tag: 'TagName=Value' (repeatable)", func(s string) error {
		tagFlags = append(tagFlags, s)
		return nil
	})

	// Edge case options
	edgeCasePercentage := flag.Int("edge-cases", 0, "Percentage of patients with edge case variations (0-100)")
	edgeCaseTypes := flag.String("edge-case-types", "special-chars,long-names,missing-tags,old-dates,varied-ids",
		"Comma-separated edge case types to enable")

	// Corruption options
	corruptTypes := flag.String("corrupt", "", "Inject vendor-specific corruption: siemens-csa,ge-private,philips-private,malformed-lengths (or 'all')")

	// Interactive wizard and config options
	interactive := flag.Bool("interactive", false, "Launch interactive wizard")
	flag.BoolVar(interactive, "i", false, "Launch interactive wizard (shortcut)")
	configFile := flag.String("config", "", "Load configuration from YAML file")
	saveConfig := flag.String("save-config", "", "Save configuration to YAML file (after generation)")

	help := flag.Bool("help", false, "Show help message")
	showVersion := flag.Bool("version", false, "Show version")

	flag.Parse()

	// Handle interactive mode
	if *interactive {
		if err := wizard.Run(""); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Handle config file loading
	if *configFile != "" {
		state, err := wizard.LoadFromYAML(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		opts, err := wizard.ToGeneratorOptions(state)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("dicomforge")
		fmt.Println("==========")
		fmt.Printf("Loading config from %s\n\n", *configFile)

		generatedFiles, err := dicom.GenerateDICOMSeries(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating DICOM series: %v\n", err)
			os.Exit(1)
		}

		if err := dicom.OrganizeFilesIntoDICOMDIR(opts.OutputDir, generatedFiles, false); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating DICOMDIR: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n✓ Generation complete!")
		fmt.Printf("  Import directory: %s\n", opts.OutputDir)
		os.Exit(0)
	}

	// Show version
	if *showVersion {
		fmt.Printf("dicomforge %s\n", version)
		os.Exit(0)
	}

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

	if *numPatients <= 0 {
		fmt.Fprintf(os.Stderr, "Error: --num-patients must be > 0\n")
		printUsage()
		os.Exit(1)
	}

	if *numPatients > *numStudies {
		fmt.Fprintf(os.Stderr, "Error: --num-patients cannot be greater than --num-studies (each patient needs at least one study)\n")
		os.Exit(1)
	}

	// Validate modality
	modalityUpper := strings.ToUpper(*modality)
	if !modalities.IsValid(modalityUpper) {
		fmt.Fprintf(os.Stderr, "Error: invalid modality %q, valid options: %v\n", *modality, modalities.AllModalities())
		os.Exit(1)
	}

	// Parse and validate study descriptions
	var parsedStudyDescriptions []string
	if *studyDescriptions != "" {
		parsedStudyDescriptions = strings.Split(*studyDescriptions, ",")
		// Trim whitespace from each description
		for i := range parsedStudyDescriptions {
			parsedStudyDescriptions[i] = strings.TrimSpace(parsedStudyDescriptions[i])
		}
		if len(parsedStudyDescriptions) != *numStudies {
			fmt.Fprintf(os.Stderr, "Error: --study-descriptions has %d descriptions but --num-studies is %d (must match)\n",
				len(parsedStudyDescriptions), *numStudies)
			os.Exit(1)
		}
	}

	// Parse priority
	parsedPriority, err := util.ParsePriority(*priority)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Parse series per study
	parsedSeriesPerStudy, err := util.ParseSeriesRange(*seriesPerStudy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Parse and validate custom tags
	parsedTags, err := util.ParseTagFlags(tagFlags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print custom tags info if specified
	if len(parsedTags) > 0 {
		fmt.Printf("Custom tags: %d specified\n", len(parsedTags))
	}

	// Parse and validate edge case config
	var edgeCaseConfig edgecases.Config
	if *edgeCasePercentage > 0 {
		types, err := edgecases.ParseTypes(*edgeCaseTypes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		edgeCaseConfig = edgecases.Config{
			Percentage: *edgeCasePercentage,
			Types:      types,
		}
		if err := edgeCaseConfig.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Edge cases: %d%% of patients with types %v\n", *edgeCasePercentage, types)
	}

	// Parse and validate corruption config
	var corruptionConfig corruption.Config
	if *corruptTypes != "" {
		types, err := corruption.ParseTypes(*corruptTypes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		corruptionConfig = corruption.Config{
			Types: types,
		}
		if err := corruptionConfig.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Corruption: injecting %v\n", types)
	}

	// Create generator options
	opts := dicom.GeneratorOptions{
		NumImages:         *numImages,
		TotalSize:         *totalSize,
		OutputDir:         *outputDir,
		Seed:              *seed,
		NumStudies:        *numStudies,
		NumPatients:       *numPatients,
		Workers:           *workers,
		Modality:          modalities.Modality(modalityUpper),
		SeriesPerStudy:    parsedSeriesPerStudy,
		StudyDescriptions: parsedStudyDescriptions,
		Institution:       *institution,
		Department:        *department,
		BodyPart:          *bodyPart,
		Priority:          parsedPriority,
		VariedMetadata:    *variedMetadata,
		CustomTags:        parsedTags,
		EdgeCaseConfig:    edgeCaseConfig,
		CorruptionConfig:  corruptionConfig,
	}

	// Generate DICOM series
	fmt.Println("dicomforge")
	fmt.Println("==========")
	fmt.Println()

	generatedFiles, err := dicom.GenerateDICOMSeries(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating DICOM series: %v\n", err)
		os.Exit(1)
	}

	// Organize into DICOMDIR structure
	if err := dicom.OrganizeFilesIntoDICOMDIR(*outputDir, generatedFiles, false); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating DICOMDIR: %v\n", err)
		os.Exit(1)
	}

	// Save config if requested
	if *saveConfig != "" {
		state := wizard.FromGeneratorOptions(opts)
		if err := wizard.SaveToYAML(state, *saveConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not save config: %v\n", err)
		} else {
			fmt.Printf("Configuration saved to %s\n", *saveConfig)
		}
	}

	fmt.Println("\n✓ Generation complete!")
	fmt.Printf("  Import directory: %s\n", *outputDir)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "\nUsage:")
	fmt.Fprintln(os.Stderr, "  dicomforge --num-images <N> --total-size <SIZE> [options]")
	fmt.Fprintln(os.Stderr, "\nRequired:")
	flag.PrintDefaults()
}

func printHelp() {
	fmt.Println("dicomforge")
	fmt.Println("==========")
	fmt.Println()
	fmt.Println("Generate valid DICOM series for testing medical platforms.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dicomforge --num-images <N> --total-size <SIZE> [options]")
	fmt.Println()
	fmt.Println("Required arguments:")
	fmt.Println("  --num-images <N>      Number of DICOM images/slices to generate")
	fmt.Println("  --total-size <SIZE>   Total size (e.g., '100MB', '1GB', '4.5GB')")
	fmt.Println()
	fmt.Println("Optional arguments:")
	fmt.Println("  --output <DIR>        Output directory (default: 'dicom_series')")
	fmt.Println("  --seed <N>            Seed for reproducibility (auto-generated if not specified)")
	fmt.Println("  --modality <MOD>      Imaging modality: MR, CT, CR, DX, US, MG (default: MR)")
	fmt.Println("  --num-studies <N>     Number of studies to generate (default: 1)")
	fmt.Println("  --study-descriptions <LIST>")
	fmt.Println("                        Comma-separated study descriptions (must match --num-studies)")
	fmt.Println("                        Example: \"IRM T0,IRM M3,IRM M6\" for 3 studies")
	fmt.Println("  --num-patients <N>    Number of patients (default: 1, studies distributed among patients)")
	fmt.Println("  --series-per-study <N|MIN-MAX>")
	fmt.Println("                        Series per study: '3' for fixed, '2-5' for random range (default: 1)")
	fmt.Printf("  --workers <N>         Number of parallel workers (default: %d = CPU cores)\n", runtime.NumCPU())
	fmt.Println()
	fmt.Println("Categorization options:")
	fmt.Println("  --institution <NAME>  Institution name (random if not specified)")
	fmt.Println("  --department <NAME>   Department name (random if not specified)")
	fmt.Println("  --body-part <PART>    Body part examined (random per modality if not specified)")
	fmt.Println("  --priority <PRIORITY> Exam priority: HIGH, ROUTINE, LOW (default: ROUTINE)")
	fmt.Println("  --varied-metadata     Generate varied institutions/physicians across studies")
	fmt.Println()
	fmt.Println("Custom tags:")
	fmt.Println("  --tag <NAME=VALUE>    Set DICOM tag value (repeatable)")
	fmt.Println("                        Example: --tag \"InstitutionName=CHU Bordeaux\"")
	fmt.Println()
	fmt.Println("Edge case options:")
	fmt.Println("  --edge-cases <N>      Percentage of patients with edge case variations (0-100)")
	fmt.Println("  --edge-case-types <T> Comma-separated types: special-chars,long-names,")
	fmt.Println("                        missing-tags,old-dates,varied-ids (default: all)")
	fmt.Println()
	fmt.Println("Corruption options (vendor-specific private tags for robustness testing):")
	fmt.Println("  --corrupt <TYPES>     Comma-separated corruption types (or 'all'):")
	fmt.Println("                        siemens-csa      - Siemens CSA private tags and crash-trigger SQ")
	fmt.Println("                        ge-private       - GE GEMS private tags")
	fmt.Println("                        philips-private  - Philips private tags and sequences")
	fmt.Println("                        malformed-lengths - Elements with incorrect VR lengths")
	fmt.Println("                        all              - All corruption types")
	fmt.Println()
	fmt.Println("  --help                Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Generate 10 MR images, 100MB total")
	fmt.Println("  dicomforge --num-images 10 --total-size 100MB")
	fmt.Println()
	fmt.Println("  # Generate CT scan with 100 slices")
	fmt.Println("  dicomforge --num-images 100 --total-size 200MB --modality CT")
	fmt.Println()
	fmt.Println("  # Generate chest X-ray (DX)")
	fmt.Println("  dicomforge --num-images 2 --total-size 50MB --modality DX --body-part CHEST")
	fmt.Println()
	fmt.Println("  # Generate ultrasound images")
	fmt.Println("  dicomforge --num-images 20 --total-size 30MB --modality US")
	fmt.Println()
	fmt.Println("  # Generate mammography images")
	fmt.Println("  dicomforge --num-images 4 --total-size 100MB --modality MG")
	fmt.Println()
	fmt.Println("  # Generate 120 images, 4.5GB, with specific seed")
	fmt.Println("  dicomforge --num-images 120 --total-size 4.5GB --seed 42")
	fmt.Println()
	fmt.Println("  # Generate 30 images across 3 studies")
	fmt.Println("  dicomforge --num-images 30 --total-size 500MB --num-studies 3")
	fmt.Println()
	fmt.Println("  # Generate 6 studies for 2 patients (3 studies each)")
	fmt.Println("  dicomforge --num-images 60 --total-size 1GB --num-studies 6 --num-patients 2")
	fmt.Println()
	fmt.Println("  # Generate with 4 parallel workers (for limited resources)")
	fmt.Println("  dicomforge --num-images 100 --total-size 1GB --workers 4")
	fmt.Println()
	fmt.Println("  # Generate MR brain study with 3-5 series (T1, T2, FLAIR, etc.)")
	fmt.Println("  dicomforge --num-images 100 --total-size 200MB --modality MR --body-part HEAD --series-per-study 3-5")
	fmt.Println()
	fmt.Println("  # Generate CT with 3 series (contrast phases)")
	fmt.Println("  dicomforge --num-images 300 --total-size 500MB --modality CT --series-per-study 3")
	fmt.Println()
	fmt.Println("  # Generate with Siemens CSA corruption (crash-trigger private tags)")
	fmt.Println("  dicomforge --num-images 10 --total-size 10MB --corrupt siemens-csa")
	fmt.Println()
	fmt.Println("  # Generate with all corruption types for robustness testing")
	fmt.Println("  dicomforge --num-images 10 --total-size 20MB --corrupt all")
	fmt.Println()
	fmt.Println("  # Combine corruption with edge cases")
	fmt.Println("  dicomforge --num-images 10 --total-size 20MB --corrupt siemens-csa --edge-cases 50")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  The program creates a DICOM series with:")
	fmt.Println("  - DICOMDIR index file")
	fmt.Println("  - PT000000/ST000000/SE000000/ hierarchy (patient/study/series)")
	fmt.Println("  - Realistic metadata (manufacturer, scanner, modality-specific parameters)")
	fmt.Println("  - Realistic patient names (80% English, 20% French)")
	fmt.Println("  - Text overlay showing 'File X/Y' on each image")
	fmt.Println()
	fmt.Println("Reproducibility:")
	fmt.Println("  Using the same seed ensures identical UIDs and patient info across runs.")
	fmt.Println("  Same output directory name also generates consistent IDs.")
}
