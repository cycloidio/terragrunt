package test_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terragrunt/test/helpers"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testFixtureDryRun = "fixtures/dry-run"
)

// TestDryRunSkipsTerraformExecution verifies that the --terragrunt-dry-run flag
// causes Terragrunt to validate configuration but skip actual Terraform execution.
func TestDryRunSkipsTerraformExecution(t *testing.T) {
	t.Parallel()

	// Set up test environment
	helpers.CleanupTerraformFolder(t, testFixtureDryRun)
	tmpEnvPath := helpers.CopyEnvironment(t, testFixtureDryRun)
	rootPath := util.JoinPath(tmpEnvPath, testFixtureDryRun, "unit-a")

	// Run terragrunt with --terragrunt-dry-run flag
	cmd := "terragrunt run plan --terragrunt-dry-run --non-interactive --working-dir " + rootPath

	stdout, stderr, err := helpers.RunTerragruntCommandWithOutput(t, cmd)
	require.NoError(t, err)

	// Verify the dry-run message is present
	combinedOutput := stdout + stderr
	assert.Contains(t, combinedOutput, "Dry-run mode enabled: Terragrunt validation complete, skipping Terraform/OpenTofu execution")

	// Verify that no terraform plan output is shown (no resource changes)
	assert.NotContains(t, combinedOutput, "Terraform will perform the following actions")
	assert.NotContains(t, combinedOutput, "Plan:")

	// Verify that the test output file was NOT created (terraform was not executed)
	testOutputFile := filepath.Join(rootPath, ".terragrunt-cache")
	entries, err := os.ReadDir(testOutputFile)
	if err == nil && len(entries) > 0 {
		// Find the cache directory and check for test-output.txt
		for _, entry := range entries {
			if entry.IsDir() {
				cacheSubdirs, _ := os.ReadDir(filepath.Join(testOutputFile, entry.Name()))
				for _, subdir := range cacheSubdirs {
					if subdir.IsDir() {
						testFile := filepath.Join(testOutputFile, entry.Name(), subdir.Name(), "test-output.txt")
						assert.NoFileExists(t, testFile, "test-output.txt should not exist because terraform was not executed")
					}
				}
			}
		}
	}
}

// TestDryRunWithAllFlag verifies that --terragrunt-dry-run works correctly with --all flag.
func TestDryRunWithAllFlag(t *testing.T) {
	t.Parallel()

	// Set up test environment
	helpers.CleanupTerraformFolder(t, testFixtureDryRun)
	tmpEnvPath := helpers.CopyEnvironment(t, testFixtureDryRun)
	rootPath := util.JoinPath(tmpEnvPath, testFixtureDryRun)

	// Run terragrunt with --all and --terragrunt-dry-run flags
	cmd := "terragrunt run plan --all --terragrunt-dry-run --non-interactive --working-dir " + rootPath

	stdout, stderr, err := helpers.RunTerragruntCommandWithOutput(t, cmd)
	require.NoError(t, err)

	// Verify the dry-run message is present (should appear for each unit)
	combinedOutput := stdout + stderr
	assert.Contains(t, combinedOutput, "Dry-run mode enabled: Terragrunt validation complete, skipping Terraform/OpenTofu execution")

	// Verify the run summary shows success
	assert.Contains(t, combinedOutput, "Succeeded")

	// Verify that no terraform plan output is shown
	assert.NotContains(t, combinedOutput, "Terraform will perform the following actions")
}

// TestDryRunApplyDoesNotModifyState verifies that --terragrunt-dry-run with apply
// does not actually create resources or modify state.
func TestDryRunApplyDoesNotModifyState(t *testing.T) {
	t.Parallel()

	// Set up test environment
	helpers.CleanupTerraformFolder(t, testFixtureDryRun)
	tmpEnvPath := helpers.CopyEnvironment(t, testFixtureDryRun)
	rootPath := util.JoinPath(tmpEnvPath, testFixtureDryRun, "unit-a")

	// Run terragrunt apply with --terragrunt-dry-run flag
	cmd := "terragrunt run apply --terragrunt-dry-run --non-interactive --working-dir " + rootPath + " -- -auto-approve"

	stdout, stderr, err := helpers.RunTerragruntCommandWithOutput(t, cmd)
	require.NoError(t, err)

	// Verify the dry-run message is present
	combinedOutput := stdout + stderr
	assert.Contains(t, combinedOutput, "Dry-run mode enabled: Terragrunt validation complete, skipping Terraform/OpenTofu execution")

	// Verify no state file was created
	stateFile := filepath.Join(rootPath, "terraform.tfstate")
	assert.NoFileExists(t, stateFile, "terraform.tfstate should not exist because terraform was not executed")
}

