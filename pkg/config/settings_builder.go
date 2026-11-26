package config

import (
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// SettingsBuilder предоставляет fluent-интерфейс для декларативной сборки
// коллекции конфигурационных параметров Settings с общими префиксами ключей и CLI-флагов.
type SettingsBuilder struct {
	keyPrefix  string
	flagPrefix string
	factories  []settingFactory
}

var _ utils.Builder[Settings] = (*SettingsBuilder)(nil)

// NewSettingsBuilder создаёт новый экземпляр SettingsBuilder с заданными базовыми
// префиксами для ключей конфигурации (разделитель ".") и CLI-флагов (разделитель "-").
func NewSettingsBuilder(keyPrefix, flagPrefix string) *SettingsBuilder {
	return &SettingsBuilder{
		keyPrefix:  keyPrefix,
		flagPrefix: flagPrefix,
		factories:  make([]settingFactory, 0),
	}
}

// WithPrefixed добавляет параметр конфигурации, автоматически объединяя базовые префиксы
// билдера с переданными суффиксами keySuffix и flagSuffix.
// Если префикс или суффикс пуст, лишний разделитель не добавляется.
func (sb *SettingsBuilder) WithPrefixed(keySuffix, flagSuffix string, defaultVal any, usage string) *SettingsBuilder {
	sb.factories = append(sb.factories, &prefixedFactory{
		keySuffix:  keySuffix,
		flagSuffix: flagSuffix,
		defaultVal: defaultVal,
		usage:      usage,
	})

	return sb
}

// WithTemplated добавляет параметр конфигурации, подставляя базовые префиксы билдера
// в форматируемые шаблоны keyTemplate, flagTemplate и usageTemplate через fmt.Sprintf.
func (sb *SettingsBuilder) WithTemplated(keyTemplate, flagTemplate string, defaultVal any, usageTemplate string) *SettingsBuilder {
	sb.factories = append(sb.factories, &templatedFactory{
		keyTemplate:   keyTemplate,
		flagTemplate:  flagTemplate,
		defaultVal:    defaultVal,
		usageTemplate: usageTemplate,
	})

	return sb
}

// WithSetting создаёт новый экземпляр Setting с явными именем ключа и флага и добавляет его в коллекцию.
func (sb *SettingsBuilder) WithSetting(key, flag string, defaultVal any, description string) *SettingsBuilder {
	return sb.With(NewSetting(key, flag, defaultVal, description))
}

// With добавляет готовый объект конфигурации Setting в коллекцию билдера,
// безопасно игнорируя nil-значения.
func (sb *SettingsBuilder) With(sett *Setting) *SettingsBuilder {
	if utils.IsNil(sett) {
		return sb
	}
	sb.factories = append(sb.factories, &fixedFactory{
		setting: sett,
	})

	return sb
}

// WithBuilder встраивает настройки из переданного SettingsBuilder «как есть» (плоское слияние),
// сохраняя его собственные ключи и флаги без добавления префиксов текущего билдера.
func (sb *SettingsBuilder) WithBuilder(builder *SettingsBuilder) *SettingsBuilder {
	if utils.IsNil(builder) || builder == sb {
		return sb
	}
	sb.factories = append(sb.factories, &flatBuilderFactory{
		builder: builder,
	})

	return sb
}

// WithSubBuilder встраивает дочерний SettingsBuilder иерархически, дополняя его
// локальные префиксы префиксами текущего родительского билдера.
func (sb *SettingsBuilder) WithSubBuilder(builder *SettingsBuilder) *SettingsBuilder {
	if utils.IsNil(builder) || builder == sb {
		return sb
	}
	sb.factories = append(sb.factories, &subBuilderFactory{
		child: builder,
	})

	return sb
}

// Build разворачивает дерево фабрик с учётом всех вложенных билдеров и накопленных префиксов,
// возвращая итоговую коллекцию Settings без изменения состояния самого билдера.
func (sb *SettingsBuilder) Build() Settings {
	return sb.buildWithPrefixes(sb.keyPrefix, sb.flagPrefix)
}

func (sb *SettingsBuilder) buildWithPrefixes(keyPrefix, flagPrefix string) Settings {
	res := make(Settings, 0, len(sb.factories))
	for _, entry := range sb.factories {
		res = append(res, entry.Build(keyPrefix, flagPrefix)...)
	}

	return res
}
