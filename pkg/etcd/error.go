package etcd

import "errors"

var (
	ErrKeyNotFound = errors.New("etcd key not found")
	ErrKeyExist    = errors.New("etcd key already exists")

	ErrInvalidPageSize = errors.New("invalid page size")
)
