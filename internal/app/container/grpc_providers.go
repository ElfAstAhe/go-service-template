package container

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/internal/facade"
	appgrpc "github.com/ElfAstAhe/go-service-template/internal/transport/grpc"
	pb "github.com/ElfAstAhe/go-service-template/pkg/api/grpc/example/v1/example-service/v1"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkggrpc "github.com/ElfAstAhe/go-service-template/pkg/transport/grpc"
	"google.golang.org/grpc"
)

func (gc *GRPCContainer) serviceRegister(server *grpc.Server) error {
	testSvcInst, err := container.GetInstance[*appgrpc.TestGRPCService](InstanceTestGRPCService)
	if err != nil {
		return errs.NewContainerError(gc.GetName(), "service register: retrieve instance failed", err)
	}
	test2SvcInst, err := container.GetInstance[*appgrpc.Test2GRPCService](InstanceTest2GRPCService)
	if err != nil {
		return errs.NewContainerError(gc.GetName(), "service register: retrieve instance failed", err)
	}

	pb.RegisterTestServiceServer(server, testSvcInst)
	pb.RegisterTest2ServiceServer(server, test2SvcInst)

	return nil
}

//goland:noinspection DuplicatedCode
func (gc *GRPCContainer) providerTestGRPCService() (any, error) {
	confInst, err := container.GetInstance[*config.Config](InstanceConfig)
	if err != nil {
		return nil, errs.NewContainerError(gc.GetName(), "provider: retrieve instance failed", err)
	}
	logInst, err := container.GetInstance[logger.Logger](InstanceLogger)
	if err != nil {
		return nil, errs.NewContainerError(gc.GetName(), "provider: retrieve instance failed", err)
	}
	testFacadeInst, err := container.GetInstance[facade.Test](InstanceTestFacade)
	if err != nil {
		return nil, errs.NewContainerError(gc.GetName(), "provider: retrieve instance failed", err)
	}

	return appgrpc.NewTestGRPCService(confInst.GRPC, testFacadeInst, logInst), nil
}

//goland:noinspection DuplicatedCode
func (gc *GRPCContainer) providerTest2GRPCService() (any, error) {
	confInst, err := container.GetInstance[*config.Config](InstanceConfig)
	if err != nil {
		return nil, errs.NewContainerError(gc.GetName(), "provider: retrieve instance failed", err)
	}
	logInst, err := container.GetInstance[logger.Logger](InstanceLogger)
	if err != nil {
		return nil, errs.NewContainerError(gc.GetName(), "provider: retrieve instance failed", err)
	}
	testFacadeInst, err := container.GetInstance[facade.Test2Facade](InstanceTest2Facade)
	if err != nil {
		return nil, errs.NewContainerError(gc.GetName(), "provider: retrieve instance failed", err)
	}

	return appgrpc.NewTest2GRPCService(confInst.GRPC, testFacadeInst, logInst), nil
}

//goland:noinspection DuplicatedCode
func (gc *GRPCContainer) providerGRPCRunner() (any, error) {
	confInst, err := container.GetInstance[*config.Config](InstanceConfig)
	if err != nil {
		return nil, errs.NewContainerError(gc.GetName(), "provider: retrieve instance failed", err)
	}
	logInst, err := container.GetInstance[logger.Logger](InstanceLogger)
	if err != nil {
		return nil, errs.NewContainerError(gc.GetName(), "provider: retrieve instance failed", err)
	}

	runner, err := pkggrpc.NewRunner(
		pkggrpc.WithName("main-grpc-server"),
		pkggrpc.WithConfig(confInst.GRPC),
		pkggrpc.WithLogger("grpc-server", logInst),
		pkggrpc.WithServiceRegister(gc.serviceRegister),
	)
	if err != nil {
		return nil, errs.NewContainerError(gc.GetName(), fmt.Sprintf("provider: create %s failed", InstanceGRPCRunner), err)
	}

	return runner, nil
}
