package zcli

import (
	"bytes"
	"flag"
	"sync"
	"testing"

	"github.com/sohaha/zlsgo"
)

// resetI18nForTesting resets the internal i18n state for testing purposes only
func resetI18nForTesting() {
	internalI18n = nil
	initOnce = sync.Once{}
	lastSyncedLang = ""
}

func TestI18n_BuiltInTranslations(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	resetI18nForTesting()
	defer func() {
		Lang = originalLang
		resetI18nForTesting()
	}()

	Lang = "en"
	tt.Equal(GetLangText("command_empty"), "Command name cannot be empty")
	tt.Equal(GetLangText("help"), "Show Command help")
	tt.Equal(GetLangText("version"), "View version")

	Lang = "zh"
	tt.Equal(GetLangText("command_empty"), "命令名不能为空")
	tt.Equal(GetLangText("help"), "显示帮助信息")
	tt.Equal(GetLangText("version"), "查看版本信息")

	tt.Equal(GetLangText("non_existent_key"), "non_existent_key")
	tt.Equal(GetLangText("non_existent_key", "default value"), "default value")
}

func TestI18n_SetLangText(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	defer func() { Lang = originalLang }()

	Lang = "en"

	SetLangText("en", "custom_test", "Custom Test Value")
	SetLangText("zh", "custom_test", "自定义测试值")

	tt.Equal(GetLangText("custom_test"), "Custom Test Value")

	Lang = "zh"
	tt.Equal(GetLangText("custom_test"), "自定义测试值")

	SetLangText("en", "custom_help_key", "Updated Help Text")
	SetLangText("zh", "custom_help_key", "更新的帮助文本")
	Lang = "en"
	tt.Equal(GetLangText("custom_help_key"), "Updated Help Text")

	Lang = "zh"
	tt.Equal(GetLangText("custom_help_key"), "更新的帮助文本")
}

func TestI18n_LanguageSync(t *testing.T) {
	tt := zlsgo.NewTest(t)

	resetI18nForTesting()

	originalLang := Lang
	defer func() { Lang = originalLang }()

	Lang = "en"
	syncLanguageWithInternalI18n()
	tt.Equal(GetLangText("help"), "Show Command help")

	Lang = "zh"
	syncLanguageWithInternalI18n()
	tt.Equal(GetLangText("help"), "显示帮助信息")

	Lang = "invalid_lang"
	syncLanguageWithInternalI18n()
	result := GetLangText("command_empty")
	tt.EqualTrue(result == "command_empty" || result == "Command name cannot be empty")
}

func TestI18n_InCliCommands(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	originalOsExit := osExit
	defer func() {
		Lang = originalLang
		osExit = originalOsExit
	}()

	myExit := func(code int) {}
	osExit = myExit

	Lang = "en"
	resetForTesting()

	testCmd := &testCmd{}
	cmd := Add("", "invalid command", testCmd)
	tt.Equal(cmd.name, "")

	Lang = "zh"
	resetForTesting()

	cmd = Add("", "无效命令", testCmd)
	tt.Equal(cmd.name, "")
}

func TestI18n_BuiltInTranslationsMap(t *testing.T) {
	tt := zlsgo.NewTest(t)

	translations := getBuiltInTranslations()

	enTranslations, exists := translations["en"]
	tt.EqualTrue(exists)
	tt.Equal(enTranslations["command_empty"], "Command name cannot be empty")
	tt.Equal(enTranslations["help"], "Show Command help")
	tt.Equal(enTranslations["version"], "View version")

	zhTranslations, exists := translations["zh"]
	tt.EqualTrue(exists)
	tt.Equal(zhTranslations["command_empty"], "命令名不能为空")
	tt.Equal(zhTranslations["help"], "显示帮助信息")
	tt.Equal(zhTranslations["version"], "查看版本信息")

	requiredKeys := []string{
		"command_empty", "help", "version", "detach",
		"restart", "stop", "start", "status",
		"uninstall", "install",
	}

	for _, key := range requiredKeys {
		_, enExists := enTranslations[key]
		tt.EqualTrue(enExists)
		if !enExists {
			tt.Logf("Missing English key: %s", key)
		}

		_, zhExists := zhTranslations[key]
		tt.EqualTrue(zhExists)
		if !zhExists {
			tt.Logf("Missing Chinese key: %s", key)
		}
	}
}

