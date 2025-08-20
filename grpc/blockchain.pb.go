package main

import (
	"context"

	"google.golang.org/grpc"
)

// Block message type
type Block struct {
	Hash         string        `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	PreviousHash string        `protobuf:"bytes,2,opt,name=previous_hash,json=previousHash,proto3" json:"previous_hash,omitempty"`
	Timestamp    int64         `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Transactions []Transaction `protobuf:"bytes,4,rep,name=transactions,proto3" json:"transactions,omitempty"`
	Nonce        int32         `protobuf:"varint,5,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Height       int32         `protobuf:"varint,6,opt,name=height,proto3" json:"height,omitempty"`
}

// Transaction message type
type Transaction struct {
	Hash      string  `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	From      string  `protobuf:"bytes,2,opt,name=from,proto3" json:"from,omitempty"`
	To        string  `protobuf:"bytes,3,opt,name=to,proto3" json:"to,omitempty"`
	Amount    float64 `protobuf:"fixed64,4,opt,name=amount,proto3" json:"amount,omitempty"`
	Timestamp int64   `protobuf:"varint,5,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Signature string  `protobuf:"bytes,6,opt,name=signature,proto3" json:"signature,omitempty"`
}

// Request/Response types
type GetBlockRequest struct {
	Hash string `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
}

type GetBlockResponse struct {
	Block *Block `protobuf:"bytes,1,opt,name=block,proto3" json:"block,omitempty"`
	Found bool   `protobuf:"varint,2,opt,name=found,proto3" json:"found,omitempty"`
}

type GetLatestBlockRequest struct{}

type GetLatestBlockResponse struct {
	Block *Block `protobuf:"bytes,1,opt,name=block,proto3" json:"block,omitempty"`
}

type AddTransactionRequest struct {
	Transaction *Transaction `protobuf:"bytes,1,opt,name=transaction,proto3" json:"transaction,omitempty"`
}

type AddTransactionResponse struct {
	Success         bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message         string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	TransactionHash string `protobuf:"bytes,3,opt,name=transaction_hash,json=transactionHash,proto3" json:"transaction_hash,omitempty"`
}

