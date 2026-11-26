package config

import (
	"fmt"
)

// settingFactory определяет внутренний узел дерева настроек (паттерн Composite)
// для отложенной сборки параметров с учётом накопленных префиксов.
type settingFactory interface {
	Build(keyPrefix, flagPrefix string) []*Setting
}

type fixedFactory struct {
	setting *Setting
}

var _ settingFactory = (*fixedFactory)(nil)

func (ff *fixedFactory) Build(_, _ string) []*Setting {
	return []*Setting{ff.setting}
}

type prefixedFactory struct {
	keySuffix  string
	flagSuffix string
	defaultVal any
	usage      string
}

var _ settingFactory = (*prefixedFactory)(nil)

func (pf *prefixedFactory) Build(keyPrefix, flagPrefix string) []*Setting {
	key := joinPrefixSuffix(keyPrefix, pf.keySuffix, ".")
	flag := joinPrefixSuffix(flagPrefix, pf.flagSuffix, "-")

	return []*Setting{NewSetting(key, flag, pf.defaultVal, pf.usage)}
}

type templatedFactory struct {
	keyTemplate   string
	flagTemplate  string
	defaultVal    any
	usageTemplate string
}

var _ settingFactory = (*templatedFactory)(nil)

func (tf *templatedFactory) Build(keyPrefix, flagPrefix string) []*Setting {
	return []*Setting{
		NewSetting(
			fmt.Sprintf(tf.keyTemplate, keyPrefix),
			fmt.Sprintf(tf.flagTemplate, flagPrefix),
			tf.defaultVal,
			fmt.Sprintf(tf.usageTemplate, flagPrefix),
		),
	}
}

type subBuilderFactory struct {
	child *SettingsBuilder
}

var _ settingFactory = (*subBuilderFactory)(nil)

func (sbf *subBuilderFactory) Build(parentKeyPrefix, parentFlagPrefix string) []*Setting {
	fullKey := joinPrefixSuffix(parentKeyPrefix, sbf.child.keyPrefix, ".")
	fullFlag := joinPrefixSuffix(parentFlagPrefix, sbf.child.flagPrefix, "-")

	return sbf.child.buildWithPrefixes(fullKey, fullFlag)
}

func joinPrefixSuffix(prefix, suffix, sep string) string {
	switch {
	case prefix == "":
		return suffix
	case suffix == "":
		return prefix
	default:
		return prefix + sep + suffix
	}
}

type flatBuilderFactory struct {
	builder *SettingsBuilder
}

var _ settingFactory = (*flatBuilderFactory)(nil)

func (fbf *flatBuilderFactory) Build(_, _ string) []*Setting {
	return fbf.builder.Build()
}
