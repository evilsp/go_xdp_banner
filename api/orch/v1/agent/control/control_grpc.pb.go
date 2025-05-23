// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.2
// source: orch/v1/agent/control/control.proto

package control

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	ControlService_Register_FullMethodName         = "/agent.control.ControlService/Register"
	ControlService_Unregister_FullMethodName       = "/agent.control.ControlService/Unregister"
	ControlService_ListRegistration_FullMethodName = "/agent.control.ControlService/ListRegistration"
	ControlService_Init_FullMethodName             = "/agent.control.ControlService/Init"
	ControlService_Enable_FullMethodName           = "/agent.control.ControlService/Enable"
	ControlService_SetConfig_FullMethodName        = "/agent.control.ControlService/SetConfig"
	ControlService_GetConfig_FullMethodName        = "/agent.control.ControlService/GetConfig"
	ControlService_SetLabels_FullMethodName        = "/agent.control.ControlService/SetLabels"
	ControlService_GetLabels_FullMethodName        = "/agent.control.ControlService/GetLabels"
	ControlService_GetStatus_FullMethodName        = "/agent.control.ControlService/GetStatus"
	ControlService_GetInfo_FullMethodName          = "/agent.control.ControlService/GetInfo"
	ControlService_GetAgent_FullMethodName         = "/agent.control.ControlService/GetAgent"
	ControlService_ListAgents_FullMethodName       = "/agent.control.ControlService/ListAgents"
)

// ControlServiceClient is the client API for ControlService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Agent Control Service
type ControlServiceClient interface {
	// Register a new agent by the admin
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	// Unregister an agent by the admin
	Unregister(ctx context.Context, in *UnRegisterRequest, opts ...grpc.CallOption) (*UnRegisterResponse, error)
	// List all registered agents
	ListRegistration(ctx context.Context, in *ListRegistrationRequest, opts ...grpc.CallOption) (*ListRegistrationResponse, error)
	// Agent initializes itself by providing the registration token
	Init(ctx context.Context, in *InitRequest, opts ...grpc.CallOption) (*InitResponse, error)
	// Enable or disable an agent by the admin
	Enable(ctx context.Context, in *EnableRequest, opts ...grpc.CallOption) (*EnableResponse, error)
	// Set the config name to an agent
	SetConfig(ctx context.Context, in *SetConfigRequest, opts ...grpc.CallOption) (*SetConfigResponse, error)
	// Get the config name of an agent
	GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error)
	// Set the labels to an agent
	SetLabels(ctx context.Context, in *SetLabelsRequest, opts ...grpc.CallOption) (*SetLabelsResponse, error)
	// Get the labels of an agent
	GetLabels(ctx context.Context, in *GetLabelsRequest, opts ...grpc.CallOption) (*GetLabelsResponse, error)
	// Get the status of an agent
	GetStatus(ctx context.Context, in *GetStatusRequest, opts ...grpc.CallOption) (*GetStatusResponse, error)
	// Get the info of an agent
	GetInfo(ctx context.Context, in *GetInfoRequest, opts ...grpc.CallOption) (*GetInfoResponse, error)
	// Get the agent by name
	GetAgent(ctx context.Context, in *GetAgentRequest, opts ...grpc.CallOption) (*GetAgentResponse, error)
	// List all agents
	ListAgents(ctx context.Context, in *ListAgentsRequest, opts ...grpc.CallOption) (*ListAgentsResponse, error)
}

type controlServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewControlServiceClient(cc grpc.ClientConnInterface) ControlServiceClient {
	return &controlServiceClient{cc}
}

