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

// ApplicationServiceClient is the client API for ApplicationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ApplicationServiceClient interface {
	UpsertApplication(ctx context.Context, in *UpsertApplicationRequest, opts ...grpc.CallOption) (*UpsertApplicationResponse, error)
	UpsertApplicationEquip(ctx context.Context, in *UpsertApplicationEquipRequest, opts ...grpc.CallOption) (*UpsertApplicationEquipResponse, error)
	DropApplicationData(ctx context.Context, in *DropApplicationDataRequest, opts ...grpc.CallOption) (*DropApplicationDataResponse, error)
	DeleteApplication(ctx context.Context, in *DeleteApplicationRequest, opts ...grpc.CallOption) (*DeleteApplicationResponse, error)
	UpsertInstance(ctx context.Context, in *UpsertInstanceRequest, opts ...grpc.CallOption) (*UpsertInstanceResponse, error)
	DeleteInstance(ctx context.Context, in *DeleteInstanceRequest, opts ...grpc.CallOption) (*DeleteInstanceResponse, error)
	ListApplications(ctx context.Context, in *ListApplicationsRequest, opts ...grpc.CallOption) (*ListApplicationsResponse, error)
	ListInstances(ctx context.Context, in *ListInstancesRequest, opts ...grpc.CallOption) (*ListInstancesResponse, error)
	// Obsolescense APIs
	ApplicationDomains(ctx context.Context, in *ApplicationDomainsRequest, opts ...grpc.CallOption) (*ApplicationDomainsResponse, error)
	ObsolescenceDomainCriticityMeta(ctx context.Context, in *DomainCriticityMetaRequest, opts ...grpc.CallOption) (*DomainCriticityMetaResponse, error)
	ObsolescenceMaintenanceCriticityMeta(ctx context.Context, in *MaintenanceCriticityMetaRequest, opts ...grpc.CallOption) (*MaintenanceCriticityMetaResponse, error)
	ObsolescenceRiskMeta(ctx context.Context, in *RiskMetaRequest, opts ...grpc.CallOption) (*RiskMetaResponse, error)
	ObsolescenceDomainCriticity(ctx context.Context, in *DomainCriticityRequest, opts ...grpc.CallOption) (*DomainCriticityResponse, error)
	PostObsolescenceDomainCriticity(ctx context.Context, in *PostDomainCriticityRequest, opts ...grpc.CallOption) (*PostDomainCriticityResponse, error)
	ObsolescenseMaintenanceCriticity(ctx context.Context, in *MaintenanceCriticityRequest, opts ...grpc.CallOption) (*MaintenanceCriticityResponse, error)
	PostObsolescenseMaintenanceCriticity(ctx context.Context, in *PostMaintenanceCriticityRequest, opts ...grpc.CallOption) (*PostMaintenanceCriticityResponse, error)
	ObsolescenseRiskMatrix(ctx context.Context, in *RiskMatrixRequest, opts ...grpc.CallOption) (*RiskMatrixResponse, error)
	PostObsolescenseRiskMatrix(ctx context.Context, in *PostRiskMatrixRequest, opts ...grpc.CallOption) (*PostRiskMatrixResponse, error)
	DropObscolenscenceData(ctx context.Context, in *DropObscolenscenceDataRequest, opts ...grpc.CallOption) (*DropObscolenscenceDataResponse, error)
	GetEquipmentsByApplication(ctx context.Context, in *GetEquipmentsByApplicationRequest, opts ...grpc.CallOption) (*GetEquipmentsByApplicationResponse, error)
}

type applicationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewApplicationServiceClient(cc grpc.ClientConnInterface) ApplicationServiceClient {
	return &applicationServiceClient{cc}
}

