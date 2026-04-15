package cache

import (
	"time"
)

type Codec[V any] interface {
	Marshal(v V, ttl time.Duration) ([]byte, error)
	Unmarshal(b []byte) (*Envelope[V], error)
}
