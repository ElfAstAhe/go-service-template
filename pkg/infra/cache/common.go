package cache

// EmptyItemFactory необходим для создания чистого экземпляра V перед десериализацией.
type EmptyItemFactory[V any] func() V

type Options struct {
	MaxSize int
}
type Option func(*Options)

func WithMaxSize(size int) Option {
	return func(o *Options) {
		o.MaxSize = size
	}
}

type Envelope[V any] struct {
	Value V
	DieAt int64
}
