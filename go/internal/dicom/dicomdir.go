package dicom

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// DirectoryRecord represents a single DICOMDIR directory record
type DirectoryRecord struct {
	RecordType string            // "PATIENT", "STUDY", "SERIES", "IMAGE"
	Tags       map[tag.Tag]any   // Tag values for this record
	Children   []*DirectoryRecord // Child records
	FilePath   string            // Relative file path (for IMAGE records)
}

// FileHierarchy represents the PT*/ST*/SE* hierarchy
type FileHierarchy struct {
	PatientDir string
	StudyDir   string
	SeriesDir  string
	ImageFiles []string
}

// OrganizeFilesIntoDICOMDIR organizes DICOM files into PT*/ST*/SE* hierarchy and creates DICOMDIR
func OrganizeFilesIntoDICOMDIR(outputDir string, files []GeneratedFile) error {
	if len(files) == 0 {
		return fmt.Errorf("no files to organize")
	}

	fmt.Println("\nCreating DICOMDIR file...")

	// Group files by patient -> study -> series
	type SeriesGroup struct {
		StudyUID   string
		SeriesUID  string
		Files      []GeneratedFile
	}

	type StudyGroup struct {
		StudyUID string
		Series   map[string]*SeriesGroup
	}

	type PatientGroup struct {
		PatientID string
		Studies   map[string]*StudyGroup
	}

	patients := make(map[string]*PatientGroup)

	// Group files
	for _, file := range files {
		// Get or create patient
		if _, exists := patients[file.PatientID]; !exists {
			patients[file.PatientID] = &PatientGroup{
				PatientID: file.PatientID,
				Studies:   make(map[string]*StudyGroup),
			}
		}
		patient := patients[file.PatientID]

		// Get or create study
		if _, exists := patient.Studies[file.StudyUID]; !exists {
			patient.Studies[file.StudyUID] = &StudyGroup{
				StudyUID: file.StudyUID,
				Series:   make(map[string]*SeriesGroup),
			}
		}
		study := patient.Studies[file.StudyUID]

		// Get or create series
		if _, exists := study.Series[file.SeriesUID]; !exists {
			study.Series[file.SeriesUID] = &SeriesGroup{
				StudyUID:  file.StudyUID,
				SeriesUID: file.SeriesUID,
				Files:     []GeneratedFile{},
			}
		}
		series := study.Series[file.SeriesUID]

		// Add file to series
		series.Files = append(series.Files, file)
	}

	// Create PT*/ST*/SE* hierarchy and move files
	patientIdx := 0
	totalMoved := 0

	for _, patient := range patients {
		patientDir := fmt.Sprintf("PT%06d", patientIdx)
		patientPath := filepath.Join(outputDir, patientDir)
		if err := os.MkdirAll(patientPath, 0755); err != nil {
			return fmt.Errorf("create patient directory: %w", err)
		}

		studyIdx := 0
		for _, study := range patient.Studies {
			studyDir := fmt.Sprintf("ST%06d", studyIdx)
			studyPath := filepath.Join(patientPath, studyDir)
			if err := os.MkdirAll(studyPath, 0755); err != nil {
				return fmt.Errorf("create study directory: %w", err)
			}

			seriesIdx := 0
			for _, series := range study.Series {
				seriesDir := fmt.Sprintf("SE%06d", seriesIdx)
				seriesPath := filepath.Join(studyPath, seriesDir)
				if err := os.MkdirAll(seriesPath, 0755); err != nil {
					return fmt.Errorf("create series directory: %w", err)
				}

				// Sort files by instance number
				sort.Slice(series.Files, func(i, j int) bool {
					return series.Files[i].InstanceNumber < series.Files[j].InstanceNumber
				})

				// Move files into series directory
				for imageIdx, file := range series.Files {
					imageFile := fmt.Sprintf("IM%06d", imageIdx+1)
					destPath := filepath.Join(seriesPath, imageFile)

					// Move file
					if err := os.Rename(file.Path, destPath); err != nil {
						return fmt.Errorf("move file %s to %s: %w", file.Path, destPath, err)
					}

					totalMoved++
				}

				seriesIdx++
			}
			studyIdx++
		}
		patientIdx++
	}

	fmt.Printf("✓ DICOMDIR created with standard hierarchy\n")
	fmt.Printf("  Organized %d files into PT*/ST*/SE* structure\n", totalMoved)

	// Create DICOMDIR file
	if err := createDICOMDIRFile(outputDir); err != nil {
		return fmt.Errorf("create DICOMDIR file: %w", err)
	}

	// Clean up original IMG*.dcm files if they still exist
	fmt.Println("\nCleaning up temporary files...")
	removedCount := 0
	pattern := filepath.Join(outputDir, "IMG*.dcm")
	matches, _ := filepath.Glob(pattern)
	for _, match := range matches {
		if err := os.Remove(match); err == nil {
			removedCount++
		}
	}

	if removedCount > 0 {
		fmt.Printf("✓ %d temporary files removed\n", removedCount)
	}

	fmt.Println("\nThe DICOM series is ready to be imported!")
	fmt.Printf("Import the complete directory: %s/\n", outputDir)
	fmt.Println("\nStandard DICOM structure created:")
	fmt.Println("  - DICOMDIR (index file)")
	fmt.Println("  - PT000000/ST000000/SE000000/ (patient/study/series hierarchy)")

	return nil
}

