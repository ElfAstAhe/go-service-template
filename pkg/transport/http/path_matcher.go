package http

import (
	"regexp"
)

// PathMatcher инкапсулирует в себе параметры маршрутизации и скомпилированные
// правила сопоставления путей (URL Path Matching) на основе регулярных выражений (RegEx).
//
// Применяется в транспортных Middleware, интерцепторах безопасности и Cors-фильтрах
// для декларативного выделения публичных роутов (например, "/api/v1/auth/.*") или исключений.
type PathMatcher struct {
	Method  string         // Целевой HTTP-метод запроса (GET, POST, PUT, DELETE)
	Path    string         // Человекочитаемый текстовый шаблон или алиас пути роута
	Pattern string         // Исходная строка регулярного выражения (RegEx паттерн)
	matcher *regexp.Regexp // Скомпилированный потокобезопасный автомат регулярного выражения
}

// NewPathMatcher — фабричный конструктор и пре-компилятор правил сопоставления путей.
// Безопасно обрабатывает синтаксические ошибки компиляции RegEx, предотвращая паники времени выполнения.
func NewPathMatcher(method, path, pattern string) *PathMatcher {
	matcher, err := regexp.Compile(pattern)
	if err != nil {
		// Отказоустойчивость: в случае синтаксической ошибки в паттерне не паникуем,
		// а выставляем nil, что приведет к контролируемому отклонению Match()
		matcher = nil
	}

	return &PathMatcher{
		Method:  method,
		Path:    path,
		Pattern: pattern,
		matcher: matcher,
	}
}

// Match выполняет потоковую конкурентно-безопасную проверку входящего HTTP-запроса
// на предмет соответствия заданному правилу маршрутизации.
// Выполняет быстрые защитные проверки инвариантов до запуска тяжелого RegEx-движка.
func (m *PathMatcher) Match(method string, path string) bool {
	// 🛡️ Защитные барьеры (Guard Clauses) для пресечения рантайм паник
	if m == nil {
		return false
	}
	if m.matcher == nil {
		return false
	}
	// Сначала сверяем дешевые строки, не нагружая CPU регулярными выражениями
	if m.Method != method {
		return false
	}

	// Выполняем сопоставление по скомпилированному автомату регулярного выражения
	return m.matcher.MatchString(path)
}
