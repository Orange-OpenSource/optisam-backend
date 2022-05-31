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

// LicenseServiceClient is the client API for LicenseService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LicenseServiceClient interface {
	GetOverAllCompliance(ctx context.Context, in *GetOverAllComplianceRequest, opts ...grpc.CallOption) (*GetOverAllComplianceResponse, error)
	ListAcqRightsForProduct(ctx context.Context, in *ListAcquiredRightsForProductRequest, opts ...grpc.CallOption) (*ListAcquiredRightsForProductResponse, error)
	ListAcqRightsForApplicationsProduct(ctx context.Context, in *ListAcqRightsForApplicationsProductRequest, opts ...grpc.CallOption) (*ListAcqRightsForApplicationsProductResponse, error)
	// ListComputationDetails
	ListComputationDetails(ctx context.Context, in *ListComputationDetailsRequest, opts ...grpc.CallOption) (*ListComputationDetailsResponse, error)
	ListAcqRightsForAggregation(ctx context.Context, in *ListAcqRightsForAggregationRequest, opts ...grpc.CallOption) (*ListAcqRightsForAggregationResponse, error)
	ProductLicensesForMetric(ctx context.Context, in *ProductLicensesForMetricRequest, opts ...grpc.CallOption) (*ProductLicensesForMetricResponse, error)
	LicensesForEquipAndMetric(ctx context.Context, in *LicensesForEquipAndMetricRequest, opts ...grpc.CallOption) (*LicensesForEquipAndMetricResponse, error)
}

type licenseServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewLicenseServiceClient(cc grpc.ClientConnInterface) LicenseServiceClient {
	return &licenseServiceClient{cc}
}

