package corruption

import (
	"bytes"
	"encoding/binary"
	"math/rand/v2"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// csaElement represents a single element in a CSA header
type csaElement struct {
	Name     string
	VM       int32
	VR       string
	SyngoDT  int32
	NumItems int32
	Values   []string
}

// buildCSAHeader encodes a list of CSA elements into the "SV10" binary format
// used by Siemens scanners.
func buildCSAHeader(elements []csaElement) []byte {
	var buf bytes.Buffer

	// Magic bytes: "SV10" followed by 0x04, 0x03, 0x02, 0x01
	buf.WriteString("SV10")
	buf.Write([]byte{0x04, 0x03, 0x02, 0x01})

	// binary.Write to bytes.Buffer never fails; discard errors explicitly.
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(elements)))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x4D))

	for _, elem := range elements {
		// Element name: 64 bytes, null-padded
		name := make([]byte, 64)
		copy(name, elem.Name)
		buf.Write(name)

		_ = binary.Write(&buf, binary.LittleEndian, elem.VM)

		// VR: 4 bytes, null-padded
		vr := make([]byte, 4)
		copy(vr, elem.VR)
		buf.Write(vr)

		_ = binary.Write(&buf, binary.LittleEndian, elem.SyngoDT)
		_ = binary.Write(&buf, binary.LittleEndian, elem.NumItems)
		_ = binary.Write(&buf, binary.LittleEndian, uint32(0x4D))

		// Write items
		for i := int32(0); i < elem.NumItems; i++ {
			var val []byte
			if i < int32(len(elem.Values)) {
				val = []byte(elem.Values[i])
			}

			// Item length (repeated 4 times per CSA format)
			itemLen := uint32(len(val))
			for j := 0; j < 4; j++ {
				_ = binary.Write(&buf, binary.LittleEndian, itemLen)
			}

			// Item data
			buf.Write(val)

			// Pad to 4-byte boundary
			if padding := (4 - len(val)%4) % 4; padding > 0 {
				buf.Write(make([]byte, padding))
			}
		}
	}

	return buf.Bytes()
}

// generateCSAImageHeader creates a realistic CSA Image Header blob
func generateCSAImageHeader(rng *rand.Rand) []byte {
	elements := []csaElement{
		{
			Name: "NumberOfImagesInMosaic", VM: 1, VR: "IS", SyngoDT: 6, NumItems: 1,
			Values: []string{"1"},
		},
		{
			Name: "SliceNormalVector", VM: 3, VR: "FD", SyngoDT: 3, NumItems: 3,
			Values: []string{"0.0", "0.0", "1.0"},
		},
		{
			Name: "DiffusionGradientDirection", VM: 3, VR: "FD", SyngoDT: 3, NumItems: 3,
			Values: []string{"0.0", "0.0", "0.0"},
		},
		{
			Name: "B_value", VM: 1, VR: "IS", SyngoDT: 6, NumItems: 1,
			Values: []string{"0"},
		},
		{
			Name: "SliceMeasurementDuration", VM: 1, VR: "DS", SyngoDT: 3, NumItems: 1,
			Values: []string{"265000.0"},
		},
		{
			Name: "BandwidthPerPixelPhaseEncode", VM: 1, VR: "FD", SyngoDT: 3, NumItems: 1,
			Values: []string{"45.455"},
		},
		{
			Name: "MosaicRefAcqTimes", VM: 1, VR: "FD", SyngoDT: 3, NumItems: 1,
			Values: []string{"0.0"},
		},
		{
			Name: "ImaRelTablePosition", VM: 3, VR: "IS", SyngoDT: 6, NumItems: 3,
			Values: []string{"0", "0", "0"},
		},
		{
			Name: "RealDwellTime", VM: 1, VR: "IS", SyngoDT: 6, NumItems: 1,
			Values: []string{"5700"},
		},
		{
			Name: "ImaCoilString", VM: 1, VR: "LO", SyngoDT: 19, NumItems: 1,
			Values: []string{"HEA;HEP"},
		},
	}

	// Add random variation in data size
	extraPadding := make([]byte, rng.IntN(2048)+1024)
	for i := range extraPadding {
		extraPadding[i] = byte(rng.IntN(256))
	}

	header := buildCSAHeader(elements)
	return append(header, extraPadding...)
}

