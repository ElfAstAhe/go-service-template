package utils

import (
	"fmt"
	"sort"
	_ "time/tzdata"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
)

var SortedTimeZones []string

var fastTimeZones map[string]struct{}

func ValidateTimeZone(tz string) error {
	if _, ok := fastTimeZones[tz]; !ok {
		return errs.NewCommonError(fmt.Sprintf("invalid time zone [%s]", tz), nil)
	}

	return nil
}

func init() {
	// init time zones
	initTimeZones()
}

func initTimeZones() {
	tz := buildEmbeddedTimeZones()
	sort.Strings(tz)
	SortedTimeZones = tz

	fastTimeZones = make(map[string]struct{}, len(fastTimeZones))
	for _, z := range SortedTimeZones {
		fastTimeZones[z] = struct{}{}
	}
}
