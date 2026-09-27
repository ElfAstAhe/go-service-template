package http

// HealthzFunc определяет сигнатуру функции-колбека для проверки общей жизнеспособности компонента (Liveness Probe).
// Используется HTTP-сервером для рапортования оркестратору Kubernetes о том, что процесс не завис и жив.
type HealthzFunc func() bool

// ReadyzFunc определяет сигнатуру функции-колбека для проверки готовности компонента к приему трафика (Readiness Probe).
// Позволяет верифицировать доступность критических сокетов (СУБД, Кафка) перед открытием балансировки на под.
type ReadyzFunc func() bool

// MapToHTTPStatusFunc определяет сигнатуру функции-транслятора абстрактных бизнес-исключенийUse Cases/DAL
// в системные числовые статус-коды спецификации протокола HTTP (например, HTTP 400 Bad Request, HTTP 500 Internal).
type MapToHTTPStatusFunc func(error) int
