package icert

import (
	"crypto/tls"
	"os"
	"xdp-banner/pkg/cert"
)

const (
	CertDir  = "/etc/xdp-banner/agent"
	CAFile   = CertDir + "/ca.pem"
	CertFile = CertDir + "/cert.pem"
	KeyFile  = CertDir + "/cert.key"
)

func GetCA() ([]byte, error) {
	return cert.ReadPemFile(CAFile)
}

func GetCert() ([]byte, error) {
	return cert.ReadPemFile(CertFile)
}

func GetCertPri() ([]byte, error) {
	return cert.ReadPemFile(KeyFile)
}

func GetCertPair() (tls.Certificate, error) {
	return tls.LoadX509KeyPair(CertFile, KeyFile)
}

func StoreCA(pemCert []byte) error {
	return cert.WritePemFile(CAFile, pemCert)
}

func StoreCert(pemCert []byte) error {
	return cert.WritePemFile(CertFile, pemCert)
}

func StoreCertPri(pemKey []byte) error {
	return cert.WritePemFile(KeyFile, pemKey)
}

func DeleteCertDir() error {
	return os.RemoveAll(CertDir)
}
