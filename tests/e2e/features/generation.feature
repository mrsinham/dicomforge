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
