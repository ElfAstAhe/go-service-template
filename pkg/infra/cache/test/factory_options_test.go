package test

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache"
	"github.com/ElfAstAhe/go-service-template/pkg/infra/cache/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewFactoryOptions_Defaults(t *testing.T) {
	opts := cache.NewFactoryOptions[string, int]()

	assert.Equal(t, cache.DefaultCacheL2, opts.L2)
	assert.Equal(t, cache.DefaultCacheShardCount, opts.ShardCount)
	assert.Equal(t, cache.DefaultCacheMaxSize, opts.MaxSize)
	assert.Equal(t, cache.DefaultCacheJanitorMaxSize, opts.JanitorMaxSize)
	assert.Nil(t, opts.Policy)
	assert.Nil(t, opts.Codec)
	assert.NotNil(t, opts.ShardFactory)
}

func TestFactoryOptions_FluentAPI(t *testing.T) {
	mockCodec := mocks.NewMockCodec[int](t)
	mockPolicy := mocks.NewMockEvictionPolicy[string](t)
	var dummyShardFactory cache.ShardFactory[string] = func(maxSize int, policy cache.EvictionPolicy[string]) cache.Storage[string] { return nil }

	opts := cache.NewFactoryOptions[string, int]()

	cache.WithCacheL2[string, int](true)(opts)
	cache.WithCacheShardCount[string, int](4)(opts)
	cache.WithCacheShardFactory[string, int](dummyShardFactory)(opts)
	cache.WithCacheMaxSize[string, int](5000)(opts)
	cache.WithCacheCustomEvictPolicy[string, int](mockPolicy)(opts)
	cache.WithCacheCodec[string, int](mockCodec)(opts)
	cache.WithCacheJanitorMaxSize[string, int](500)(opts)

	assert.True(t, opts.L2)
	assert.Equal(t, uint64(4), opts.ShardCount)
	assert.NotNil(t, opts.ShardFactory)
	assert.Equal(t, 5000, opts.MaxSize)
	assert.Equal(t, mockPolicy, opts.Policy)
	assert.Equal(t, mockCodec, opts.Codec)
	assert.Equal(t, 500, opts.JanitorMaxSize)
}

func TestFactoryOptions_EvictionPoliciesFluent(t *testing.T) {
	optsLRU := cache.NewFactoryOptions[string, int]()
	cache.WithCacheLRUEvictPolicy[string, int]()(optsLRU)
	assert.NotNil(t, optsLRU.Policy)

	optsLFU := cache.NewFactoryOptions[string, int]()
	cache.WithCacheLFUEvictPolicy[string, int]()(optsLFU)
	assert.NotNil(t, optsLFU.Policy)

	optsFIFO := cache.NewFactoryOptions[string, int]()
	cache.WithCacheFIFOEvictPolicy[string, int]()(optsFIFO)
	assert.NotNil(t, optsFIFO.Policy)
}

func TestFactoryOptions_Validate(t *testing.T) {
	mockCodec := mocks.NewMockCodec[int](t)
	//	var dummyShardFactory ShardFactory[string] = func(maxSize int, policy EvictionPolicy[string]) Storage[string] { return nil }

	tests := []struct {
		name    string
		modify  func(*cache.FactoryOptions[string, int])
		wantErr bool
	}{
		{
			name: "valid options",
			modify: func(o *cache.FactoryOptions[string, int]) {
				o.Codec = mockCodec
			},
			wantErr: false,
		},
		{
			name: "zero shard count",
			modify: func(o *cache.FactoryOptions[string, int]) {
				o.ShardCount = 0
				o.Codec = mockCodec
			},
			wantErr: true,
		},
		{
			name: "nil shard factory",
			modify: func(o *cache.FactoryOptions[string, int]) {
				o.ShardFactory = nil
				o.Codec = mockCodec
			},
			wantErr: true,
		},
		{
			name: "negative max size",
			modify: func(o *cache.FactoryOptions[string, int]) {
				o.MaxSize = -1
				o.Codec = mockCodec
			},
			wantErr: true,
		},
		{
			name: "nil codec",
			modify: func(o *cache.FactoryOptions[string, int]) {
				o.Codec = nil
			},
			wantErr: true,
		},
		{
			name: "zero janitor max size",
			modify: func(o *cache.FactoryOptions[string, int]) {
				o.JanitorMaxSize = 0
				o.Codec = mockCodec
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := cache.NewFactoryOptions[string, int]()
			tt.modify(opts)
			err := opts.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
