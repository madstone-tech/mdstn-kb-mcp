package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/config"
)

func TestNewProfileCmd(t *testing.T) {
	cmd := newProfileCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "profile", cmd.Use)
	assert.True(t, cmd.HasSubCommands())
}

func TestProfileListCmd(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cmd := newProfileListCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "list", cmd.Use)

	// Test table output (default)
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "ACTIVE")
	assert.Contains(t, output, "STORAGE")
}

func TestProfileListCmd_JSON(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cmd := newProfileListCmd()

	// Set JSON output flag
	err := cmd.Flags().Set("output", "json")
	require.NoError(t, err)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.RunE(cmd, []string{})
	assert.NoError(t, err)

	// Verify it's valid JSON
	var profiles []config.ProfileInfo
	err = json.Unmarshal(buf.Bytes(), &profiles)
	assert.NoError(t, err)
	assert.Len(t, profiles, 1) // Should have default profile
	assert.Equal(t, "default", profiles[0].Name)
}

func TestProfileCreateCmd(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cmd := newProfileCreateCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "create <profile-name>", cmd.Use)

	// Test basic profile creation
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Mock stdin to respond "n" to switch prompt
	stdin := strings.NewReader("n\n")
	cmd.SetIn(stdin)

	err := cmd.RunE(cmd, []string{"test-profile"})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Profile 'test-profile' created successfully")
}

func TestProfileCreateCmd_WithFlags(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cmd := newProfileCreateCmd()

	// Set flags
	err := cmd.Flags().Set("storage-type", "s3")
	require.NoError(t, err)
	err = cmd.Flags().Set("s3-bucket", "test-bucket")
	require.NoError(t, err)
	err = cmd.Flags().Set("s3-region", "us-east-1")
	require.NoError(t, err)
	err = cmd.Flags().Set("vault-name", "test-vault")
	require.NoError(t, err)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Mock stdin to respond "n" to switch prompt
	stdin := strings.NewReader("n\n")
	cmd.SetIn(stdin)

	err = cmd.RunE(cmd, []string{"s3-profile"})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Profile 's3-profile' created successfully")

	// Verify the profile was created with correct settings
	pm, err := config.NewProfileManager()
	require.NoError(t, err)

	profileConfig, err := pm.GetConfig("s3-profile")
	assert.NoError(t, err)
	assert.Equal(t, "test-vault", profileConfig.Vault.Name)
	assert.Equal(t, "s3", string(profileConfig.Storage.Type))
	assert.Equal(t, "test-bucket", profileConfig.Storage.S3.Bucket)
	assert.Equal(t, "us-east-1", profileConfig.Storage.S3.Region)
}

func TestProfileCreateCmd_InvalidArgs(t *testing.T) {
	cmd := newProfileCreateCmd()

	// Test with no arguments - should validate args first
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err)
}

func TestProfileDeleteCmd(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// First create a profile to delete
	pm, err := config.NewProfileManager()
	require.NoError(t, err)
	err = pm.CreateProfile("delete-me", nil)
	require.NoError(t, err)

	cmd := newProfileDeleteCmd()

	// Set force flag to avoid confirmation prompt
	err = cmd.Flags().Set("force", "true")
	require.NoError(t, err)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.RunE(cmd, []string{"delete-me"})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Profile 'delete-me' deleted successfully")
}

func TestProfileDeleteCmd_WithConfirmation(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// First create a profile to delete
	pm, err := config.NewProfileManager()
	require.NoError(t, err)
	err = pm.CreateProfile("delete-me-confirm", nil)
	require.NoError(t, err)

	cmd := newProfileDeleteCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Mock stdin to respond "y" to confirmation
	stdin := strings.NewReader("y\n")
	cmd.SetIn(stdin)

	err = cmd.RunE(cmd, []string{"delete-me-confirm"})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Profile 'delete-me-confirm' deleted successfully")
}

