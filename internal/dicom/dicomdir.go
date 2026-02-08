package dicom

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// DirectoryRecord represents a single DICOMDIR directory record
type DirectoryRecord struct {
	RecordType string              // "PATIENT", "STUDY", "SERIES", "IMAGE"
	Tags       map[tag.Tag]any     // Tag values for this record
	Children   []*DirectoryRecord  // Child records
	FilePath   string              // Relative file path (for IMAGE records)
}

// FileHierarchy represents the PT*/ST*/SE* hierarchy
type FileHierarchy struct {
	PatientDir string
	StudyDir   string
	SeriesDir  string
	ImageFiles []string
}

// OrganizeFilesIntoDICOMDIR organizes DICOM files into PT*/ST*/SE* hierarchy and creates DICOMDIR
func OrganizeFilesIntoDICOMDIR(outputDir string, files []GeneratedFile, quiet bool) error {
	if len(files) == 0 {
		return fmt.Errorf("no files to organize")
	}

	if !quiet {
		fmt.Println("\nCreating DICOMDIR file...")
	}

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

	if !quiet {
		fmt.Printf("✓ DICOMDIR created with standard hierarchy\n")
		fmt.Printf("  Organized %d files into PT*/ST*/SE* structure\n", totalMoved)
	}

	// Create DICOMDIR file with directory records
	if err := createDICOMDIRFile(outputDir); err != nil {
		return fmt.Errorf("create DICOMDIR file: %w", err)
	}

	// Clean up original IMG*.dcm files if they still exist
	if !quiet {
		fmt.Println("\nCleaning up temporary files...")
	}
	removedCount := 0
	pattern := filepath.Join(outputDir, "IMG*.dcm")
	matches, _ := filepath.Glob(pattern)
	for _, match := range matches {
		if err := os.Remove(match); err == nil {
			removedCount++
		}
	}

	if !quiet {
		if removedCount > 0 {
			fmt.Printf("✓ %d temporary files removed\n", removedCount)
		}

		fmt.Println("\nThe DICOM series is ready to be imported!")
		fmt.Printf("Import the complete directory: %s/\n", outputDir)
		fmt.Println("\nStandard DICOM structure created:")
		fmt.Println("  - DICOMDIR (index file)")
		fmt.Println("  - PT*/ST*/SE*/ (patient/study/series hierarchy)")
	}

	return nil
}

// getStringValue safely extracts a string value from a dataset
func getStringValue(ds dicom.Dataset, t tag.Tag) []string {
	elem, err := ds.FindElementByTag(t)
	if err != nil || elem == nil {
		return []string{""}
	}
	str := strings.Trim(elem.Value.String(), " []")
	return []string{str}
}

// parseDICOMTolerant parses a DICOM file element-by-element, tolerating errors
// in individual elements (e.g., malformed VR lengths from corruption).
// It collects all successfully parsed elements and returns them as a dataset.
func parseDICOMTolerant(filepath string) (dicom.Dataset, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return dicom.Dataset{}, err
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return dicom.Dataset{}, err
	}

	p, err := dicom.NewParser(f, info.Size(), nil, dicom.SkipPixelData())
	if err != nil {
		return dicom.Dataset{}, err
	}

	var elements []*dicom.Element
	for {
		elem, err := p.Next()
		if err != nil {
			// Stop on any error - we've collected what we can
			break
		}
		elements = append(elements, elem)
	}

	if len(elements) == 0 {
		return dicom.Dataset{}, fmt.Errorf("no elements parsed")
	}

	ds := dicom.Dataset{Elements: elements}
	// Include metadata elements from the parser
	meta := p.GetMetadata()
	ds.Elements = append(meta.Elements, ds.Elements...)

	return ds, nil
}

