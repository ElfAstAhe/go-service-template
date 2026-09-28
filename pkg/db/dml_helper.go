package db

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

// BuildRepeatedObjectName генерирует уникальное имя для повторяющихся или секционированных объектов БД (например, партиций).
func BuildRepeatedObjectName(objectName string, index string) string {
	return fmt.Sprintf("%s_%s", objectName, index)
}

// BuildPKConstraintName формирует каноничное системное имя для ограничения первичного ключа (Primary Key Constraint).
// На выходе: "[имя_таблицы]_pk".
func BuildPKConstraintName(tableName string) string {
	return fmt.Sprintf("%s_pk", tableName)
}

// BuildUKConstraintName генерирует детерминированное имя для ограничения уникальности (Unique Key Constraint).
// Для предотвращения превышения лимита длины идентификатора СУБД (63 байта в Postgres) составные ключи хэшируются.
func BuildUKConstraintName(tableName string, fieldNames ...string) string {
	if len(fieldNames) == 0 {
		return fmt.Sprintf("%s_uk", tableName)
	}

	return fmt.Sprintf("%s_%s_uk", tableName, buildFieldNamesHash(fieldNames...))
}

// BuildFKConstraintName генерирует детерминированное имя для ограничения внешнего ключа (Foreign Key Constraint).
func BuildFKConstraintName(tableName string, fieldNames ...string) string {
	if len(fieldNames) == 0 {
		return fmt.Sprintf("%s_fk", tableName)
	}

	return fmt.Sprintf("%s_%s_fk", tableName, buildFieldNamesHash(fieldNames...))
}

// buildFieldNamesHash вычисляет компактный детерминированный маркер для массива полей.
// Если поле одно — возвращает его имя. Если полей несколько — сворачивает их в MD5-хэш фиксированной длины.
func buildFieldNamesHash(fieldNames ...string) string {
	if len(fieldNames) == 0 {
		return ""
	}
	// Fast-Path: для одиночного индекса сохраняем прозрачное человекочитаемое имя
	if len(fieldNames) == 1 {
		return fieldNames[0]
	}

	// Оптимизированная сборка строки без лишних аллокаций памяти в куче
	builder := strings.Builder{}
	for _, field := range fieldNames {
		builder.WriteString(field)
	}

	// Хэшируем склеенный payload для получения константной длины строки
	hasher := md5.New()
	hasher.Write([]byte(builder.String()))

	return hex.EncodeToString(hasher.Sum(nil))
}
