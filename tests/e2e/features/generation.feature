Feature: DICOM Generation
  As a user of dicomforge
  I want to generate valid DICOM files
  So that I can test medical imaging platforms

  Scenario: Generate minimal MRI series
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --output {tmpdir}"
    Then the exit code should be 0
    And the output should contain "3 DICOM files created"
    And "{tmpdir}" should contain 3 DICOM files
    And dcmdump should successfully parse all files in "{tmpdir}"

  Scenario: Generate with DICOMDIR structure
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}/DICOMDIR" should exist
    And "{tmpdir}" should have patient/study/series hierarchy

  Scenario: Missing required flag total-size
    Given dicomforge is built
    When I run dicomforge with "--num-images 5"
    Then the exit code should be 1
    And the output should contain "total-size is required"

  Scenario: Missing required flag num-images
    Given dicomforge is built
    When I run dicomforge with "--total-size 200KB"
    Then the exit code should be 1
    And the output should contain "num-images must be"

  # Seed and reproducibility
  Scenario: Generate with seed for reproducibility
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --seed 42 --output {tmpdir}"
    Then the exit code should be 0
    And the output should contain "Using seed: 42"
    And "{tmpdir}" should contain 3 DICOM files

  Scenario: Same seed produces same patient name
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --seed 12345 --output {tmpdir}"
    Then the exit code should be 0
    And DICOM tag "PatientName" in "{tmpdir}" should match across all files

  # Multiple studies
  Scenario: Generate multiple studies
    Given dicomforge is built
    When I run dicomforge with "--num-images 6 --total-size 400KB --num-studies 2 --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 2 study directories
    And "{tmpdir}" should contain 6 DICOM files

  Scenario: Generate multiple studies for multiple patients
    Given dicomforge is built
    When I run dicomforge with "--num-images 6 --total-size 400KB --num-studies 2 --num-patients 2 --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 2 patient directories
    And "{tmpdir}" should contain 6 DICOM files

  # Workers option
  Scenario: Generate with limited workers
    Given dicomforge is built
    When I run dicomforge with "--num-images 4 --total-size 300KB --workers 2 --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 4 DICOM files

  # Categorization options
  Scenario: Generate with custom institution
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --institution TestHospital --output {tmpdir}"
    Then the exit code should be 0
    And DICOM tag "InstitutionName" in "{tmpdir}" should contain "TestHospital"

  Scenario: Generate with custom department
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --department RadiologyDept --output {tmpdir}"
    Then the exit code should be 0
    And DICOM tag "InstitutionalDepartmentName" in "{tmpdir}" should contain "RadiologyDept"

  Scenario: Generate with specific body part
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --body-part HEAD --output {tmpdir}"
    Then the exit code should be 0
    And DICOM tag "BodyPartExamined" in "{tmpdir}" should contain "HEAD"

  Scenario: Generate with high priority
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --priority HIGH --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 2 DICOM files

  Scenario: Generate with varied metadata
    Given dicomforge is built
    When I run dicomforge with "--num-images 4 --total-size 300KB --num-studies 2 --varied-metadata --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 4 DICOM files

  # Custom tags
  Scenario: Generate with custom tag
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --tag InstitutionName=CHUBordeaux --output {tmpdir}"
    Then the exit code should be 0
    And DICOM tag "InstitutionName" in "{tmpdir}" should contain "CHUBordeaux"

  Scenario: Generate with multiple custom tags
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --tag InstitutionName=MyHospital --tag StationName=Scanner1 --output {tmpdir}"
    Then the exit code should be 0
    And DICOM tag "InstitutionName" in "{tmpdir}" should contain "MyHospital"
    And DICOM tag "StationName" in "{tmpdir}" should contain "Scanner1"

  # Edge cases
  Scenario: Generate with edge cases enabled
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --edge-cases 100 --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 3 DICOM files
    And dcmdump should successfully parse all files in "{tmpdir}"

  Scenario: Generate with specific edge case types
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --edge-cases 100 --edge-case-types special-chars,long-names --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 3 DICOM files

  # Error cases
  Scenario: Invalid edge case percentage
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --edge-cases 150 --output {tmpdir}"
    Then the exit code should be 1
    And the output should contain "must be 0-100"

  Scenario: Invalid edge case type
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --edge-cases 50 --edge-case-types invalid-type --output {tmpdir}"
    Then the exit code should be 1
    And the output should contain "unknown edge case type"

  Scenario: Invalid priority value
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --priority INVALID --output {tmpdir}"
    Then the exit code should be 1
    And the output should contain "invalid priority"

  Scenario: Num-patients without num-studies fails
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --num-patients 2 --output {tmpdir}"
    Then the exit code should be 1
    And the output should contain "num-studies"

  # Modality options
  Scenario: Generate CT series
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --modality CT --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 3 DICOM files
    And DICOM tag "Modality" in "{tmpdir}" should contain "CT"
    And dcmdump should successfully parse all files in "{tmpdir}"

  Scenario: Generate CT with correct SOP Class
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --modality CT --output {tmpdir}"
    Then the exit code should be 0
    And DICOM tag "SOPClassUID" in "{tmpdir}" should contain "CTImageStorage"

  Scenario: Generate MR with correct SOP Class
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --modality MR --output {tmpdir}"
    Then the exit code should be 0
    And DICOM tag "SOPClassUID" in "{tmpdir}" should contain "MRImageStorage"

  Scenario: Generate CT with Hounsfield unit tags
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --modality CT --output {tmpdir}"
    Then the exit code should be 0
    And DICOM tag "RescaleIntercept" in "{tmpdir}" should contain "-1024"
    And DICOM tag "RescaleType" in "{tmpdir}" should contain "HU"

  Scenario: Invalid modality value
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --modality INVALID --output {tmpdir}"
    Then the exit code should be 1
    And the output should contain "invalid modality"

  # Additional Modalities

  Scenario: Generate CR series
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --modality CR --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 3 DICOM files
    And DICOM tag "Modality" in "{tmpdir}" should contain "CR"
    And dcmdump should successfully parse all files in "{tmpdir}"

  Scenario: Generate CR with correct SOP Class
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --modality CR --output {tmpdir}"
    Then the exit code should be 0
    And DICOM tag "SOPClassUID" in "{tmpdir}" should contain "ComputedRadiographyImageStorage"

  Scenario: Generate DX series
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --modality DX --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 3 DICOM files
    And DICOM tag "Modality" in "{tmpdir}" should contain "DX"
    And dcmdump should successfully parse all files in "{tmpdir}"

  Scenario: Generate DX with correct SOP Class
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --modality DX --output {tmpdir}"
    Then the exit code should be 0
    And DICOM tag "SOPClassUID" in "{tmpdir}" should contain "DigitalXRayImageStorageForPresentation"

  Scenario: Generate US series
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --modality US --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 3 DICOM files
    And DICOM tag "Modality" in "{tmpdir}" should contain "US"
    And dcmdump should successfully parse all files in "{tmpdir}"

  Scenario: Generate US with correct SOP Class
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --modality US --output {tmpdir}"
    Then the exit code should be 0
    And DICOM tag "SOPClassUID" in "{tmpdir}" should contain "UltrasoundImageStorage"

  Scenario: Generate MG series
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --modality MG --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 3 DICOM files
    And DICOM tag "Modality" in "{tmpdir}" should contain "MG"
    And dcmdump should successfully parse all files in "{tmpdir}"

  Scenario: Generate MG with correct SOP Class
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --modality MG --output {tmpdir}"
    Then the exit code should be 0
    And DICOM tag "SOPClassUID" in "{tmpdir}" should contain "DigitalMammographyXRayImageStorageForPresentation"

  # Multi-series studies
  Scenario: Generate study with multiple series
    Given dicomforge is built
    When I run dicomforge with "--num-images 6 --total-size 400KB --series-per-study 3 --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 6 DICOM files
    And "{tmpdir}" should contain 3 series directories
    And dcmdump should successfully parse all files in "{tmpdir}"

  Scenario: Generate study with series range
    Given dicomforge is built
    When I run dicomforge with "--num-images 8 --total-size 500KB --series-per-study 2-4 --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 8 DICOM files
    And dcmdump should successfully parse all files in "{tmpdir}"

  Scenario: Generate MR brain with multiple series
    Given dicomforge is built
    When I run dicomforge with "--num-images 12 --total-size 600KB --modality MR --body-part HEAD --series-per-study 3 --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 12 DICOM files
    And "{tmpdir}" should contain 3 series directories
    And DICOM tag "Modality" in "{tmpdir}" should contain "MR"

  Scenario: Generate CT with multiple contrast phases
    Given dicomforge is built
    When I run dicomforge with "--num-images 9 --total-size 500KB --modality CT --series-per-study 3 --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 9 DICOM files
    And "{tmpdir}" should contain 3 series directories
    And DICOM tag "Modality" in "{tmpdir}" should contain "CT"

  Scenario: Series share same FrameOfReferenceUID
    Given dicomforge is built
    When I run dicomforge with "--num-images 6 --total-size 400KB --series-per-study 2 --seed 42 --output {tmpdir}"
    Then the exit code should be 0
    And DICOM tag "FrameOfReferenceUID" in "{tmpdir}" should match across all files

  Scenario: Invalid series range
    Given dicomforge is built
    When I run dicomforge with "--num-images 6 --total-size 400KB --series-per-study 5-2 --output {tmpdir}"
    Then the exit code should be 1
    And the output should contain "max"

  # Custom study descriptions
  Scenario: Generate with custom study descriptions
    Given dicomforge is built
    When I run dicomforge with "--num-images 6 --total-size 400KB --num-studies 3 --study-descriptions IRM_T0,IRM_M3,IRM_M6 --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 6 DICOM files
    And DICOM tag "StudyDescription" in "{tmpdir}" should have value "IRM_T0" in some file
    And DICOM tag "StudyDescription" in "{tmpdir}" should have value "IRM_M3" in some file
    And DICOM tag "StudyDescription" in "{tmpdir}" should have value "IRM_M6" in some file

  Scenario: Study descriptions count mismatch
    Given dicomforge is built
    When I run dicomforge with "--num-images 6 --total-size 400KB --num-studies 3 --study-descriptions IRM_T0,IRM_M3 --output {tmpdir}"
    Then the exit code should be 1
    And the output should contain "must match"

  # YAML Configuration

  Scenario: Generate from YAML config file
    Given dicomforge is built
    And a config file "{tmpdir}/config.yaml" with:
      """
      global:
        modality: MR
        total_images: 4
        total_size: "200KB"
        output: {tmpdir}/output
      patients:
        - name: "DUPONT^JEAN"
          id: "P12345"
          birth_date: "19850315"
          sex: "M"
          studies:
            - description: "IRM Cerebrale"
              date: "20240120"
              institution: "CHU Test"
              body_part: "HEAD"
              series:
                - description: "T1"
                  images: 2
                - description: "T2"
                  images: 2
      """
    When I run dicomforge with "--config {tmpdir}/config.yaml"
    Then the exit code should be 0
    And the output should contain "Loading config from"
    And "{tmpdir}/output" should contain 4 DICOM files
    And DICOM tag "PatientName" in "{tmpdir}/output" should contain "DUPONT^JEAN"
    And DICOM tag "PatientID" in "{tmpdir}/output" should contain "P12345"
    And DICOM tag "StudyDescription" in "{tmpdir}/output" should contain "IRM Cerebrale"
    And DICOM tag "InstitutionName" in "{tmpdir}/output" should contain "CHU Test"

  Scenario: Generate from config with multiple patients
    Given dicomforge is built
    And a config file "{tmpdir}/config.yaml" with:
      """
      global:
        modality: CT
        total_images: 4
        total_size: "200KB"
        output: {tmpdir}/output
      patients:
        - name: "MARTIN^PIERRE"
          id: "P001"
          sex: "M"
          studies:
            - description: "CT Thorax"
              series:
                - description: "Sans contraste"
                  images: 2
        - name: "DURAND^MARIE"
          id: "P002"
          sex: "F"
          studies:
            - description: "CT Abdominal"
              series:
                - description: "Portal"
                  images: 2
      """
    When I run dicomforge with "--config {tmpdir}/config.yaml"
    Then the exit code should be 0
    And "{tmpdir}/output" should contain 4 DICOM files
    And "{tmpdir}/output" should contain 2 patient directories

  Scenario: Config file not found
    Given dicomforge is built
    When I run dicomforge with "--config {tmpdir}/nonexistent.yaml"
    Then the exit code should be 1
    And the output should contain "Error loading config"

  Scenario: Save config after generation
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 200KB --modality CT --num-patients 2 --num-studies 2 --output {tmpdir}/output --save-config {tmpdir}/saved.yaml"
    Then the exit code should be 0
    And the output should contain "Configuration saved to"
    And "{tmpdir}/saved.yaml" should exist
    And "{tmpdir}/saved.yaml" should contain "modality: CT"
    And "{tmpdir}/saved.yaml" should contain "total_images: 3"
    And "{tmpdir}/saved.yaml" should contain "num_patients: 2"

  Scenario: Config with custom series protocols
    Given dicomforge is built
    And a config file "{tmpdir}/config.yaml" with:
      """
      global:
        modality: MR
        total_images: 3
        total_size: "200KB"
        output: {tmpdir}/output
      patients:
        - name: "TEST^PATIENT"
          id: "P999"
          studies:
            - description: "Brain MRI"
              body_part: "HEAD"
              priority: "HIGH"
              series:
                - description: "FLAIR"
                  protocol: "T2_FLAIR_AX"
                  images: 3
      """
    When I run dicomforge with "--config {tmpdir}/config.yaml"
    Then the exit code should be 0
    And "{tmpdir}/output" should contain 3 DICOM files
    And DICOM tag "BodyPartExamined" in "{tmpdir}/output" should contain "HEAD"
    And dcmdump should successfully parse all files in "{tmpdir}/output"

  Scenario: Config with minimal global settings
    Given dicomforge is built
    And a config file "{tmpdir}/config.yaml" with:
      """
      global:
        total_images: 2
        total_size: "200KB"
        output: {tmpdir}/output
      patients:
        - name: "SIMPLE^TEST"
          studies:
            - description: "Simple Study"
              series:
                - description: "Series 1"
                  images: 2
      """
    When I run dicomforge with "--config {tmpdir}/config.yaml"
    Then the exit code should be 0
    And "{tmpdir}/output" should contain 2 DICOM files
    And DICOM tag "PatientName" in "{tmpdir}/output" should contain "SIMPLE^TEST"

  # Vendor corruption tags

  Scenario: Generate with Siemens CSA corruption
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --corrupt siemens-csa --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 2 DICOM files
    And dcmdump should successfully parse all files in "{tmpdir}"
    And dcmdump output for "{tmpdir}" should contain "SIEMENS CSA HEADER"
    And dcmdump output for "{tmpdir}" should contain "CSAImageHeaderInfo"
    And dcmdump output for "{tmpdir}" should contain "CSASeriesHeaderInfo"
    And dcmdump output for "{tmpdir}" should contain "(0029,1102) SQ"

  Scenario: Generate with GE private corruption
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --corrupt ge-private --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 2 DICOM files
    And dcmdump should successfully parse all files in "{tmpdir}"
    And dcmdump output for "{tmpdir}" should contain "GEMS_IDEN_01"
    And dcmdump output for "{tmpdir}" should contain "GEMS_PARM_01"
    And dcmdump output for "{tmpdir}" should contain "(0043,1039) IS"

  Scenario: Generate with Philips private corruption
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --corrupt philips-private --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 2 DICOM files
    And dcmdump should successfully parse all files in "{tmpdir}"
    And dcmdump output for "{tmpdir}" should contain "Philips Imaging DD 001"
    And dcmdump output for "{tmpdir}" should contain "Philips MR Imaging DD 001"
    And dcmdump output for "{tmpdir}" should contain "(2005,100e) SQ"

  Scenario: Generate with malformed lengths corruption
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --corrupt malformed-lengths --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 2 DICOM files
    And dcmdump warnings for "{tmpdir}" should contain "Length of element (0070,0253) is not a multiple of 4"

  Scenario: Invalid corruption type
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --corrupt invalid-type --output {tmpdir}"
    Then the exit code should be 1
    And the output should contain "unknown corruption type"

  Scenario: Corruption all shorthand
    Given dicomforge is built
    When I run dicomforge with "--num-images 2 --total-size 200KB --corrupt all --output {tmpdir}"
    Then the exit code should be 0
    And "{tmpdir}" should contain 2 DICOM files
