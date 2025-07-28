# Comprehensive QA Test Commands for Session 5

Here's a systematic test suite to validate all the profile and multi-storage functionality:

## 1. Profile Management Tests

```bash
# Check current active profile
cat ~/.kbvault/active_profile

# List existing profiles
ls -la ~/.kbvault/profiles/

# View profile configurations
cat ~/.kbvault/profiles/personal.toml | grep -A5 "\[storage"
cat ~/.kbvault/profiles/work.toml | grep -A5 "\[storage"

# Test profile switching with --profile flag
./kbvault --profile personal config show storage.type
./kbvault --profile work config show storage.type
```

## 2. Note Creation Tests

```bash
# Test local storage (personal profile)
./kbvault --profile personal new "Local Note Test 1"
./kbvault --profile personal new "Local Note Test 2" --tags personal,test
./kbvault --profile personal new --title "Local Note Test 3" --tags work,important

# Test S3 storage (work profile) 
./kbvault --profile work new "S3 Note Test 1"
./kbvault --profile work new "S3 Note Test 2" --tags cloud,test
./kbvault --profile work new --title "S3 Note Test 3" --tags meeting,important

# Test with default profile (should use active profile)
./kbvault new "Default Profile Test"
```

## 3. Note Listing Tests

```bash
# Test listing with different profiles
./kbvault --profile personal list
./kbvault --profile work list
./kbvault list  # Should use active profile

# Test with different output formats
./kbvault --profile personal list --format json
./kbvault --profile work list --format yaml

# Test with sorting and filtering
./kbvault --profile personal list --sort-by created --reverse
./kbvault --profile work list --tags test
```

## 4. Configuration Tests

```bash
# Test configuration viewing
./kbvault --profile personal config show
./kbvault --profile work config show
./kbvault config show  # Should use active profile

# Test specific config keys
./kbvault --profile personal config show vault.name
./kbvault --profile work config show storage.type
./kbvault --profile personal config show storage.local.path
./kbvault --profile work config show storage.s3.bucket

# Test configuration validation
./kbvault --profile personal config validate
./kbvault --profile work config validate

# Test config path
./kbvault --profile personal config path
./kbvault --profile work config path
```

## 5. Search Tests (if implemented)

```bash
# Test search across different storage backends
./kbvault --profile personal search "test"
./kbvault --profile work search "test"
./kbvault search "note"  # Should use active profile

# Test with filters
./kbvault --profile personal search --tags personal
./kbvault --profile work search --tags cloud
```

## 6. Error Handling Tests

```bash
# Test with non-existent profile
./kbvault --profile nonexistent new "Should fail"

# Test with invalid configuration
# (Temporarily break a config file to test error handling)
cp ~/.kbvault/profiles/work.toml ~/.kbvault/profiles/work.toml.backup
echo "invalid_toml" >> ~/.kbvault/profiles/work.toml
./kbvault --profile work new "Should fail"
mv ~/.kbvault/profiles/work.toml.backup ~/.kbvault/profiles/work.toml

# Test S3 access without credentials (should gracefully handle)
./kbvault --profile work new "S3 Test Without Creds"
```

## 7. Storage Backend Validation

```bash
# Verify local storage creates files in correct location
./kbvault --profile personal new "Local Storage Test"
ls -la ~/code-notes/notes/  # Check if file exists

# Check that notes show correct storage backend in output
./kbvault --profile personal new "Check Storage Type Local"
./kbvault --profile work new "Check Storage Type S3"
```

## 8. Profile Configuration Management

```bash
# Test creating new profiles (interactive - will require input)
./kbvault configure create test-profile

# Test interactive configuration (will require input)
./kbvault --profile personal configure
./kbvault --profile work configure

# Test profile listing (if command works correctly)
./kbvault configure list-profiles
```

## 9. Integration Tests

```bash
# Test workflow: create note with one profile, try to access with another
./kbvault --profile personal new "Personal Note"
./kbvault --profile work new "Work Note" 

# Verify they don't interfere with each other
./kbvault --profile personal list
./kbvault --profile work list
```

## 10. Build and Development Tests

```bash
# Verify build still works
go build ./cmd/kbvault

# Run tests (if any exist)
go test ./... -v

# Check for import issues
go mod tidy
go mod verify

# Test linting
golangci-lint run ./cmd/kbvault/
```

## Expected Outcomes

### Personal Profile (Local Storage)
- ✅ Should create files in `~/code-notes/notes/`
- ✅ Should show `💾 Storage: local` in output
- ✅ Should use local filesystem operations

### Work Profile (S3 Storage)  
- ✅ Should show `💾 Storage: s3` in output
- ✅ Should attempt S3 operations (may fail without proper AWS credentials)
- ✅ Should use S3 configuration from work.toml

### Error Cases to Watch For
- ❌ Profile not found errors
- ❌ Configuration validation failures  
- ❌ Storage backend initialization failures
- ❌ Missing required S3 configuration (region, bucket)

## Quick Test Script

Save this as `qa-test.sh` for rapid testing:

```bash
#!/bin/bash
echo "=== Profile Management Tests ==="
./kbvault --profile personal config show storage.type
./kbvault --profile work config show storage.type

echo -e "\n=== Note Creation Tests ==="
./kbvault --profile personal new "QA Test Personal $(date)"
./kbvault --profile work new "QA Test Work $(date)"

echo -e "\n=== List Tests ==="
./kbvault --profile personal list | head -3
./kbvault --profile work list | head -3

echo -e "\n=== Validation Tests ==="
./kbvault --profile personal config validate
./kbvault --profile work config validate

echo -e "\n=== Test Complete ==="
```

Run with: `chmod +x qa-test.sh && ./qa-test.sh`

## Test Results Log

Use this section to document your QA test results:

### Date: ________________
### Tester: ______________

| Test Category | Status | Notes |
|---------------|--------|-------|
| Profile Management | ⬜ Pass / ⬜ Fail | |
| Note Creation | ⬜ Pass / ⬜ Fail | |
| Note Listing | ⬜ Pass / ⬜ Fail | |
| Configuration | ⬜ Pass / ⬜ Fail | |
| Search | ⬜ Pass / ⬜ Fail | |
| Error Handling | ⬜ Pass / ⬜ Fail | |
| Storage Validation | ⬜ Pass / ⬜ Fail | |
| Profile Config Mgmt | ⬜ Pass / ⬜ Fail | |
| Integration | ⬜ Pass / ⬜ Fail | |
| Build/Development | ⬜ Pass / ⬜ Fail | |

### Issues Found:
- [ ] Issue 1: ________________________________
- [ ] Issue 2: ________________________________
- [ ] Issue 3: ________________________________

### Recommendations:
1. ________________________________________
2. ________________________________________
3. ________________________________________