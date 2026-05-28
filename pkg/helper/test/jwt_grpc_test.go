package test

import (
	"context"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestJWTGRPCHelper(t *testing.T) {
	// Базовый хелпер (можно реальный, так как он простой, или мок)
	baseHelper := helper.NewDefaultJWTHelper("secret")
	grpcHelper := helper.NewJWTGRPCHelper(baseHelper)
	mdName := "authorization"

	t.Run("ExtractString: успешно извлекает с Bearer", func(t *testing.T) {
		md := metadata.Pairs(mdName, helper.TokenPrefix+"my-token")

		res, err := grpcHelper.ExtractTokenStringFromMetadata(mdName, md)

		assert.NoError(t, err)
		assert.Equal(t, "my-token", res)
	})

	t.Run("ExtractString: успешно извлекает БЕЗ Bearer", func(t *testing.T) {
		md := metadata.Pairs(mdName, "raw-token")

		res, err := grpcHelper.ExtractTokenStringFromMetadata(mdName, md)

		assert.NoError(t, err)
		assert.Equal(t, "raw-token", res)
	})

	t.Run("FromContext: ошибка если метаданных нет", func(t *testing.T) {
		ctx := context.Background() // Пустой контекст без MD

		_, err := grpcHelper.ExtractTokenStringFromContext(mdName, ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no metadata")
	})

	t.Run("Full Cycle: извлечение из контекста через MD", func(t *testing.T) {
		tokenStr, _ := baseHelper.BuildTokenString("1", "sub", "type", false)
		md := metadata.Pairs(mdName, helper.TokenPrefix+tokenStr)

		// Помещаем MD в контекст (имитация входящего вызова)
		ctx := metadata.NewIncomingContext(context.Background(), md)

		token, err := grpcHelper.ExtractTokenFromContext(mdName, ctx)

		require.NoError(t, err)
		assert.True(t, token.Valid)
	})
}
