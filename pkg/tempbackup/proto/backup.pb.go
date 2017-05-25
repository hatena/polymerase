// Code generated by protoc-gen-go. DO NOT EDIT.
// source: tempbackup/proto/backup.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	tempbackup/proto/backup.proto

It has these top-level messages:
	FullBackupContentStream
	IncBackupContentStream
	BackupReply
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type FullBackupContentStream struct {
	Db      string `protobuf:"bytes,1,opt,name=db" json:"db,omitempty"`
	Content []byte `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
}

func (m *FullBackupContentStream) Reset()                    { *m = FullBackupContentStream{} }
func (m *FullBackupContentStream) String() string            { return proto1.CompactTextString(m) }
func (*FullBackupContentStream) ProtoMessage()               {}
func (*FullBackupContentStream) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *FullBackupContentStream) GetDb() string {
	if m != nil {
		return m.Db
	}
	return ""
}

func (m *FullBackupContentStream) GetContent() []byte {
	if m != nil {
		return m.Content
	}
	return nil
}

type IncBackupContentStream struct {
	Db      string `protobuf:"bytes,1,opt,name=db" json:"db,omitempty"`
	Lsn     string `protobuf:"bytes,2,opt,name=lsn" json:"lsn,omitempty"`
	Content []byte `protobuf:"bytes,3,opt,name=content,proto3" json:"content,omitempty"`
}

func (m *IncBackupContentStream) Reset()                    { *m = IncBackupContentStream{} }
func (m *IncBackupContentStream) String() string            { return proto1.CompactTextString(m) }
func (*IncBackupContentStream) ProtoMessage()               {}
func (*IncBackupContentStream) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *IncBackupContentStream) GetDb() string {
	if m != nil {
		return m.Db
	}
	return ""
}

func (m *IncBackupContentStream) GetLsn() string {
	if m != nil {
		return m.Lsn
	}
	return ""
}

func (m *IncBackupContentStream) GetContent() []byte {
	if m != nil {
		return m.Content
	}
	return nil
}

type BackupReply struct {
	Message string `protobuf:"bytes,1,opt,name=message" json:"message,omitempty"`
}

func (m *BackupReply) Reset()                    { *m = BackupReply{} }
func (m *BackupReply) String() string            { return proto1.CompactTextString(m) }
func (*BackupReply) ProtoMessage()               {}
func (*BackupReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *BackupReply) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto1.RegisterType((*FullBackupContentStream)(nil), "proto.FullBackupContentStream")
	proto1.RegisterType((*IncBackupContentStream)(nil), "proto.IncBackupContentStream")
	proto1.RegisterType((*BackupReply)(nil), "proto.BackupReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for BackupTransferService service

type BackupTransferServiceClient interface {
	TransferFullBackup(ctx context.Context, opts ...grpc.CallOption) (BackupTransferService_TransferFullBackupClient, error)
	TransferIncBackup(ctx context.Context, opts ...grpc.CallOption) (BackupTransferService_TransferIncBackupClient, error)
}

type backupTransferServiceClient struct {
	cc *grpc.ClientConn
}

func NewBackupTransferServiceClient(cc *grpc.ClientConn) BackupTransferServiceClient {
	return &backupTransferServiceClient{cc}
}

func (c *backupTransferServiceClient) TransferFullBackup(ctx context.Context, opts ...grpc.CallOption) (BackupTransferService_TransferFullBackupClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_BackupTransferService_serviceDesc.Streams[0], c.cc, "/proto.BackupTransferService/TransferFullBackup", opts...)
	if err != nil {
		return nil, err
	}
	x := &backupTransferServiceTransferFullBackupClient{stream}
	return x, nil
}

type BackupTransferService_TransferFullBackupClient interface {
	Send(*FullBackupContentStream) error
	CloseAndRecv() (*BackupReply, error)
	grpc.ClientStream
}

type backupTransferServiceTransferFullBackupClient struct {
	grpc.ClientStream
}

func (x *backupTransferServiceTransferFullBackupClient) Send(m *FullBackupContentStream) error {
	return x.ClientStream.SendMsg(m)
}

func (x *backupTransferServiceTransferFullBackupClient) CloseAndRecv() (*BackupReply, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(BackupReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *backupTransferServiceClient) TransferIncBackup(ctx context.Context, opts ...grpc.CallOption) (BackupTransferService_TransferIncBackupClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_BackupTransferService_serviceDesc.Streams[1], c.cc, "/proto.BackupTransferService/TransferIncBackup", opts...)
	if err != nil {
		return nil, err
	}
	x := &backupTransferServiceTransferIncBackupClient{stream}
	return x, nil
}

type BackupTransferService_TransferIncBackupClient interface {
	Send(*IncBackupContentStream) error
	CloseAndRecv() (*BackupReply, error)
	grpc.ClientStream
}

type backupTransferServiceTransferIncBackupClient struct {
	grpc.ClientStream
}

func (x *backupTransferServiceTransferIncBackupClient) Send(m *IncBackupContentStream) error {
	return x.ClientStream.SendMsg(m)
}

func (x *backupTransferServiceTransferIncBackupClient) CloseAndRecv() (*BackupReply, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(BackupReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for BackupTransferService service

type BackupTransferServiceServer interface {
	TransferFullBackup(BackupTransferService_TransferFullBackupServer) error
	TransferIncBackup(BackupTransferService_TransferIncBackupServer) error
}

func RegisterBackupTransferServiceServer(s *grpc.Server, srv BackupTransferServiceServer) {
	s.RegisterService(&_BackupTransferService_serviceDesc, srv)
}

func _BackupTransferService_TransferFullBackup_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BackupTransferServiceServer).TransferFullBackup(&backupTransferServiceTransferFullBackupServer{stream})
}

type BackupTransferService_TransferFullBackupServer interface {
	SendAndClose(*BackupReply) error
	Recv() (*FullBackupContentStream, error)
	grpc.ServerStream
}

type backupTransferServiceTransferFullBackupServer struct {
	grpc.ServerStream
}

func (x *backupTransferServiceTransferFullBackupServer) SendAndClose(m *BackupReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *backupTransferServiceTransferFullBackupServer) Recv() (*FullBackupContentStream, error) {
	m := new(FullBackupContentStream)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _BackupTransferService_TransferIncBackup_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BackupTransferServiceServer).TransferIncBackup(&backupTransferServiceTransferIncBackupServer{stream})
}

type BackupTransferService_TransferIncBackupServer interface {
	SendAndClose(*BackupReply) error
	Recv() (*IncBackupContentStream, error)
	grpc.ServerStream
}

type backupTransferServiceTransferIncBackupServer struct {
	grpc.ServerStream
}

func (x *backupTransferServiceTransferIncBackupServer) SendAndClose(m *BackupReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *backupTransferServiceTransferIncBackupServer) Recv() (*IncBackupContentStream, error) {
	m := new(IncBackupContentStream)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _BackupTransferService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.BackupTransferService",
	HandlerType: (*BackupTransferServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "TransferFullBackup",
			Handler:       _BackupTransferService_TransferFullBackup_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "TransferIncBackup",
			Handler:       _BackupTransferService_TransferIncBackup_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "tempbackup/proto/backup.proto",
}

func init() { proto1.RegisterFile("tempbackup/proto/backup.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 227 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2d, 0x49, 0xcd, 0x2d,
	0x48, 0x4a, 0x4c, 0xce, 0x2e, 0x2d, 0xd0, 0x2f, 0x28, 0xca, 0x2f, 0xc9, 0xd7, 0x87, 0x70, 0xf4,
	0xc0, 0x1c, 0x21, 0x56, 0x30, 0xa5, 0xe4, 0xcc, 0x25, 0xee, 0x56, 0x9a, 0x93, 0xe3, 0x04, 0x96,
	0x72, 0xce, 0xcf, 0x2b, 0x49, 0xcd, 0x2b, 0x09, 0x2e, 0x29, 0x4a, 0x4d, 0xcc, 0x15, 0xe2, 0xe3,
	0x62, 0x4a, 0x49, 0x92, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0c, 0x62, 0x4a, 0x49, 0x12, 0x92, 0xe0,
	0x62, 0x4f, 0x86, 0x28, 0x90, 0x60, 0x52, 0x60, 0xd4, 0xe0, 0x09, 0x82, 0x71, 0x95, 0x42, 0xb8,
	0xc4, 0x3c, 0xf3, 0x92, 0x89, 0x31, 0x43, 0x80, 0x8b, 0x39, 0xa7, 0x38, 0x0f, 0xac, 0x9f, 0x33,
	0x08, 0xc4, 0x44, 0x36, 0x95, 0x19, 0xd5, 0x54, 0x75, 0x2e, 0x6e, 0x88, 0x91, 0x41, 0xa9, 0x05,
	0x39, 0x95, 0x20, 0x85, 0xb9, 0xa9, 0xc5, 0xc5, 0x89, 0xe9, 0xa9, 0x50, 0xf3, 0x60, 0x5c, 0xa3,
	0x8d, 0x8c, 0x5c, 0xa2, 0x10, 0x95, 0x21, 0x45, 0x89, 0x79, 0xc5, 0x69, 0xa9, 0x45, 0xc1, 0xa9,
	0x45, 0x65, 0x99, 0xc9, 0xa9, 0x42, 0x3e, 0x5c, 0x42, 0x30, 0x21, 0x84, 0x2f, 0x85, 0xe4, 0x20,
	0x41, 0xa0, 0x87, 0xc3, 0xe3, 0x52, 0x42, 0x50, 0x79, 0x24, 0xdb, 0x95, 0x18, 0x34, 0x18, 0x85,
	0xbc, 0xb8, 0x04, 0x61, 0xa6, 0xc1, 0xbd, 0x2b, 0x24, 0x0b, 0x55, 0x8c, 0x3d, 0x00, 0x70, 0x99,
	0x95, 0xc4, 0x06, 0x16, 0x36, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0x42, 0xd4, 0xef, 0x3f, 0xa6,
	0x01, 0x00, 0x00,
}