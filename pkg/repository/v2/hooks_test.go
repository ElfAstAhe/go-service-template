package v2

import (
	"context"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/db"
	mocks2 "github.com/ElfAstAhe/go-service-template/pkg/db/mocks"
	"github.com/ElfAstAhe/go-service-template/pkg/repository/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHooks(t *testing.T) {
	hooks := NewHooks[*dummyModel, string]()

	require.NotNil(t, hooks)
	assert.Nil(t, hooks.DispatchMapper)
	assert.Nil(t, hooks.NewEntityFactory)
}

func TestHooks_Validate(t *testing.T) {
	// Инициализируем минимальные рабочие заглушки с правильными сигнатурами
	var (
		dummyDispatchMapper DispatchMapperFunc[*dummyModel, string] = func(ctx context.Context, mapperKey string, scannable Scannable, params ...any) (*dummyModel, error) {
			return &dummyModel{}, nil
		}
		dummyNewEntityFactory NewEntityFactory[*dummyModel, string] = func() *dummyModel {
			return &dummyModel{}
		}
	)

	tests := []struct {
		name        string
		setup       func(*Hooks[*dummyModel, string])
		expectedErr string
	}{
		{
			name: "error when DispatchMapper is nil",
			setup: func(h *Hooks[*dummyModel, string]) {
				h.NewEntityFactory = dummyNewEntityFactory
			},
			expectedErr: "DispatchMapper required",
		},
		{
			name: "error when NewEntityFactory is nil",
			setup: func(h *Hooks[*dummyModel, string]) {
				h.DispatchMapper = dummyDispatchMapper
			},
			expectedErr: "NewEntityFactory required",
		},
		{
			name: "success when all required fields are present",
			setup: func(h *Hooks[*dummyModel, string]) {
				h.DispatchMapper = dummyDispatchMapper
				h.NewEntityFactory = dummyNewEntityFactory
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hooks := NewHooks[*dummyModel, string]()
			tt.setup(hooks)

			err := hooks.Validate()

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHooks_FluentAPI(t *testing.T) {
	t.Run("should correctly apply all hook functions and execute them", func(t *testing.T) {
		hooks := NewHooks[*dummyModel, string]()
		ctx := context.Background()
		model := &dummyModel{ID: "test-id"}

		// Инициализируем мок Scannable, сгенерированный с помощью mockery
		mockScannable := mocks.NewMockScannable(t)
		mockQuerier := mocks2.NewMockQuerier(t)

		// Настраиваем строго типизированные заглушки для проверки вызовов Fluent API
		fnDispatchMapper := func(ctx context.Context, mapperKey string, scannable Scannable, params ...any) (*dummyModel, error) {
			return model, nil
		}
		fnNewEntityFactory := func() *dummyModel {
			return model
		}
		fnValidateID := func(ctx context.Context, id string) error {
			return nil
		}
		fnAfterFind := func(ctx context.Context, entity *dummyModel, params ...any) (*dummyModel, error) {
			return entity, nil
		}
		fnListFilterYield := func(ctx context.Context, entity *dummyModel, params ...any) (*dummyModel, YieldAction, error) {
			return entity, ListYieldActionAppend, nil
		}
		fnValidate := func(ctx context.Context, entity *dummyModel, params ...any) error {
			return nil
		}
		fnBefore := func(ctx context.Context, entity *dummyModel, params ...any) (*dummyModel, error) {
			return entity, nil
		}
		fnModificator := func(ctx context.Context, querier db.Querier, entity *dummyModel, params ...any) (Scannable, error) {
			return mockScannable, nil
		}

		// Применяем хуки через Fluent API цепочку
		withHookDispatchMapper[*dummyModel, string](fnDispatchMapper)(hooks)
		WithHookNewEntityFactory[*dummyModel, string](fnNewEntityFactory)(hooks)
		WithValidateFind[*dummyModel, string](fnValidateID)(hooks)
		WithHookAfterFind[*dummyModel, string](fnAfterFind)(hooks)
		WithHookListFilterYield[*dummyModel, string](fnListFilterYield)(hooks)
		WithHookValidateCreate[*dummyModel, string](fnValidate)(hooks)
		WithHookBeforeCreate[*dummyModel, string](fnBefore)(hooks)
		WithHookCreator[*dummyModel, string](fnModificator)(hooks)
		WithHookValidateChange[*dummyModel, string](fnValidate)(hooks)
		WithHookBeforeChange[*dummyModel, string](fnBefore)(hooks)
		WithHookChanger[*dummyModel, string](fnModificator)(hooks)
		WithHookValidateDelete[*dummyModel, string](fnValidateID)(hooks)

		// 1. Проверяем, что все поля успешно заполнились
		assert.NotNil(t, hooks.DispatchMapper)
		assert.NotNil(t, hooks.NewEntityFactory)
		assert.NotNil(t, hooks.ValidateFind)
		assert.NotNil(t, hooks.AfterFind)
		assert.NotNil(t, hooks.ListFilterYield)
		assert.NotNil(t, hooks.ValidateCreate)
		assert.NotNil(t, hooks.BeforeCreate)
		assert.NotNil(t, hooks.Creator)
		assert.NotNil(t, hooks.ValidateChange)
		assert.NotNil(t, hooks.BeforeChange)
		assert.NotNil(t, hooks.Changer)
		assert.NotNil(t, hooks.ValidateDelete)

		// 2. Делаем тестовые вызовы, проверяя корректность проброса типов в рантайме
		resEntity, err := hooks.DispatchMapper(ctx, "key", mockScannable)
		assert.NoError(t, err)
		assert.Equal(t, model, resEntity)

		resScannable, err := hooks.Creator(ctx, mockQuerier, model)
		assert.NoError(t, err)
		assert.Equal(t, mockScannable, resScannable)

		actionEntity, action, err := hooks.ListFilterYield(ctx, model)
		assert.NoError(t, err)
		assert.Equal(t, ListYieldActionAppend, action)
		assert.Equal(t, model, actionEntity)
	})
}
