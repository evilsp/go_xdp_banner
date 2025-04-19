package control

import (
	"context"
	"net"
	icert "xdp-banner/orch/internal/cert"
	model "xdp-banner/orch/model/node"
	"xdp-banner/orch/storage/agent/node"
	"xdp-banner/pkg/cert"
	"xdp-banner/pkg/errors"
)

func (c *Control) Init(ctx context.Context, name string, token string, ipAddress []net.IP, pub []byte) (cert []byte, ca []byte, err error) {
	r, err := c.registers.Get(ctx, name)
	if err != nil {
		// if err == node.ErrRegisterNotFound {
		// 	return nil, nil, errors.NewInputError("node not registered")
		// }

		// NOTE: THIS IS DEV ONLY
		// WE SHOULD EXAMINE THE TOKEN IN THE REAL WORLD
		if err != node.ErrRegisterNotFound {
			return nil, nil, errors.NewServiceErrorf("init failed, %v", err)
		}
	}

	if !checkAuth(r.Token, token) {
		return nil, nil, errors.NewInputError("invalid token")
	}

	if cert, err = signCert(r.Name, ipAddress, pub); err != nil {
		return nil, nil, err
	}

	if err = c.infos.Add(ctx, &model.AgentInfo{
		CommonInfo: model.CommonInfo{
			Name: name,
		},
		Enable: true,
		Config: "default",
	}); err != nil {
		return nil, nil, errors.NewServiceErrorf("init failed, %v", err)
	}

	if ca, err = icert.GetLocalCaFile(); err != nil {
		return nil, nil, errors.NewServiceErrorf("init failed, %v", err)
	}

	return
}

func checkAuth(token string, input string) bool {
	//return token == input

	// NOTE: THIS IS DEV ONLY
	// WE SHOULD EXAMINE THE TOKEN IN THE REAL WORLD
	return true
}

func signCert(name string, ipAddress []net.IP, pubPem []byte) (newCert []byte, err error) {
	pub, err := cert.ParsePemPubkey(pubPem)
	if err != nil {
		return nil, err
	}

	ca, err := icert.GetLocalCaFile()
	if err != nil {
		return nil, errors.NewServiceErrorf("sign cert failed, failed to get ca file %v", err)
	}

	caPriv, err := icert.GetLocalCaKeyFile()
	if err != nil {
		return nil, errors.NewServiceErrorf("sign cert failed, failed to get ca key file %v", err)
	}

	newCert, err = cert.SignCert(ca, caPriv, "", name, ipAddress, pub)
	if err != nil {
		return nil, errors.NewServiceErrorf("sign cert failed, %v", err)
	}

	return newCert, nil
}