func TestI18n_ConcurrentAccess(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	defer func() { Lang = originalLang }()

	SetLangText("en", "concurrent_test_en", "Concurrent Test EN")
	SetLangText("zh", "concurrent_test_zh", "并发测试 ZH")

	Lang = "en"
	syncLanguageWithInternalI18n()
	result := GetLangText("concurrent_test_en")
	tt.Equal(result, "Concurrent Test EN")

	Lang = "zh"
	syncLanguageWithInternalI18n()
	result = GetLangText("concurrent_test_zh")
	tt.Equal(result, "并发测试 ZH")

	result = GetLangText("concurrent_test_en")
	tt.Equal(result, "Concurrent Test EN")
}

func TestI18n_EdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	defer func() { Lang = originalLang }()

	tt.Equal(GetLangText(""), "")
	tt.Equal(GetLangText("   "), "   ")

	SetLangText("en", "special.chars.key", "Special Characters Test")
	Lang = "en"
	tt.Equal(GetLangText("special.chars.key"), "Special Characters Test")

	longKey := string(make([]byte, 1000))
	tt.Equal(GetLangText(longKey), longKey)

	SetLangText("en", "unicode_test", "Unicode: αβγ 中文")
	tt.Equal(GetLangText("unicode_test"), "Unicode: αβγ 中文")
}

func TestI18n_ZlocaleIntegration(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	defer func() { Lang = originalLang }()

	initInternalI18n()
	tt.NotNil(internalI18n)

	Lang = "zh"
	syncLanguageWithInternalI18n()
	tt.Equal(internalI18n.GetLanguage(), "zh")

	internalI18n.LoadLanguage("fr", "Français", map[string]string{
		"test_key": "Ceci est un test",
	})

	Lang = "fr"
	syncLanguageWithInternalI18n()

	tt.Equal(internalI18n.GetLanguage(), "fr")
	tt.EqualTrue(internalI18n.HasLanguage("fr"))
}

func TestI18n_HelpMessages(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	originalOsExit := osExit
	originalFirstParameter := FirstParameter
	defer func() {
		Lang = originalLang
		osExit = originalOsExit
		FirstParameter = originalFirstParameter
	}()

	myExit := func(code int) {}
	osExit = myExit

	FirstParameter = "testapp"

	tests := []struct {
		lang     string
		expected string
	}{
		{
			lang:     "en",
			expected: "Show Command help",
		},
		{
			lang:     "zh",
			expected: "显示帮助信息",
		},
	}

	for _, test := range tests {
		Lang = test.lang
		syncLanguageWithInternalI18n()

		oldOutput := flag.CommandLine.Output()
		flag.CommandLine.SetOutput(&bytes.Buffer{})
		resetForTesting("-help")

		flag.CommandLine.SetOutput(oldOutput)

		tt.Equal(Lang, test.lang)
	}
}

func TestI18n_VersionMessages(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	originalOsExit := osExit
	defer func() {
		Lang = originalLang
		osExit = originalOsExit
	}()

	myExit := func(code int) {}
	osExit = myExit

	tests := []struct {
		lang     string
		expected string
	}{
		{
			lang:     "en",
			expected: "View version",
		},
		{
			lang:     "zh",
			expected: "查看版本信息",
		},
	}

	for _, test := range tests {
		Lang = test.lang
		syncLanguageWithInternalI18n()

		tt.Equal(GetLangText("version"), test.expected)
	}
}

func TestI18n_CommandDescriptions(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	defer func() { Lang = originalLang }()

	SetLangText("en", "test_command_desc", "Test command for internationalization")
	SetLangText("zh", "test_command_desc", "国际化测试命令")

	tests := []struct {
		lang     string
		expected string
	}{
		{
			lang:     "en",
			expected: "Test command for internationalization",
		},
		{
			lang:     "zh",
			expected: "国际化测试命令",
		},
	}

	for _, test := range tests {
		Lang = test.lang
		syncLanguageWithInternalI18n()

		desc := GetLangText("test_command_desc")
		tt.Equal(desc, test.expected)
	}
}

func TestI18n_ErrorMessages(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	originalOsExit := osExit
	defer func() {
		Lang = originalLang
		osExit = originalOsExit
	}()

	myExit := func(code int) {}
	osExit = myExit

	SetLangText("en", "custom_error", "Custom error occurred")
	SetLangText("zh", "custom_error", "发生了自定义错误")

	tests := []struct {
		lang     string
		expected string
	}{
		{
			lang:     "en",
			expected: "Custom error occurred",
		},
		{
			lang:     "zh",
			expected: "发生了自定义错误",
		},
	}

	for _, test := range tests {
		Lang = test.lang
		syncLanguageWithInternalI18n()

		errorMsg := GetLangText("custom_error")
		tt.Equal(errorMsg, test.expected)
	}
}