// generateCSASeriesHeader creates a realistic CSA Series Header blob
func generateCSASeriesHeader(rng *rand.Rand) []byte {
	elements := []csaElement{
		{
			Name: "UsedPatientWeight", VM: 1, VR: "DS", SyngoDT: 3, NumItems: 1,
			Values: []string{"70.0"},
		},
		{
			Name: "MrProtocolVersion", VM: 1, VR: "IS", SyngoDT: 6, NumItems: 1,
			Values: []string{"1"},
		},
		{
			Name: "DataFileName", VM: 1, VR: "LO", SyngoDT: 19, NumItems: 1,
			Values: []string{"%ScanProtocol%_PROT"},
		},
		{
			Name: "MrProtocol", VM: 1, VR: "LO", SyngoDT: 19, NumItems: 1,
			Values: []string{"### ASCCONV BEGIN ###"},
		},
		{
			Name: "Isocentered", VM: 1, VR: "IS", SyngoDT: 6, NumItems: 1,
			Values: []string{"1"},
		},
		{
			Name: "CoilForGradient", VM: 1, VR: "LO", SyngoDT: 19, NumItems: 1,
			Values: []string{"AS"},
		},
		{
			Name: "CoilForGradient2", VM: 1, VR: "LO", SyngoDT: 19, NumItems: 1,
			Values: []string{""},
		},
		{
			Name: "TablePositionOrigin", VM: 3, VR: "FD", SyngoDT: 3, NumItems: 3,
			Values: []string{"0.0", "0.0", "0.0"},
		},
	}

	// Add random variation
	extraPadding := make([]byte, rng.IntN(1024)+512)
	for i := range extraPadding {
		extraPadding[i] = byte(rng.IntN(256))
	}

	header := buildCSAHeader(elements)
	return append(header, extraPadding...)
}

// generateCrashTriggerSequence creates a private sequence at (0029,1102) that
// mimics the problematic Siemens private sequences known to crash fragile DICOM readers.
func generateCrashTriggerSequence(rng *rand.Rand) *dicom.Element {
	// Build nested items with private sub-elements
	nestedData := make([]byte, rng.IntN(4096)+5120)
	for i := range nestedData {
		nestedData[i] = byte(rng.IntN(256))
	}

	item := []*dicom.Element{
		mustNewPrivateElement(tag.Tag{Group: 0x0029, Element: 0x0011}, "LO", []string{"SIEMENS CSA NON-IMAGE"}),
		mustNewPrivateElement(tag.Tag{Group: 0x0029, Element: 0x1100}, "OB", nestedData),
	}

	return mustNewPrivateElement(tag.Tag{Group: 0x0029, Element: 0x1102}, "SQ", [][]*dicom.Element{item})
}

// generateSiemensCSAElements generates all Siemens CSA private elements.
func generateSiemensCSAElements(rng *rand.Rand) []*dicom.Element {
	csaImageHeader := generateCSAImageHeader(rng)
	csaSeriesHeader := generateCSASeriesHeader(rng)

	return []*dicom.Element{
		// Private creator block
		mustNewPrivateElement(tag.Tag{Group: 0x0029, Element: 0x0010}, "LO", []string{"SIEMENS CSA HEADER"}),
		// CSA Image Header
		mustNewPrivateElement(tag.Tag{Group: 0x0029, Element: 0x1010}, "OB", csaImageHeader),
		// CSA Series Header
		mustNewPrivateElement(tag.Tag{Group: 0x0029, Element: 0x1020}, "OB", csaSeriesHeader),
		// Crash-trigger sequence
		generateCrashTriggerSequence(rng),
	}
}