// TestDryRunShowsOnlyOneMessagePerUnit verifies that the dry-run message
// is shown only once per unit (not duplicated for init).
func TestDryRunShowsOnlyOneMessagePerUnit(t *testing.T) {
	t.Parallel()

	// Set up test environment
	helpers.CleanupTerraformFolder(t, testFixtureDryRun)
	tmpEnvPath := helpers.CopyEnvironment(t, testFixtureDryRun)
	rootPath := util.JoinPath(tmpEnvPath, testFixtureDryRun, "unit-a")

	// Clean any existing .terragrunt-cache to force init
	cacheDir := filepath.Join(rootPath, ".terragrunt-cache")
	_ = os.RemoveAll(cacheDir)

	// Run terragrunt with --terragrunt-dry-run flag
	cmd := "terragrunt run plan --terragrunt-dry-run --non-interactive --working-dir " + rootPath

	stdout, stderr, err := helpers.RunTerragruntCommandWithOutput(t, cmd)
	require.NoError(t, err)

	// Count occurrences of the dry-run message
	combinedOutput := stdout + stderr
	dryRunMessage := "Dry-run mode enabled: Terragrunt validation complete, skipping Terraform/OpenTofu execution"
	count := strings.Count(combinedOutput, dryRunMessage)

	// Should appear exactly once
	assert.Equal(t, 1, count, "Dry-run message should appear exactly once, but appeared %d times", count)
}

// TestDryRunWithEnvVar verifies that the TERRAGRUNT_DRY_RUN environment variable works.
// Note: This test cannot be parallel because it uses t.Setenv
func TestDryRunWithEnvVar(t *testing.T) {
	// Set up test environment
	helpers.CleanupTerraformFolder(t, testFixtureDryRun)
	tmpEnvPath := helpers.CopyEnvironment(t, testFixtureDryRun)
	rootPath := util.JoinPath(tmpEnvPath, testFixtureDryRun, "unit-a")

	// Set the environment variable
	t.Setenv("TG_TERRAGRUNT_DRY_RUN", "true")

	// Run terragrunt without the flag (should still use dry-run from env var)
	cmd := "terragrunt run plan --non-interactive --working-dir " + rootPath

	stdout, stderr, err := helpers.RunTerragruntCommandWithOutput(t, cmd)
	require.NoError(t, err)

	// Verify the dry-run message is present
	combinedOutput := stdout + stderr
	assert.Contains(t, combinedOutput, "Dry-run mode enabled: Terragrunt validation complete, skipping Terraform/OpenTofu execution")
}

// TestDryRunStillRunsTerraformInit verifies that dry-run still runs terraform init
// (to validate the configuration) but skips the actual command.
func TestDryRunStillRunsTerraformInit(t *testing.T) {
	t.Parallel()

	// Set up test environment
	helpers.CleanupTerraformFolder(t, testFixtureDryRun)
	tmpEnvPath := helpers.CopyEnvironment(t, testFixtureDryRun)
	rootPath := util.JoinPath(tmpEnvPath, testFixtureDryRun, "unit-a")

	// Clean any existing .terragrunt-cache to force init
	cacheDir := filepath.Join(rootPath, ".terragrunt-cache")
	_ = os.RemoveAll(cacheDir)

	// Run terragrunt with --terragrunt-dry-run flag
	cmd := "terragrunt run plan --terragrunt-dry-run --non-interactive --working-dir " + rootPath

	stdout, stderr, err := helpers.RunTerragruntCommandWithOutput(t, cmd)
	require.NoError(t, err)

	// Verify that terraform init was executed (initialization messages present)
	combinedOutput := stdout + stderr
	assert.Contains(t, combinedOutput, "Initializing")

	// But the actual plan was skipped
	assert.Contains(t, combinedOutput, "Dry-run mode enabled")
}
