package cache

import (
	"bytes"
	"encoding/json"
	"sync"
	"time"
)

type JSONCodec[V any] struct {
	pool             *sync.Pool
	emptyItemFactory EmptyItemFactory[V]
}

func NewJSONCodec[V any](factory EmptyItemFactory[V]) *JSONCodec[V] {
	return &JSONCodec[V]{
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		emptyItemFactory: factory,
	}
}

func (jc *JSONCodec[V]) Marshal(value V, ttl time.Duration) ([]byte, error) {
	buf := jc.pool.Get().(*bytes.Buffer)
	buf.Reset()
	defer jc.pool.Put(buf)

	var dieAt int64 = 0
	if ttl > 0 {
		dieAt = time.Now().Add(ttl).UnixNano()
	}
	env := &Envelope[V]{
		Value: value,
		DieAt: dieAt,
	}
	if err := json.NewEncoder(buf).Encode(env); err != nil {
		return nil, err
	}

	res := make([]byte, buf.Len())
	copy(res, buf.Bytes())

	return res, nil
}

func (jc *JSONCodec[V]) Unmarshal(buf []byte) (*Envelope[V], error) {
	env := &Envelope[V]{
		Value: jc.emptyItemFactory(),
	}
	err := json.NewDecoder(bytes.NewReader(buf)).Decode(env)

	return env, err
}
