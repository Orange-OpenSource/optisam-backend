// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// ReportServiceClient is the client API for ReportService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ReportServiceClient interface {
	SubmitReport(ctx context.Context, in *SubmitReportRequest, opts ...grpc.CallOption) (*SubmitReportResponse, error)
	ListReport(ctx context.Context, in *ListReportRequest, opts ...grpc.CallOption) (*ListReportResponse, error)
	DownloadReport(ctx context.Context, in *DownloadReportRequest, opts ...grpc.CallOption) (*DownloadReportResponse, error)
	ListReportType(ctx context.Context, in *ListReportTypeRequest, opts ...grpc.CallOption) (*ListReportTypeResponse, error)
	DropReportData(ctx context.Context, in *DropReportDataRequest, opts ...grpc.CallOption) (*DropReportDataResponse, error)
}

type reportServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewReportServiceClient(cc grpc.ClientConnInterface) ReportServiceClient {
	return &reportServiceClient{cc}
}

func (c *reportServiceClient) SubmitReport(ctx context.Context, in *SubmitReportRequest, opts ...grpc.CallOption) (*SubmitReportResponse, error) {
	out := new(SubmitReportResponse)
	err := c.cc.Invoke(ctx, "/optisam.reports.v1.ReportService/SubmitReport", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *reportServiceClient) ListReport(ctx context.Context, in *ListReportRequest, opts ...grpc.CallOption) (*ListReportResponse, error) {
	out := new(ListReportResponse)
	err := c.cc.Invoke(ctx, "/optisam.reports.v1.ReportService/ListReport", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *reportServiceClient) DownloadReport(ctx context.Context, in *DownloadReportRequest, opts ...grpc.CallOption) (*DownloadReportResponse, error) {
	out := new(DownloadReportResponse)
	err := c.cc.Invoke(ctx, "/optisam.reports.v1.ReportService/DownloadReport", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *reportServiceClient) ListReportType(ctx context.Context, in *ListReportTypeRequest, opts ...grpc.CallOption) (*ListReportTypeResponse, error) {
	out := new(ListReportTypeResponse)
	err := c.cc.Invoke(ctx, "/optisam.reports.v1.ReportService/ListReportType", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *reportServiceClient) DropReportData(ctx context.Context, in *DropReportDataRequest, opts ...grpc.CallOption) (*DropReportDataResponse, error) {
	out := new(DropReportDataResponse)
	err := c.cc.Invoke(ctx, "/optisam.reports.v1.ReportService/DropReportData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReportServiceServer is the server API for ReportService service.
// All implementations should embed UnimplementedReportServiceServer
// for forward compatibility
type ReportServiceServer interface {
	SubmitReport(context.Context, *SubmitReportRequest) (*SubmitReportResponse, error)
	ListReport(context.Context, *ListReportRequest) (*ListReportResponse, error)
	DownloadReport(context.Context, *DownloadReportRequest) (*DownloadReportResponse, error)
	ListReportType(context.Context, *ListReportTypeRequest) (*ListReportTypeResponse, error)
	DropReportData(context.Context, *DropReportDataRequest) (*DropReportDataResponse, error)
}

// UnimplementedReportServiceServer should be embedded to have forward compatible implementations.
type UnimplementedReportServiceServer struct {
}

func (UnimplementedReportServiceServer) SubmitReport(context.Context, *SubmitReportRequest) (*SubmitReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitReport not implemented")
}
func (UnimplementedReportServiceServer) ListReport(context.Context, *ListReportRequest) (*ListReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListReport not implemented")
}
func (UnimplementedReportServiceServer) DownloadReport(context.Context, *DownloadReportRequest) (*DownloadReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DownloadReport not implemented")
}
func (UnimplementedReportServiceServer) ListReportType(context.Context, *ListReportTypeRequest) (*ListReportTypeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListReportType not implemented")
}
func (UnimplementedReportServiceServer) DropReportData(context.Context, *DropReportDataRequest) (*DropReportDataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DropReportData not implemented")
}

// UnsafeReportServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ReportServiceServer will
// result in compilation errors.
type UnsafeReportServiceServer interface {
	mustEmbedUnimplementedReportServiceServer()
}

func RegisterReportServiceServer(s grpc.ServiceRegistrar, srv ReportServiceServer) {
	s.RegisterService(&_ReportService_serviceDesc, srv)
}

func _ReportService_SubmitReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReportServiceServer).SubmitReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.reports.v1.ReportService/SubmitReport",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReportServiceServer).SubmitReport(ctx, req.(*SubmitReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReportService_ListReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReportServiceServer).ListReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.reports.v1.ReportService/ListReport",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReportServiceServer).ListReport(ctx, req.(*ListReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReportService_DownloadReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DownloadReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReportServiceServer).DownloadReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.reports.v1.ReportService/DownloadReport",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReportServiceServer).DownloadReport(ctx, req.(*DownloadReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReportService_ListReportType_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListReportTypeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReportServiceServer).ListReportType(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.reports.v1.ReportService/ListReportType",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReportServiceServer).ListReportType(ctx, req.(*ListReportTypeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReportService_DropReportData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DropReportDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReportServiceServer).DropReportData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.reports.v1.ReportService/DropReportData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReportServiceServer).DropReportData(ctx, req.(*DropReportDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ReportService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "optisam.reports.v1.ReportService",
	HandlerType: (*ReportServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitReport",
			Handler:    _ReportService_SubmitReport_Handler,
		},
		{
			MethodName: "ListReport",
			Handler:    _ReportService_ListReport_Handler,
		},
		{
			MethodName: "DownloadReport",
			Handler:    _ReportService_DownloadReport_Handler,
		},
		{
			MethodName: "ListReportType",
			Handler:    _ReportService_ListReportType_Handler,
		},
		{
			MethodName: "DropReportData",
			Handler:    _ReportService_DropReportData_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "report.proto",
}