// createDICOMDIRFile creates a complete DICOMDIR file with directory record sequence
func createDICOMDIRFile(outputDir string) error {
	dicomdirPath := filepath.Join(outputDir, "DICOMDIR")

	// Collect all DICOM files organized by hierarchy
	type ImageInfo struct {
		RelPath        string
		SOPClassUID    string
		SOPInstanceUID string
	}

	type SeriesInfo struct {
		SeriesUID    string
		SeriesNumber string
		Modality     string
		Images       []ImageInfo
	}

	type StudyInfo struct {
		StudyUID  string
		StudyID   string
		StudyDate string
		StudyTime string
		Series    []SeriesInfo
	}

	type PatientInfo struct {
		PatientID   string
		PatientName string
		Studies     []StudyInfo
	}

	var patients []PatientInfo

	// Walk the PT*/ST*/SE* hierarchy
	patientDirs, _ := filepath.Glob(filepath.Join(outputDir, "PT*"))
	sort.Strings(patientDirs)

	for _, patientDir := range patientDirs {
		patient := PatientInfo{
			Studies: []StudyInfo{},
		}

		studyDirs, _ := filepath.Glob(filepath.Join(patientDir, "ST*"))
		sort.Strings(studyDirs)

		for _, studyDir := range studyDirs {
			study := StudyInfo{
				Series: []SeriesInfo{},
			}

			seriesDirs, _ := filepath.Glob(filepath.Join(studyDir, "SE*"))
			sort.Strings(seriesDirs)

			for _, seriesDir := range seriesDirs {
				series := SeriesInfo{
					Images: []ImageInfo{},
				}

				imageFiles, _ := filepath.Glob(filepath.Join(seriesDir, "IM*"))
				sort.Strings(imageFiles)

				for _, imageFile := range imageFiles {
					// Parse DICOM file with tolerance for malformed elements.
					// Uses element-by-element parsing to handle files with intentionally
					// corrupted tags (e.g., from --corrupt malformed-lengths).
					ds, err := parseDICOMTolerant(imageFile)
					if err != nil {
						continue
					}

					// Get relative path from outputDir
					relPath, _ := filepath.Rel(outputDir, imageFile)

					// Extract metadata
					sopClass := getStringValue(ds, tag.SOPClassUID)
					sopInstance := getStringValue(ds, tag.SOPInstanceUID)

					image := ImageInfo{
						RelPath:        filepath.ToSlash(relPath),
						SOPClassUID:    sopClass[0],
						SOPInstanceUID: sopInstance[0],
					}
					series.Images = append(series.Images, image)

					// Get series info from first image
					if len(series.Images) == 1 {
						series.SeriesUID = getStringValue(ds, tag.SeriesInstanceUID)[0]
						series.SeriesNumber = getStringValue(ds, tag.SeriesNumber)[0]
						series.Modality = getStringValue(ds, tag.Modality)[0]
					}

					// Get study info from first image
					if len(study.Series) == 0 && len(series.Images) == 1 {
						study.StudyUID = getStringValue(ds, tag.StudyInstanceUID)[0]
						study.StudyID = getStringValue(ds, tag.StudyID)[0]
						study.StudyDate = getStringValue(ds, tag.StudyDate)[0]
						study.StudyTime = getStringValue(ds, tag.StudyTime)[0]
					}

					// Get patient info from first image of this patient
					if patient.PatientID == "" && len(series.Images) == 1 {
						patient.PatientID = getStringValue(ds, tag.PatientID)[0]
						patient.PatientName = getStringValue(ds, tag.PatientName)[0]
					}
				}

				if len(series.Images) > 0 {
					study.Series = append(study.Series, series)
				}
			}

			if len(study.Series) > 0 {
				patient.Studies = append(patient.Studies, study)
			}
		}

		if len(patient.Studies) > 0 {
			patients = append(patients, patient)
		}
	}

	// Build directory record sequence
	// Each record is a []*Element, and we collect them into [][]*Element
	var recordItems [][]*dicom.Element

	for _, patient := range patients {
		// PATIENT record - create element list
		patientElements := []*dicom.Element{
			mustNewElement(tag.OffsetOfTheNextDirectoryRecord, []int{0}), // Will be updated during write
			mustNewElement(tag.RecordInUseFlag, []int{0xFFFF}),           // 0xFFFF means record is in use
			mustNewElement(tag.OffsetOfReferencedLowerLevelDirectoryEntity, []int{0}), // Points to first STUDY
			mustNewElement(tag.DirectoryRecordType, []string{"PATIENT"}),
			mustNewElement(tag.PatientID, []string{patient.PatientID}),
			mustNewElement(tag.PatientName, []string{patient.PatientName}),
		}
		recordItems = append(recordItems, patientElements)

		for _, study := range patient.Studies {
			// STUDY record
			studyElements := []*dicom.Element{
				mustNewElement(tag.OffsetOfTheNextDirectoryRecord, []int{0}), // Will be updated
				mustNewElement(tag.RecordInUseFlag, []int{0xFFFF}),           // 0xFFFF means record is in use
				mustNewElement(tag.OffsetOfReferencedLowerLevelDirectoryEntity, []int{0}), // Points to first SERIES
				mustNewElement(tag.DirectoryRecordType, []string{"STUDY"}),
				mustNewElement(tag.StudyInstanceUID, []string{study.StudyUID}),
				mustNewElement(tag.StudyID, []string{study.StudyID}),
				mustNewElement(tag.StudyDate, []string{study.StudyDate}),
				mustNewElement(tag.StudyTime, []string{study.StudyTime}),
			}
			recordItems = append(recordItems, studyElements)

			for _, series := range study.Series {
				// SERIES record
				seriesElements := []*dicom.Element{
					mustNewElement(tag.OffsetOfTheNextDirectoryRecord, []int{0}), // Will be updated
					mustNewElement(tag.RecordInUseFlag, []int{0xFFFF}),           // 0xFFFF means record is in use
					mustNewElement(tag.OffsetOfReferencedLowerLevelDirectoryEntity, []int{0}), // Points to first IMAGE
					mustNewElement(tag.DirectoryRecordType, []string{"SERIES"}),
					mustNewElement(tag.Modality, []string{series.Modality}),
					mustNewElement(tag.SeriesInstanceUID, []string{series.SeriesUID}),
					mustNewElement(tag.SeriesNumber, []string{series.SeriesNumber}),
				}
				recordItems = append(recordItems, seriesElements)

				for _, image := range series.Images {
					// IMAGE record
					// Split path into components for ReferencedFileID
					pathParts := strings.Split(image.RelPath, "/")

					imageElements := []*dicom.Element{
						mustNewElement(tag.OffsetOfTheNextDirectoryRecord, []int{0}), // Will be updated
						mustNewElement(tag.RecordInUseFlag, []int{0xFFFF}),           // 0xFFFF means record is in use
						mustNewElement(tag.OffsetOfReferencedLowerLevelDirectoryEntity, []int{0}), // No children for IMAGE
						mustNewElement(tag.DirectoryRecordType, []string{"IMAGE"}),
						mustNewElement(tag.ReferencedFileID, pathParts),
						mustNewElement(tag.ReferencedSOPClassUIDInFile, []string{image.SOPClassUID}),
						mustNewElement(tag.ReferencedSOPInstanceUIDInFile, []string{image.SOPInstanceUID}),
						mustNewElement(tag.ReferencedTransferSyntaxUIDInFile, []string{"1.2.840.10008.1.2.1"}),
					}
					recordItems = append(recordItems, imageElements)
				}
			}
		}
	}

	// Create DICOMDIR dataset
	ds := &dicom.Dataset{
		Elements: []*dicom.Element{},
	}

	// File Meta Information (must be first)
	ds.Elements = append(ds.Elements,
		mustNewElement(tag.TransferSyntaxUID, []string{"1.2.840.10008.1.2.1"}), // Explicit VR Little Endian
		mustNewElement(tag.MediaStorageSOPClassUID, []string{"1.2.840.10008.1.3.10"}), // Media Storage Directory Storage
		mustNewElement(tag.MediaStorageSOPInstanceUID, []string{"1.2.826.0.1.3680043.8.498.1"}),
		mustNewElement(tag.ImplementationClassUID, []string{"1.2.826.0.1.3680043.8.498"}),
	)

	// FileSet Identification
	filesetID := filepath.Base(outputDir)
	if len(filesetID) > 16 {
		filesetID = filesetID[:16]
	}
	ds.Elements = append(ds.Elements,
		mustNewElement(tag.FileSetID, []string{filesetID}),
		// Directory record offsets - these should be byte offsets but we set to 0
		// A proper implementation would calculate these during write
		mustNewElement(tag.OffsetOfTheFirstDirectoryRecordOfTheRootDirectoryEntity, []int{0}),
		mustNewElement(tag.OffsetOfTheLastDirectoryRecordOfTheRootDirectoryEntity, []int{0}),
		// FileSet Consistency Flag - 0 means no known inconsistencies
		mustNewElement(tag.FileSetConsistencyFlag, []int{0}),
	)

	// Add Directory Record Sequence
	// recordItems is [][]*Element, which NewElement will convert to SequenceItemValue automatically
	if len(recordItems) > 0 {
		seqElem, err := dicom.NewElement(tag.DirectoryRecordSequence, recordItems)
		if err != nil {
			return fmt.Errorf("create directory record sequence: %w", err)
		}
		ds.Elements = append(ds.Elements, seqElem)
	}

	// Write DICOMDIR (first pass with offsets at 0)
	if err := writeDatasetToFile(dicomdirPath, *ds); err != nil {
		return fmt.Errorf("write DICOMDIR: %w", err)
	}

	// Second pass: update offsets with correct byte positions
	if err := updateDICOMDIROffsets(dicomdirPath); err != nil {
		return fmt.Errorf("update DICOMDIR offsets: %w", err)
	}

	return nil
}

