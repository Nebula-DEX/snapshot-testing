package networkutils

import (
	"fmt"
	"os"
	"path/filepath"
)

type PathManager struct {
	workDir string
}

func NewPathManager(workDir string) PathManager {
	return PathManager{
		workDir: workDir,
	}
}

func (pm PathManager) WorkDir() string {
	return pm.workDir
}

func (pm PathManager) Logs() string {
	return filepath.Join(pm.workDir, "logs")
}

func (pm PathManager) Binaries() string {
	return filepath.Join(pm.workDir, "bins")
}

func (pm PathManager) VegaHome() string {
	return filepath.Join(pm.workDir, "vega_home")
}

func (pm PathManager) VisorHome() string {
	return filepath.Join(pm.workDir, "visor_home")
}

func (pm PathManager) TendermintHome() string {
	return filepath.Join(pm.workDir, "tendermint_home")
}

func (pm PathManager) VegaBin() string {
	return filepath.Join(pm.Binaries(), "vega")
}

func (pm PathManager) VisorBin() string {
	return filepath.Join(pm.Binaries(), "visor")
}

func (pm PathManager) LogFile(fileName string) string {
	return filepath.Join(pm.Logs(), fileName)
}

func (pm PathManager) Results() string {
	return filepath.Join(pm.workDir, "results.json")
}

func (pm PathManager) fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}

	return err == nil
}

func (pm PathManager) CreateDirectoryStructure() error {
	// logger.Sugar().Infof("Ensure the working directory(%s) exists", path) // TODO: We have no loger at this point
	if err := os.MkdirAll(pm.workDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create working directory")
	}

	// logger.Sugar().Infof("Ensure the directory for logs (%s) exists", logsDir)
	if err := os.MkdirAll(pm.Logs(), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create logs directory")
	}

	return nil
}

func (pm PathManager) IsNodeInitialized() bool {
	return pm.fileExists(pm.Binaries()) &&
		pm.fileExists(pm.VegaBin()) &&
		pm.fileExists(pm.VisorBin()) &&
		pm.fileExists(pm.VegaHome()) &&
		pm.fileExists(pm.TendermintHome()) &&
		pm.fileExists(pm.VisorHome())
}

func (pm PathManager) AreBinariesDownloaded() bool {
	return pm.fileExists(pm.Binaries()) &&
		pm.fileExists(pm.VegaBin()) &&
		pm.fileExists(pm.VisorBin())
}
