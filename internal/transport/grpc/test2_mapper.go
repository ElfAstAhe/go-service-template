package grpc

import (
	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	grpcdto "github.com/ElfAstAhe/go-service-template/pkg/api/grpc/example/v1/example-service/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapTest2GRPCToDto(src *grpcdto.Test2) *dto.Test2 {
	if src == nil {
		return nil
	}

	res := &dto.Test2{
		ID:          src.GetId(),
		Code:        src.GetCode(),
		Name:        src.GetName(),
		Description: src.GetDescription(),
		CreatedAt:   src.GetCreatedAt().AsTime(),
		ModifiedAt:  src.GetCreatedAt().AsTime(),
	}

	return res
}

func MapTest2DtoToGRPC(src *dto.Test2) *grpcdto.Test2 {
	if src == nil {
		return nil
	}

	res := grpcdto.Test2_builder{
		Id:          src.ID,
		Code:        src.Code,
		Name:        src.Name,
		Description: src.Description,
		CreatedAt:   timestamppb.New(src.CreatedAt),
		ModifiedAt:  timestamppb.New(src.ModifiedAt),
	}.Build()

	return res
}

func MapTest2DtosToGRPCs(src []*dto.Test2) []*grpcdto.Test2 {
	res := make([]*grpcdto.Test2, len(src))
	if len(src) == 0 {
		return res
	}

	for i, item := range src {
		res[i] = MapTest2DtoToGRPC(item)
	}

	return res
}