// updateDICOMDIROffsets reads a DICOMDIR file and updates the offset tags with correct byte positions
func updateDICOMDIROffsets(dicomdirPath string) error {
	// Read the entire DICOMDIR file
	data, err := os.ReadFile(dicomdirPath)
	if err != nil {
		return fmt.Errorf("read DICOMDIR: %w", err)
	}

	// Find the Directory Record Sequence (0004,1220)
	// We need to find where each Item (Directory Record) starts
	recordPositions, err := findDirectoryRecordPositions(data)
	if err != nil {
		return fmt.Errorf("find record positions: %w", err)
	}

	if len(recordPositions) == 0 {
		return fmt.Errorf("no directory records found")
	}

	// Now update the offset values in the file
	// We need to update:
	// 1. OffsetOfTheFirstDirectoryRecordOfTheRootDirectoryEntity (0004,1200)
	// 2. OffsetOfTheLastDirectoryRecordOfTheRootDirectoryEntity (0004,1202)
	// 3. OffsetOfTheNextDirectoryRecord (0004,1400) in each record
	// 4. OffsetOfReferencedLowerLevelDirectoryEntity (0004,1420) in each record

	// Update file with calculated offsets
	if err := updateOffsetsInFile(dicomdirPath, data, recordPositions); err != nil {
		return fmt.Errorf("update offsets in file: %w", err)
	}

	return nil
}

