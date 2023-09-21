// unzip is a package for extracting ZIP files
package unzip

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// Max zip file filename length
const maxLength = 150

// Given a source filename and a destination path, extract the ZIP archive
func Extract(zipFilename, destPath string) error {
	// Extract the ZIP file and don't filter out any files
	return FilterExtract(zipFilename, destPath, func(_ string) bool {
		return true
	})
}

// Given a source filename and a destination path, extract the ZIP archive.
// The filter function can be used to avoid extracting some filenames;
// when filterFunc returns true, the file is extracted.
func FilterExtract(zipFilename, destPath string, filterFunc func(string) bool) error {

	// Open the source filename for reading
	zipReader, err := zip.OpenReader(zipFilename)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	// For each file in the archive
	for _, archiveReader := range zipReader.File {

		// Open the file in the archive
		archiveFile, err := archiveReader.Open()
		if err != nil {
			return err
		}
		defer archiveFile.Close()

		// Prepare to write the file
		finalPath := filepath.Join(destPath, archiveReader.Name)

		// Check if the file to extract is just a directory
		if archiveReader.FileInfo().IsDir() {
			err = os.MkdirAll(finalPath, 0755)
			if err != nil {
				return err
			}
			// Continue to the next file in the archive
			continue
		}

		if !filterFunc(finalPath) {
			// Skip this file
			continue
		}

		if len(archiveReader.Name) >= maxLength {
			return errors.New("Too long filename: " + archiveReader.Name)
		}

		// Create all needed directories
		if os.MkdirAll(filepath.Dir(finalPath), 0755) != nil {
			return err
		}

		// Prepare to write the destination file
		destinationFile, err := os.OpenFile(finalPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, archiveReader.Mode())
		if err != nil {
			return err
		}
		defer destinationFile.Close()

		// Write the destination file
		if _, err = io.Copy(destinationFile, archiveFile); err != nil {
			return err
		}
	}

	return nil
}
