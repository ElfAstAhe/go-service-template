package utils

import (
	"fmt"
	"reflect"
	"strings"
)

// GetTypeName вычисляет и возвращает короткое текстовое наименование типа переданного объекта.
//
// Метод автоматически разыменовывает указатели любой вложенности (например, **Type -> Type)
// и очищает строки анонимных типов или коллекций от длинных путей импорта (://github.com...).
// Активно используется для динамической раздачи имен метрикам Prometheus и спанам OpenTelemetry.
func GetTypeName(instance any) string {
	if instance == nil {
		return "nil"
	}

	t := reflect.TypeOf(instance)

	// Уходим от указателей любой вложенности (разматываем **Type до базовой структуры)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	// 1. Пытаемся извлечь чистое, явное имя типа (например, "LoginAttemptsRepository")
	name := t.Name()
	if name != "" {
		return name
	}

	// 2. Если имени нет (анонимная структура, map, slice) — берем строковое представление.
	// t.String() вернет мета-описание: "struct { ID string }", "[]domain.Audit" или "map[string]int"
	res := t.String()

	// 3. Срезаем длинные пути пакетов для лаконичности лейблов мониторинга.
	// Было: "://github.comElfAstAhe/go-service-template/pkg/domain.Test" -> Стало: "domain.Test"
	if lastSlash := strings.LastIndex(res, "/"); lastSlash != -1 {
		res = res[lastSlash+1:]
	}

	return res
}

// GetFullTypeName возвращает абсолютное, уникальное имя типа, включая полный путь импорта пакета.
// Применяется для исключения коллизий имен при логировании или регистрации полиморфных сущностей в DI.
func GetFullTypeName(instance any) string {
	if instance == nil {
		return "nil"
	}

	t := reflect.TypeOf(instance)

	// Разыменовываем указатели любой вложенности
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	// 1. Извлекаем имя типа и его абсолютный пакетный путь
	name := t.Name()
	pkg := t.PkgPath()

	if name != "" {
		if pkg != "" {
			// Формируем уникальный паспорт типа (например, "://github.comElfAstAhe/pkg/domain.User")
			return fmt.Sprintf("%s.%s", pkg, name)
		}

		return name
	}

	// 2. Для анонимных структур, слайсов и мап возвращаем стандартную OTel-совместимую сигнатуру
	return t.String()
}

// IsNil производит глубокую атомарную проверку значения на предмет равенства nil.
//
// ⚠️ Внимание (Инженерный нюанс Go):
// Обычное сравнение (val == nil) вернет false, если в интерфейс упакован nil-указатель
// на конкретную структуру (Typed nil interface ловушка). Этот метод использует рефлексию
// для раскрытия внутреннего состояния интерфейса, гарантируя 100% защиту от паник nil pointer dereference.
func IsNil(val any) bool {
	if val == nil {
		return true
	}

	v := reflect.ValueOf(val)
	// Проверяем только те типы данных, которые физически могут быть неинициализированными (указатели, каналы, мапы)
	switch v.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Chan, reflect.Interface, reflect.Func:
		return v.IsNil()
	default:
		// Примитивные типы (int, string, bool) или плоские структуры никогда не бывают nil
		return false
	}
}