func (c *applicationServiceClient) UpsertApplication(ctx context.Context, in *UpsertApplicationRequest, opts ...grpc.CallOption) (*UpsertApplicationResponse, error) {
	out := new(UpsertApplicationResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/UpsertApplication", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) UpsertApplicationEquip(ctx context.Context, in *UpsertApplicationEquipRequest, opts ...grpc.CallOption) (*UpsertApplicationEquipResponse, error) {
	out := new(UpsertApplicationEquipResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/UpsertApplicationEquip", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) DropApplicationData(ctx context.Context, in *DropApplicationDataRequest, opts ...grpc.CallOption) (*DropApplicationDataResponse, error) {
	out := new(DropApplicationDataResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/DropApplicationData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) DeleteApplication(ctx context.Context, in *DeleteApplicationRequest, opts ...grpc.CallOption) (*DeleteApplicationResponse, error) {
	out := new(DeleteApplicationResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/DeleteApplication", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) UpsertInstance(ctx context.Context, in *UpsertInstanceRequest, opts ...grpc.CallOption) (*UpsertInstanceResponse, error) {
	out := new(UpsertInstanceResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/UpsertInstance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) DeleteInstance(ctx context.Context, in *DeleteInstanceRequest, opts ...grpc.CallOption) (*DeleteInstanceResponse, error) {
	out := new(DeleteInstanceResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/DeleteInstance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) ListApplications(ctx context.Context, in *ListApplicationsRequest, opts ...grpc.CallOption) (*ListApplicationsResponse, error) {
	out := new(ListApplicationsResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/ListApplications", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) ListInstances(ctx context.Context, in *ListInstancesRequest, opts ...grpc.CallOption) (*ListInstancesResponse, error) {
	out := new(ListInstancesResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/ListInstances", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) ApplicationDomains(ctx context.Context, in *ApplicationDomainsRequest, opts ...grpc.CallOption) (*ApplicationDomainsResponse, error) {
	out := new(ApplicationDomainsResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/ApplicationDomains", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) ObsolescenceDomainCriticityMeta(ctx context.Context, in *DomainCriticityMetaRequest, opts ...grpc.CallOption) (*DomainCriticityMetaResponse, error) {
	out := new(DomainCriticityMetaResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/ObsolescenceDomainCriticityMeta", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) ObsolescenceMaintenanceCriticityMeta(ctx context.Context, in *MaintenanceCriticityMetaRequest, opts ...grpc.CallOption) (*MaintenanceCriticityMetaResponse, error) {
	out := new(MaintenanceCriticityMetaResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/ObsolescenceMaintenanceCriticityMeta", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) ObsolescenceRiskMeta(ctx context.Context, in *RiskMetaRequest, opts ...grpc.CallOption) (*RiskMetaResponse, error) {
	out := new(RiskMetaResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/ObsolescenceRiskMeta", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) ObsolescenceDomainCriticity(ctx context.Context, in *DomainCriticityRequest, opts ...grpc.CallOption) (*DomainCriticityResponse, error) {
	out := new(DomainCriticityResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/ObsolescenceDomainCriticity", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) PostObsolescenceDomainCriticity(ctx context.Context, in *PostDomainCriticityRequest, opts ...grpc.CallOption) (*PostDomainCriticityResponse, error) {
	out := new(PostDomainCriticityResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/PostObsolescenceDomainCriticity", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) ObsolescenseMaintenanceCriticity(ctx context.Context, in *MaintenanceCriticityRequest, opts ...grpc.CallOption) (*MaintenanceCriticityResponse, error) {
	out := new(MaintenanceCriticityResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/ObsolescenseMaintenanceCriticity", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) PostObsolescenseMaintenanceCriticity(ctx context.Context, in *PostMaintenanceCriticityRequest, opts ...grpc.CallOption) (*PostMaintenanceCriticityResponse, error) {
	out := new(PostMaintenanceCriticityResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/PostObsolescenseMaintenanceCriticity", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) ObsolescenseRiskMatrix(ctx context.Context, in *RiskMatrixRequest, opts ...grpc.CallOption) (*RiskMatrixResponse, error) {
	out := new(RiskMatrixResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/ObsolescenseRiskMatrix", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) PostObsolescenseRiskMatrix(ctx context.Context, in *PostRiskMatrixRequest, opts ...grpc.CallOption) (*PostRiskMatrixResponse, error) {
	out := new(PostRiskMatrixResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/PostObsolescenseRiskMatrix", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) DropObscolenscenceData(ctx context.Context, in *DropObscolenscenceDataRequest, opts ...grpc.CallOption) (*DropObscolenscenceDataResponse, error) {
	out := new(DropObscolenscenceDataResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/DropObscolenscenceData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationServiceClient) GetEquipmentsByApplication(ctx context.Context, in *GetEquipmentsByApplicationRequest, opts ...grpc.CallOption) (*GetEquipmentsByApplicationResponse, error) {
	out := new(GetEquipmentsByApplicationResponse)
	err := c.cc.Invoke(ctx, "/optisam.applications.v1.ApplicationService/GetEquipmentsByApplication", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ApplicationServiceServer is the server API for ApplicationService service.
// All implementations should embed UnimplementedApplicationServiceServer
// for forward compatibility
type ApplicationServiceServer interface {
	UpsertApplication(context.Context, *UpsertApplicationRequest) (*UpsertApplicationResponse, error)
	UpsertApplicationEquip(context.Context, *UpsertApplicationEquipRequest) (*UpsertApplicationEquipResponse, error)
	DropApplicationData(context.Context, *DropApplicationDataRequest) (*DropApplicationDataResponse, error)
	DeleteApplication(context.Context, *DeleteApplicationRequest) (*DeleteApplicationResponse, error)
	UpsertInstance(context.Context, *UpsertInstanceRequest) (*UpsertInstanceResponse, error)
	DeleteInstance(context.Context, *DeleteInstanceRequest) (*DeleteInstanceResponse, error)
	ListApplications(context.Context, *ListApplicationsRequest) (*ListApplicationsResponse, error)
	ListInstances(context.Context, *ListInstancesRequest) (*ListInstancesResponse, error)
	// Obsolescense APIs
	ApplicationDomains(context.Context, *ApplicationDomainsRequest) (*ApplicationDomainsResponse, error)
	ObsolescenceDomainCriticityMeta(context.Context, *DomainCriticityMetaRequest) (*DomainCriticityMetaResponse, error)
	ObsolescenceMaintenanceCriticityMeta(context.Context, *MaintenanceCriticityMetaRequest) (*MaintenanceCriticityMetaResponse, error)
	ObsolescenceRiskMeta(context.Context, *RiskMetaRequest) (*RiskMetaResponse, error)
	ObsolescenceDomainCriticity(context.Context, *DomainCriticityRequest) (*DomainCriticityResponse, error)
	PostObsolescenceDomainCriticity(context.Context, *PostDomainCriticityRequest) (*PostDomainCriticityResponse, error)
	ObsolescenseMaintenanceCriticity(context.Context, *MaintenanceCriticityRequest) (*MaintenanceCriticityResponse, error)
	PostObsolescenseMaintenanceCriticity(context.Context, *PostMaintenanceCriticityRequest) (*PostMaintenanceCriticityResponse, error)
	ObsolescenseRiskMatrix(context.Context, *RiskMatrixRequest) (*RiskMatrixResponse, error)
	PostObsolescenseRiskMatrix(context.Context, *PostRiskMatrixRequest) (*PostRiskMatrixResponse, error)
	DropObscolenscenceData(context.Context, *DropObscolenscenceDataRequest) (*DropObscolenscenceDataResponse, error)
	GetEquipmentsByApplication(context.Context, *GetEquipmentsByApplicationRequest) (*GetEquipmentsByApplicationResponse, error)
}

// UnimplementedApplicationServiceServer should be embedded to have forward compatible implementations.
type UnimplementedApplicationServiceServer struct {
}

func (UnimplementedApplicationServiceServer) UpsertApplication(context.Context, *UpsertApplicationRequest) (*UpsertApplicationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpsertApplication not implemented")
}
func (UnimplementedApplicationServiceServer) UpsertApplicationEquip(context.Context, *UpsertApplicationEquipRequest) (*UpsertApplicationEquipResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpsertApplicationEquip not implemented")
}
func (UnimplementedApplicationServiceServer) DropApplicationData(context.Context, *DropApplicationDataRequest) (*DropApplicationDataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DropApplicationData not implemented")
}
func (UnimplementedApplicationServiceServer) DeleteApplication(context.Context, *DeleteApplicationRequest) (*DeleteApplicationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteApplication not implemented")
}
func (UnimplementedApplicationServiceServer) UpsertInstance(context.Context, *UpsertInstanceRequest) (*UpsertInstanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpsertInstance not implemented")
}
func (UnimplementedApplicationServiceServer) DeleteInstance(context.Context, *DeleteInstanceRequest) (*DeleteInstanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteInstance not implemented")
}
func (UnimplementedApplicationServiceServer) ListApplications(context.Context, *ListApplicationsRequest) (*ListApplicationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListApplications not implemented")
}
func (UnimplementedApplicationServiceServer) ListInstances(context.Context, *ListInstancesRequest) (*ListInstancesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListInstances not implemented")
}
func (UnimplementedApplicationServiceServer) ApplicationDomains(context.Context, *ApplicationDomainsRequest) (*ApplicationDomainsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApplicationDomains not implemented")
}
func (UnimplementedApplicationServiceServer) ObsolescenceDomainCriticityMeta(context.Context, *DomainCriticityMetaRequest) (*DomainCriticityMetaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ObsolescenceDomainCriticityMeta not implemented")
}
func (UnimplementedApplicationServiceServer) ObsolescenceMaintenanceCriticityMeta(context.Context, *MaintenanceCriticityMetaRequest) (*MaintenanceCriticityMetaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ObsolescenceMaintenanceCriticityMeta not implemented")
}
func (UnimplementedApplicationServiceServer) ObsolescenceRiskMeta(context.Context, *RiskMetaRequest) (*RiskMetaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ObsolescenceRiskMeta not implemented")
}
func (UnimplementedApplicationServiceServer) ObsolescenceDomainCriticity(context.Context, *DomainCriticityRequest) (*DomainCriticityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ObsolescenceDomainCriticity not implemented")
}
func (UnimplementedApplicationServiceServer) PostObsolescenceDomainCriticity(context.Context, *PostDomainCriticityRequest) (*PostDomainCriticityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostObsolescenceDomainCriticity not implemented")
}
func (UnimplementedApplicationServiceServer) ObsolescenseMaintenanceCriticity(context.Context, *MaintenanceCriticityRequest) (*MaintenanceCriticityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ObsolescenseMaintenanceCriticity not implemented")
}
func (UnimplementedApplicationServiceServer) PostObsolescenseMaintenanceCriticity(context.Context, *PostMaintenanceCriticityRequest) (*PostMaintenanceCriticityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostObsolescenseMaintenanceCriticity not implemented")
}
func (UnimplementedApplicationServiceServer) ObsolescenseRiskMatrix(context.Context, *RiskMatrixRequest) (*RiskMatrixResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ObsolescenseRiskMatrix not implemented")
}
func (UnimplementedApplicationServiceServer) PostObsolescenseRiskMatrix(context.Context, *PostRiskMatrixRequest) (*PostRiskMatrixResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostObsolescenseRiskMatrix not implemented")
}
func (UnimplementedApplicationServiceServer) DropObscolenscenceData(context.Context, *DropObscolenscenceDataRequest) (*DropObscolenscenceDataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DropObscolenscenceData not implemented")
}
func (UnimplementedApplicationServiceServer) GetEquipmentsByApplication(context.Context, *GetEquipmentsByApplicationRequest) (*GetEquipmentsByApplicationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetEquipmentsByApplication not implemented")
}

// UnsafeApplicationServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ApplicationServiceServer will
// result in compilation errors.
type UnsafeApplicationServiceServer interface {
	mustEmbedUnimplementedApplicationServiceServer()
}

func RegisterApplicationServiceServer(s grpc.ServiceRegistrar, srv ApplicationServiceServer) {
	s.RegisterService(&_ApplicationService_serviceDesc, srv)
}

func _ApplicationService_UpsertApplication_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpsertApplicationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).UpsertApplication(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/UpsertApplication",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).UpsertApplication(ctx, req.(*UpsertApplicationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_UpsertApplicationEquip_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpsertApplicationEquipRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).UpsertApplicationEquip(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/UpsertApplicationEquip",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).UpsertApplicationEquip(ctx, req.(*UpsertApplicationEquipRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_DropApplicationData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DropApplicationDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).DropApplicationData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/DropApplicationData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).DropApplicationData(ctx, req.(*DropApplicationDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_DeleteApplication_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteApplicationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).DeleteApplication(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/DeleteApplication",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).DeleteApplication(ctx, req.(*DeleteApplicationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_UpsertInstance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpsertInstanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).UpsertInstance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/UpsertInstance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).UpsertInstance(ctx, req.(*UpsertInstanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_DeleteInstance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteInstanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).DeleteInstance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/DeleteInstance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).DeleteInstance(ctx, req.(*DeleteInstanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_ListApplications_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListApplicationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).ListApplications(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/ListApplications",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).ListApplications(ctx, req.(*ListApplicationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_ListInstances_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListInstancesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).ListInstances(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/ListInstances",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).ListInstances(ctx, req.(*ListInstancesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_ApplicationDomains_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ApplicationDomainsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).ApplicationDomains(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/ApplicationDomains",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).ApplicationDomains(ctx, req.(*ApplicationDomainsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_ObsolescenceDomainCriticityMeta_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DomainCriticityMetaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).ObsolescenceDomainCriticityMeta(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/ObsolescenceDomainCriticityMeta",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).ObsolescenceDomainCriticityMeta(ctx, req.(*DomainCriticityMetaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_ObsolescenceMaintenanceCriticityMeta_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MaintenanceCriticityMetaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).ObsolescenceMaintenanceCriticityMeta(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/ObsolescenceMaintenanceCriticityMeta",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).ObsolescenceMaintenanceCriticityMeta(ctx, req.(*MaintenanceCriticityMetaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_ObsolescenceRiskMeta_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RiskMetaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).ObsolescenceRiskMeta(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/ObsolescenceRiskMeta",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).ObsolescenceRiskMeta(ctx, req.(*RiskMetaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_ObsolescenceDomainCriticity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DomainCriticityRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).ObsolescenceDomainCriticity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/ObsolescenceDomainCriticity",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).ObsolescenceDomainCriticity(ctx, req.(*DomainCriticityRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_PostObsolescenceDomainCriticity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostDomainCriticityRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).PostObsolescenceDomainCriticity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/PostObsolescenceDomainCriticity",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).PostObsolescenceDomainCriticity(ctx, req.(*PostDomainCriticityRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_ObsolescenseMaintenanceCriticity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MaintenanceCriticityRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).ObsolescenseMaintenanceCriticity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/ObsolescenseMaintenanceCriticity",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).ObsolescenseMaintenanceCriticity(ctx, req.(*MaintenanceCriticityRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_PostObsolescenseMaintenanceCriticity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostMaintenanceCriticityRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).PostObsolescenseMaintenanceCriticity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/PostObsolescenseMaintenanceCriticity",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).PostObsolescenseMaintenanceCriticity(ctx, req.(*PostMaintenanceCriticityRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_ObsolescenseRiskMatrix_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RiskMatrixRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).ObsolescenseRiskMatrix(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/ObsolescenseRiskMatrix",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).ObsolescenseRiskMatrix(ctx, req.(*RiskMatrixRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_PostObsolescenseRiskMatrix_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostRiskMatrixRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).PostObsolescenseRiskMatrix(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/PostObsolescenseRiskMatrix",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).PostObsolescenseRiskMatrix(ctx, req.(*PostRiskMatrixRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_DropObscolenscenceData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DropObscolenscenceDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).DropObscolenscenceData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/DropObscolenscenceData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).DropObscolenscenceData(ctx, req.(*DropObscolenscenceDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplicationService_GetEquipmentsByApplication_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetEquipmentsByApplicationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationServiceServer).GetEquipmentsByApplication(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/optisam.applications.v1.ApplicationService/GetEquipmentsByApplication",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationServiceServer).GetEquipmentsByApplication(ctx, req.(*GetEquipmentsByApplicationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ApplicationService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "optisam.applications.v1.ApplicationService",
	HandlerType: (*ApplicationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UpsertApplication",
			Handler:    _ApplicationService_UpsertApplication_Handler,
		},
		{
			MethodName: "UpsertApplicationEquip",
			Handler:    _ApplicationService_UpsertApplicationEquip_Handler,
		},
		{
			MethodName: "DropApplicationData",
			Handler:    _ApplicationService_DropApplicationData_Handler,
		},
		{
			MethodName: "DeleteApplication",
			Handler:    _ApplicationService_DeleteApplication_Handler,
		},
		{
			MethodName: "UpsertInstance",
			Handler:    _ApplicationService_UpsertInstance_Handler,
		},
		{
			MethodName: "DeleteInstance",
			Handler:    _ApplicationService_DeleteInstance_Handler,
		},
		{
			MethodName: "ListApplications",
			Handler:    _ApplicationService_ListApplications_Handler,
		},
		{
			MethodName: "ListInstances",
			Handler:    _ApplicationService_ListInstances_Handler,
		},
		{
			MethodName: "ApplicationDomains",
			Handler:    _ApplicationService_ApplicationDomains_Handler,
		},
		{
			MethodName: "ObsolescenceDomainCriticityMeta",
			Handler:    _ApplicationService_ObsolescenceDomainCriticityMeta_Handler,
		},
		{
			MethodName: "ObsolescenceMaintenanceCriticityMeta",
			Handler:    _ApplicationService_ObsolescenceMaintenanceCriticityMeta_Handler,
		},
		{
			MethodName: "ObsolescenceRiskMeta",
			Handler:    _ApplicationService_ObsolescenceRiskMeta_Handler,
		},
		{
			MethodName: "ObsolescenceDomainCriticity",
			Handler:    _ApplicationService_ObsolescenceDomainCriticity_Handler,
		},
		{
			MethodName: "PostObsolescenceDomainCriticity",
			Handler:    _ApplicationService_PostObsolescenceDomainCriticity_Handler,
		},
		{
			MethodName: "ObsolescenseMaintenanceCriticity",
			Handler:    _ApplicationService_ObsolescenseMaintenanceCriticity_Handler,
		},
		{
			MethodName: "PostObsolescenseMaintenanceCriticity",
			Handler:    _ApplicationService_PostObsolescenseMaintenanceCriticity_Handler,
		},
		{
			MethodName: "ObsolescenseRiskMatrix",
			Handler:    _ApplicationService_ObsolescenseRiskMatrix_Handler,
		},
		{
			MethodName: "PostObsolescenseRiskMatrix",
			Handler:    _ApplicationService_PostObsolescenseRiskMatrix_Handler,
		},
		{
			MethodName: "DropObscolenscenceData",
			Handler:    _ApplicationService_DropObscolenscenceData_Handler,
		},
		{
			MethodName: "GetEquipmentsByApplication",
			Handler:    _ApplicationService_GetEquipmentsByApplication_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "application.proto",
}
