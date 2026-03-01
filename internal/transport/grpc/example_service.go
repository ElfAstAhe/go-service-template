package grpc

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/internal/facade"
	pb "github.com/ElfAstAhe/go-service-template/pkg/api/grpc/example/v1"
	conf "github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ExampleGRPCService struct {
	pb.UnimplementedExampleServiceServer
	testFacade facade.TestFacade
	log        logger.Logger
	config     conf.GRPCConfig
}

func NewExampleGRPCService(config conf.GRPCConfig, testFacade facade.TestFacade, logger logger.Logger) *ExampleGRPCService {
	return &ExampleGRPCService{
		testFacade: testFacade,
		log:        logger.GetLogger("example gRPC service"),
		config:     config,
	}
}

func (es *ExampleGRPCService) Find(ctx context.Context, req *pb.ExampleServiceFindRequest) (*pb.ExampleServiceInstanceResponse, error) {
	dtoRes, err := es.testFacade.Get(ctx, req.GetId())
	if err != nil {

	}

	return pb.ExampleServiceInstanceResponse_builder{
		Instance: MapTestDtoToGRPC(dtoRes),
	}.Build(), nil
}

func (es *ExampleGRPCService) FindByCode(context.Context, *pb.ExampleServiceFindByCodeRequest) (*pb.ExampleServiceInstanceResponse, error) {
	// ToDo: implement

	return nil, errs.NewNotImplementedError(nil)
}

func (es *ExampleGRPCService) List(context.Context, *pb.ExampleServiceListRequest) (*pb.ExampleServiceInstancesResponse, error) {
	// ToDo: implement

	return nil, errs.NewNotImplementedError(nil)
}

func (es *ExampleGRPCService) Save(context.Context, *pb.ExampleServiceSaveRequest) (*pb.ExampleServiceInstanceResponse, error) {
	// ToDo: implement

	return nil, errs.NewNotImplementedError(nil)
}

func (es *ExampleGRPCService) Delete(context.Context, *pb.ExampleServiceDeleteRequest) (*emptypb.Empty, error) {
	// ToDo: implement

	return nil, errs.NewNotImplementedError(nil)
}
