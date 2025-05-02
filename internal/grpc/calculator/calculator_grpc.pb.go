
package calculator

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	CalculatorAgentService_GetTask_FullMethodName      = "/calculator.CalculatorAgentService/GetTask"
	CalculatorAgentService_SubmitResult_FullMethodName = "/calculator.CalculatorAgentService/SubmitResult"
)

type CalculatorAgentServiceClient interface {
	GetTask(ctx context.Context, in *GetTaskRequest, opts ...grpc.CallOption) (*GetTaskResponse, error)
	SubmitResult(ctx context.Context, in *SubmitResultRequest, opts ...grpc.CallOption) (*SubmitResultResponse, error)
}

type calculatorAgentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCalculatorAgentServiceClient(cc grpc.ClientConnInterface) CalculatorAgentServiceClient {
	return &calculatorAgentServiceClient{cc}
}

func (c *calculatorAgentServiceClient) GetTask(ctx context.Context, in *GetTaskRequest, opts ...grpc.CallOption) (*GetTaskResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetTaskResponse)
	err := c.cc.Invoke(ctx, CalculatorAgentService_GetTask_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *calculatorAgentServiceClient) SubmitResult(ctx context.Context, in *SubmitResultRequest, opts ...grpc.CallOption) (*SubmitResultResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SubmitResultResponse)
	err := c.cc.Invoke(ctx, CalculatorAgentService_SubmitResult_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type CalculatorAgentServiceServer interface {
	GetTask(context.Context, *GetTaskRequest) (*GetTaskResponse, error)
	SubmitResult(context.Context, *SubmitResultRequest) (*SubmitResultResponse, error)
	mustEmbedUnimplementedCalculatorAgentServiceServer()
}

type UnimplementedCalculatorAgentServiceServer struct{}

func (UnimplementedCalculatorAgentServiceServer) GetTask(context.Context, *GetTaskRequest) (*GetTaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTask not implemented")
}
func (UnimplementedCalculatorAgentServiceServer) SubmitResult(context.Context, *SubmitResultRequest) (*SubmitResultResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitResult not implemented")
}
func (UnimplementedCalculatorAgentServiceServer) mustEmbedUnimplementedCalculatorAgentServiceServer() {
}
func (UnimplementedCalculatorAgentServiceServer) testEmbeddedByValue() {}

type UnsafeCalculatorAgentServiceServer interface {
	mustEmbedUnimplementedCalculatorAgentServiceServer()
}

func RegisterCalculatorAgentServiceServer(s grpc.ServiceRegistrar, srv CalculatorAgentServiceServer) {
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&CalculatorAgentService_ServiceDesc, srv)
}

func _CalculatorAgentService_GetTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CalculatorAgentServiceServer).GetTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CalculatorAgentService_GetTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CalculatorAgentServiceServer).GetTask(ctx, req.(*GetTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CalculatorAgentService_SubmitResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitResultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CalculatorAgentServiceServer).SubmitResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CalculatorAgentService_SubmitResult_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CalculatorAgentServiceServer).SubmitResult(ctx, req.(*SubmitResultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var CalculatorAgentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "calculator.CalculatorAgentService",
	HandlerType: (*CalculatorAgentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetTask",
			Handler:    _CalculatorAgentService_GetTask_Handler,
		},
		{
			MethodName: "SubmitResult",
			Handler:    _CalculatorAgentService_SubmitResult_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "calculator.proto",
}