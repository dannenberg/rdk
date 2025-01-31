package registry

import (
	"context"
	"testing"

	"github.com/edaniels/golog"
	"github.com/jhump/protoreflect/grpcreflect"
	pb "go.viam.com/api/robot/v1"
	"go.viam.com/test"
	"go.viam.com/utils/rpc"

	"go.viam.com/rdk/config"
	"go.viam.com/rdk/discovery"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/subtype"
)

var (
	button      = resource.SubtypeName("button")
	acme        = resource.NewName(resource.Namespace("acme"), resource.ResourceTypeComponent, button, "button1")
	nav         = resource.SubtypeName("navigation")
	testService = resource.NewName(resource.Namespace("rdk"), resource.ResourceTypeComponent, nav, "nav1")
)

func TestComponentRegistry(t *testing.T) {
	rf := func(ctx context.Context, deps Dependencies, config config.Component, logger golog.Logger) (interface{}, error) {
		return 1, nil
	}
	modelName := "x"
	test.That(t, func() { RegisterComponent(acme.Subtype, modelName, Component{}) }, test.ShouldPanic)
	RegisterComponent(acme.Subtype, modelName, Component{Constructor: rf})

	creator := ComponentLookup(acme.Subtype, modelName)
	test.That(t, creator, test.ShouldNotBeNil)
	test.That(t, ComponentLookup(acme.Subtype, "z"), test.ShouldBeNil)
	test.That(t, creator.Constructor, test.ShouldEqual, rf)
}

func TestResourceSubtypeRegistry(t *testing.T) {
	rf := func(r interface{}, n resource.Name) (resource.Reconfigurable, error) {
		return nil, nil
	}
	statf := func(context.Context, interface{}) (interface{}, error) {
		return map[string]interface{}{}, nil
	}
	sf := func(ctx context.Context, rpcServer rpc.Server, subtypeSvc subtype.Service) error {
		return nil
	}
	rcf := func(ctx context.Context, conn rpc.ClientConn, name string, logger golog.Logger) interface{} {
		return nil
	}
	test.That(t, func() { RegisterResourceSubtype(acme.Subtype, ResourceSubtype{}) }, test.ShouldPanic)

	RegisterResourceSubtype(acme.Subtype, ResourceSubtype{
		Reconfigurable: rf, Status: statf, RegisterRPCService: sf, RPCServiceDesc: &pb.RobotService_ServiceDesc,
	})
	creator := ResourceSubtypeLookup(acme.Subtype)
	test.That(t, creator, test.ShouldNotBeNil)
	test.That(t, creator.Reconfigurable, test.ShouldEqual, rf)
	test.That(t, creator.Status, test.ShouldEqual, statf)
	test.That(t, creator.RegisterRPCService, test.ShouldEqual, sf)
	test.That(t, creator.RPCClient, test.ShouldBeNil)

	subtype2 := resource.NewSubtype(resource.Namespace("acme2"), resource.ResourceTypeComponent, button)
	test.That(t, ResourceSubtypeLookup(subtype2), test.ShouldBeNil)

	RegisterResourceSubtype(subtype2, ResourceSubtype{
		RegisterRPCService: sf, RPCClient: rcf, RPCServiceDesc: &pb.RobotService_ServiceDesc,
	})
	creator = ResourceSubtypeLookup(subtype2)
	test.That(t, creator, test.ShouldNotBeNil)
	test.That(t, creator.Status, test.ShouldBeNil)
	test.That(t, creator.RegisterRPCService, test.ShouldEqual, sf)
	test.That(t, creator.RPCClient, test.ShouldEqual, rcf)
	test.That(t, creator.RPCServiceDesc, test.ShouldEqual, &pb.RobotService_ServiceDesc)

	reflectSvcDesc, err := grpcreflect.LoadServiceDescriptor(creator.RPCServiceDesc)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, creator.ReflectRPCServiceDesc, test.ShouldResemble, reflectSvcDesc)

	subtype3 := resource.NewSubtype(resource.Namespace("acme3"), resource.ResourceTypeComponent, button)
	test.That(t, ResourceSubtypeLookup(subtype3), test.ShouldBeNil)

	RegisterResourceSubtype(subtype3, ResourceSubtype{RPCClient: rcf})
	creator = ResourceSubtypeLookup(subtype3)
	test.That(t, creator, test.ShouldNotBeNil)
	test.That(t, creator.RPCClient, test.ShouldEqual, rcf)

	subtype4 := resource.NewSubtype(resource.Namespace("acme4"), resource.ResourceTypeComponent, button)
	test.That(t, ResourceSubtypeLookup(subtype4), test.ShouldBeNil)
	test.That(t, func() {
		RegisterResourceSubtype(subtype4, ResourceSubtype{
			RegisterRPCService: sf, RPCClient: rcf,
		})
	}, test.ShouldPanic)
}

func TestDiscoveryFunctionRegistry(t *testing.T) {
	df := func(ctx context.Context) (interface{}, error) { return []discovery.Discovery{}, nil }
	invalidSubtypeQuery := discovery.NewQuery(resource.SubtypeName("some subtype"), "some model")
	test.That(t, func() { RegisterDiscoveryFunction(invalidSubtypeQuery, df) }, test.ShouldPanic)

	validSubtypeQuery := discovery.NewQuery(acme.ResourceSubtype, "some model")
	_, ok := DiscoveryFunctionLookup(validSubtypeQuery)
	test.That(t, ok, test.ShouldBeFalse)

	test.That(t, func() { RegisterDiscoveryFunction(validSubtypeQuery, df) }, test.ShouldNotPanic)
	acmeDF, ok := DiscoveryFunctionLookup(validSubtypeQuery)
	test.That(t, ok, test.ShouldBeTrue)
	test.That(t, acmeDF, test.ShouldEqual, df)
}

func TestServiceRegistry(t *testing.T) {
	rf := func(ctx context.Context, deps Dependencies, config config.Service, logger golog.Logger) (interface{}, error) {
		return 1, nil
	}
	modelName := resource.DefaultModelName
	test.That(t, func() { RegisterService(testService.Subtype, modelName, Service{}) }, test.ShouldPanic)
	RegisterService(testService.Subtype, modelName, Service{Constructor: rf})

	creator := ServiceLookup(testService.Subtype, modelName)
	test.That(t, creator, test.ShouldNotBeNil)
	test.That(t, ServiceLookup(testService.Subtype, "z"), test.ShouldBeNil)
	test.That(t, creator.Constructor, test.ShouldEqual, rf)
}

func TestFindValidServiceModels(t *testing.T) {
	rf := func(ctx context.Context, deps Dependencies, config config.Service, logger golog.Logger) (interface{}, error) {
		return 1, nil
	}
	RegisterService(testService.Subtype, "testModel1", Service{Constructor: rf})
	RegisterService(testService.Subtype, "testModel2", Service{Constructor: rf})
	modelList := FindValidServiceModels(testService)
	test.That(t, modelList, test.ShouldContain, "testModel1")
	test.That(t, modelList, test.ShouldContain, "testModel2")
}