func TestProfileDeleteCmd_DefaultProfile(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cmd := newProfileDeleteCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Mock stdin to respond "y" to confirmation so we get the actual delete error
	stdin := strings.NewReader("y\n")
	cmd.SetIn(stdin)

	err := cmd.RunE(cmd, []string{"default"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete the default profile")
}

func TestProfileSwitchCmd(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// First create a profile to switch to
	pm, err := config.NewProfileManager()
	require.NoError(t, err)
	err = pm.CreateProfile("switch-to-me", nil)
	require.NoError(t, err)

	cmd := newProfileSwitchCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.RunE(cmd, []string{"switch-to-me"})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Switched to profile 'switch-to-me'")

	// Verify the switch worked by creating a new ProfileManager instance
	pm2, err := config.NewProfileManager()
	require.NoError(t, err)
	assert.Equal(t, "switch-to-me", pm2.GetActiveProfile())
}

func TestProfileShowCmd(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cmd := newProfileShowCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Show default profile
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Profile: default")
}

func TestProfileShowCmd_JSON(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cmd := newProfileShowCmd()

	// Set JSON output flag
	err := cmd.Flags().Set("output", "json")
	require.NoError(t, err)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.RunE(cmd, []string{})
	assert.NoError(t, err)

	// Verify it's valid JSON
	var config map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &config)
	assert.NoError(t, err)
	assert.Contains(t, config, "vault")
	assert.Contains(t, config, "storage")
}

func TestProfileCopyCmd(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// First create a source profile
	pm, err := config.NewProfileManager()
	require.NoError(t, err)
	err = pm.CreateProfile("source", nil)
	require.NoError(t, err)

	cmd := newProfileCopyCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.RunE(cmd, []string{"source", "target"})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Profile 'source' copied to 'target' successfully")

	// Verify the copy exists
	profiles, err := pm.ListProfiles()
	assert.NoError(t, err)
	found := false
	for _, p := range profiles {
		if p.Name == "target" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestProfileSetGetCmd(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// First create a profile
	pm, err := config.NewProfileManager()
	require.NoError(t, err)
	err = pm.CreateProfile("test-set-get", nil)
	require.NoError(t, err)

	// Test set command
	setCmd := newProfileSetCmd()
	var setBuf bytes.Buffer
	setCmd.SetOut(&setBuf)

	err = setCmd.RunE(setCmd, []string{"test-set-get", "vault.name", "test-vault"})
	assert.NoError(t, err)

	output := setBuf.String()
	assert.Contains(t, output, "Set test-set-get.vault.name = test-vault")

	// Test get command
	getCmd := newProfileGetCmd()
	var getBuf bytes.Buffer
	getCmd.SetOut(&getBuf)

	err = getCmd.RunE(getCmd, []string{"test-set-get", "vault.name"})
	assert.NoError(t, err)

	output = getBuf.String()
	assert.Contains(t, output, "test-vault")
}

func TestStorageTypeValue(t *testing.T) {
	var st storageTypeValue

	// Test valid values
	err := st.Set("local")
	assert.NoError(t, err)
	assert.Equal(t, "local", st.String())

	err = st.Set("s3")
	assert.NoError(t, err)
	assert.Equal(t, "s3", st.String())

	// Test invalid value
	err = st.Set("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid storage type")

	// Test type method
	assert.Equal(t, "storage-type", st.Type())
}

func TestPrintFunctions(t *testing.T) {
	// Test print functions with sample data
	profiles := []config.ProfileInfo{
		{
			Name:        "default",
			IsActive:    true,
			IsDefault:   true,
			StorageType: "local",
		},
		{
			Name:        "work",
			IsActive:    false,
			IsDefault:   false,
			StorageType: "s3",
		},
	}

	// Test table output
	var buf bytes.Buffer
	err := printProfilesTable(&buf, profiles)
	assert.NoError(t, err)

	// Test JSON output
	buf.Reset()
	err = printProfilesJSON(&buf, profiles)
	assert.NoError(t, err)
}

// Integration test for the complete profile workflow
func TestProfileWorkflow(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create profile manager
	pm, err := config.NewProfileManager()
	require.NoError(t, err)

	// 1. List profiles (should only have default)
	profiles, err := pm.ListProfiles()
	assert.NoError(t, err)
	assert.Len(t, profiles, 1)
	assert.Equal(t, "default", profiles[0].Name)

	// 2. Create a new profile
	err = pm.CreateProfile("workflow-test", &config.CreateProfileOptions{
		StorageType: "s3",
		S3Bucket:    "workflow-bucket",
		S3Region:    "us-east-1",
		VaultName:   "workflow-vault",
	})
	assert.NoError(t, err)

	// 3. List profiles again (should have 2)
	profiles, err = pm.ListProfiles()
	assert.NoError(t, err)
	assert.Len(t, profiles, 2)

	// 4. Switch to new profile
	err = pm.SwitchProfile("workflow-test")
	assert.NoError(t, err)
	assert.Equal(t, "workflow-test", pm.GetActiveProfile())

	// 5. Verify configuration
	config, err := pm.GetConfig("workflow-test")
	assert.NoError(t, err)
	assert.Equal(t, "workflow-vault", config.Vault.Name)
	assert.Equal(t, "s3", string(config.Storage.Type))
	assert.Equal(t, "workflow-bucket", config.Storage.S3.Bucket)

	// 6. Copy profile
	err = pm.CopyProfile("workflow-test", "workflow-copy")
	assert.NoError(t, err)

	// 7. Verify copy has same config
	copyConfig, err := pm.GetConfig("workflow-copy")
	assert.NoError(t, err)
	assert.Equal(t, config.Vault.Name, copyConfig.Vault.Name)
	assert.Equal(t, config.Storage.S3.Bucket, copyConfig.Storage.S3.Bucket)

	// 8. Delete the copy
	err = pm.DeleteProfile("workflow-copy")
	assert.NoError(t, err)

	// 9. Verify it's gone
	profiles, err = pm.ListProfiles()
	assert.NoError(t, err)
	assert.Len(t, profiles, 2) // default + workflow-test
}
