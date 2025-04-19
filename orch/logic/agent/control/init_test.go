package control

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"
	localcert "xdp-banner/orch/internal/cert"
	"xdp-banner/pkg/node"
)

func TestInit(t *testing.T) {
	// 生成 ecdsa 私钥
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ecdsa private key: %v", err)
	}

	pub := priv.PublicKey

	// encode pub to der
	pubDer, err := x509.MarshalPKIXPublicKey(&pub)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}

	// encode pubDer to pem
	pubPem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDer})

	// sign cert
	name, err := node.Name()
	if err != nil {
		t.Fatalf("failed to get local name: %v", err)
	}
	certPem, err := signCert(name, nil, pubPem)
	if err != nil {
		t.Fatalf("failed to sign cert: %v", err)
	}

	t.Logf("cert: %s", certPem)

	// check cert valid
	caCertPem, err := localcert.GetLocalCaFile()
	if err != nil {
		t.Fatalf("failed to get ca cert: %v", err)
	}

	caBlock, _ := pem.Decode([]byte(caCertPem))
	if caBlock == nil {
		t.Fatalf("failed to parse CA certificate PEM")
	}
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse CA certificate: %v", err)
	}

	// 解析待验证的证书
	certBlock, _ := pem.Decode(certPem)
	if certBlock == nil {
		t.Fatalf("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	// 创建证书池并添加 CA 证书
	certPool := x509.NewCertPool()
	certPool.AddCert(caCert)

	// 验证证书
	opts := x509.VerifyOptions{
		Roots: certPool,
	}
	if _, err := cert.Verify(opts); err != nil {
		t.Fatalf("failed to verify certificate: %v", err)
	}

	t.Log("certificate verified successfully")

}
