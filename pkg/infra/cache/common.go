package cache

// EmptyItemFactory необходим для создания чистого экземпляра V перед десериализацией.
type EmptyItemFactory[V any] func() V

type Envelope[V any] struct {
	Value V
	DieAt int64
}