func (c *licenseServiceClient) GetOverAllCompliance(ctx context.Context, in *GetOverAllComplianceRequest, opts ...grpc.CallOption) (*GetOverAllComplianceResponse, error) {
	out := new(GetOverAllComplianceResponse)
	err := c.cc.Invoke(ctx, "/optisam.license.v1.LicenseService/GetOverAllCompliance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *licenseServiceClient) ListAcqRightsForProduct(ctx context.Context, in *ListAcquiredRightsForProductRequest, opts ...grpc.CallOption) (*ListAcquiredRightsForProductResponse, error) {
	out := new(ListAcquiredRightsForProductResponse)
	err := c.cc.Invoke(ctx, "/optisam.license.v1.LicenseService/ListAcqRightsForProduct", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *licenseServiceClient) ListAcqRightsForApplicationsProduct(ctx context.Context, in *ListAcqRightsForApplicationsProductRequest, opts ...grpc.CallOption) (*ListAcqRightsForApplicationsProductResponse, error) {
	out := new(ListAcqRightsForApplicationsProductResponse)
	err := c.cc.Invoke(ctx, "/optisam.license.v1.LicenseService/ListAcqRightsForApplicationsProduct", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *licenseServiceClient) ListComputationDetails(ctx context.Context, in *ListComputationDetailsRequest, opts ...grpc.CallOption) (*ListComputationDetailsResponse, error) {
	out := new(ListComputationDetailsResponse)
	err := c.cc.Invoke(ctx, "/optisam.license.v1.LicenseService/ListComputationDetails", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *licenseServiceClient) ListAcqRightsForAggregation(ctx context.Context, in *ListAcqRightsForAggregationRequest, opts ...grpc.CallOption) (*ListAcqRightsForAggregationResponse, error) {
	out := new(ListAcqRightsForAggregationResponse)
	err := c.cc.Invoke(ctx, "/optisam.license.v1.LicenseService/ListAcqRightsForAggregation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *licenseServiceClient) ProductLicensesForMetric(ctx context.Context, in *ProductLicensesForMetricRequest, opts ...grpc.CallOption) (*ProductLicensesForMetricResponse, error) {
	out := new(ProductLicensesForMetricResponse)
	err := c.cc.Invoke(ctx, "/optisam.license.v1.LicenseService/ProductLicensesForMetric", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *licenseServiceClient) LicensesForEquipAndMetric(ctx context.Context, in *LicensesForEquipAndMetricRequest, opts ...grpc.CallOption) (*LicensesForEquipAndMetricResponse, error) {
	out := new(LicensesForEquipAndMetricResponse)
	err := c.cc.Invoke(ctx, "/optisam.license.v1.LicenseService/LicensesForEquipAndMetric", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LicenseServiceServer is the server API for LicenseService service.
// All implementations should embed UnimplementedLicenseServiceServer
// for forward compatibility
type LicenseServiceServer interface {
	GetOverAllCompliance(context.Context, *GetOverAllComplianceRequest) (*GetOverAllComplianceResponse, error)
	ListAcqRightsForProduct(context.Context, *ListAcquiredRightsForProductRequest) (*ListAcquiredRightsForProductResponse, error)
	ListAcqRightsForApplicationsProduct(context.Context, *ListAcqRightsForApplicationsProductRequest) (*ListAcqRightsForApplicationsProductResponse, error)
	// ListComputationDetails
	ListComputationDetails(context.Context, *ListComputationDetailsRequest) (*ListComputationDetailsResponse, error)
	ListAcqRightsForAggregation(context.Context, *ListAcqRightsForAggregationRequest) (*ListAcqRightsForAggregationResponse, error)
	ProductLicensesForMetric(context.Context, *ProductLicensesForMetricRequest) (*ProductLicensesForMetricResponse, error)
	LicensesForEquipAndMetric(context.Context, *LicensesForEquipAndMetricRequest) (*LicensesForEquipAndMetricResponse, error)
}

// UnimplementedLicenseServiceServer should be embedded to have forward compatible implementations.
type UnimplementedLicenseServiceServer struct {
}

func (UnimplementedLicenseServiceServer) GetOverAllCompliance(context.Context, *GetOverAllComplianceRequest) (*GetOverAllComplianceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOverAllCompliance not implemented")
}
func (UnimplementedLicenseServiceServer) ListAcqRightsForProduct(context.Context, *ListAcquiredRightsForProductRequest) (*ListAcquiredRightsForProductResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAcqRightsForProduct not implemented")
}
func (UnimplementedLicenseServiceServer) ListAcqRightsForApplicationsProduct(context.Context, *ListAcqRightsForApplicationsProductRequest) (*ListAcqRightsForApplicationsProductResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAcqRightsForApplicationsProduct not implemented")
}
func (UnimplementedLicenseServiceServer) ListComputationDetails(context.Context, *ListComputationDetailsRequest) (*ListComputationDetailsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListComputationDetails not implemented")
}
func (UnimplementedLicenseServiceServer) ListAcqRightsForAggregation(context.Context, *ListAcqRightsForAggregationRequest) (*ListAcqRightsForAggregationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAcqRightsForAggregation not implemented")
}
func (UnimplementedLicenseServiceServer) ProductLicensesForMetric(context.Context, *ProductLicensesForMetricRequest) (*ProductLicensesForMetricResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProductLicensesForMetric not implemented")
}
func (UnimplementedLicenseServiceServer) LicensesForEquipAndMetric(context.Context, *LicensesForEquipAndMetricRequest) (*LicensesForEquipAndMetricResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LicensesForEquipAndMetric not implemented")
}

// UnsafeLicenseServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LicenseServiceServer will
// result in compilation errors.
type UnsafeLicenseServiceServer interface {
	mustEmbedUnimplementedLicenseServiceServer()
}

func RegisterLicenseServiceServer(s grpc.ServiceRegistrar, srv LicenseServiceServer) {
	s.RegisterService(&_LicenseService_serviceDesc, srv)
}

func _LicenseService_GetOverAllCompliance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOverAllComplianceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LicenseServiceServer).GetOverAllCompliance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.license.v1.LicenseService/GetOverAllCompliance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LicenseServiceServer).GetOverAllCompliance(ctx, req.(*GetOverAllComplianceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LicenseService_ListAcqRightsForProduct_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAcquiredRightsForProductRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LicenseServiceServer).ListAcqRightsForProduct(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.license.v1.LicenseService/ListAcqRightsForProduct",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LicenseServiceServer).ListAcqRightsForProduct(ctx, req.(*ListAcquiredRightsForProductRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LicenseService_ListAcqRightsForApplicationsProduct_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAcqRightsForApplicationsProductRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LicenseServiceServer).ListAcqRightsForApplicationsProduct(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.license.v1.LicenseService/ListAcqRightsForApplicationsProduct",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LicenseServiceServer).ListAcqRightsForApplicationsProduct(ctx, req.(*ListAcqRightsForApplicationsProductRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LicenseService_ListComputationDetails_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListComputationDetailsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LicenseServiceServer).ListComputationDetails(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.license.v1.LicenseService/ListComputationDetails",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LicenseServiceServer).ListComputationDetails(ctx, req.(*ListComputationDetailsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LicenseService_ListAcqRightsForAggregation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAcqRightsForAggregationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LicenseServiceServer).ListAcqRightsForAggregation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.license.v1.LicenseService/ListAcqRightsForAggregation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LicenseServiceServer).ListAcqRightsForAggregation(ctx, req.(*ListAcqRightsForAggregationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LicenseService_ProductLicensesForMetric_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProductLicensesForMetricRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LicenseServiceServer).ProductLicensesForMetric(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.license.v1.LicenseService/ProductLicensesForMetric",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LicenseServiceServer).ProductLicensesForMetric(ctx, req.(*ProductLicensesForMetricRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LicenseService_LicensesForEquipAndMetric_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LicensesForEquipAndMetricRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LicenseServiceServer).LicensesForEquipAndMetric(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.license.v1.LicenseService/LicensesForEquipAndMetric",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LicenseServiceServer).LicensesForEquipAndMetric(ctx, req.(*LicensesForEquipAndMetricRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _LicenseService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "optisam.license.v1.LicenseService",
	HandlerType: (*LicenseServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetOverAllCompliance",
			Handler:    _LicenseService_GetOverAllCompliance_Handler,
		},
		{
			MethodName: "ListAcqRightsForProduct",
			Handler:    _LicenseService_ListAcqRightsForProduct_Handler,
		},
		{
			MethodName: "ListAcqRightsForApplicationsProduct",
			Handler:    _LicenseService_ListAcqRightsForApplicationsProduct_Handler,
		},
		{
			MethodName: "ListComputationDetails",
			Handler:    _LicenseService_ListComputationDetails_Handler,
		},
		{
			MethodName: "ListAcqRightsForAggregation",
			Handler:    _LicenseService_ListAcqRightsForAggregation_Handler,
		},
		{
			MethodName: "ProductLicensesForMetric",
			Handler:    _LicenseService_ProductLicensesForMetric_Handler,
		},
		{
			MethodName: "LicensesForEquipAndMetric",
			Handler:    _LicenseService_LicensesForEquipAndMetric_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "license.proto",
}
