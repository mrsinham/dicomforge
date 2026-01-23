package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

// binaryPath holds the path to the compiled binary (set once in TestMain)
var binaryPath string

// testContext holds state for a single scenario
type testContext struct {
	tmpDir   string
	exitCode int
	output   string
}

// buildBinary compiles the dicomforge binary once
func buildBinary() (string, error) {
	tmpFile, err := os.CreateTemp("", "dicomforge-test-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpFile.Close()

	// Get the directory of this test file to find the project root
	_, thisFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")

	cmd := exec.Command("go", "build", "-o", tmpFile.Name(), "./cmd/dicomforge")
	cmd.Dir = projectRoot
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build failed: %w\n%s", err, stderr.String())
	}

	return tmpFile.Name(), nil
}

// TestMain compiles the binary once before running all tests
func TestMain(m *testing.M) {
	var err error
	binaryPath, err = buildBinary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build binary: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(binaryPath)

	code := m.Run()
	os.Exit(code)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(sc *godog.ScenarioContext) {
	tc := &testContext{}

	// Setup: create temp directory before each scenario
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		tmpDir, err := os.MkdirTemp("", "dicomforge-e2e-*")
		if err != nil {
			return ctx, err
		}
		tc.tmpDir = tmpDir
		return ctx, nil
	})

	// Teardown: cleanup temp directory after each scenario
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if tc.tmpDir != "" {
			os.RemoveAll(tc.tmpDir)
		}
		return ctx, nil
	})

	// Step definitions
	sc.Step(`^dicomforge is built$`, tc.dicomforgeIsBuilt)
	sc.Step(`^I run dicomforge with "([^"]*)"$`, tc.iRunDicomforgeWith)
	sc.Step(`^the exit code should be (\d+)$`, tc.theExitCodeShouldBe)
	sc.Step(`^the output should contain "([^"]*)"$`, tc.theOutputShouldContain)
	sc.Step(`^"([^"]*)" should contain (\d+) DICOM files$`, tc.shouldContainDICOMFiles)
	sc.Step(`^dcmdump should successfully parse all files in "([^"]*)"$`, tc.dcmdumpShouldParse)
	sc.Step(`^"([^"]*)" should exist$`, tc.shouldExist)
	sc.Step(`^"([^"]*)" should have patient/study/series hierarchy$`, tc.shouldHaveHierarchy)
}

func (tc *testContext) dicomforgeIsBuilt() error {
	if binaryPath == "" {
		return fmt.Errorf("binary not built")
	}
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("binary does not exist at %s", binaryPath)
	}
	return nil
}

func (tc *testContext) iRunDicomforgeWith(args string) error {
	// Replace {tmpdir} placeholder with actual temp directory
	args = strings.ReplaceAll(args, "{tmpdir}", tc.tmpDir)

	// Split args respecting quotes
	argList := splitArgs(args)

	cmd := exec.Command(binaryPath, argList...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	tc.output = output.String()

	if exitErr, ok := err.(*exec.ExitError); ok {
		tc.exitCode = exitErr.ExitCode()
	} else if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	} else {
		tc.exitCode = 0
	}

	return nil
}

func (tc *testContext) theExitCodeShouldBe(expected int) error {
	if tc.exitCode != expected {
		return fmt.Errorf("expected exit code %d, got %d\nOutput:\n%s", expected, tc.exitCode, tc.output)
	}
	return nil
}

func (tc *testContext) theOutputShouldContain(expected string) error {
	if !strings.Contains(tc.output, expected) {
		return fmt.Errorf("output does not contain %q\nOutput:\n%s", expected, tc.output)
	}
	return nil
}

func (tc *testContext) shouldContainDICOMFiles(path string, count int) error {
	path = strings.ReplaceAll(path, "{tmpdir}", tc.tmpDir)

	files, err := findDICOMFiles(path)
	if err != nil {
		return fmt.Errorf("failed to find DICOM files: %w", err)
	}

	if len(files) != count {
		return fmt.Errorf("expected %d DICOM files, found %d", count, len(files))
	}
	return nil
}

func (tc *testContext) dcmdumpShouldParse(path string) error {
	path = strings.ReplaceAll(path, "{tmpdir}", tc.tmpDir)

	files, err := findDICOMFiles(path)
	if err != nil {
		return fmt.Errorf("failed to find DICOM files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no DICOM files found in %s", path)
	}

	for _, file := range files {
		cmd := exec.Command("dcmdump", "-q", file)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("dcmdump failed for %s: %w\n%s", file, err, stderr.String())
		}
	}
	return nil
}

func (tc *testContext) shouldExist(path string) error {
	path = strings.ReplaceAll(path, "{tmpdir}", tc.tmpDir)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}
	return nil
}

func (tc *testContext) shouldHaveHierarchy(path string) error {
	path = strings.ReplaceAll(path, "{tmpdir}", tc.tmpDir)

	// Check for PT*/ST*/SE* structure
	ptDirs, err := filepath.Glob(filepath.Join(path, "PT*"))
	if err != nil || len(ptDirs) == 0 {
		return fmt.Errorf("no patient directories (PT*) found in %s", path)
	}

	for _, ptDir := range ptDirs {
		stDirs, err := filepath.Glob(filepath.Join(ptDir, "ST*"))
		if err != nil || len(stDirs) == 0 {
			return fmt.Errorf("no study directories (ST*) found in %s", ptDir)
		}

		for _, stDir := range stDirs {
			seDirs, err := filepath.Glob(filepath.Join(stDir, "SE*"))
			if err != nil || len(seDirs) == 0 {
				return fmt.Errorf("no series directories (SE*) found in %s", stDir)
			}
		}
	}
	return nil
}

// findDICOMFiles finds all DICOM image files (IM*) recursively
func findDICOMFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasPrefix(info.Name(), "IM") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// splitArgs splits a command line string into arguments
func splitArgs(s string) []string {
	var args []string
	var current strings.Builder
	inQuote := false

	for _, r := range s {
		switch {
		case r == '"':
			inQuote = !inQuote
		case r == ' ' && !inQuote:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}
