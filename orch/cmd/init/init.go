package init

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"xdp-banner/orch/cmd/global"

	certi "xdp-banner/orch/internal/cert"

	mnode "xdp-banner/orch/model/node"

	certs "xdp-banner/orch/storage/orch/cert"
	nodes "xdp-banner/orch/storage/orch/node"

	"xdp-banner/pkg/cert"
	"xdp-banner/pkg/errors"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/node"

	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

func Cmd(parentOpt *global.ControllerOptions) *cobra.Command {
	opt := DefaultOption(parentOpt)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opt.Check(); err != nil {
				return err
			}

			orchdInit(opt)
			return nil
		},
	}

	opt.SetFlags(cmd)

	return cmd
}

type initCmd struct {
	cli   etcd.Client
	certs certs.Storage
	infos nodes.InfoStorage

	token string
}

func orchdInit(opt *Option) {

	var cli etcd.Client
	var err error

	// Cli init
	if opt.Parent.Etcd.Authentication.Enabled {
		cli, err = etcd.New(clientv3.Config{
			Endpoints:   opt.Parent.Etcd.Endpoints,
			Username:    opt.Parent.Etcd.Authentication.Username,
			Password:    opt.Parent.Etcd.Authentication.Password,
			DialTimeout: opt.Parent.Etcd.DialTimeout,
		})
	} else {
		cli, err = etcd.New(clientv3.Config{
			Endpoints:   opt.Parent.Etcd.Endpoints,
			DialTimeout: opt.Parent.Etcd.DialTimeout,
		})
	}

	if err != nil {
		log.Fatal("Connect etcd failed", zap.String("err", err.Error()))
	}

	cs := certs.New(cli)
	ois := nodes.NewInfoStorage(cli)

	ic := &initCmd{
		cli:   cli,
		certs: cs,
		infos: ois,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ic.init(ctx)
}

func (ic *initCmd) init(ctx context.Context) {
	// 检查 cluster 是否已存在
	log.Info("Checking cluster exist")
	if exist, err := ic.certs.CAExists(ctx); err != nil {
		log.Fatal("Check cluster exist failed", zap.String("err", err.Error()))
	} else if exist {
		log.Fatal("Cluster already exists, please run orch reset to clean up.")
	}

	// start init
	err := ic.startInit(ctx)
	if err != nil {
		log.Info("Init failed, undoing")
		errs := ic.undoInit(ctx)
		for _, e := range errs {
			log.Error("Undo failed", zap.String("err", e.Error()))
		}
		log.Fatal("Init failed", zap.String("err", err.Error()))
	}

	ic.printResult()
}

func (ic *initCmd) startInit(ctx context.Context) error {
	if token, err := generateRandomToken(32); err != nil {
		return fmt.Errorf("failed to generate random token: %w", err)
	} else {
		ic.token = token
	}

	ca, priv, err := ic.generateClusterCA(ctx)
	if err != nil {
		return err
	}

	if err = ic.generateNodeCert(ca, priv); err != nil {
		return err
	}

	if err = ic.registerNode(ctx); err != nil {
		return fmt.Errorf("failed to update node info: %w", err)
	}

	return nil
}

func (ic *initCmd) generateClusterCA(ctx context.Context) (ca []byte, priv []byte, err error) {
	// generate ca
	log.Info("Generating CA")
	ca, priv, err = cert.GenerateCA(ic.token)
	if err != nil {
		return nil, nil, fmt.Errorf("generate CA: %w", err)
	}
	// upload CA
	log.Info("Uploading CA")
	if err = ic.certs.UploadCAPair(ctx, ca, priv); err != nil {
		return nil, nil, fmt.Errorf("upload CA: %w", err)
	}
	// save CA
	log.Info("Saving CA")
	if err = certi.StoreLocalCaFile(ca); err != nil {
		return nil, nil, fmt.Errorf("save CA: %w", err)
	}
	// ignore error, since encrypt is successful.
	priv, _ = cert.DecryptPem(priv, ic.token)
	if err = certi.StoreLocalCaKeyFile(priv); err != nil {
		return nil, nil, fmt.Errorf("failed to save key: %w", err)
	}

	return ca, priv, nil
}

func (ic *initCmd) generateNodeCert(ca, priv []byte) error {
	log.Info("Generating node cert")
	name, err := node.Name()
	if err != nil {
		return fmt.Errorf("failed to get node name: %w", err)
	}

	nCert, nKey, err := cert.GenerateCert(ca, priv, "", name, "", nil)
	if err != nil {
		return fmt.Errorf("failed to generate node cert: %w", err)
	}

	log.Info("Saving node cert")
	if err = certi.StoreLocalCertFile(nCert); err != nil {
		return fmt.Errorf("failed to save node cert: %w", err)
	}
	if err = certi.StoreLocalKeyFile(nKey); err != nil {
		return fmt.Errorf("failed to save node key: %w", err)
	}

	return nil
}

func (ic *initCmd) printResult() {
	log.Info("===================Init Success======================")
	log.Info("Please run the following command to join the cluster:")
	log.Info("orch join" + " --etcd-endpoints " + ic.cli.Endpoints()[0] + " --token " + ic.token)
	log.Info("")
	log.Info("*Please remember the token, it's the only way to join the cluster")
	log.Info("=====================================================")
}

func (ic *initCmd) undoInit(ctx context.Context) []error {
	return errors.GatherErrors(
		func() error {
			return ic.certs.DeleteDir(ctx)
		},
		func() error {
			return certi.DeleteLocalCertDir()
		},
	)
}

func generateRandomToken(size int) (string, error) {

	token := make([]byte, size)
	// 生成随机字节并填满 token
	_, err := rand.Read(token)
	if err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// 二进制转 16 进制
	return hex.EncodeToString(token), nil
}

// Register a new etcd key at /orch/node/info/{hostname} with value
// {"name":"{hostname}"},"labels":null}
func (ic *initCmd) registerNode(ctx context.Context) error {
	log.Info("Registering current node")

	hostname, err := node.Name()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	info := &mnode.OrchInfo{
		CommonInfo: mnode.CommonInfo{
			Name: hostname,
		},
	}

	err = ic.infos.Add(ctx, info)
	if err != nil {
		return err
	}

	return nil
}
