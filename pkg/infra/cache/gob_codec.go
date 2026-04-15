package cache

import (
	"bytes"
	"encoding/gob"
	"sync"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

type GobCodec[V any] struct {
	pool             *sync.Pool
	emptyItemFactory EmptyItemFactory[V]
}

func NewGobCodec[V any](factory EmptyItemFactory[V]) *GobCodec[V] {
	return &GobCodec[V]{
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		emptyItemFactory: factory,
	}
}

func (gc *GobCodec[V]) Marshal(value V, ttl time.Duration) ([]byte, error) {
	if utils.IsNil(value) {
		return nil, nil
	}
	buf := gc.pool.Get().(*bytes.Buffer)
	buf.Reset()
	defer gc.pool.Put(buf)

	var dieAt int64 = 0
	if ttl > 0 {
		dieAt = time.Now().Add(ttl).UnixNano()
	}
	env := &Envelope[V]{
		Value: value,
		DieAt: dieAt,
	}
	if err := gob.NewEncoder(buf).Encode(env); err != nil {
		return nil, err
	}

	res := make([]byte, buf.Len())
	copy(res, buf.Bytes())

	return res, nil
}

func (gc *GobCodec[V]) Unmarshal(buf []byte) (*Envelope[V], error) {
	if len(buf) == 0 {
		return &Envelope[V]{}, nil
	}
	env := &Envelope[V]{
		Value: gc.emptyItemFactory(),
	}
	err := gob.NewDecoder(bytes.NewReader(buf)).Decode(env)

	return env, err
}
