package cache

import (
	"time"
)

// Константы рантайм-настроек защиты двухконтурного кэша.
const (
	// negativeGetTTL определяет время жизни (TTL) для «негативного кэша».
	// Защищает базу данных от лавинообразных холостых запросов при отсутствии ключа (Cache Penetration).
	negativeGetTTL time.Duration = 2 * time.Minute
)

// L2Manager реализует расширенное поведение дженерик-кэша, наследуя базовый функционал Manager.
//
// 💡 Архитектурный паттерн (Negative Caching / Cache Stampede Protection):
// При отсутствии ключа в хранилище менеджер перехватывает промах (Cache Miss) и принудительно
// фиксирует в памяти пустое (zero-value) значение типа V на короткий интервал времени.
// Это предотвращает утилизацию процессора и дисков СУБД при агрессивном конкурентном поиске несуществующих данных.
type L2Manager[K comparable, V any] struct {
	*Manager[K, V] // Встраиваем базовый потокобезопасный менеджер кэша
}

// Гарантируем соответствие контракту Cache[K, V] на этапе компиляции
var _ Cache[string, string] = (*L2Manager[string, string])(nil)

// NewL2 — фабричный конструктор двухконтурного защищенного менеджера кэша.
func NewL2[K comparable, V any](
	storage Storage[K],
	codec Codec[V],
	janitorMaxSize int,
) *L2Manager[K, V] {
	return &L2Manager[K, V]{
		Manager: New(storage, codec, janitorMaxSize),
	}
}

// Get извлекает строго типизированное значение по его ключу.
// В случае промаха кэша атомарно фиксирует негативный результат в памяти для снижения нагрузки на DAL-слой.
func (l2m *L2Manager[K, V]) Get(key K) (V, bool, error) {
	// Вызываем базовое извлечение из нижележащих шардов памяти
	res, ok, err := l2m.Manager.Get(key)
	if err != nil {
		return res, ok, err
	}

	// Если элемент отсутствует в кэше — консервируем zero-value «заглушку» на 2 минуты
	if !ok {
		// Игнорируем ошибку записи, так как приоритетом является возврат исходного статуса ok
		_ = l2m.Manager.Set(key, res, negativeGetTTL)
	}

	return res, ok, nil
}
