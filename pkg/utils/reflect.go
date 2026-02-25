package utils

import (
	"reflect"
	"strings"
)

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
