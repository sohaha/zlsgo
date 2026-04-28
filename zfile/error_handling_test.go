package zfile

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWriteFileErrorHandling verifies that WriteFile properly handles errors
func TestWriteFileErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setup       func() string
		expectError bool
		description string
		useAppend   bool
	}{
		{
			name: "successful write",
			setup: func() string {
				return filepath.Join(tmpDir, "success.txt")
			},
			expectError: false,
			description: "normal file write should succeed",
		},
		{
			name: "write to read-only file",
			setup: func() string {
				filePath := filepath.Join(tmpDir, "readonly.txt")
				file, _ := os.Create(filePath)
				file.Close()
				os.Chmod(filePath, 0o444)
				return filePath
			},
			expectError: true,
			description: "writing to read-only file should fail",
		},
		{
			name: "write to invalid path",
			setup: func() string {
				if _, err := os.Stat("/dev/full"); err == nil {
					return "/dev/full/test.txt"
				}
				return filepath.Join(tmpDir, "\x00invalid.txt")
			},
			expectError: true,
			description: "writing to invalid path should fail",
		},
		{
			name: "append to existing file",
			setup: func() string {
				filePath := filepath.Join(tmpDir, "append.txt")
				WriteFile(filePath, []byte("initial"))
				return filePath
			},
			expectError: false,
			description: "appending to existing file should succeed",
			useAppend:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup()

			var err error
			if tt.useAppend {
				err = WriteFile(filePath, []byte("test content"), true)
			} else {
				err = WriteFile(filePath, []byte("test content"))
			}

			if tt.expectError {
				if err == nil {
					t.Errorf("%s: expected error but got none", tt.description)
				}
				t.Logf("Expected error occurred: %v", err)
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.description, err)
				}
				if tt.name == "successful write" || tt.name == "append to existing file" {
					content, err := os.ReadFile(filePath)
					if err != nil {
						t.Errorf("Failed to read written file: %v", err)
					} else {
						expectedContent := "test content"
						if tt.name == "append to existing file" {
							expectedContent = "initialtest content"
						}
						if string(content) != expectedContent {
							t.Errorf("Expected content '%s', got '%s'", expectedContent, string(content))
						}
					}
				}
			}
		})
	}
}

// TestPutAppendErrorHandling verifies that PutAppend properly handles errors
func TestPutAppendErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setup       func() string
		expectError bool
	}{
		{
			name: "append to new file",
			setup: func() string {
				return filepath.Join(tmpDir, "new.txt")
			},
			expectError: false,
		},
		{
			name: "append to existing file",
			setup: func() string {
				filePath := filepath.Join(tmpDir, "existing.txt")
				WriteFile(filePath, []byte("initial"))
				return filePath
			},
			expectError: false,
		},
		{
			name: "append to read-only file",
			setup: func() string {
				filePath := filepath.Join(tmpDir, "readonly.txt")
				file, _ := os.Create(filePath)
				file.Close()
				os.Chmod(filePath, 0o444)
				return filePath
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup()

			err := PutAppend(filePath, []byte("appended"))

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				t.Logf("Expected error occurred: %v", err)
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				content, _ := os.ReadFile(filePath)
				switch tt.name {
				case "append to new file":
					if string(content) != "appended" {
						t.Errorf("Expected 'appended', got '%s'", string(content))
					}
				case "append to existing file":
					if string(content) != "initialappended" {
						t.Errorf("Expected 'initialappended', got '%s'", string(content))
					}
				}
			}
		})
	}
}

// TestWriteFileDirectoryCreation verifies that WriteFile creates parent directories
func TestWriteFileDirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "level1", "level2", "file.txt")

	err := WriteFile(filePath, []byte("test"))
	if err != nil {
		t.Errorf("Failed to create file with parent directories: %v", err)
	}

	if !FileExist(filePath) {
		t.Error("File was not created")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("Failed to read created file: %v", err)
	}
	if string(content) != "test" {
		t.Errorf("Expected content 'test', got '%s'", string(content))
	}
}

// TestWriteFileWithLargeData verifies error handling with large writes
func TestWriteFileWithLargeData(t *testing.T) {
	tmpDir := t.TempDir()

	largeData := make([]byte, 10*1024*1024) // 10MB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	filePath := filepath.Join(tmpDir, "large.bin")

	err := WriteFile(filePath, largeData)
	if err != nil {
		t.Errorf("Failed to write large file: %v", err)
	}

	info, _ := os.Stat(filePath)
	if info.Size() != int64(len(largeData)) {
		t.Errorf("Expected file size %d, got %d", len(largeData), info.Size())
	}
}
