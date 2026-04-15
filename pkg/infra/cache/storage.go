package cache

// Storage - cache low level data storage
type Storage[K comparable] interface {
	Get(key K) ([]byte, bool)
	Set(key K, b []byte)
	Delete(key K)
	Range(fn func(key K, value []byte) bool)
	Has(key K) bool
	Len() int
	Clear()
}
