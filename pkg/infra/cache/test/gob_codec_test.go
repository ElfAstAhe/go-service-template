package test

import (
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestManager_GobCodec_Integration(t *testing.T) {
	// Тип для теста
	type UserProfile struct {
		Name string
		Age  int
	}

	// 1. Регистрация типа обязательна для Gob, если работаем через интерфейсы/any
	// Но так как у нас Envelope[V] и V конкретный, Gob может справиться сам.
	// Однако для сложных структур лучше перестраховаться.

	factory := func() UserProfile { return UserProfile{} }

	mStorage := mocks.NewMockStorage[string](t)
	codec := cache.NewGobCodec[UserProfile](factory)
	mgr := cache.New[string, UserProfile](mStorage, codec, 0)

	key := "user:1"
	val := UserProfile{Name: "Test Robot", Age: 25}

	// Тестируем цикл Set -> Get
	var captured []byte
	mStorage.On("Set", key, mock.Anything).Run(func(args mock.Arguments) {
		captured = args.Get(1).([]byte)
	}).Return()

	err := mgr.Set(key, val, time.Minute)
	assert.NoError(t, err)

	mStorage.On("Get", key).Return(captured, true)

	res, ok, err := mgr.Get(key)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, val, res)
}

func TestManager_GobCodec_Slices(t *testing.T) {
	// Проверяем работу со слайсами через фабрику
	factory := func() []string { return make([]string, 0) }

	mStorage := mocks.NewMockStorage[string](t)
	codec := cache.NewGobCodec[[]string](factory)
	mgr := cache.New[string, []string](mStorage, codec, 0)

	val := []string{"apple", "banana"}

	var captured []byte
	mStorage.On("Set", "k", mock.Anything).Run(func(args mock.Arguments) {
		captured = args.Get(1).([]byte)
	}).Return()

	_ = mgr.Set("k", val, time.Minute)
	mStorage.On("Get", "k").Return(captured, true)

	res, ok, _ := mgr.Get("k")
	assert.True(t, ok)
	assert.Equal(t, val, res)
}
