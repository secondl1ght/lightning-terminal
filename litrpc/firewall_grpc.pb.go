// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package litrpc

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

// FirewallClient is the client API for Firewall service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FirewallClient interface {
	ListActions(ctx context.Context, in *ListActionsRequest, opts ...grpc.CallOption) (*ListActionsResponse, error)
}

type firewallClient struct {
	cc grpc.ClientConnInterface
}

func NewFirewallClient(cc grpc.ClientConnInterface) FirewallClient {
	return &firewallClient{cc}
}

func (c *firewallClient) ListActions(ctx context.Context, in *ListActionsRequest, opts ...grpc.CallOption) (*ListActionsResponse, error) {
	out := new(ListActionsResponse)
	err := c.cc.Invoke(ctx, "/litrpc.Firewall/ListActions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FirewallServer is the server API for Firewall service.
// All implementations must embed UnimplementedFirewallServer
// for forward compatibility
type FirewallServer interface {
	ListActions(context.Context, *ListActionsRequest) (*ListActionsResponse, error)
	mustEmbedUnimplementedFirewallServer()
}

// UnimplementedFirewallServer must be embedded to have forward compatible implementations.
type UnimplementedFirewallServer struct {
}

func (UnimplementedFirewallServer) ListActions(context.Context, *ListActionsRequest) (*ListActionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListActions not implemented")
}
func (UnimplementedFirewallServer) mustEmbedUnimplementedFirewallServer() {}

// UnsafeFirewallServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FirewallServer will
// result in compilation errors.
type UnsafeFirewallServer interface {
	mustEmbedUnimplementedFirewallServer()
}

func RegisterFirewallServer(s grpc.ServiceRegistrar, srv FirewallServer) {
	s.RegisterService(&Firewall_ServiceDesc, srv)
}

func _Firewall_ListActions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListActionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FirewallServer).ListActions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/litrpc.Firewall/ListActions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FirewallServer).ListActions(ctx, req.(*ListActionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Firewall_ServiceDesc is the grpc.ServiceDesc for Firewall service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Firewall_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "litrpc.Firewall",
	HandlerType: (*FirewallServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListActions",
			Handler:    _Firewall_ListActions_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "firewall.proto",
}
