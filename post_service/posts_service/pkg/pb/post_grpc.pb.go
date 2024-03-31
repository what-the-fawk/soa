// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.12.4
// source: post.proto

package pb

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	PostService_NewPost_FullMethodName    = "/post.PostService/NewPost"
	PostService_UpdatePost_FullMethodName = "/post.PostService/UpdatePost"
	PostService_DeletePost_FullMethodName = "/post.PostService/DeletePost"
	PostService_GetPost_FullMethodName    = "/post.PostService/GetPost"
	PostService_GetPosts_FullMethodName   = "/post.PostService/GetPosts"
)

// PostServiceClient is the client API for PostService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PostServiceClient interface {
	NewPost(ctx context.Context, in *PostInfo, opts ...grpc.CallOption) (*PostID, error)
	UpdatePost(ctx context.Context, in *PostInfo, opts ...grpc.CallOption) (*empty.Empty, error)
	DeletePost(ctx context.Context, in *PostID, opts ...grpc.CallOption) (*empty.Empty, error)
	GetPost(ctx context.Context, in *PostID, opts ...grpc.CallOption) (*Post, error)
	GetPosts(ctx context.Context, in *PaginationInfo, opts ...grpc.CallOption) (PostService_GetPostsClient, error)
}

type postServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPostServiceClient(cc grpc.ClientConnInterface) PostServiceClient {
	return &postServiceClient{cc}
}

func (c *postServiceClient) NewPost(ctx context.Context, in *PostInfo, opts ...grpc.CallOption) (*PostID, error) {
	out := new(PostID)
	err := c.cc.Invoke(ctx, PostService_NewPost_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *postServiceClient) UpdatePost(ctx context.Context, in *PostInfo, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, PostService_UpdatePost_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *postServiceClient) DeletePost(ctx context.Context, in *PostID, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, PostService_DeletePost_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *postServiceClient) GetPost(ctx context.Context, in *PostID, opts ...grpc.CallOption) (*Post, error) {
	out := new(Post)
	err := c.cc.Invoke(ctx, PostService_GetPost_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *postServiceClient) GetPosts(ctx context.Context, in *PaginationInfo, opts ...grpc.CallOption) (PostService_GetPostsClient, error) {
	stream, err := c.cc.NewStream(ctx, &PostService_ServiceDesc.Streams[0], PostService_GetPosts_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &postServiceGetPostsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type PostService_GetPostsClient interface {
	Recv() (*Post, error)
	grpc.ClientStream
}

type postServiceGetPostsClient struct {
	grpc.ClientStream
}

func (x *postServiceGetPostsClient) Recv() (*Post, error) {
	m := new(Post)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// PostServiceServer is the server API for PostService service.
// All implementations must embed UnimplementedPostServiceServer
// for forward compatibility
type PostServiceServer interface {
	NewPost(context.Context, *PostInfo) (*PostID, error)
	UpdatePost(context.Context, *PostInfo) (*empty.Empty, error)
	DeletePost(context.Context, *PostID) (*empty.Empty, error)
	GetPost(context.Context, *PostID) (*Post, error)
	GetPosts(*PaginationInfo, PostService_GetPostsServer) error
	mustEmbedUnimplementedPostServiceServer()
}

// UnimplementedPostServiceServer must be embedded to have forward compatible implementations.
type UnimplementedPostServiceServer struct {
}

func (UnimplementedPostServiceServer) NewPost(context.Context, *PostInfo) (*PostID, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewPost not implemented")
}
func (UnimplementedPostServiceServer) UpdatePost(context.Context, *PostInfo) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatePost not implemented")
}
func (UnimplementedPostServiceServer) DeletePost(context.Context, *PostID) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeletePost not implemented")
}
func (UnimplementedPostServiceServer) GetPost(context.Context, *PostID) (*Post, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPost not implemented")
}
func (UnimplementedPostServiceServer) GetPosts(*PaginationInfo, PostService_GetPostsServer) error {
	return status.Errorf(codes.Unimplemented, "method GetPosts not implemented")
}
func (UnimplementedPostServiceServer) mustEmbedUnimplementedPostServiceServer() {}

// UnsafePostServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PostServiceServer will
// result in compilation errors.
type UnsafePostServiceServer interface {
	mustEmbedUnimplementedPostServiceServer()
}

func RegisterPostServiceServer(s grpc.ServiceRegistrar, srv PostServiceServer) {
	s.RegisterService(&PostService_ServiceDesc, srv)
}

func _PostService_NewPost_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PostServiceServer).NewPost(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PostService_NewPost_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PostServiceServer).NewPost(ctx, req.(*PostInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _PostService_UpdatePost_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PostServiceServer).UpdatePost(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PostService_UpdatePost_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PostServiceServer).UpdatePost(ctx, req.(*PostInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _PostService_DeletePost_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PostServiceServer).DeletePost(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PostService_DeletePost_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PostServiceServer).DeletePost(ctx, req.(*PostID))
	}
	return interceptor(ctx, in, info, handler)
}

func _PostService_GetPost_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PostServiceServer).GetPost(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PostService_GetPost_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PostServiceServer).GetPost(ctx, req.(*PostID))
	}
	return interceptor(ctx, in, info, handler)
}

func _PostService_GetPosts_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(PaginationInfo)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PostServiceServer).GetPosts(m, &postServiceGetPostsServer{stream})
}

type PostService_GetPostsServer interface {
	Send(*Post) error
	grpc.ServerStream
}

type postServiceGetPostsServer struct {
	grpc.ServerStream
}

func (x *postServiceGetPostsServer) Send(m *Post) error {
	return x.ServerStream.SendMsg(m)
}

// PostService_ServiceDesc is the grpc.ServiceDesc for PostService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PostService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "post.PostService",
	HandlerType: (*PostServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NewPost",
			Handler:    _PostService_NewPost_Handler,
		},
		{
			MethodName: "UpdatePost",
			Handler:    _PostService_UpdatePost_Handler,
		},
		{
			MethodName: "DeletePost",
			Handler:    _PostService_DeletePost_Handler,
		},
		{
			MethodName: "GetPost",
			Handler:    _PostService_GetPost_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetPosts",
			Handler:       _PostService_GetPosts_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "post.proto",
}