package cert

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
	"xdp-banner/pkg/errors"
)

// GenerateCA generates a new self-signed CA certificate and private key in pem format.
// if password is not empty, the private key will be encrypted with the password.
func GenerateCA(password string) (cert []byte, key []byte, err error) {
	priv, err := GeneratePrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ca private key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), // 唯一的序列号
		Subject: pkix.Name{
			Organization: []string{"xdp-banner"},
			Country:      []string{"CN"},
			Province:     []string{"Chongqing"},
			Locality:     []string{"Chongqing"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 有效期 1 年
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true, // 自签名证书是 CA
	}

	certBytes, err := createCert(template, template, priv, &priv.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create ca certificate: %w", err)
	}

	return pairToPem(certBytes, priv, password)
}

// GenerateCert generates a new certificate signed by the CA certificate and private key in pem format.
// if password is not empty, the private key will be encrypted with the password.
func GenerateCert(ca, caPriv []byte, capass, name, password string, ipAddress []net.IP) (cert []byte, key []byte, err error) {
	// 生成子证书的私钥
	priv, err := GeneratePrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("generate private key: %w", err)
	}

	certPem, err := SignCert(ca, caPriv, capass, name, ipAddress, &priv.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("create  certificate: %w", err)
	}

	privPem, err := PrivToPem(priv, password)
	if err != nil {
		return nil, nil, fmt.Errorf("convert private key to PEM format:%w", err)
	}

	return certPem, privPem, nil
}

// SignCert signs a certificate with the CA certificate and private key in pem format.
func SignCert(ca, caPriv []byte, capass string, name string, ipAddress []net.IP, pub any) (certPem []byte, err error) {
	caCert, caKey, err := parsePem(ca, caPriv, capass)
	if err != nil {
		return nil, err
	}

	ipAddress = append([]net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}, ipAddress...)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2), // 唯一的序列号
		Subject: pkix.Name{
			CommonName:   name,
			Organization: []string{"xdp-banner"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour), // 有效期 1 年
		KeyUsage:    x509.KeyUsageDigitalSignature,
		DNSNames:    []string{"localhost", "*.joshua.su", "*.302.kim", "*.xdp-banner.svc.cluster.local", "*.evilsp4.ltd"},
		IPAddresses: ipAddress,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	certBytes, err := createCert(template, caCert, caKey, pub)
	if err != nil {
		return nil, err
	}

	certPem, err = CertToPem(certBytes)
	if err != nil {
		return nil, err
	}

	return certPem, nil
}

func GeneratePrivateKey() (key *ecdsa.PrivateKey, err error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func createCert(template *x509.Certificate, parent *x509.Certificate, parentKey *ecdsa.PrivateKey, pub any) (certBytes []byte, err error) {
	return x509.CreateCertificate(rand.Reader, template, parent, pub, parentKey)
}

func pairToPem(certBytes []byte, priv any, password string) (cert []byte, key []byte, err error) {
	cert, err = CertToPem(certBytes)
	if err != nil {
		return nil, nil, err
	}

	key, err = PrivToPem(priv, password)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

func CertToPem(certBytes []byte) (certPem []byte, err error) {
	var buf bytes.Buffer
	pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	return buf.Bytes(), nil
}

func PrivToPem(pri any, password string) (keyPem []byte, err error) {
	var keyBytes []byte

	switch pri := pri.(type) {
	case *ecdsa.PrivateKey:
		if keyBytes, err = x509.MarshalECPrivateKey(pri); err != nil {
			return nil, fmt.Errorf("failed to marshal private key: %w", err)
		}
	default:
		return nil, errors.NewInputError("not a valid private key")
	}

	var keyBuf bytes.Buffer
	var block *pem.Block
	if password != "" {
		block, err = x509.EncryptPEMBlock(rand.Reader, "EC PRIVATE KEY", keyBytes, []byte(password), x509.PEMCipherAES256)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt ca private key: %w", err)
		}
	} else {
		block = &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}
	}
	pem.Encode(&keyBuf, block)

	return keyBuf.Bytes(), nil
}

func PubKeyToPem(pub any) (pubPem []byte, err error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	var buf bytes.Buffer
	pem.Encode(&buf, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	return buf.Bytes(), nil
}

func parsePem(cert []byte, key []byte, password string) (ca *x509.Certificate, priv *ecdsa.PrivateKey, err error) {
	certBlock, _ := pem.Decode(cert)
	if certBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode certificate pem")
	}
	ca, err = x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	var keyBytes []byte
	keyBlock, _ := pem.Decode(key)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode key pem")
	}
	if x509.IsEncryptedPEMBlock(keyBlock) {
		keyBytes, err = x509.DecryptPEMBlock(keyBlock, []byte(password))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decrypt private key: %w", err)
		}
	} else {
		keyBytes = keyBlock.Bytes
	}
	priv, err = x509.ParseECPrivateKey(keyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return
}

// DecryptPem decrypts the private key in pem format with the password and returns the decrypted private key in pem format.
func DecryptPem(key []byte, password string) (keyBytes []byte, err error) {
	keyBlock, _ := pem.Decode(key)
	if keyBlock == nil {
		return nil, fmt.Errorf("not a valid PEM")
	}
	if !x509.IsEncryptedPEMBlock(keyBlock) {
		return keyBlock.Bytes, nil
	}
	keyBytes, err = x509.DecryptPEMBlock(keyBlock, []byte(password))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	var buf bytes.Buffer

	newBlock := &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}
	if err = pem.Encode(&buf, newBlock); err != nil {
		return nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	return buf.Bytes(), nil
}

func ParsePemPubkey(pemData []byte) (pub any, err error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.NewInputError("not a valid PEM format public key")
	}

	pub, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.NewInputErrorf("not a valid public key: %v", err)
	}

	return pub, nil
}