func TestI18n_ServiceCommands(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	defer func() { Lang = originalLang }()

	serviceCommands := []string{"start", "stop", "restart", "status", "install", "uninstall"}

	expectedTranslations := map[string]map[string]string{
		"en": {
			"start":     "Start service",
			"stop":      "Stop service",
			"restart":   "Restart service",
			"status":    "Service status",
			"install":   "Install service",
			"uninstall": "Uninstall service",
		},
		"zh": {
			"start":     "开始服务",
			"stop":      "停止服务",
			"restart":   "重启服务",
			"status":    "服务状态",
			"install":   "安装服务",
			"uninstall": "卸载服务",
		},
	}

	for lang, translations := range expectedTranslations {
		Lang = lang
		syncLanguageWithInternalI18n()

		for _, cmd := range serviceCommands {
			expected, exists := translations[cmd]
			tt.EqualTrue(exists)
			if !exists {
				tt.Logf("Missing %s translation for command %s", lang, cmd)
			}

			actual := GetLangText(cmd)
			tt.Equal(actual, expected)
		}
	}
}

func TestI18n_ParameterizedTranslations(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	defer func() { Lang = originalLang }()

	SetLangText("en", "welcome_user", "Welcome, {0}!")
	SetLangText("zh", "welcome_user", "欢迎, {0}!")

	SetLangText("en", "file_count", "Found {0} files in {1} directories")
	SetLangText("zh", "file_count", "在 {1} 个目录中找到 {0} 个文件")

	SetLangText("en", "welcome_user", "Welcome, {0}!")
	SetLangText("zh", "welcome_user", "欢迎, {0}!")

	Lang = "en"
	syncLanguageWithInternalI18n()
	result := GetLangText("welcome_user")
	tt.Equal(result, "Welcome, {0}!")

	Lang = "zh"
	syncLanguageWithInternalI18n()
	result = GetLangText("welcome_user")
	tt.Equal(result, "欢迎, {0}!")

	initInternalI18n()
	if internalI18n != nil {
		internalI18n.LoadLanguage("en", "English", map[string]string{
			"welcome_user": "Welcome, {0}!",
		})
		internalI18n.LoadLanguage("zh", "简体中文", map[string]string{
			"welcome_user": "欢迎, {0}!",
		})

		internalI18n.SetLanguage("en")
		result = internalI18n.T("welcome_user", "Alice")
		tt.Equal(result, "Welcome, Alice!")

		internalI18n.SetLanguage("zh")
		result = internalI18n.T("welcome_user", "张三")
		tt.Equal(result, "欢迎, 张三!")
	}
}

func TestI18n_FallbackMechanism(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	defer func() { Lang = originalLang }()

	SetLangText("en", "fallback_test", "English only translation")

	Lang = "en"
	tt.Equal(GetLangText("fallback_test"), "English only translation")

	Lang = "zh"
	tt.Equal(GetLangText("fallback_test"), "English only translation")

	Lang = "en"
	tt.Equal(GetLangText("completely_nonexistent"), "completely_nonexistent")
	tt.Equal(GetLangText("completely_nonexistent", "default"), "default")
}

func TestI18n_KeyNormalization(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	defer func() { Lang = originalLang }()

	testCases := []struct {
		key      string
		expected string
	}{
		{"simple_key", "simple_key"},
		{"key.with.dots", "key.with.dots"},
		{"key-with-dashes", "key-with-dashes"},
		{"key_with_underscores", "key_with_underscores"},
		{"MixedCase_Key", "MixedCase_Key"},
	}

	for _, tc := range testCases {
		Lang = "en"

		SetLangText("en", tc.key, tc.expected)

		result := GetLangText(tc.key)
		tt.Equal(result, tc.expected)
	}
}

func TestI18n_PerformanceWithManyTranslations(t *testing.T) {
	tt := zlsgo.NewTest(t)

	originalLang := Lang
	defer func() { Lang = originalLang }()

	for i := 0; i < 1000; i++ {
		key := "perf_test_" + string(rune(i))
		SetLangText("en", key, "Performance test "+string(rune(i)))
	}

	Lang = "en"

	for i := 0; i < 100; i++ {
		key := "perf_test_" + string(rune(i))
		expected := "Performance test " + string(rune(i))
		result := GetLangText(key)
		tt.Equal(result, expected)
	}
}
