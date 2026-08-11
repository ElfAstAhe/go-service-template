package utils

import (
	"fmt"
	"sort"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"golang.org/x/text/language"
)

var SortedLanguages []string

var fastLanguages map[string]struct{}

func BuildLanguages() []string {
	res := make([]string, 0, 700)
	for i := 'a'; i <= 'z'; i++ {
		for j := 'a'; j <= 'z'; j++ {
			lang := string(i) + string(j)
			tag, err := language.Parse(lang)
			if err == nil && tag.String() == lang {
				res = append(res, lang)
			}
		}
	}

	return res
}

func ValidateLanguage(lang string) error {
	if _, ok := fastLanguages[lang]; !ok {
		return errs.NewCommonError(fmt.Sprintf("invalid language [%s]", lang), nil)
	}

	return nil
}

func init() {
	// build langs list
	SortedLanguages = BuildLanguages()
	sort.Strings(SortedLanguages)

	fastLanguages = make(map[string]struct{}, len(SortedLanguages))

	for _, lang := range SortedLanguages {
		fastLanguages[lang] = struct{}{}
	}
}