func (c *controlServiceClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, ControlService_Register_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) Unregister(ctx context.Context, in *UnRegisterRequest, opts ...grpc.CallOption) (*UnRegisterResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UnRegisterResponse)
	err := c.cc.Invoke(ctx, ControlService_Unregister_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) ListRegistration(ctx context.Context, in *ListRegistrationRequest, opts ...grpc.CallOption) (*ListRegistrationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListRegistrationResponse)
	err := c.cc.Invoke(ctx, ControlService_ListRegistration_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) Init(ctx context.Context, in *InitRequest, opts ...grpc.CallOption) (*InitResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(InitResponse)
	err := c.cc.Invoke(ctx, ControlService_Init_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) Enable(ctx context.Context, in *EnableRequest, opts ...grpc.CallOption) (*EnableResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EnableResponse)
	err := c.cc.Invoke(ctx, ControlService_Enable_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) SetConfig(ctx context.Context, in *SetConfigRequest, opts ...grpc.CallOption) (*SetConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SetConfigResponse)
	err := c.cc.Invoke(ctx, ControlService_SetConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetConfigResponse)
	err := c.cc.Invoke(ctx, ControlService_GetConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) SetLabels(ctx context.Context, in *SetLabelsRequest, opts ...grpc.CallOption) (*SetLabelsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SetLabelsResponse)
	err := c.cc.Invoke(ctx, ControlService_SetLabels_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) GetLabels(ctx context.Context, in *GetLabelsRequest, opts ...grpc.CallOption) (*GetLabelsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetLabelsResponse)
	err := c.cc.Invoke(ctx, ControlService_GetLabels_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) GetStatus(ctx context.Context, in *GetStatusRequest, opts ...grpc.CallOption) (*GetStatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetStatusResponse)
	err := c.cc.Invoke(ctx, ControlService_GetStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) GetInfo(ctx context.Context, in *GetInfoRequest, opts ...grpc.CallOption) (*GetInfoResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetInfoResponse)
	err := c.cc.Invoke(ctx, ControlService_GetInfo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) GetAgent(ctx context.Context, in *GetAgentRequest, opts ...grpc.CallOption) (*GetAgentResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetAgentResponse)
	err := c.cc.Invoke(ctx, ControlService_GetAgent_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controlServiceClient) ListAgents(ctx context.Context, in *ListAgentsRequest, opts ...grpc.CallOption) (*ListAgentsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListAgentsResponse)
	err := c.cc.Invoke(ctx, ControlService_ListAgents_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ControlServiceServer is the server API for ControlService service.
// All implementations must embed UnimplementedControlServiceServer
// for forward compatibility.
//
// Agent Control Service
type ControlServiceServer interface {
	// Register a new agent by the admin
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	// Unregister an agent by the admin
	Unregister(context.Context, *UnRegisterRequest) (*UnRegisterResponse, error)
	// List all registered agents
	ListRegistration(context.Context, *ListRegistrationRequest) (*ListRegistrationResponse, error)
	// Agent initializes itself by providing the registration token
	Init(context.Context, *InitRequest) (*InitResponse, error)
	// Enable or disable an agent by the admin
	Enable(context.Context, *EnableRequest) (*EnableResponse, error)
	// Set the config name to an agent
	SetConfig(context.Context, *SetConfigRequest) (*SetConfigResponse, error)
	// Get the config name of an agent
	GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error)
	// Set the labels to an agent
	SetLabels(context.Context, *SetLabelsRequest) (*SetLabelsResponse, error)
	// Get the labels of an agent
	GetLabels(context.Context, *GetLabelsRequest) (*GetLabelsResponse, error)
	// Get the status of an agent
	GetStatus(context.Context, *GetStatusRequest) (*GetStatusResponse, error)
	// Get the info of an agent
	GetInfo(context.Context, *GetInfoRequest) (*GetInfoResponse, error)
	// Get the agent by name
	GetAgent(context.Context, *GetAgentRequest) (*GetAgentResponse, error)
	// List all agents
	ListAgents(context.Context, *ListAgentsRequest) (*ListAgentsResponse, error)
	mustEmbedUnimplementedControlServiceServer()
}

// UnimplementedControlServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedControlServiceServer struct{}

func (UnimplementedControlServiceServer) Register(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedControlServiceServer) Unregister(context.Context, *UnRegisterRequest) (*UnRegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Unregister not implemented")
}
func (UnimplementedControlServiceServer) ListRegistration(context.Context, *ListRegistrationRequest) (*ListRegistrationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListRegistration not implemented")
}
func (UnimplementedControlServiceServer) Init(context.Context, *InitRequest) (*InitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Init not implemented")
}
func (UnimplementedControlServiceServer) Enable(context.Context, *EnableRequest) (*EnableResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Enable not implemented")
}
func (UnimplementedControlServiceServer) SetConfig(context.Context, *SetConfigRequest) (*SetConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetConfig not implemented")
}
func (UnimplementedControlServiceServer) GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetConfig not implemented")
}
func (UnimplementedControlServiceServer) SetLabels(context.Context, *SetLabelsRequest) (*SetLabelsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetLabels not implemented")
}
func (UnimplementedControlServiceServer) GetLabels(context.Context, *GetLabelsRequest) (*GetLabelsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLabels not implemented")
}
func (UnimplementedControlServiceServer) GetStatus(context.Context, *GetStatusRequest) (*GetStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStatus not implemented")
}
func (UnimplementedControlServiceServer) GetInfo(context.Context, *GetInfoRequest) (*GetInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInfo not implemented")
}
func (UnimplementedControlServiceServer) GetAgent(context.Context, *GetAgentRequest) (*GetAgentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAgent not implemented")
}
func (UnimplementedControlServiceServer) ListAgents(context.Context, *ListAgentsRequest) (*ListAgentsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAgents not implemented")
}
func (UnimplementedControlServiceServer) mustEmbedUnimplementedControlServiceServer() {}
func (UnimplementedControlServiceServer) testEmbeddedByValue()                        {}

// UnsafeControlServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ControlServiceServer will
// result in compilation errors.
type UnsafeControlServiceServer interface {
	mustEmbedUnimplementedControlServiceServer()
}

func RegisterControlServiceServer(s grpc.ServiceRegistrar, srv ControlServiceServer) {
	// If the following call pancis, it indicates UnimplementedControlServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ControlService_ServiceDesc, srv)
}

func _ControlService_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControlService_Register_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_Unregister_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UnRegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).Unregister(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControlService_Unregister_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).Unregister(ctx, req.(*UnRegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_ListRegistration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRegistrationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).ListRegistration(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControlService_ListRegistration_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).ListRegistration(ctx, req.(*ListRegistrationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_Init_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).Init(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControlService_Init_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).Init(ctx, req.(*InitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_Enable_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EnableRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).Enable(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControlService_Enable_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).Enable(ctx, req.(*EnableRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_SetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).SetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControlService_SetConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).SetConfig(ctx, req.(*SetConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControlService_GetConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).GetConfig(ctx, req.(*GetConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_SetLabels_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetLabelsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).SetLabels(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControlService_SetLabels_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).SetLabels(ctx, req.(*SetLabelsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_GetLabels_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLabelsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).GetLabels(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControlService_GetLabels_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).GetLabels(ctx, req.(*GetLabelsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_GetStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).GetStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControlService_GetStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).GetStatus(ctx, req.(*GetStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_GetInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).GetInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControlService_GetInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).GetInfo(ctx, req.(*GetInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_GetAgent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAgentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).GetAgent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControlService_GetAgent_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).GetAgent(ctx, req.(*GetAgentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControlService_ListAgents_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAgentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControlServiceServer).ListAgents(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControlService_ListAgents_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControlServiceServer).ListAgents(ctx, req.(*ListAgentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ControlService_ServiceDesc is the grpc.ServiceDesc for ControlService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ControlService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "agent.control.ControlService",
	HandlerType: (*ControlServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _ControlService_Register_Handler,
		},
		{
			MethodName: "Unregister",
			Handler:    _ControlService_Unregister_Handler,
		},
		{
			MethodName: "ListRegistration",
			Handler:    _ControlService_ListRegistration_Handler,
		},
		{
			MethodName: "Init",
			Handler:    _ControlService_Init_Handler,
		},
		{
			MethodName: "Enable",
			Handler:    _ControlService_Enable_Handler,
		},
		{
			MethodName: "SetConfig",
			Handler:    _ControlService_SetConfig_Handler,
		},
		{
			MethodName: "GetConfig",
			Handler:    _ControlService_GetConfig_Handler,
		},
		{
			MethodName: "SetLabels",
			Handler:    _ControlService_SetLabels_Handler,
		},
		{
			MethodName: "GetLabels",
			Handler:    _ControlService_GetLabels_Handler,
		},
		{
			MethodName: "GetStatus",
			Handler:    _ControlService_GetStatus_Handler,
		},
		{
			MethodName: "GetInfo",
			Handler:    _ControlService_GetInfo_Handler,
		},
		{
			MethodName: "GetAgent",
			Handler:    _ControlService_GetAgent_Handler,
		},
		{
			MethodName: "ListAgents",
			Handler:    _ControlService_ListAgents_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "orch/v1/agent/control/control.proto",
}
