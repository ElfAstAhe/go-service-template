package grpc

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/internal/facade"
	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	pb "github.com/ElfAstAhe/go-service-template/pkg/api/grpc/example/v1"
	conf "github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ExampleGRPCService struct {
	pb.UnimplementedExampleServiceServer
	testFacade facade.TestFacade
	log        logger.Logger
	config     *conf.GRPCConfig
}

func NewExampleGRPCService(config *conf.GRPCConfig, testFacade facade.TestFacade, logger logger.Logger) *ExampleGRPCService {
	return &ExampleGRPCService{
		testFacade: testFacade,
		log:        logger.GetLogger("example gRPC service"),
		config:     config,
	}
}

func (es *ExampleGRPCService) Find(ctx context.Context, req *pb.ExampleServiceFindRequest) (*pb.ExampleServiceInstanceResponse, error) {
	dtoRes, err := es.testFacade.Get(ctx, req.GetId())
	if err != nil {
		return nil, MapToGrpcError(err)
	}

	return pb.ExampleServiceInstanceResponse_builder{
		Instance: MapTestDtoToGRPC(dtoRes),
	}.Build(), nil
}

func (es *ExampleGRPCService) FindByCode(ctx context.Context, req *pb.ExampleServiceFindByCodeRequest) (*pb.ExampleServiceInstanceResponse, error) {
	dtoRes, err := es.testFacade.GetByCode(ctx, req.GetCode())
	if err != nil {
		return nil, MapToGrpcError(err)
	}

	return pb.ExampleServiceInstanceResponse_builder{
		Instance: MapTestDtoToGRPC(dtoRes),
	}.Build(), nil
}

func (es *ExampleGRPCService) List(ctx context.Context, req *pb.ExampleServiceListRequest) (*pb.ExampleServiceInstancesResponse, error) {
	dtosRes, err := es.testFacade.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, MapToGrpcError(err)
	}

	return pb.ExampleServiceInstancesResponse_builder{
		Data: MapTestDtosToGRPCs(dtosRes),
	}.Build(), nil
}

func (es *ExampleGRPCService) Save(ctx context.Context, req *pb.ExampleServiceSaveRequest) (*pb.ExampleServiceInstanceResponse, error) {
	income := MapTestGRPCToDto(req.GetInstance())
	var dtoRes *dto.TestDTO
	var err error
	if income.ID == "" {
		dtoRes, err = es.testFacade.Create(ctx, income)
	} else {
		dtoRes, err = es.testFacade.Change(ctx, income.ID, income)
	}
	if err != nil {
		return nil, MapToGrpcError(err)
	}

	return pb.ExampleServiceInstanceResponse_builder{
		Instance: MapTestDtoToGRPC(dtoRes),
	}.Build(), nil
}

func (es *ExampleGRPCService) Delete(ctx context.Context, req *pb.ExampleServiceDeleteRequest) (*emptypb.Empty, error) {
	err := es.testFacade.Delete(ctx, req.GetId())
	if err != nil {
		return nil, MapToGrpcError(err)
	}

	return &emptypb.Empty{}, nil
}
