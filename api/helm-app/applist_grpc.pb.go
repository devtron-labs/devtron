// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package client

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ApplicationServiceClient is the client API for ApplicationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ApplicationServiceClient interface {
	ListApplications(ctx context.Context, in *AppListRequest, opts ...grpc.CallOption) (ApplicationService_ListApplicationsClient, error)
	GetAppDetail(ctx context.Context, in *AppDetailRequest, opts ...grpc.CallOption) (*AppDetail, error)
	Hibernate(ctx context.Context, in *HibernateRequest, opts ...grpc.CallOption) (*HibernateResponse, error)
	UnHibernate(ctx context.Context, in *HibernateRequest, opts ...grpc.CallOption) (*HibernateResponse, error)
	GetDeploymentHistory(ctx context.Context, in *AppDetailRequest, opts ...grpc.CallOption) (*HelmAppDeploymentHistory, error)
	GetValuesYaml(ctx context.Context, in *AppDetailRequest, opts ...grpc.CallOption) (*ReleaseInfo, error)
}

type applicationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewApplicationServiceClient(cc grpc.ClientConnInterface) ApplicationServiceClient {
	return &applicationServiceClient{cc}
}

func (c *applicationServiceClient) ListApplications(ctx context.Context, in *AppListRequest, opts ...grpc.CallOption) (ApplicationService_ListApplicationsClient, error) {
	stream, err := c.cc.NewStream(ctx, &ApplicationService_ServiceDesc.Streams[0], "/ApplicationService/ListApplications", opts...)
	if err != nil {
		return nil, err
	}
	x := &applicationServiceListApplicationsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ApplicationService_ListApplicationsClient interface {
	Recv() (*DeployedAppList, error)
	grpc.ClientStream
}

type applicationServiceListApplicationsClient struct {
	grpc.ClientStream
}

func (x *applicationServiceListApplicationsClient) Recv() (*DeployedAppList, error) {
	m := new(DeployedAppList)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *applicationServiceClient) GetAppDetail(ctx context.Context, in *AppDetailRequest, opts ...grpc.CallOption) (*AppDetail, error) {
	out := new(AppDetail)
	err := c.cc.Invoke(ctx, "/ApplicationService/GetAppDetail", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) Hibernate(ctx context.Context, in *HibernateRequest, opts ...grpc.CallOption) (*HibernateResponse, error) {
	out := new(HibernateResponse)
	err := c.cc.Invoke(ctx, "/ApplicationService/Hibernate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) UnHibernate(ctx context.Context, in *HibernateRequest, opts ...grpc.CallOption) (*HibernateResponse, error) {
	out := new(HibernateResponse)
	err := c.cc.Invoke(ctx, "/ApplicationService/UnHibernate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) GetDeploymentHistory(ctx context.Context, in *AppDetailRequest, opts ...grpc.CallOption) (*HelmAppDeploymentHistory, error) {
	out := new(HelmAppDeploymentHistory)
	err := c.cc.Invoke(ctx, "/ApplicationService/GetDeploymentHistory", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) GetValuesYaml(ctx context.Context, in *AppDetailRequest, opts ...grpc.CallOption) (*ReleaseInfo, error) {
	out := new(ReleaseInfo)
	err := c.cc.Invoke(ctx, "/ApplicationService/GetValuesYaml", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ApplicationServiceServer is the server API for ApplicationService service.
// All implementations must embed UnimplementedApplicationServiceServer
// for forward compatibility
type ApplicationServiceServer interface {
	ListApplications(*AppListRequest, ApplicationService_ListApplicationsServer) error
	GetAppDetail(context.Context, *AppDetailRequest) (*AppDetail, error)
	Hibernate(context.Context, *HibernateRequest) (*HibernateResponse, error)
	UnHibernate(context.Context, *HibernateRequest) (*HibernateResponse, error)
	GetDeploymentHistory(context.Context, *AppDetailRequest) (*HelmAppDeploymentHistory, error)
	GetValuesYaml(context.Context, *AppDetailRequest) (*ReleaseInfo, error)
	mustEmbedUnimplementedApplicationServiceServer()
}

// UnimplementedApplicationServiceServer must be embedded to have forward compatible implementations.
type UnimplementedApplicationServiceServer struct {
}

func (UnimplementedApplicationServiceServer) ListApplications(*AppListRequest, ApplicationService_ListApplicationsServer) error {
	return status.Errorf(codes.Unimplemented, "method ListApplications not implemented")
}
func (UnimplementedApplicationServiceServer) GetAppDetail(context.Context, *AppDetailRequest) (*AppDetail, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAppDetail not implemented")
}
func (UnimplementedApplicationServiceServer) Hibernate(context.Context, *HibernateRequest) (*HibernateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Hibernate not implemented")
}
func (UnimplementedApplicationServiceServer) UnHibernate(context.Context, *HibernateRequest) (*HibernateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnHibernate not implemented")
}
func (UnimplementedApplicationServiceServer) GetDeploymentHistory(context.Context, *AppDetailRequest) (*HelmAppDeploymentHistory, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDeploymentHistory not implemented")
}
func (UnimplementedApplicationServiceServer) GetValuesYaml(context.Context, *AppDetailRequest) (*ReleaseInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetValuesYaml not implemented")
}
func (UnimplementedApplicationServiceServer) mustEmbedUnimplementedApplicationServiceServer() {}

// UnsafeApplicationServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ApplicationServiceServer will
// result in compilation errors.
type UnsafeApplicationServiceServer interface {
	mustEmbedUnimplementedApplicationServiceServer()
}

func RegisterApplicationServiceServer(s grpc.ServiceRegistrar, srv ApplicationServiceServer) {
	s.RegisterService(&ApplicationService_ServiceDesc, srv)
}

func _ApplicationService_ListApplications_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(AppListRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ApplicationServiceServer).ListApplications(m, &applicationServiceListApplicationsServer{stream})
}

type ApplicationService_ListApplicationsServer interface {
	Send(*DeployedAppList) error
	grpc.ServerStream
}

type applicationServiceListApplicationsServer struct {
	grpc.ServerStream
}

func (x *applicationServiceListApplicationsServer) Send(m *DeployedAppList) error {
	return x.ServerStream.SendMsg(m)
}

func _ApplicationService_GetAppDetail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppDetailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).GetAppDetail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ApplicationService/GetAppDetail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).GetAppDetail(ctx, req.(*AppDetailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_Hibernate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HibernateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).Hibernate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ApplicationService/Hibernate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).Hibernate(ctx, req.(*HibernateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_UnHibernate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HibernateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).UnHibernate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ApplicationService/UnHibernate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).UnHibernate(ctx, req.(*HibernateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_GetDeploymentHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppDetailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).GetDeploymentHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ApplicationService/GetDeploymentHistory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).GetDeploymentHistory(ctx, req.(*AppDetailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_GetValuesYaml_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppDetailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).GetValuesYaml(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ApplicationService/GetValuesYaml",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).GetValuesYaml(ctx, req.(*AppDetailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ApplicationService_ServiceDesc is the grpc.ServiceDesc for ApplicationService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ApplicationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ApplicationService",
	HandlerType: (*ApplicationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAppDetail",
			Handler:    _ApplicationService_GetAppDetail_Handler,
		},
		{
			MethodName: "Hibernate",
			Handler:    _ApplicationService_Hibernate_Handler,
		},
		{
			MethodName: "UnHibernate",
			Handler:    _ApplicationService_UnHibernate_Handler,
		},
		{
			MethodName: "GetDeploymentHistory",
			Handler:    _ApplicationService_GetDeploymentHistory_Handler,
		},
		{
			MethodName: "GetValuesYaml",
			Handler:    _ApplicationService_GetValuesYaml_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ListApplications",
			Handler:       _ApplicationService_ListApplications_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "grpc/applist.proto",
}
