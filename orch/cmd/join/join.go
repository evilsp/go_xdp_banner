package join

import (
	"context"
	"fmt"
	"xdp-banner/orch/cmd/global"

	icert "xdp-banner/orch/internal/cert"

	mnode "xdp-banner/orch/model/node"

	certs "xdp-banner/orch/storage/orch/cert"
	nodes "xdp-banner/orch/storage/orch/node"

	"time"
	"xdp-banner/pkg/cert"
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/node"

	"github.com/spf13/cobra"

	"go.uber.org/zap"
)

func Cmd(parentOpt *global.ControllerOptions) *cobra.Command {
	opt := DefaultOption(parentOpt)

	cmd := &cobra.Command{
		Use:   "join",
		Short: "join an exist cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opt.Check(); err != nil {
				return err
			}

			joinCluster(opt)
			return nil
		},
	}

	opt.SetFlags(cmd)

	return cmd
}

func joinCluster(opt *Option) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Cli init
	err := global.CreateGlobalEtcdInstance(opt.Parent)

	if err != nil {
		log.Fatal("Connect etcd failed", zap.String("err", err.Error()))
	}

	joinCmd := &joinCmd{
		certs: certs.New(global.Cli),
		infos: nodes.NewInfoStorage(global.Cli),
	}

	ca := joinCmd.getAndStoreCert(ctx)
	priv := joinCmd.getAndStoreCertKey(ctx, opt.Token)

	// Use the Cluster CA which initialized in init phase
	generateNodeCert(ca, priv, opt.Token)

	if err := joinCmd.registerNode(ctx); err != nil {
		log.Fatal("Register node failed", zap.String("err", err.Error()))
	}

}

type joinCmd struct {
	certs certs.Storage
	infos nodes.InfoStorage
}

func (j *joinCmd) getAndStoreCert(ctx context.Context) []byte {
	log.Info("Getting cluster ca from etcd")
	ca, err := j.certs.GetCA(ctx)
	if err != nil {
		log.Fatal("Get cluster cert failed", zap.String("err", err.Error()))
	}

	if err = icert.StoreLocalCaFile(ca); err != nil {
		log.Fatal("Store cluster cert failed", zap.String("err", err.Error()))
	}

	return ca
}

func (j *joinCmd) getAndStoreCertKey(ctx context.Context, passwd string) []byte {
	log.Info("Getting cluster private key from etcd")

	privEn, err := j.certs.GetCAPrivate(ctx)
	if err != nil {
		log.Fatal("Get cluster private key failed", zap.String("err", err.Error()))
	}

	log.Info("Decrypting cluster private key")
	priv, err := cert.DecryptPem(privEn, passwd)
	if err != nil {
		log.Fatal("Decrypt cluster private key failed, maybe token is incorrect.", zap.String("err", err.Error()))
	}

	log.Info("Storing cluster private key to local")
	if err := icert.StoreLocalCaKeyFile(priv); err != nil {
		log.Fatal("Store cluster private key failed", zap.String("err", err.Error()))
	}

	return priv
}

func generateNodeCert(ca, priv []byte, capass string) {
	log.Info("Generating node cert")

	name, err := node.Name()
	if err != nil {
		log.FatalE("Get node name failed", err)
	}

	nodeCert, nodeKey, err := cert.GenerateCert(ca, priv, capass, name, "", nil)
	if err != nil {
		log.Fatal("Generate node cert failed", zap.String("err", err.Error()))
	}

	if err = icert.StoreLocalCertFile(nodeCert); err != nil {
		log.Fatal("Store node cert failed", zap.String("err", err.Error()))
	}

	if err = icert.StoreLocalKeyFile(nodeKey); err != nil {
		log.Fatal("Store node key failed", zap.String("err", err.Error()))
	}
}

func (j *joinCmd) registerNode(ctx context.Context) error {
	log.Info("Registering current node")

	name, err := node.Name()
	if err != nil {
		return fmt.Errorf("failed to get hostname from current machine: %w", err)
	}

	info := &mnode.OrchInfo{
		CommonInfo: mnode.CommonInfo{
			Name: name,
		},
	}

	err = j.infos.Add(ctx, info)
	if err != nil {
		return err
	}

	return nil
}
