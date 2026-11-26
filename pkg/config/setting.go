package config

import (
	"fmt"
	"reflect"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Setting инкапсулирует метаданные одного параметра конфигурации,
// связывая ключ Viper, CLI-флаг, значение по умолчанию и описание флага.
type Setting struct {
	Key        string
	Flag       string
	Default    any
	Usage      string
	defValType reflect.Type
}

// NewSetting создает новый экземпляр описания настройки Setting и вычисляет её reflect.Type.
func NewSetting(key string, flag string, defaultVal any, usage string) *Setting {
	return &Setting{
		Key:        key,
		Flag:       flag,
		Default:    defaultVal,
		Usage:      usage,
		defValType: reflect.TypeOf(defaultVal),
	}
}

// InitKey иниуиализация ключей и дефоль значений
func (s *Setting) InitKey(v *viper.Viper) error {
	if s.defValType == nil {
		return errs.NewConfigError(fmt.Sprintf("setting %s: default value type is nil", s.Key), nil)
	}

	v.SetDefault(s.Key, s.Default)

	return nil
}

// InitFlag автоматически регистрирует флаг в pflag.FlagSet на основе вычисленного типа данных.
// Поддерживает типы: string, int, int32, int64, float32, float64, bool, []string, time.Duration и time.Time.
func (s *Setting) InitFlag(fs *pflag.FlagSet) error {
	if s.defValType == nil {
		return errs.NewConfigError(fmt.Sprintf("setting %s: default value type is nil", s.Flag), nil)
	}

	switch s.defValType.Kind() {
	case reflect.String:
		fs.String(s.Flag, s.Default.(string), s.Usage)
	case reflect.Int:
		fs.Int(s.Flag, s.Default.(int), s.Usage)
	case reflect.Int32:
		fs.Int32(s.Flag, s.Default.(int32), s.Usage)
	case reflect.Bool:
		fs.Bool(s.Flag, s.Default.(bool), s.Usage)
	case reflect.Float32:
		fs.Float32(s.Flag, s.Default.(float32), s.Usage)
	case reflect.Float64:
		fs.Float64(s.Flag, s.Default.(float64), s.Usage)
	case reflect.Slice:
		if s.defValType.Elem().Kind() == reflect.String {
			fs.StringSlice(s.Flag, s.Default.([]string), s.Usage)
		} else {
			return errs.NewConfigError(fmt.Sprintf("setting %s: unsupported slice type []%s", s.Flag, s.defValType.Elem().Kind().String()), nil)
		}
	case reflect.Struct:
		if s.defValType == reflect.TypeFor[time.Time]() {
			t := s.Default.(time.Time)
			fs.String(s.Flag, t.Format(time.RFC3339), s.Usage)
		} else {
			return errs.NewConfigError(fmt.Sprintf("setting %s: unsupported struct type %s", s.Flag, s.defValType.String()), nil)
		}
	case reflect.Int64:
		if s.defValType == reflect.TypeFor[time.Duration]() {
			fs.Duration(s.Flag, s.Default.(time.Duration), s.Usage)
		} else {
			fs.Int64(s.Flag, s.Default.(int64), s.Usage)
		}
	default:
		return errs.NewConfigError(fmt.Sprintf("setting %s: unsupported flag type %s", s.Flag, s.defValType.String()), nil)
	}

	return nil
}
