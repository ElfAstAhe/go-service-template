package http

import (
	"strings"
)

// PathMatchers инкапсулирует в себе иерархический реестр сгруппированных по HTTP-методам
// правил маршрутизации (Path Matchers Collection).
//
// Выступает в роли эффективного селектора роутов первого уровня, позволяя транспортным Middleware
// быстро проверять входящие пути на предмет совпадения с вайтлистами или Cors-политиками.
type PathMatchers struct {
	WatchPaths map[string][]*PathMatcher // Карта бакетов, сопоставляющая HTTP-метод со срезом его RegEx-паттернов
}

// NewHTTPPathMatchers — фабричный конструктор коллективного реестра селекторов путей.
// Выполняет автоматическую дедупликацию идентичных паттернов на этапе сборки графа зависимостей.
func NewHTTPPathMatchers(matchers []*PathMatcher) *PathMatchers {
	watchPaths := make(map[string][]*PathMatcher)

	for _, matcher := range matchers {
		slice, ok := watchPaths[matcher.Method]
		if !ok {
			slice = make([]*PathMatcher, 0)
		}

		// Предохранитель: если такой паттерн для данного метода уже зарегистрирован — игнорируем дубликат
		if watchPathExists(slice, matcher.Method, matcher.Path) {
			continue
		}

		// Перезаписываем бакет мапы новой ссылкой на срез, возвращенной функцией append
		watchPaths[matcher.Method] = append(slice, matcher)
	}

	return &PathMatchers{
		WatchPaths: watchPaths,
	}
}

// Match выполняет быстрый иерархический поиск по реестру, проверяя, удовлетворяет ли
// входящий URL-путь хотя бы одному зарегистрированному регулярному выражению для данного HTTP-метода.
func (hpm *PathMatchers) Match(method string, path string) bool {
	pms, ok := hpm.WatchPaths[method]
	if !ok {
		return false // Fast-Path: если для такого HTTP-метода роутов нет, мгновенно выходим
	}

	// Переходим к линейному обходу пре-компилированных автоматов регулярных выражений внутри бакета
	for _, pm := range pms {
		if pm.Match(method, path) {
			return true // Мгновенный возврат при первом же совпадении (Short-Circuit evaluation)
		}
	}

	return false
}

// GetPathMatcher производит точечный поиск и извлечение исходного объекта PathMatcher
// по его точному строковому совпадению метода и человекочитаемого алиаса пути (без прогона через RegEx-движок).
func (hpm *PathMatchers) GetPathMatcher(method string, path string) *PathMatcher {
	slice, ok := hpm.WatchPaths[method]
	if !ok {
		return nil
	}

	for _, item := range slice {
		// Очищаем пробельные символы перед точной сверкой инвариантов путей
		if strings.TrimSpace(method) == item.Method && strings.TrimSpace(path) == item.Path {
			return item
		}
	}

	return nil
}

// watchPathExists — служебная (неэкспортируемая) функция линейного поиска дубликатов на этапе сборки коллекции.
func watchPathExists(src []*PathMatcher, method string, path string) bool {
	if strings.TrimSpace(path) == "" || strings.TrimSpace(method) == "" || len(src) == 0 {
		return false
	}

	for _, item := range src {
		if item.Method == method && item.Path == path {
			return true
		}
	}

	return false
}
