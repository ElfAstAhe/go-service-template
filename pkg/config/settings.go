package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

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
