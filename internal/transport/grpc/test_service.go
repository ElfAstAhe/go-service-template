package grpc

import (
	"context"

	"github.com/ElfAstAhe/go-service-template/internal/facade"
	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	pb "github.com/ElfAstAhe/go-service-template/pkg/api/grpc/example/v1/example-service/v1"
	conf "github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	"github.com/ElfAstAhe/go-service-template/pkg/transport/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TestGRPCService struct {
	pb.UnimplementedTestServiceServer
	testFacade facade.Test
	log        logger.Logger
	config     *conf.GRPCConfig
}

var _ pb.TestServiceServer = (*TestGRPCService)(nil)

func NewTestGRPCService(config *conf.GRPCConfig, testFacade facade.Test, logger logger.Logger) *TestGRPCService {
	return &TestGRPCService{
		testFacade: testFacade,
		log:        logger.GetLogger("example gRPC service"),
		config:     config,
	}
}

func (es *TestGRPCService) Find(ctx context.Context, req *pb.TestServiceFindRequest) (*pb.TestServiceInstanceResponse, error) {
	dtoRes, err := es.testFacade.Get(ctx, req.GetId())
	if err != nil {
		return nil, grpc.MapToGrpcError(err)
	}

	return pb.TestServiceInstanceResponse_builder{
		Instance: MapTestDtoToGRPC(dtoRes),
	}.Build(), nil
}

func (es *TestGRPCService) FindByCode(ctx context.Context, req *pb.TestServiceFindByCodeRequest) (*pb.TestServiceInstanceResponse, error) {
	dtoRes, err := es.testFacade.GetByCode(ctx, req.GetCode())
	if err != nil {
		return nil, grpc.MapToGrpcError(err)
	}

	return pb.TestServiceInstanceResponse_builder{
		Instance: MapTestDtoToGRPC(dtoRes),
	}.Build(), nil
}

func (es *TestGRPCService) List(ctx context.Context, req *pb.TestServiceListRequest) (*pb.TestServiceInstancesResponse, error) {
	dtosRes, err := es.testFacade.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, grpc.MapToGrpcError(err)
	}

	return pb.TestServiceInstancesResponse_builder{
		Data: MapTestDtosToGRPCs(dtosRes),
	}.Build(), nil
}

func (es *TestGRPCService) Save(ctx context.Context, req *pb.TestServiceSaveRequest) (*pb.TestServiceInstanceResponse, error) {
	income := MapTestGRPCToDto(req.GetInstance())
	var dtoRes *dto.Test
	var err error
	if income.ID == "" {
		dtoRes, err = es.testFacade.Create(ctx, income)
	} else {
		dtoRes, err = es.testFacade.Change(ctx, income.ID, income)
	}
	if err != nil {
		return nil, grpc.MapToGrpcError(err)
	}

	return pb.TestServiceInstanceResponse_builder{
		Instance: MapTestDtoToGRPC(dtoRes),
	}.Build(), nil
}

func (es *TestGRPCService) Delete(ctx context.Context, req *pb.TestServiceDeleteRequest) (*emptypb.Empty, error) {
	err := es.testFacade.Delete(ctx, req.GetId())
	if err != nil {
		return nil, grpc.MapToGrpcError(err)
	}

	return &emptypb.Empty{}, nil
}
