package join

import (
	"crypto/tls"
	"xdp-banner/agent/cmd/global"
	"xdp-banner/agent/internal/icert"
	"xdp-banner/api/orch/v1/agent/control"
	"xdp-banner/pkg/cert"
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/node"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func Cmd(parentOpt *global.Option) *cobra.Command {
	opt := DefaultOption(parentOpt)

	cmd := &cobra.Command{
		Use:   "join",
		Short: "join to the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opt.Check(); err != nil {
				return err
			}

			join(opt)
			return nil
		},
	}

	opt.SetFlags(cmd)

	return cmd
}

func join(opt *Option) {
	ctx, cancel := context.WithTimeout(context.Background(), opt.Timeout)
	defer cancel()

	conn := newGrpcServer(ctx, opt.Parent.Orch.Endpoints)
	defer conn.Close()

	pubPem, priPem, err := genKeyPairPem()
	if err != nil {
		log.FatalE("gen key pair failed", err)
	}

	name, err := node.Name()

	if err != nil {
		log.FatalE("get node name failed", err)
	}

	caPem, certPem, err := joinToCluster(ctx, conn, name, opt.Token, pubPem)
	if err != nil {
		log.FatalE("join to cluster failed", err)
	}

	if storeCert(caPem, certPem, priPem) != nil {
		log.FatalE("store cert failed", err)
	}
}

func newGrpcServer(_ context.Context, endpoints string) *grpc.ClientConn {
	cred := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})

	conn, err := grpc.NewClient(
		endpoints,
		grpc.WithTransportCredentials(cred),
	)
	if err != nil {
		log.FatalE("create grpc client failed", err)
	}

	return conn
}

func genKeyPairPem() (pub []byte, priv []byte, err error) {
	log.Info("generating key pair")
	pri, err := cert.GeneratePrivateKey()
	if err != nil {
		return
	}

	if pub, err = cert.PubKeyToPem(&pri.PublicKey); err != nil {
		return
	}

	if priv, err = cert.PrivToPem(pri, ""); err != nil {
		return
	}

	return
}

func joinToCluster(ctx context.Context, conn *grpc.ClientConn, name string, token string, pubPem []byte) (caPem, certPem []byte, err error) {
	log.Info("joining to the cluster")

	ip, err := node.DefaultIP()
	if err != nil {
		return nil, nil, err
	}

	cli := control.NewControlServiceClient(conn)
	resp, err := cli.Init(ctx, &control.InitRequest{
		Name:      name,
		Token:     token,
		PubKeyPem: pubPem,
		IpAddresses: []string{
			ip.String(),
		},
	})
	if err != nil {
		log.FatalE("init failed", err)
	}

	return resp.Ca, resp.Cert, nil
}

func storeCert(ca, cert, key []byte) (err error) {
	log.Info("storing cert")

	if err = icert.StoreCA(ca); err != nil {
		return
	}
	if err = icert.StoreCert(cert); err != nil {
		return
	}
	if err = icert.StoreCertPri(key); err != nil {
		return
	}

	return
}
