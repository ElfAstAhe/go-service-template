package utils

import (
	"fmt"
	"sort"
	// Встраиваем базу данных IANA таймзон прямо в бинарный файл приложения.
	// Гарантирует автономную работу валидатора в пустых Docker-контейнерах (scratch / alpine).
	_ "time/tzdata"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

var (
	// SortedTimeZones содержит алфавитный перечень всех валидных мировых таймзон (например, "America/New_York", "Europe/Moscow").
	SortedTimeZones []string

	// fastTimeZones — высокоскоростная хэш-мапа для мгновенной O(1) проверки существования зоны без аллокаций.
	fastTimeZones map[string]struct{}
)

// ValidateTimeZone осуществляет атомарную O(1) проверку переданного строкового идентификатора таймзоны.
// Возвращает бизнес-ошибку, если зона не соответствует мировому стандарту базы данных IANA.
func ValidateTimeZone(tz string) error {
	if _, ok := fastTimeZones[tz]; !ok {
		return errs.NewCommonError(fmt.Sprintf("invalid time zone [%s]", tz), nil)
	}

	return nil
}

// init координирует запуск ленивой инициализации базы зон при старте текущего процесса.
func init() {
	initTimeZones()
}

// initTimeZones выполняет генерацию, алфавитную сортировку и хэш-индексацию реестра мировых таймзон.
func initTimeZones() {
	// Извлекаем встроенный массив зон
	tz := buildEmbeddedTimeZones()
	sort.Strings(tz)
	SortedTimeZones = tz

	// ИСПРАВЛЕНО: Аллоцируем мапу точного размера len(SortedTimeZones) для полного исключения рехашинга
	fastTimeZones = make(map[string]struct{}, len(SortedTimeZones))
	for _, z := range SortedTimeZones {
		fastTimeZones[z] = struct{}{}
	}
}
