package rawpanelproc

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/s00500/env_logger"
)

//go:embed embedded
var embeddedFS embed.FS

var UseEmbeddedFS bool = true

func GetIconFS() http.FileSystem {
	fSys, err := fs.Sub(embeddedFS, "embedded/gfx/icons")
	log.Must(err)
	return http.FS(fSys)
}

/************************
 Embedded / Ordinary file system dual functions
************************/

// Read from embedded file system:
func ReadEmbeddedFile(fileName string) []byte {
	fileName = strings.ReplaceAll(fileName, "\\", "/")
	byteValue, err := embeddedFS.ReadFile(fileName)
	log.Should(err)
	return byteValue
}
func ReadEmbeddedFileWithError(fileName string) ([]byte, error) {
	fileName = strings.ReplaceAll(fileName, "\\", "/")
	return embeddedFS.ReadFile(fileName)
}

// Check if ordinary or embedded file exists
func ResourceFileExist(fileName string) bool {
	if fileName == "" {
		return false
	}

	if UseEmbeddedFS && strings.HasPrefix(fileName, "embedded") {
		fileName = strings.ReplaceAll(fileName, "\\", "/")

		f, err := embeddedFS.Open(fileName)
		if err != nil {
			return false
		}

		f.Close()
		return true
	}

	f, err := os.Open(fileName)
	if err != nil {
		return false
	}

	f.Close()
	return true
}

// Check if ordinary or embedded file exists
func ResourceExistIgnoreFinalCase(fileName string) (actualName string, exists bool) {
	if fileName == "" {
		return "", false
	}

	if UseEmbeddedFS && strings.HasPrefix(fileName, "embedded") {
		fileDir := filepath.Dir(fileName)
		fileName = strings.ReplaceAll(fileName, "\\", "/")
		fileDir = strings.ReplaceAll(fileDir, "\\", "/")

		// List base
		entries, err := embeddedFS.ReadDir(fileDir)
		if err != nil {
			return "", false
		}

		for _, dir := range entries {
			if strings.EqualFold(fileName, strings.Join([]string{fileDir, dir.Name()}, "/")) {
				return strings.Join([]string{fileDir, dir.Name()}, "/"), true
			}
		}

		return "", false
	}

	// List base
	entries, err := os.ReadDir(filepath.Dir(fileName))
	if err != nil {
		return "", false
	}

	basePath := filepath.Join(filepath.Dir(fileName))
	for _, dir := range entries {
		if strings.EqualFold(fileName, filepath.Join(basePath, dir.Name())) {
			return filepath.Join(filepath.Dir(fileName), dir.Name()), true
		}
	}
	return "", false
}

// Check if ordinary or embedded directory exists
func ResourceDirExist(fileName string) bool {
	if fileName == "" {
		return false
	}

	if UseEmbeddedFS && strings.HasPrefix(fileName, "embedded") {
		fileName = strings.ReplaceAll(fileName, "\\", "/")
		_, err := embeddedFS.ReadDir(fileName)
		return err == nil
	}

	info, err := os.Stat(fileName)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// Read contents from ordinary or embedded file
func ReadResourceFile(fileName string) []byte {
	byteValue, err := ReadResourceFileWithError(fileName)
	log.Should(err)
	return byteValue
}

// Read contents from ordinary or embedded file
func ReadResourceFileWithError(fileName string) ([]byte, error) {
	if UseEmbeddedFS && strings.HasPrefix(fileName, "embedded") {
		fileName = strings.ReplaceAll(fileName, "\\", "/")
		byteValue, err := embeddedFS.ReadFile(fileName)
		return byteValue, err
	}

	return os.ReadFile(fileName)
}

// Read directory entries from ordinary or embedded FS
func readResourceDirectory(path string) []fs.DirEntry {
	if UseEmbeddedFS && strings.HasPrefix(path, "embedded") {
		path = strings.ReplaceAll(path, "\\", "/")
		directories, err := embeddedFS.ReadDir(path)
		if err != nil {
			return []fs.DirEntry{}
		}
		return directories
	}

	directories, err := os.ReadDir(path)
	if err != nil {
		log.Error(err)
		return []fs.DirEntry{}
	}
	return directories
}
