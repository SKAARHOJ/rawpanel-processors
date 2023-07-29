package rawpanelproc

import (
	"embed"
	"os"
	"strings"

	log "github.com/s00500/env_logger"
)

//go:embed embedded
var embeddedFS embed.FS

var UseEmbeddedFS bool = true

/************************
 Embedded / Ordinary file system dual functions
************************/

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
