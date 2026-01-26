package help

// HelpText contains information about a field
type HelpText struct {
	Title       string
	Description string
	Details     string
}

// Texts contains help information for all wizard fields
var Texts = map[string]HelpText{
	"modality": {
		Title:       "MODALITY",
		Description: "Imaging modality type.",
		Details: `MR (Magnetic Resonance) - Brain, joints, soft tissues
CT (Computed Tomography) - Cross-sectional X-ray imaging
CR (Computed Radiography) - Digital X-ray plates
DX (Digital X-Ray) - Direct digital radiography
US (Ultrasound) - Sound wave imaging
MG (Mammography) - Breast X-ray imaging`,
	},
	"total_images": {
		Title:       "TOTAL IMAGES",
		Description: "Number of DICOM images to generate.",
		Details:     "Each image becomes a separate .dcm file in the output directory.",
	},
	"total_size": {
		Title:       "TOTAL SIZE",
		Description: "Total size of all generated files.",
		Details:     "Format: number + unit (e.g., 100MB, 1GB, 500KB)",
	},
	"output": {
		Title:       "OUTPUT DIRECTORY",
		Description: "Directory where DICOM files will be created.",
		Details:     "Will be created if it doesn't exist. Contains DICOMDIR and PT/ST/SE hierarchy.",
	},
	"num_patients": {
		Title:       "NUMBER OF PATIENTS",
		Description: "How many unique patients to generate.",
		Details:     "Each patient gets their own directory (PT000000, PT000001, etc.)",
	},
	"studies_per_patient": {
		Title:       "STUDIES PER PATIENT",
		Description: "Number of studies for each patient.",
		Details:     "A study represents one exam session (e.g., a hospital visit).",
	},
	"series_per_study": {
		Title:       "SERIES PER STUDY",
		Description: "Number of series in each study.",
		Details:     "MR: different sequences (T1, T2, FLAIR). CT: contrast phases.",
	},
	"patient_name": {
		Title:       "PATIENT NAME",
		Description: "Patient name in DICOM format.",
		Details: `Format: FAMILY^Given (e.g., SMITH^John)
Special characters allowed: accents, apostrophes, hyphens
Maximum 64 characters`,
	},
	"patient_id": {
		Title:       "PATIENT ID",
		Description: "Unique patient identifier.",
		Details:     "Usually assigned by the hospital information system.",
	},
	"birth_date": {
		Title:       "BIRTH DATE",
		Description: "Patient's date of birth.",
		Details:     "Format: YYYY-MM-DD",
	},
	"sex": {
		Title:       "SEX",
		Description: "Patient's sex.",
		Details:     "M = Male, F = Female, O = Other",
	},
	"study_description": {
		Title:       "STUDY DESCRIPTION",
		Description: "Human-readable study description.",
		Details:     "Examples: Brain MRI Routine, CT Chest with Contrast",
	},
	"study_date": {
		Title:       "STUDY DATE",
		Description: "Date the study was performed.",
		Details:     "Format: YYYY-MM-DD",
	},
	"accession": {
		Title:       "ACCESSION NUMBER",
		Description: "Unique identifier for this study order.",
		Details:     "Assigned by the radiology information system (RIS).",
	},
	"institution": {
		Title:       "INSTITUTION",
		Description: "Name of the hospital or imaging center.",
		Details:     "Examples: General Hospital, CHU Bordeaux",
	},
	"department": {
		Title:       "DEPARTMENT",
		Description: "Department within the institution.",
		Details:     "Examples: Radiology, Emergency, Cardiology",
	},
	"body_part": {
		Title:       "BODY PART",
		Description: "Anatomical region examined.",
		Details: `Standard DICOM values:
HEAD, BRAIN, NECK, CHEST, ABDOMEN, PELVIS
SPINE, CSPINE, TSPINE, LSPINE
SHOULDER, ELBOW, HAND, HIP, KNEE, ANKLE, FOOT`,
	},
	"priority": {
		Title:       "PRIORITY",
		Description: "Exam priority level.",
		Details:     "HIGH = Urgent, ROUTINE = Normal, LOW = Non-urgent",
	},
	"referring_physician": {
		Title:       "REFERRING PHYSICIAN",
		Description: "Name of the ordering physician.",
		Details:     "Format: DR FAMILY^Given or FAMILY^Given",
	},
	"series_description": {
		Title:       "SERIES DESCRIPTION",
		Description: "Description of the acquisition sequence.",
		Details: `MR examples: T1 SAG, T2 AX, FLAIR COR, DWI, T1+C
CT examples: Without contrast, Arterial, Portal, Delayed`,
	},
	"protocol": {
		Title:       "PROTOCOL NAME",
		Description: "Acquisition protocol identifier.",
		Details:     "Examples: BRAIN_T1_SAG, CHEST_CT_ROUTINE",
	},
	"orientation": {
		Title:       "ORIENTATION",
		Description: "Image orientation plane.",
		Details: `AXIAL - Horizontal slices (top to bottom)
SAGITTAL - Side view slices (left to right)
CORONAL - Front view slices (front to back)`,
	},
	"images_in_series": {
		Title:       "IMAGES IN SERIES",
		Description: "Number of images in this series.",
		Details:     "Images are distributed across series to reach total count.",
	},
	"custom_tags": {
		Title:       "CUSTOM TAGS",
		Description: "Override or add DICOM tag values.",
		Details:     "Format: TagName=Value (e.g., InstitutionName=CHU Bordeaux)",
	},
	"accept_defaults": {
		Title:       "ACCEPT DEFAULTS",
		Description: "Skip individual study configuration for this patient.",
		Details:     "If checked, all studies for this patient will use default values automatically.",
	},
	"bulk_choice": {
		Title:       "BULK PATIENT CONFIGURATION",
		Description: "Choose how to handle remaining patients.",
		Details: `Generate automatically: Random names/IDs based on patient index
Configure each one: Step through each patient's configuration screen`,
	},
}
