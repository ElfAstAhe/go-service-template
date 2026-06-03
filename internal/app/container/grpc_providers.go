package container

import (
	"fmt"

	"github.com/ElfAstAhe/go-service-template/internal/config"
	"github.com/ElfAstAhe/go-service-template/internal/facade"
	grpcsvc "github.com/ElfAstAhe/go-service-template/internal/transport/grpc"
	pb "github.com/ElfAstAhe/go-service-template/pkg/api/grpc/example/v1"
	"github.com/ElfAstAhe/go-service-template/pkg/container"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/grpc"

	libgrpc "google.golang.org/grpc"
)

func (gc *GRPCContainer) serviceRegister(server *libgrpc.Server) error {
	serviceInst, err := container.GetInstance[*grpcsvc.ExampleGRPCService](InstanceGRPCService)
	if err != nil {
		return errs.NewContainerError(gc.GetName(), "service register: retrieve instance failed", err)
	}

	pb.RegisterExampleServiceServer(server, serviceInst)

	return nil
}

//goland:noinspection DuplicatedCode
func (gc *GRPCContainer) providerGRPCService() (any, error) {
	confInst, err := container.GetInstance[*config.Config](InstanceConfig)
	if err != nil {
		return nil, errs.NewContainerError(gc.GetName(), "provider: retrieve instance failed", err)
	}
	logInst, err := container.GetInstance[logger.Logger](InstanceLogger)
	if err != nil {
		return nil, errs.NewContainerError(gc.GetName(), "provider: retrieve instance failed", err)
	}
	testFacadeInst, err := container.GetInstance[facade.TestFacade](InstanceTestFacade)
	if err != nil {
		return nil, errs.NewContainerError(gc.GetName(), "provider: retrieve instance failed", err)
	}

	return grpcsvc.NewExampleGRPCService(confInst.GRPC, testFacadeInst, logInst), nil
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

	runner, err := grpc.NewRunner(
		grpc.WithName("main-grpc-server"),
		grpc.WithConfig(confInst.GRPC),
		grpc.WithLogger("grpc_server", logInst),
		grpc.WithServiceRegister(gc.serviceRegister),
	)
	if err != nil {
		return nil, errs.NewContainerError(gc.GetName(), fmt.Sprintf("provider: create %s failed", InstanceGRPCRunner), err)
	}

	return runner, nil
}