// findDirectoryRecordPositions scans the DICOMDIR binary data to find the byte position of each Directory Record
func findDirectoryRecordPositions(data []byte) ([]int64, error) {
	var positions []int64

	// Look for Item tags (FFFE,E000) which indicate the start of each Directory Record
	// In DICOM binary: Tag is little-endian, so (FFFE,E000) = 0xE0 0x00 0xFE 0xFF in bytes
	itemTag := []byte{0xFE, 0xFF, 0x00, 0xE0}

	// Start searching after the file meta information
	// Skip preamble (128 bytes) + "DICM" (4 bytes) = 132 bytes minimum
	searchStart := 132

	for i := searchStart; i < len(data)-4; i++ {
		if bytes.Equal(data[i:i+4], itemTag) {
			// Found an item tag, this could be a Directory Record
			// Verify it's within the Directory Record Sequence by checking context
			positions = append(positions, int64(i))
		}
	}

	return positions, nil
}

// updateOffsetsInFile updates the offset values in the DICOMDIR file
func updateOffsetsInFile(path string, data []byte, recordPositions []int64) error {
	// Parse the DICOMDIR to understand the structure
	ds, err := dicom.ParseFile(path, nil)
	if err != nil {
		return fmt.Errorf("parse DICOMDIR: %w", err)
	}

	// Get the Directory Record Sequence
	seqElem, err := ds.FindElementByTag(tag.DirectoryRecordSequence)
	if err != nil {
		return fmt.Errorf("find directory record sequence: %w", err)
	}

	seqItems := seqElem.Value.GetValue().([]*dicom.SequenceItemValue)

	// We need to map which record position corresponds to which record in the hierarchy
	// For simplicity, we'll build a mapping based on the order

	// Count records by type to understand the hierarchy
	var recordInfos []RecordInfo
	for i, item := range seqItems {
		if i >= len(recordPositions) {
			break
		}
		elements := item.GetValue().([]*dicom.Element)
		recordType := ""
		for _, elem := range elements {
			if elem.Tag == tag.DirectoryRecordType {
				recordType = elem.Value.GetValue().([]string)[0]
				break
			}
		}
		recordInfos = append(recordInfos, RecordInfo{
			Type:     recordType,
			Index:    i,
			Position: recordPositions[i],
		})
	}

	// Now update the offsets
	// Strategy: Open file for read/write and update specific offset fields
	f, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("open file for update: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Update FirstDirectoryRecordOffset and LastDirectoryRecordOffset in the header
	if len(recordPositions) > 0 {
		firstOffset := uint32(recordPositions[0])
		lastOffset := uint32(recordPositions[len(recordPositions)-1])

		// Find and update (0004,1200) - FirstDirectoryRecordOffset
		if pos := findTagPosition(data, 0x0004, 0x1200); pos >= 0 {
			if err := updateUInt32At(f, pos+8, firstOffset); err != nil {
				return fmt.Errorf("update first offset: %w", err)
			}
		}

		// Find and update (0004,1202) - LastDirectoryRecordOffset
		if pos := findTagPosition(data, 0x0004, 0x1202); pos >= 0 {
			if err := updateUInt32At(f, pos+8, lastOffset); err != nil {
				return fmt.Errorf("update last offset: %w", err)
			}
		}
	}

	// Build parent-child relationships to calculate proper offsets
	hierarchy := buildHierarchy(recordInfos)

	// Update offsets within each Directory Record with proper hierarchy
	for i, info := range recordInfos {
		basePos := info.Position

		// Calculate OffsetOfTheNextDirectoryRecord
		// This should point to the next sibling (same parent, same level)
		nextOffset := hierarchy[i].NextSibling

		// Calculate OffsetOfReferencedLowerLevelDirectoryEntity
		// This should point to the first CHILD record
		lowerOffset := hierarchy[i].FirstChild

		// Update OffsetOfTheNextDirectoryRecord (0004,1400) in this record
		if pos := findTagPositionAfter(data, int(basePos), 0x0004, 0x1400); pos >= 0 {
			if err := updateUInt32At(f, int64(pos+8), nextOffset); err != nil {
				return fmt.Errorf("update next offset at record %d: %w", i, err)
			}
		}

		// Update OffsetOfReferencedLowerLevelDirectoryEntity (0004,1420) in this record
		if pos := findTagPositionAfter(data, int(basePos), 0x0004, 0x1420); pos >= 0 {
			if err := updateUInt32At(f, int64(pos+8), lowerOffset); err != nil {
				return fmt.Errorf("update lower offset at record %d: %w", i, err)
			}
		}
	}

	return nil
}

// RecordInfo holds information about a directory record
type RecordInfo struct {
	Type     string
	Index    int
	Position int64
}

// HierarchyInfo holds offset information for a record
type HierarchyInfo struct {
	NextSibling uint32
	FirstChild  uint32
}

// buildHierarchy analyzes the record list and builds parent-child relationships
func buildHierarchy(records []RecordInfo) map[int]HierarchyInfo {
	result := make(map[int]HierarchyInfo)

	// Track hierarchy levels: when we see a record, remember where we are
	// We process records in order, maintaining a stack of "current" items at each level
	type LevelState struct {
		Type     string
		Index    int
		Children []int // indices of direct children
	}

	var stack []*LevelState       // stack of current items at each hierarchy level
	var rootRecords []int         // indices of root-level records (PATIENT)

	for i, record := range records {
		// Pop stack until we find where this record belongs
		level := getHierarchyLevel(record.Type)

		// Pop items from stack that are at >= this level (we're back up the tree)
		for len(stack) > level {
			stack = stack[:len(stack)-1]
		}

		// If stack is not empty, this record is a child of the top item
		if len(stack) > 0 {
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, i)
		} else {
			// This is a root-level record (PATIENT)
			// Link to previous root record if exists
			if len(rootRecords) > 0 {
				prevRootIdx := rootRecords[len(rootRecords)-1]
				info := result[prevRootIdx]
				info.NextSibling = uint32(records[i].Position)
				result[prevRootIdx] = info
			}
			rootRecords = append(rootRecords, i)
		}

		// Push this record onto the stack
		stack = append(stack, &LevelState{
			Type:     record.Type,
			Index:    i,
			Children: []int{},
		})

		// Now calculate offsets for all completed siblings
		// When we add a new item at a level, we can finalize the previous item's NextSibling
		if len(stack) >= 2 {
			parentLevel := stack[len(stack)-2]
			if len(parentLevel.Children) >= 2 {
				// There are at least 2 children, so we can link them
				prevChildIdx := parentLevel.Children[len(parentLevel.Children)-2]
				currChildIdx := parentLevel.Children[len(parentLevel.Children)-1]

				// Previous child's NextSibling points to current child
				info := result[prevChildIdx]
				info.NextSibling = uint32(records[currChildIdx].Position)
				result[prevChildIdx] = info
			}

			// First child: parent's FirstChild points to it
			if len(parentLevel.Children) == 1 {
				childIdx := parentLevel.Children[0]
				parentIdx := parentLevel.Index
				info := result[parentIdx]
				info.FirstChild = uint32(records[childIdx].Position)
				result[parentIdx] = info
			}
		}
	}

	// Final pass: ensure all indices have an entry (even if both offsets are 0)
	for i := range records {
		if _, exists := result[i]; !exists {
			result[i] = HierarchyInfo{}
		}
	}

	return result
}

