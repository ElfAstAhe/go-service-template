package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
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

// Settings представляет собой коллекцию элементов конфигурации с методами массовой обработки.
type Settings []*Setting

// InitFlags регистрирует все флаги коллекции в переданном pflag.FlagSet.
// Пропускает элементы, у которых имя флага не задано (равно пустой строке).
func (ss Settings) InitFlags(fs *pflag.FlagSet) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Проверяем, является ли r ошибкой
			recoveryErr, ok := r.(error)
			if !ok {
				// Если это строка или что-то другое, приводим к виду error вручную
				recoveryErr = errs.NewConfigError(fmt.Sprintf("panic [%v] recovery", r), nil)
			}

			err = errs.NewConfigError("parse cli flags panic", recoveryErr)
		}
	}()

	for _, setting := range ss {
		if strings.TrimSpace(setting.Flag) == "" {
			continue
		}
		if err := setting.InitFlag(fs); err != nil {
			return err
		}
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		return errs.NewConfigError("flags parse failed", err)
	}

	return nil
}

// InitKeys инициализирует все ключи в коллекции.
func (ss Settings) InitKeys(v *viper.Viper) error {
	for _, setting := range ss {
		if strings.TrimSpace(setting.Key) == "" {
			continue
		}
		if err := setting.InitKey(v); err != nil {
			return err
		}
	}

	return nil
}

// BindKeysAndFlags инициирует массовую установку значений по умолчанию и связывание флагов в Viper.
func (ss Settings) BindKeysAndFlags(v *viper.Viper, fs *pflag.FlagSet) error {
	for _, setting := range ss {
		if strings.TrimSpace(setting.Key) == "" || strings.TrimSpace(setting.Flag) == "" {
			continue
		}
		flag := fs.Lookup(setting.Flag)
		if flag != nil {
			if err := v.BindPFlag(setting.Key, flag); err != nil {
				return errs.NewConfigError(fmt.Sprintf("failed to bind flag %s to key %s", setting.Flag, setting.Key), err)
			}
		}
	}

	return nil
}

// Len возвращает количество элементов в коллекции Settings.
func (ss Settings) Len() int {
	return len(ss)
}
