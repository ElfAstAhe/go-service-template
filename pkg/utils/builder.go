package utils

// Builder — маркерный интерфейс, задающий единый стандарт для всех строителей объектов типа T.
type Builder[T any] interface {
	// Build конструирует и возвращает финальный объект.
	Build() T
}