// getHierarchyLevel returns the hierarchy level (0=PATIENT, 1=STUDY, 2=SERIES, 3=IMAGE)
func getHierarchyLevel(recordType string) int {
	switch recordType {
	case "PATIENT":
		return 0
	case "STUDY":
		return 1
	case "SERIES":
		return 2
	case "IMAGE":
		return 3
	default:
		return -1
	}
}

// findTagPosition finds the byte position of a DICOM tag in the data
func findTagPosition(data []byte, group, element uint16) int64 {
	// DICOM tags are stored as: group (2 bytes LE) + element (2 bytes LE)
	tagBytes := make([]byte, 4)
	binary.LittleEndian.PutUint16(tagBytes[0:2], group)
	binary.LittleEndian.PutUint16(tagBytes[2:4], element)

	for i := 0; i < len(data)-4; i++ {
		if bytes.Equal(data[i:i+4], tagBytes) {
			return int64(i)
		}
	}
	return -1
}

// findTagPositionAfter finds the byte position of a DICOM tag after a given position
func findTagPositionAfter(data []byte, startPos int, group, element uint16) int64 {
	tagBytes := make([]byte, 4)
	binary.LittleEndian.PutUint16(tagBytes[0:2], group)
	binary.LittleEndian.PutUint16(tagBytes[2:4], element)

	for i := startPos; i < len(data)-4 && i < startPos+500; i++ { // Search within 500 bytes
		if bytes.Equal(data[i:i+4], tagBytes) {
			return int64(i)
		}
	}
	return -1
}

// updateUInt32At writes a uint32 value at the specified position in the file
func updateUInt32At(f io.WriteSeeker, pos int64, value uint32) error {
	if _, err := f.Seek(pos, io.SeekStart); err != nil {
		return err
	}
	return binary.Write(f, binary.LittleEndian, value)
}
