package utils

import (
	"testing"
)

func TestValidateLanguage(t *testing.T) {
	// Проверяем, что списки не пустые после init()
	if len(SortedLanguages) == 0 {
		t.Fatal("SortedLanguages is empty after initialization")
	}

	tests := []struct {
		name    string
		lang    string
		wantErr bool
	}{
		{
			name:    "Valid Russian language code",
			lang:    "ru",
			wantErr: false,
		},
		{
			name:    "Valid English language code",
			lang:    "en",
			wantErr: false,
		},
		{
			name:    "Invalid short code",
			lang:    "xx",
			wantErr: true,
		},
		{
			name:    "Invalid uppercase code (strict check)",
			lang:    "RU",
			wantErr: true,
		},
		{
			name:    "Empty string",
			lang:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLanguage(tt.lang)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLanguage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLangSortedListAreActuallySorted(t *testing.T) {
	// Проверяем корректность сортировки языков
	for i := 1; i < len(SortedLanguages); i++ {
		if SortedLanguages[i-1] > SortedLanguages[i] {
			t.Errorf("SortedLanguages is not sorted at index %d: %s > %s", i, SortedLanguages[i-1], SortedLanguages[i])
		}
	}
}
