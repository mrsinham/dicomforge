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
