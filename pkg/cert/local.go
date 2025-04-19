package cert

import (
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func CheckPemValidity(cert []byte) bool {
	if b, _ := pem.Decode(cert); b == nil {
		return false
	}
	return true
}

func WritePemFile(path string, pem []byte) error {
	if !CheckPemValidity(pem) {
		return fmt.Errorf("not a valid PEM certificate")
	}

	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0770); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, pem, 0600); err != nil {
		return fmt.Errorf("failed to write certificate to %s: %w", path, err)
	}

	return nil
}

func ReadPemFile(path string) ([]byte, error) {
	cert, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if !CheckPemValidity(cert) {
		return nil, fmt.Errorf("invalid PEM certificate")
	}
	return cert, nil
}
