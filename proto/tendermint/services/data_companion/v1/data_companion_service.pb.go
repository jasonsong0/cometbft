// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: tendermint/services/data_companion/v1/data_companion_service.proto

package v1

import (
	context "context"
	fmt "fmt"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

func init() {
	proto.RegisterFile("tendermint/services/data_companion/v1/data_companion_service.proto", fileDescriptor_f80d6e527d0b2bf0)
}

var fileDescriptor_f80d6e527d0b2bf0 = []byte{
	// 229 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x72, 0x2a, 0x49, 0xcd, 0x4b,
	0x49, 0x2d, 0xca, 0xcd, 0xcc, 0x2b, 0xd1, 0x2f, 0x4e, 0x2d, 0x2a, 0xcb, 0x4c, 0x4e, 0x2d, 0xd6,
	0x4f, 0x49, 0x2c, 0x49, 0x8c, 0x4f, 0xce, 0xcf, 0x2d, 0x48, 0xcc, 0xcb, 0xcc, 0xcf, 0xd3, 0x2f,
	0x33, 0x44, 0x13, 0x89, 0x87, 0xaa, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x52, 0x45, 0x98,
	0xa1, 0x07, 0x33, 0x43, 0x0f, 0x55, 0x87, 0x5e, 0x99, 0xa1, 0x94, 0x15, 0x39, 0x56, 0x41, 0xac,
	0x30, 0xda, 0xc7, 0xc4, 0x25, 0xe2, 0x92, 0x58, 0x92, 0xe8, 0x0c, 0x13, 0x0f, 0x86, 0x18, 0x20,
	0x34, 0x81, 0x91, 0x8b, 0xdf, 0x3d, 0xb5, 0x24, 0x28, 0xb5, 0x24, 0x31, 0x33, 0xcf, 0x23, 0x35,
	0x33, 0x3d, 0xa3, 0x44, 0xc8, 0x56, 0x8f, 0x28, 0x07, 0xe9, 0xa1, 0xe9, 0x0b, 0x4a, 0x2d, 0x2c,
	0x4d, 0x2d, 0x2e, 0x91, 0xb2, 0x23, 0x57, 0x7b, 0x71, 0x41, 0x7e, 0x5e, 0x71, 0xaa, 0xd0, 0x24,
	0x46, 0x2e, 0xfe, 0x60, 0x32, 0x9d, 0x14, 0x4c, 0x99, 0x93, 0x82, 0xb1, 0x3b, 0x49, 0x89, 0xc1,
	0x29, 0xed, 0xc4, 0x23, 0x39, 0xc6, 0x0b, 0x8f, 0xe4, 0x18, 0x1f, 0x3c, 0x92, 0x63, 0x9c, 0xf0,
	0x58, 0x8e, 0xe1, 0xc2, 0x63, 0x39, 0x86, 0x1b, 0x8f, 0xe5, 0x18, 0xa2, 0x7c, 0xd2, 0x33, 0x4b,
	0x32, 0x4a, 0x93, 0xf4, 0x92, 0xf3, 0x73, 0xf5, 0x91, 0x62, 0x08, 0x89, 0x09, 0x8e, 0x02, 0x7d,
	0xa2, 0x62, 0x2f, 0x89, 0x0d, 0xac, 0xd8, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0x6e, 0x7c, 0xee,
	0xd0, 0x58, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// DataCompanionServiceClient is the client API for DataCompanionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type DataCompanionServiceClient interface {
	// GetRetainHeight returns the retain height set by the companion and that
	// set by the application. This can give the companion an indication as to
	// which heights' data are currently available.
	GetRetainHeight(ctx context.Context, in *GetRetainHeightRequest, opts ...grpc.CallOption) (*GetRetainHeightResponse, error)
	// SetRetainHeight notifies the node of the minimum height whose data must
	// be retained by the node. This data includes block data and block
	// execution results.
	//
	// Setting a retain height lower than a previous setting will result in an
	// error.
	//
	// The lower of this retain height and that set by the application in its
	// Commit response will be used by the node to determine which heights' data
	// can be pruned.
	SetRetainHeight(ctx context.Context, in *SetRetainHeightRequest, opts ...grpc.CallOption) (*SetRetainHeightResponse, error)
}

type dataCompanionServiceClient struct {
	cc grpc1.ClientConn
}

func NewDataCompanionServiceClient(cc grpc1.ClientConn) DataCompanionServiceClient {
	return &dataCompanionServiceClient{cc}
}

func (c *dataCompanionServiceClient) GetRetainHeight(ctx context.Context, in *GetRetainHeightRequest, opts ...grpc.CallOption) (*GetRetainHeightResponse, error) {
	out := new(GetRetainHeightResponse)
	err := c.cc.Invoke(ctx, "/tendermint.services.data_companion.v1.DataCompanionService/GetRetainHeight", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dataCompanionServiceClient) SetRetainHeight(ctx context.Context, in *SetRetainHeightRequest, opts ...grpc.CallOption) (*SetRetainHeightResponse, error) {
	out := new(SetRetainHeightResponse)
	err := c.cc.Invoke(ctx, "/tendermint.services.data_companion.v1.DataCompanionService/SetRetainHeight", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DataCompanionServiceServer is the server API for DataCompanionService service.
type DataCompanionServiceServer interface {
	// GetRetainHeight returns the retain height set by the companion and that
	// set by the application. This can give the companion an indication as to
	// which heights' data are currently available.
	GetRetainHeight(context.Context, *GetRetainHeightRequest) (*GetRetainHeightResponse, error)
	// SetRetainHeight notifies the node of the minimum height whose data must
	// be retained by the node. This data includes block data and block
	// execution results.
	//
	// Setting a retain height lower than a previous setting will result in an
	// error.
	//
	// The lower of this retain height and that set by the application in its
	// Commit response will be used by the node to determine which heights' data
	// can be pruned.
	SetRetainHeight(context.Context, *SetRetainHeightRequest) (*SetRetainHeightResponse, error)
}

// UnimplementedDataCompanionServiceServer can be embedded to have forward compatible implementations.
type UnimplementedDataCompanionServiceServer struct {
}

func (*UnimplementedDataCompanionServiceServer) GetRetainHeight(ctx context.Context, req *GetRetainHeightRequest) (*GetRetainHeightResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRetainHeight not implemented")
}
func (*UnimplementedDataCompanionServiceServer) SetRetainHeight(ctx context.Context, req *SetRetainHeightRequest) (*SetRetainHeightResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetRetainHeight not implemented")
}

func RegisterDataCompanionServiceServer(s grpc1.Server, srv DataCompanionServiceServer) {
	s.RegisterService(&_DataCompanionService_serviceDesc, srv)
}

func _DataCompanionService_GetRetainHeight_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRetainHeightRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataCompanionServiceServer).GetRetainHeight(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tendermint.services.data_companion.v1.DataCompanionService/GetRetainHeight",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataCompanionServiceServer).GetRetainHeight(ctx, req.(*GetRetainHeightRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DataCompanionService_SetRetainHeight_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetRetainHeightRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataCompanionServiceServer).SetRetainHeight(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tendermint.services.data_companion.v1.DataCompanionService/SetRetainHeight",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataCompanionServiceServer).SetRetainHeight(ctx, req.(*SetRetainHeightRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _DataCompanionService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "tendermint.services.data_companion.v1.DataCompanionService",
	HandlerType: (*DataCompanionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetRetainHeight",
			Handler:    _DataCompanionService_GetRetainHeight_Handler,
		},
		{
			MethodName: "SetRetainHeight",
			Handler:    _DataCompanionService_SetRetainHeight_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tendermint/services/data_companion/v1/data_companion_service.proto",
}