package utils

import (
	"fmt"
	"sort"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"golang.org/x/text/language"
)

var (
	// SortedLanguages содержит отсортированный по алфавиту срез всех валидных двухбуквенных языковых кодов (ISO 639-1).
	SortedLanguages []string

	// fastLanguages — высокоскоростная индексированная хэш-мапа для мгновенной O(1) проверки существования языка.
	fastLanguages map[string]struct{}
)

// BuildLanguages генерирует полный перечень валидных двухбуквенных языковых тегов.
// Перебирает комбинации символов от 'aa' до 'zz' и верифицирует их через парсер библиотеки golang.org/x/text/language.
func BuildLanguages() []string {
	res := make([]string, 0, 700) // Аллоцируем буфер с запасом под потенциальный объем языковых кодов
	for i := 'a'; i <= 'z'; i++ {
		for j := 'a'; j <= 'z'; j++ {
			lang := string(i) + string(j)
			tag, err := language.Parse(lang)
			// Если парсер принял тег и он совпадает с исходной строкой (исключаем дефолтные маппинги) — фиксируем его
			if err == nil && tag.String() == lang {
				res = append(res, lang)
			}
		}
	}

	return res
}

// ValidateLanguage выполняет атомарную O(1) проверку переданного языкового кода.
// Возвращает понятную бизнес-ошибку, если код отсутствует в словаре валидных стандартов ISO.
func ValidateLanguage(lang string) error {
	if _, ok := fastLanguages[lang]; !ok {
		return errs.NewCommonError(fmt.Sprintf("invalid language [%s]", lang), nil)
	}

	return nil
}

// init осуществляет ленивую генерацию, сортировку и хэш-индексацию языковой матрицы в момент старта приложения.
func init() {
	// Сборка и алфавитная сортировка исходного массива
	SortedLanguages = BuildLanguages()
	sort.Strings(SortedLanguages)

	// Аллоцируем мапу точного размера во избежание рехашинга при заполнении
	fastLanguages = make(map[string]struct{}, len(SortedLanguages))

	// Переносим данные в мапу пустых структур struct{}{} (аллокация памяти под значение равна 0 байт)
	for _, lang := range SortedLanguages {
		fastLanguages[lang] = struct{}{}
	}
}
