package cert

import (
	"context"
	"fmt"
	"xdp-banner/orch/storage/orch"
	"xdp-banner/pkg/etcd"
)

var (
	EtcdDir          = etcd.Join(orch.EtcdDir, "cert")
	EtcdKeyCA        = etcd.Join(EtcdDir, "ca")
	EtcdKeyCAPrivate = etcd.Join(EtcdDir, "ca-private")
)

type Storage struct {
	client etcd.Client
}

func New(cli etcd.Client) Storage {
	return Storage{
		client: cli,
	}
}

// CAExists checks if the cluster certificate exists in etcd.
func (s Storage) CAExists(ctx context.Context) (bool, error) {
	return s.client.Exist(ctx, EtcdKeyCA)
}

// DeleteClusterCert deletes the cluster certificate from etcd.
func (s Storage) DeleteCAPair(ctx context.Context) error {
	return s.client.DeleteWithPrefix(ctx, EtcdDir)
}

// UploadCertCluster uploads the cluster certificate and private key to etcd.
func (s Storage) UploadCAPair(ctx context.Context, cert, key []byte) error {
	err1 := s.UploadCA(ctx, cert)
	err2 := s.UploadCAPrivate(ctx, key)

	if err1 != nil && err2 != nil {
		return fmt.Errorf("failed to upload cluster certificate: %w, %w", err1, err2)
	} else if err1 != nil {
		return fmt.Errorf("failed to upload cluster certificate: %w", err1)
	} else if err2 != nil {
		return fmt.Errorf("failed to upload cluster key: %w", err2)
	}

	return nil
}

// UploadCertClusterCert uploads the cluster certificate to etcd.
func (s Storage) UploadCA(ctx context.Context, cert []byte) error {
	return s.client.Create(ctx, EtcdKeyCA, string(cert))
}

// UploadCertClusterKey uploads the cluster private key to etcd.
func (s Storage) UploadCAPrivate(ctx context.Context, key []byte) error {
	return s.client.Create(ctx, EtcdKeyCAPrivate, string(key))
}

var ErrClusterCertNotFound = fmt.Errorf("cluster certificate not found")

// GetClusterCert gets the cluster certificate from etcd.
func (s Storage) GetCA(ctx context.Context) ([]byte, error) {
	rsp, err := s.client.GetMustExist(ctx, EtcdKeyCA)
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return nil, ErrClusterCertNotFound
		}
		return nil, fmt.Errorf("failed to get cluster certificate: %w", err)
	}

	return []byte(rsp.Kvs[0].Value), nil
}

var ErrClusterKeyNotFound = fmt.Errorf("cluster key not found")

// GetClusterKey gets the cluster private key from etcd.
func (s Storage) GetCAPrivate(ctx context.Context) ([]byte, error) {
	rsp, err := s.client.GetMustExist(ctx, EtcdKeyCAPrivate)
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return nil, ErrClusterKeyNotFound
		}
		return nil, fmt.Errorf("failed to get cluster key: %w", err)
	}

	return []byte(rsp.Kvs[0].Value), nil
}

func (s Storage) DeleteDir(ctx context.Context) error {
	return s.client.DeleteWithPrefix(ctx, EtcdDir)
}
