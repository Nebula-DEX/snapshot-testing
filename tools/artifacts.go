package tools

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DownloadFile(url string, outputFile string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to send get http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid response status code: got %d, expected 200", resp.StatusCode)
	}

	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file(%s): %w", outputFile, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy bytes from http response to output file: %w", err)
	}

	return nil
}

func UnzipFile(zipFile string, outputFolder string) error {
	archive, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(outputFolder, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(outputFolder)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path for file %s: %s", f.Name, filePath)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to extract directory(%s): %w", f.Name, err)
			}

			continue
		}

		dirPath := filepath.Dir(filePath)
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create parent dir for extracting file(%s): %w", dirPath, err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file(%s): %w", filePath, err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			dstFile.Close()
			return fmt.Errorf("failed to open file in archive: %w", err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			dstFile.Close()
			fileInArchive.Close()
			return fmt.Errorf("failed to extract file %s: %w", filePath, err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}

	return nil
}
