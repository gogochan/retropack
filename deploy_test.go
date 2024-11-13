package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadAndSaveURL(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatalf("error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test server to serve a test file
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".zip") {
			http.ServeFile(w, r, "testdata/test.zip")
		} else {
			w.Write([]byte("test content"))
		}
	}))
	defer testServer.Close()

	// Define test cases
	tests := []struct {
		name       string
		artifact   Artifact
		expectErr  bool
		verifyFunc func(t *testing.T, artifact Artifact)
	}{
		{
			name: "Download and save non-zip file",
			artifact: Artifact{
				Url:        testServer.URL,
				TargetPath: tempDir,
			},
			expectErr: false,
			verifyFunc: func(t *testing.T, artifact Artifact) {
				finalPath := filepath.Join(artifact.TargetPath, filepath.Base(artifact.Url))
				if _, err := os.Stat(finalPath); os.IsNotExist(err) {
					t.Errorf("expected file to be downloaded and saved at %s", finalPath)
				}
			},
		},
		{
			name: "Download and extract zip file",
			artifact: Artifact{
				Url:        testServer.URL + "/test.zip",
				TargetPath: tempDir,
			},
			expectErr: false,
			verifyFunc: func(t *testing.T, artifact Artifact) {
				// Verify that the zip file was extracted
				extractedFilePath := filepath.Join(artifact.TargetPath, "test")
				if _, err := os.Stat(extractedFilePath); os.IsNotExist(err) {
					t.Errorf("expected zip file to be extracted at %s", extractedFilePath)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.artifact.DownloadAndSaveURL()
			if (err != nil) != tt.expectErr {
				t.Errorf("DownloadAndSaveURL() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if tt.verifyFunc != nil {
				tt.verifyFunc(t, tt.artifact)
			}
		})
	}
}
func TestRenderPowershellSetPathCommand(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		scope  string
		expect string
	}{
		{
			name:  "Machine scope",
			path:  `C:\example\path`,
			scope: "Machine",
			expect: `
	# Persistently add the target path to the PATH environment variable
	$curPath = [Environment]::GetEnvironmentVariable('Path', 'Machine')
	[Environment]::SetEnvironmentVariable('Path', $curPath + ';C:\example\path', 'Machine')

	# Add the target path to the PATH environment variable for the current session
	$env:Path += $env:Path + ';C:\example\path'
	`,
		},
		{
			name:  "User scope",
			path:  `D:\another\path`,
			scope: "User",
			expect: `
	# Persistently add the target path to the PATH environment variable
	$curPath = [Environment]::GetEnvironmentVariable('Path', 'User')
	[Environment]::SetEnvironmentVariable('Path', $curPath + ';D:\another\path', 'User')

	# Add the target path to the PATH environment variable for the current session
	$env:Path += $env:Path + ';D:\another\path'
	`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderPowershellSetPathCommand(tt.path, tt.scope)
			if got != tt.expect {
				t.Errorf("renderPowershellSetPathCommand() = %v, want %v", got, tt.expect)
			}
		})
	}
}
