package transport

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
)

// Глобальные переменные рантайма для генерации распределенных идентификаторов.
var (
	// prefix пакует в себя имя хоста и случайную соль, уникальную для текущего инстанса приложения.
	prefix string
	// reqID — потокобезопасный атомарный счетчик инкремента запросов.
	reqID atomic.Uint64
)

// GetPrefix возвращает неизменяемую статическую часть (префикс) текущего пода/инстанса.
func GetPrefix() string {
	return prefix
}

// NextReqID атомарно увеличивает счетчик на уровне инструкций CPU и возвращает следующий ID.
// Работает без блокировок (Lock-free), обеспечивая экстремальную скорость в конкурентной среде.
func NextReqID() uint64 {
	return reqID.Add(1)
}

// init инициализирует уникальный паспорт (префикс) текущего процесса при старте приложения.
func init() {
	// Извлекаем имя хоста (в Kubernetes это имя конкретного пода)
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}

	var buf [12]byte
	var b64 string

	// Цикл гарантирует генерацию чистой base64-строки длиной не менее 10 символов
	// после вырезания спецсимволов, потенциально ломающих HTTP-заголовки.
	for len(b64) < 10 {
		// Используем криптографически стойкий генератор rand вместо math/rand
		_, _ = rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		// Вычищаем символы, которые могут быть невалидными для сетевых протоколов
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	// Формируем итоговый шаблон: "hostname/salt_10_chars" (например: "tiny-auth-pod-xyz/aB3dE6gH1j")
	prefix = fmt.Sprintf("%s/%s", hostname, b64[0:10])
}
