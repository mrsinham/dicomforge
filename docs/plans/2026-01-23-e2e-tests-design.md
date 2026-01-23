# E2E Tests with Godog/Gherkin

**Date:** 2026-01-23
**Status:** Approved

## Goal

Add end-to-end integration tests that:
- Compile the binary once and reuse it across scenarios
- Execute the CLI and validate results
- Use DCMTK (`dcmdump`) to validate generated DICOM files
- Run in GitHub Actions

## Structure

```
tests/e2e/
├── features/
│   └── generation.feature      # Gherkin scenarios
├── steps_test.go               # Step definitions + TestMain
└── testdata/                   # Temporary files (gitignored)
```

## Scenarios (Base)

```gherkin
Feature: DICOM Generation

  Scenario: Generate minimal MRI series
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 50KB --output {tmpdir}"
    Then the exit code should be 0
    And the output should contain "3 DICOM files created"
    And "{tmpdir}" should contain 3 DICOM files
    And dcmdump should successfully parse all files in "{tmpdir}"

  Scenario: Generate with DICOMDIR structure
    Given dicomforge is built
    When I run dicomforge with "--num-images 3 --total-size 50KB --output {tmpdir}"
    Then "{tmpdir}/DICOMDIR" should exist
    And "{tmpdir}" should have patient/study/series hierarchy

  Scenario: Missing required flags
    Given dicomforge is built
    When I run dicomforge with "--num-images 5"
    Then the exit code should be 1
    And the output should contain "total-size is required"
```

## GitHub Actions

Add e2e job to `.github/workflows/ci.yml`:

```yaml
e2e:
  name: E2E Tests
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: "1.24"
    - name: Install DCMTK
      run: sudo apt-get update && sudo apt-get install -y dcmtk
    - name: Run E2E tests
      run: go test ./tests/e2e/... -v
```

Job is required (not optional) for PR merging.

## Implementation Notes

- Binary compiled once in `TestMain()`, reused by all scenarios
- Small files (50KB, 3 images) to respect GitHub Actions resource limits
- `{tmpdir}` replaced dynamically with unique temp directory per scenario
- `dcmdump -q` returns 0 if file is valid DICOM

## Dependencies

- `github.com/cucumber/godog`
- DCMTK package (`apt-get install dcmtk`)
