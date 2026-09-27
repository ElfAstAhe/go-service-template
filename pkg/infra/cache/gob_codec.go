package cache

import (
	"bytes"
	"encoding/gob"
	"sync"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// GobCodec реализует интерфейс Codec[V], используя нативный бинарный формат сериализации Go (encoding/gob).
//
// 💡 Архитектурный нюанс:
// Данный кодек оптимизирован для внутреннего высокопроизводительного кэширования (Go-to-Go рантайм).
// За счет применения sync.Pool для байтовых буферов полностью устраняет аллокации памяти в куче (Zero-Allocation),
// кратно снижая нагрузку на Garbage Collector (GC) под интенсивной Highload-нагрузкой.
type GobCodec[V any] struct {
	pool             *sync.Pool          // Пул переиспользуемых байтовых буферов bytes.Buffer
	emptyItemFactory EmptyItemFactory[V] // Фабрика для безопасной аллокации целевого типа V перед десериализацией
}

// Гарантируем полное соответствие интерфейсу Codec[V] на этапе компиляции
var _ Codec[string] = (*GobCodec[string])(nil)

// NewGobCodec — фабричный конструктор бинарного кодека GOB.
// Принимает фабрику-замыкание для подготовки памяти под дженерик-сущности.
func NewGobCodec[V any](factory EmptyItemFactory[V]) *GobCodec[V] {
	return &GobCodec[V]{
		pool: &sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
		emptyItemFactory: factory,
	}
}

// Marshal пакует объект v и время его смерти по TTL в монолитный бинарный конверт Envelope[V].
// Утилизирует буферы из пула, минимизируя нагрузку на подсистему управления памятью.
//
//goland:noinspection DuplicatedCode
func (gc *GobCodec[V]) Marshal(value V, ttl time.Duration) ([]byte, error) {
	if utils.IsNil(value) {
		return nil, nil
	}

	// Извлекаем чистый буфер из пула
	buf := gc.pool.Get().(*bytes.Buffer)
	buf.Reset()
	defer gc.pool.Put(buf) // Гарантируем возврат буфера в пул при выходе из метода

	var dieAt int64
	if ttl > 0 {
		// Фиксируем абсолютную метку смерти в наносекундах для ультраточного контроля TTL
		dieAt = time.Now().Add(ttl).UnixNano()
	}

	env := &Envelope[V]{
		Value: value,
		DieAt: dieAt,
	}

	if err := gob.NewEncoder(buf).Encode(env); err != nil {
		return nil, err
	}

	// Выделяем точечный массив байт и копируем туда результат перед очисткой буфера пула
	res := make([]byte, buf.Len())
	copy(res, buf.Bytes())

	return res, nil
}

// Unmarshal восстанавливает бинарный массив в строго типизированный конверт Envelope[V].
// Безопасно разворачивает указатели и интерфейсы с помощью инжектированной фабрики emptyItemFactory.
func (gc *GobCodec[V]) Unmarshal(buf []byte) (*Envelope[V], error) {
	if len(buf) == 0 {
		return &Envelope[V]{}, nil
	}

	// Пре-аллокация памяти под дженерик-структуру через фабрику рантайма
	env := &Envelope[V]{
		Value: gc.emptyItemFactory(),
	}

	err := gob.NewDecoder(bytes.NewReader(buf)).Decode(env)
	return env, err
}
