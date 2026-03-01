package grpc

import (
	"github.com/ElfAstAhe/go-service-template/internal/facade/dto"
	grpcdto "github.com/ElfAstAhe/go-service-template/pkg/api/grpc/example/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapTestGRPCToDto(src *grpcdto.Test) *dto.TestDTO {
	if src == nil {
		return nil
	}

	res := &dto.TestDTO{
		ID:           src.GetId(),
		Code:         src.GetCode(),
		Name:         src.GetName(),
		Description:  src.GetDescription(),
		RegisteredAt: src.GetCreatedAt().AsTime(),
		UpdatedAt:    src.GetCreatedAt().AsTime(),
	}

	return res
}

func MapTestDtoToGRPC(src *dto.TestDTO) *grpcdto.Test {
	if src == nil {
		return nil
	}

	res := grpcdto.Test_builder{
		Id:          src.ID,
		Code:        src.Code,
		Name:        src.Name,
		Description: src.Description,
		CreatedAt:   timestamppb.New(src.RegisteredAt),
		ModifiedAt:  timestamppb.New(src.UpdatedAt),
	}.Build()

	return res
}

func MapTestDtosToGRPCs(src []*dto.TestDTO) []*grpcdto.Test {
	res := make([]*grpcdto.Test, len(src))
	if len(src) == 0 {
		return res
	}

	for i, item := range src {
		res[i] = MapTestDtoToGRPC(item)
	}

	return res
}