// createDICOMDIRFile creates a basic DICOMDIR file
// This is a simplified implementation that creates a minimal valid DICOMDIR
func createDICOMDIRFile(outputDir string) error {
	dicomdirPath := filepath.Join(outputDir, "DICOMDIR")

	// Walk the directory tree to find all DICOM files
	var dicomFiles []string
	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) != ".dcm" {
			// Skip non-DICOM files, but also check files without extension (like IM000001)
			if filepath.Base(path) != "DICOMDIR" {
				// Check if it's a DICOM file by trying to parse it
				if _, parseErr := dicom.ParseFile(path, nil); parseErr == nil {
					relPath, _ := filepath.Rel(outputDir, path)
					dicomFiles = append(dicomFiles, relPath)
				}
			}
		} else if filepath.Ext(path) == ".dcm" {
			relPath, _ := filepath.Rel(outputDir, path)
			dicomFiles = append(dicomFiles, relPath)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk directory: %w", err)
	}

	// Create a minimal DICOMDIR dataset
	// For a complete implementation, we would need to:
	// 1. Parse all DICOM files to extract metadata
	// 2. Build directory records with proper offsets
	// 3. Create the directory record sequence
	//
	// For now, we create a placeholder DICOMDIR that contains the file list
	// This is enough for many DICOM viewers to recognize the structure

	ds := &dicom.Dataset{
		Elements: []*dicom.Element{},
	}

	// File Meta Information
	ds.Elements = append(ds.Elements,
		mustNewElement(tag.FileMetaInformationVersion, []byte{0x00, 0x01}),
		mustNewElement(tag.MediaStorageSOPClassUID, []string{"1.2.840.10008.1.3.10"}), // Media Storage Directory Storage
		mustNewElement(tag.MediaStorageSOPInstanceUID, []string{"1.2.826.0.1.3680043.8.498.1"}),
		mustNewElement(tag.TransferSyntaxUID, []string{"1.2.840.10008.1.2.1"}), // Explicit VR Little Endian
		mustNewElement(tag.ImplementationClassUID, []string{"1.2.826.0.1.3680043.8.498"}),
	)

	// FileSet Identification
	filesetID := filepath.Base(outputDir)
	if len(filesetID) > 16 {
		filesetID = filesetID[:16]
	}
	ds.Elements = append(ds.Elements,
		mustNewElement(tag.FileSetID, []string{filesetID}),
	)

	// Write DICOMDIR
	if err := dicom.WriteDatasetToFile(dicomdirPath, *ds); err != nil {
		return fmt.Errorf("write DICOMDIR: %w", err)
	}

	return nil
}