type GetTransactionRequest struct {
	Hash string `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
}

type GetTransactionResponse struct {
	Transaction *Transaction `protobuf:"bytes,1,opt,name=transaction,proto3" json:"transaction,omitempty"`
	Found       bool         `protobuf:"varint,2,opt,name=found,proto3" json:"found,omitempty"`
}

type SubscribeBlocksRequest struct{}

// Service interface
type BlockchainServiceServer interface {
	GetBlock(context.Context, *GetBlockRequest) (*GetBlockResponse, error)
	GetLatestBlock(context.Context, *GetLatestBlockRequest) (*GetLatestBlockResponse, error)
	AddTransaction(context.Context, *AddTransactionRequest) (*AddTransactionResponse, error)
	GetTransaction(context.Context, *GetTransactionRequest) (*GetTransactionResponse, error)
	SubscribeBlocks(*SubscribeBlocksRequest, BlockchainService_SubscribeBlocksServer) error
}

type BlockchainService_SubscribeBlocksServer interface {
	Send(*Block) error
	Context() context.Context
}

type UnimplementedBlockchainServiceServer struct{}

func (UnimplementedBlockchainServiceServer) GetBlock(context.Context, *GetBlockRequest) (*GetBlockResponse, error) {
	return nil, nil
}
func (UnimplementedBlockchainServiceServer) GetLatestBlock(context.Context, *GetLatestBlockRequest) (*GetLatestBlockResponse, error) {
	return nil, nil
}
func (UnimplementedBlockchainServiceServer) AddTransaction(context.Context, *AddTransactionRequest) (*AddTransactionResponse, error) {
	return nil, nil
}
func (UnimplementedBlockchainServiceServer) GetTransaction(context.Context, *GetTransactionRequest) (*GetTransactionResponse, error) {
	return nil, nil
}
func (UnimplementedBlockchainServiceServer) SubscribeBlocks(*SubscribeBlocksRequest, BlockchainService_SubscribeBlocksServer) error {
	return nil
}

// Client interface
type BlockchainServiceClient interface {
	GetBlock(ctx context.Context, in *GetBlockRequest, opts ...grpc.CallOption) (*GetBlockResponse, error)
	GetLatestBlock(ctx context.Context, in *GetLatestBlockRequest, opts ...grpc.CallOption) (*GetLatestBlockResponse, error)
	AddTransaction(ctx context.Context, in *AddTransactionRequest, opts ...grpc.CallOption) (*AddTransactionResponse, error)
	GetTransaction(ctx context.Context, in *GetTransactionRequest, opts ...grpc.CallOption) (*GetTransactionResponse, error)
	SubscribeBlocks(ctx context.Context, in *SubscribeBlocksRequest, opts ...grpc.CallOption) (BlockchainService_SubscribeBlocksClient, error)
}

type blockchainServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBlockchainServiceClient(cc grpc.ClientConnInterface) BlockchainServiceClient {
	return &blockchainServiceClient{cc}
}

func (c *blockchainServiceClient) GetBlock(ctx context.Context, in *GetBlockRequest, opts ...grpc.CallOption) (*GetBlockResponse, error) {
	out := new(GetBlockResponse)
	err := c.cc.Invoke(ctx, "/blockchain.BlockchainService/GetBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockchainServiceClient) GetLatestBlock(ctx context.Context, in *GetLatestBlockRequest, opts ...grpc.CallOption) (*GetLatestBlockResponse, error) {
	out := new(GetLatestBlockResponse)
	err := c.cc.Invoke(ctx, "/blockchain.BlockchainService/GetLatestBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockchainServiceClient) AddTransaction(ctx context.Context, in *AddTransactionRequest, opts ...grpc.CallOption) (*AddTransactionResponse, error) {
	out := new(AddTransactionResponse)
	err := c.cc.Invoke(ctx, "/blockchain.BlockchainService/AddTransaction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockchainServiceClient) GetTransaction(ctx context.Context, in *GetTransactionRequest, opts ...grpc.CallOption) (*GetTransactionResponse, error) {
	out := new(GetTransactionResponse)
	err := c.cc.Invoke(ctx, "/blockchain.BlockchainService/GetTransaction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockchainServiceClient) SubscribeBlocks(ctx context.Context, in *SubscribeBlocksRequest, opts ...grpc.CallOption) (BlockchainService_SubscribeBlocksClient, error) {
	stream, err := c.cc.NewStream(ctx, &BlockchainService_ServiceDesc.Streams[0], "/blockchain.BlockchainService/SubscribeBlocks", opts...)
	if err != nil {
		return nil, err
	}
	x := &blockchainServiceSubscribeBlocksClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type BlockchainService_SubscribeBlocksClient interface {
	Recv() (*Block, error)
	grpc.ClientStream
}

type blockchainServiceSubscribeBlocksClient struct {
	grpc.ClientStream
}

func (x *blockchainServiceSubscribeBlocksClient) Recv() (*Block, error) {
	m := new(Block)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server registration
func RegisterBlockchainServiceServer(s grpc.ServiceRegistrar, srv BlockchainServiceServer) {
	s.RegisterService(&BlockchainService_ServiceDesc, srv)
}

var BlockchainService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "blockchain.BlockchainService",
	HandlerType: (*BlockchainServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetBlock",
			Handler:    _BlockchainService_GetBlock_Handler,
		},
		{
			MethodName: "GetLatestBlock",
			Handler:    _BlockchainService_GetLatestBlock_Handler,
		},
		{
			MethodName: "AddTransaction",
			Handler:    _BlockchainService_AddTransaction_Handler,
		},
		{
			MethodName: "GetTransaction",
			Handler:    _BlockchainService_GetTransaction_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SubscribeBlocks",
			Handler:       _BlockchainService_SubscribeBlocks_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "blockchain.proto",
}

func _BlockchainService_GetBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockchainServiceServer).GetBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/blockchain.BlockchainService/GetBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockchainServiceServer).GetBlock(ctx, req.(*GetBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockchainService_GetLatestBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLatestBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockchainServiceServer).GetLatestBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/blockchain.BlockchainService/GetLatestBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockchainServiceServer).GetLatestBlock(ctx, req.(*GetLatestBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockchainService_AddTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddTransactionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockchainServiceServer).AddTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/blockchain.BlockchainService/AddTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockchainServiceServer).AddTransaction(ctx, req.(*AddTransactionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockchainService_GetTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTransactionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockchainServiceServer).GetTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/blockchain.BlockchainService/GetTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockchainServiceServer).GetTransaction(ctx, req.(*GetTransactionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockchainService_SubscribeBlocks_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SubscribeBlocksRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BlockchainServiceServer).SubscribeBlocks(m, &blockchainServiceSubscribeBlocksServer{stream})
}

type blockchainServiceSubscribeBlocksServer struct {
	grpc.ServerStream
}

func (x *blockchainServiceSubscribeBlocksServer) Send(m *Block) error {
	return x.ServerStream.SendMsg(m)
}

func (x *blockchainServiceSubscribeBlocksServer) Context() context.Context {
	return x.ServerStream.Context()
}