package cert

import (
	"crypto/tls"
	"os"
	"xdp-banner/pkg/cert"
)

const (
	LocalCertDir   = "/etc/xdp-banner/orch"
	LocalCaFile    = LocalCertDir + "/ca.pem"
	LocalCaKeyFile = LocalCertDir + "/ca.key"
	LocalCertFile  = LocalCertDir + "/cert.pem"
	LocalKeyFile   = LocalCertDir + "/cert.key"
)

func GetLocalCaFile() ([]byte, error) {
	return cert.ReadPemFile(LocalCaFile)
}

func GetLocalCaKeyFile() ([]byte, error) {
	return cert.ReadPemFile(LocalCaKeyFile)
}

func GetLocalCertFile() ([]byte, error) {
	return cert.ReadPemFile(LocalCertFile)
}

func GetLocalKeyFile() ([]byte, error) {
	return cert.ReadPemFile(LocalKeyFile)
}

func GetLocalCertPair() (tls.Certificate, error) {
	return tls.LoadX509KeyPair(LocalCertFile, LocalKeyFile)
}

func StoreLocalCaFile(pemCert []byte) error {
	return cert.WritePemFile(LocalCaFile, pemCert)
}

func StoreLocalCaKeyFile(pemKey []byte) error {
	return cert.WritePemFile(LocalCaKeyFile, pemKey)
}

func StoreLocalCertFile(pemCert []byte) error {
	return cert.WritePemFile(LocalCertFile, pemCert)
}

func StoreLocalKeyFile(pemKey []byte) error {
	return cert.WritePemFile(LocalKeyFile, pemKey)
}

func DeleteLocalCertDir() error {
	return os.RemoveAll(LocalCertDir)
}
