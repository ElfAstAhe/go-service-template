package config

import (
	"time"
)

// PoolConfig инкапсулирует конфигурационные параметры емкости и политик останова пула воркеров.
type PoolConfig struct {
	WorkerCount     int           // Количество параллельно запущенных горутин-обработчиков
	DataCapacity    int           // Буферная емкость внутреннего канала задач (Backpressure window)
	CompleteProcess bool          // Флаг: вычитывать ли буфер до конца при закрытии канала (true) или тушить экстренно (false)
	StopTimeout     time.Duration // Временной лимит (таймаут) на мягкое завершение обработки перед принудительным выходом
}
