package utils

import (
	"reflect"
	"strings"
)

// GetTypeName возвращает наименование типа хоть в каком-либо виде
func GetTypeName(instance any) string {
	if instance == nil {
		return "nil"
	}

	t := reflect.TypeOf(instance)

	// Уходим от указателей (даже если их несколько: **Type)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 1. Пытаемся взять чистое имя типа (например, "TestRepositoryImpl")
	name := t.Name()
	if name != "" {
		return name
	}

	// 2. Если имени нет (анонимная структура, map, slice), берем строковое представление
	// t.String() вернет что-то вроде "struct { ID string }", "[]domain.Test" или "map[string]int"
	res := t.String()

	// 3. Убираем длинные пути пакетов для лаконичности (оставляем только последний сегмент)
	// Было: "://github.com" -> Стало: "domain.Test"
	if lastSlash := strings.LastIndex(res, "/"); lastSlash != -1 {
		res = res[lastSlash+1:]
	}

	return res
}

// IsNil определяет, является ли параметр nil по значению
//
// Параметры:
//   - val: любое значение
//
// Например: интерфейс после присвоения, допустим структуры, уже точно не будет nil, даже если значение структуры nil
func IsNil(val any) bool {
	if val == nil {
		return true
	}

	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Chan, reflect.Interface, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}
