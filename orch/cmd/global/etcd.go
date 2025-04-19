package global

import (
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/log"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

var Cli etcd.Client
var err error

func CreateGlobalEtcdInstance(opt *ControllerOptions) error {
	if opt.Etcd.Authentication.Enabled {
		Cli, err = etcd.New(clientv3.Config{
			Endpoints:   opt.Etcd.Endpoints,
			Username:    opt.Etcd.Authentication.Username,
			Password:    opt.Etcd.Authentication.Password,
			DialTimeout: opt.Etcd.DialTimeout,
		})
	} else {
		Cli, err = etcd.New(clientv3.Config{
			Endpoints:   opt.Etcd.Endpoints,
			DialTimeout: opt.Etcd.DialTimeout,
		})
	}
	if err != nil {
		log.Error("Etcd Connect failed: ", zap.Error(err))
		return err
	}
	return nil
}
