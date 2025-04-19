package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"xdp-banner/orch/internal/cert"

	"google.golang.org/grpc/credentials"
)

func NewCredits() (credentials.TransportCredentials, error) {
	// load ca
	caCert, err := cert.GetLocalCaFile()
	if err != nil {
		return nil, fmt.Errorf("read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("add CA certificate")
	}

	// load cert
	cert, err := cert.GetLocalCertPair()
	if err != nil {
		return nil, fmt.Errorf("load node cert pair: %w", err)
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ClientCAs:    caCertPool,
		ClientAuth:   tls.VerifyClientCertIfGiven,
	}), nil
}

func NewCreditsInsecure() (credentials.TransportCredentials, error) {
	// load ca
	caCert, err := cert.GetLocalCaFile()
	if err != nil {
		return nil, fmt.Errorf("read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("add CA certificate")
	}

	// load cert
	cert, err := cert.GetLocalCertPair()
	if err != nil {
		return nil, fmt.Errorf("load node cert pair: %w", err)
	}

	return credentials.NewTLS(&tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		RootCAs:            caCertPool,
		ClientCAs:          caCertPool,
		ClientAuth:         tls.VerifyClientCertIfGiven,
	}), nil
}
