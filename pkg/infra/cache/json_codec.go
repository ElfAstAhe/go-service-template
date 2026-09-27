package cache

import (
	"bytes"
	"encoding/json"
	"sync"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// JSONCodec реализует интерфейс Codec[V], используя каноничный текстовый формат сериализации JSON.
//
// 💡 Архитектурный нюанс:
// В отличие от GobCodec, данный маршалер ориентирован на открытое межсервисное L2-кэширование (Interoperability).
// Текстовый payload JSON-конверта легко читается любыми сторонними языками программирования в кластере.
// Использование sync.Pool для байтовых буферов обеспечивает высокую скорость работы и защиту от нагрузок на GC.
type JSONCodec[V any] struct {
	pool             *sync.Pool          // Пул переиспользуемых байтовых буферов bytes.Buffer (Zero-Allocation)
	emptyItemFactory EmptyItemFactory[V] // Фабрика рантайма для аллокации памяти под дженерик-тип V перед парсингом
}

// Гарантируем полное соответствие интерфейсу Codec[V] на этапе компиляции
var _ Codec[string] = (*JSONCodec[string])(nil)

// NewJSONCodec — фабричный конструктор текстового кодека JSON.
func NewJSONCodec[V any](factory EmptyItemFactory[V]) *JSONCodec[V] {
	return &JSONCodec[V]{
		pool: &sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
		emptyItemFactory: factory,
	}
}

// Marshal упаковывает доменный объект v и время его смерти по TTL в текстовый JSON-конверт Envelope[V].
// Оптимизирует операции ввода-вывода (I/O) за счет стримингового кодирования прямо в буфер пула.
//
//goland:noinspection DuplicatedCode
func (jc *JSONCodec[V]) Marshal(value V, ttl time.Duration) ([]byte, error) {
	if utils.IsNil(value) {
		return nil, nil
	}

	// Извлекаем переиспользуемый буфер из пула, сбрасывая его внутренние указатели
	buf := jc.pool.Get().(*bytes.Buffer)
	buf.Reset()
	defer jc.pool.Put(buf) // Гарантируем возврат ресурса в пул для повторного использования

	var dieAt int64
	if ttl > 0 {
		// Фиксируем Unix Timestamp смерти в наносекундах для сквозной точности
		dieAt = time.Now().Add(ttl).UnixNano()
	}

	env := &Envelope[V]{
		Value: value,
		DieAt: dieAt,
	}

	// Стриминговое потоковое кодирование без лишних промежуточных аллокаций строк
	if err := json.NewEncoder(buf).Encode(env); err != nil {
		return nil, err
	}

	// Нарезаем точный массив байт под размер буфера и копируем данные
	res := make([]byte, buf.Len())
	copy(res, buf.Bytes())

	return res, nil
}

// Unmarshal выполняет чтение байтового JSON-массива и восстанавливает структуру Envelope[V].
// Безопасно разворачивает вложенные структуры и карты за счет пре-аллокации памяти через emptyItemFactory.
func (jc *JSONCodec[V]) Unmarshal(buf []byte) (*Envelope[V], error) {
	if len(buf) == 0 {
		return &Envelope[V]{}, nil
	}

	// Аллоцируем чистую память под тип V перед запуском парсера
	env := &Envelope[V]{
		Value: jc.emptyItemFactory(),
	}

	err := json.NewDecoder(bytes.NewReader(buf)).Decode(env)
	return env, err
}
