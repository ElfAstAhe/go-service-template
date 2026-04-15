package cache

// EvictionPolicy - policy to make decision to delete item from cache storage
type EvictionPolicy[K comparable] interface {
	OnGet(key K)    // Сигнал: данные прочитаны (важно для LRU/LFU)
	OnSet(key K)    // Сигнал: данные записаны/обновлены (для всех стратегий)
	OnRemove(key K) // Сигнал: данные удалены вручную (очистка индексов стратегии)
	Reset()
	Evict() (K, bool) // Удаляем жертву
}
