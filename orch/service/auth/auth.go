package auth

import (
	"context"
	"time"
	"xdp-banner/api/orch/v1/auth"
	"xdp-banner/pkg/server/common"

	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type AuthService struct {
	auth.UnimplementedAuthServiceServer
}

func (s *AuthService) RegisterGrpcService(gs grpc.ServiceRegistrar) {
	auth.RegisterAuthServiceServer(gs, s)
}

func (s *AuthService) PublicGrpcMethods() []string {
	return nil
}

func (s *AuthService) RegisterHttpService(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return auth.RegisterAuthServiceHandler(ctx, mux, conn)
}

func New() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Login(ctx context.Context, r *auth.LoginRequest) (*auth.LoginResponse, error) {
	switch {
	case r.Username == "":
		return nil, common.InvalidArgumentError("账号是必须的")
	case r.Password == "":
		return nil, common.InvalidArgumentError("密码是必须的")
	}

	if r.Username != "i@joshua.su" && r.Password != "i@joshua.su1" {
		return nil, common.InvalidArgumentError("账号或密码错误")
	}

	token, err := GenToken(r.Username)

	if err != nil {
		return nil, common.InvalidArgumentError("生成token失败")
	}

	return &auth.LoginResponse{
		AccessToken: token,
	}, nil
}

func GenToken(u string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": u,
		"nbf":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Minute * 300).Unix(),
	})

	return token.SignedString([]byte("fasthub-center"))
}

func (s *AuthService) Me(ctx context.Context, _ *auth.MeRequest) (*auth.MeResponse, error) {
	return &auth.MeResponse{
		Id:          "8864c717-587d-472a-929a-8e5f298024da-0",
		DisplayName: "Joshua Su",
		PhotoURL:    "http://localhost:7272/assets/images/avatar/avatar-25.webp",
		PhoneNumber: "+86 10086",
		Country:     "China",
		Address:     "cqupt",
		State:       "Chongqing",
		City:        "Chongqing",
		ZipCode:     "400000",
		About:       "this is about",
		Role:        "admin",
		IsPublic:    true,
		Email:       "i@joshua.su",
		Password:    "i@joshua.su1",
	}, nil
}
