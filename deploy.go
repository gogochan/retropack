package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var cmdDeploy = &cli.Command{
	Name:      "deploy",
	Usage:     "deploy artifacts described in the spec-file",
	UsageText: "deploy [options] spec-file",
	Action: func(ctx *cli.Context) error {
		specFilePath := ctx.Args().Get(0)

		var specSrc io.Reader
		if specFilePath == "" {
			return fmt.Errorf("spec-file is required")
		} else if specFilePath == "-" {
			specSrc = os.Stdin
		} else {
			file, err := os.Open(specFilePath)
			if err != nil {
				return fmt.Errorf("error opening spec-file: %w", err)
			}
			defer file.Close()
			specSrc = file
		}

		var specs Artifacts
		if err := yaml.NewDecoder(specSrc).Decode(&specs); err != nil {
			return err
		}
		for name, artifact := range specs {
			fmt.Println("Processing artifact: ", name)
			err := artifact.DownloadAndSaveURL()
			if err != nil {
				return err
			}
		}

		return nil
	},
}

type Artifacts map[string]Artifact

type Artifact struct {
	Url        string `yaml:"url"`
	TargetPath string `yaml:"target_path"`
}

func (a Artifact) DownloadAndSaveURL() error {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "download-*")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Download the file
	resp, err := http.Get(a.Url)
	if err != nil {
		return fmt.Errorf("error downloading file: %w", err)
	}
	defer resp.Body.Close()

	// Write the response body to the temporary file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing to temporary file: %w", err)
	}

	// Check if the file is a zip file
	if filepath.Ext(a.Url) == ".zip" {
		// Extract the zip file
		if err := unzip(tempFile.Name(), a.TargetPath); err != nil {
			return fmt.Errorf("error extracting zip file: %w", err)
		}
	} else {
		// Move the file to the desired location
		finalPath := filepath.Join(a.TargetPath, filepath.Base(a.Url))
		if err := os.Rename(tempFile.Name(), finalPath); err != nil {
			return fmt.Errorf("error moving file to final destination: %w", err)
		}
	}

	return nil
}

func unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
