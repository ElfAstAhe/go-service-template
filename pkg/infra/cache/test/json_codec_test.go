package test

import (
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestManager_Codecs_Integration(t *testing.T) {
	// Фабрика для тестов
	userFactory := func() TestUser { return TestUser{} }

	// Список кодеков для прогона тестов в цикле
	testCases := []struct {
		name  string
		codec cache.Codec[TestUser]
	}{
		{
			name:  "JSON_Codec",
			codec: cache.NewJSONCodec[TestUser](userFactory),
		},
		{
			name:  "Gob_Codec",
			codec: cache.NewGobCodec[TestUser](userFactory),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mStorage := mocks.NewMockStorage[string](t)
			mgr := cache.New[string, TestUser](mStorage, tc.codec, 0)

			key := "user:42"
			val := TestUser{ID: 42, Name: "Test"}

			// 1. Тестируем Set (проверяем, что кодек выдал хоть какие-то байты)
			var captured []byte
			mStorage.On("Set", key, mock.MatchedBy(func(b []byte) bool {
				return len(b) > 0
			})).Run(func(args mock.Arguments) {
				captured = args.Get(1).([]byte)
			}).Return().Once()

			err := mgr.Set(key, val, time.Minute)
			assert.NoError(t, err)

			// 2. Тестируем Get (проверяем корректность восстановления данных)
			mStorage.On("Get", key).Return(captured, true).Once()

			res, ok, err := mgr.Get(key)
			assert.NoError(t, err)
			assert.True(t, ok)
			assert.Equal(t, val, res)
		})
	}
}

func TestManager_JSON_NullValue(t *testing.T) {
	// Проверим специфику JSON: как он ест пустую фабрику для указателей
	ptrFactory := func() *TestUser { return nil }
	mStorage := mocks.NewMockStorage[string](t)
	codec := cache.NewJSONCodec[*TestUser](ptrFactory)
	mgr := cache.New[string, *TestUser](mStorage, codec, 0)

	key := "null-ptr"
	mStorage.On("Set", key, mock.Anything).Return().Once()

	// Пытаемся сохранить nil
	err := mgr.Set(key, nil, time.Minute)
	assert.NoError(t, err)

	// При вычитке JSON должен вернуть nil (для указателя)
	// Но так как в сторадже лежат байты, нам нужно их "замокать"
	// (в реальном JSON там будет лежать Envelope с Value: null)
}
