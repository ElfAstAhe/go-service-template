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

type Test2GRPCService struct {
	pb.UnimplementedTest2ServiceServer
	facade facade.Test2
	log    logger.Logger
	config *conf.GRPCConfig
}

var _ pb.Test2ServiceServer = (*Test2GRPCService)(nil)

func NewTest2GRPCService(config *conf.GRPCConfig, test2Facade facade.Test2, logger logger.Logger) *Test2GRPCService {
	return &Test2GRPCService{
		facade: test2Facade,
		log:    logger.GetLogger("example gRPC service"),
		config: config,
	}
}

func (t2s *Test2GRPCService) Find(ctx context.Context, req *pb.Test2ServiceFindRequest) (*pb.Test2ServiceInstanceResponse, error) {
	dtoRes, err := t2s.facade.Get(ctx, req.GetId())
	if err != nil {
		return nil, grpc.MapToGrpcError(err)
	}

	return pb.Test2ServiceInstanceResponse_builder{
		Instance: MapTest2DtoToGRPC(dtoRes),
	}.Build(), nil
}

func (t2s *Test2GRPCService) FindByCode(ctx context.Context, req *pb.Test2ServiceFindByCodeRequest) (*pb.Test2ServiceInstanceResponse, error) {
	dtoRes, err := t2s.facade.GetByCode(ctx, req.GetCode())
	if err != nil {
		return nil, grpc.MapToGrpcError(err)
	}

	return pb.Test2ServiceInstanceResponse_builder{
		Instance: MapTest2DtoToGRPC(dtoRes),
	}.Build(), nil
}

func (t2s *Test2GRPCService) List(ctx context.Context, req *pb.Test2ServiceListRequest) (*pb.Test2ServiceInstancesResponse, error) {
	dtosRes, err := t2s.facade.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, grpc.MapToGrpcError(err)
	}

	return pb.Test2ServiceInstancesResponse_builder{
		Data: MapTest2DtosToGRPCs(dtosRes),
	}.Build(), nil
}

func (t2s *Test2GRPCService) Save(ctx context.Context, req *pb.Test2ServiceSaveRequest) (*pb.Test2ServiceInstanceResponse, error) {
	income := MapTest2GRPCToDto(req.GetInstance())
	var dtoRes *dto.Test2
	var err error
	if income.ID == "" {
		dtoRes, err = t2s.facade.Create(ctx, income)
	} else {
		dtoRes, err = t2s.facade.Change(ctx, income.ID, income)
	}
	if err != nil {
		return nil, grpc.MapToGrpcError(err)
	}

	return pb.Test2ServiceInstanceResponse_builder{
		Instance: MapTest2DtoToGRPC(dtoRes),
	}.Build(), nil
}

func (t2s *Test2GRPCService) Delete(ctx context.Context, req *pb.Test2ServiceDeleteRequest) (*emptypb.Empty, error) {
	err := t2s.facade.Delete(ctx, req.GetId())
	if err != nil {
		return nil, grpc.MapToGrpcError(err)
	}

	return &emptypb.Empty{}, nil
}
