package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Тестовые структуры и кастомные типы
type TestStruct struct {
	ID int
}

type CustomInt int

// Имитируем сгенерированный через mockery мок для демонстрации
type MockRepository struct {
	mock.Mock
}

func TestGetTypeName_WithTestify(t *testing.T) {
	str := "hello"
	pStr := &str
	ppStr := &pStr // Двойной указатель

	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"Nil value", nil, "nil"},
		{"Simple string", "hello", "string"},
		{"Simple int", 42, "int"},
		{"Custom int type", CustomInt(10), "CustomInt"},
		{"Struct", TestStruct{}, "TestStruct"},
		{"Pointer to struct", &TestStruct{}, "TestStruct"},
		{"Double pointer to string", ppStr, "string"},
		{"Slice of structs", []TestStruct{}, "[]utils.TestStruct"},
		{"Map", map[string]int{}, "map[string]int"},
		{"Anonymous struct", struct{ Name string }{}, "struct { Name string }"},
		{"Mockery generated object", &MockRepository{}, "MockRepository"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Используем testify assert вместо ручных if-условий
			assert.Equal(t, tt.expected, GetTypeName(tt.input))
		})
	}
}

func TestGetFullTypeName_WithTestify(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"Nil value", nil, "nil"},
		{"Simple int", 42, "int"},
		{"Custom struct with package path", TestStruct{}, "github.com/ElfAstAhe/go-service-template/pkg/utils.TestStruct"},
		{"Pointer to custom struct", &TestStruct{}, "github.com/ElfAstAhe/go-service-template/pkg/utils.TestStruct"},
		{"Slice of custom structs", []TestStruct{}, "[]utils.TestStruct"},
		{"Anonymous struct", struct{ X int }{}, "struct { X int }"},
		{"Mock object full name", &MockRepository{}, "github.com/ElfAstAhe/go-service-template/pkg/utils.MockRepository"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetFullTypeName(tt.input))
		})
	}
}

func TestIsNil_WithTestify(t *testing.T) {
	var emptyPtr *TestStruct = nil
	var anyWithNilPtr any = emptyPtr // Ловушка Go: интерфейс хранит тип (*TestStruct) и nil-значение

	var emptySlice []int = nil
	var emptyMap map[string]int = nil
	var emptyChan chan int = nil
	var emptyFunc func() = nil

	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"Pure nil", nil, true},
		{"Nil pointer", emptyPtr, true},
		{"Interface containing nil pointer", anyWithNilPtr, true}, // IsNil должен раскусить это
		{"Nil slice", emptySlice, true},
		{"Nil map", emptyMap, true},
		{"Nil chan", emptyChan, true},
		{"Nil func", emptyFunc, true},
		{"Not nil struct", TestStruct{}, false},
		{"Not nil pointer to struct", &TestStruct{}, false},
		{"Not nil slice", []int{1, 2}, false},
		{"Not nil map", map[string]int{"a": 1}, false},
		{"Int zero value", 0, false},
		{"String zero value", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected {
				assert.True(t, IsNil(tt.input), "Expected %s to be recognized as nil", tt.name)
			} else {
				assert.False(t, IsNil(tt.input), "Expected %s to NOT be recognized as nil", tt.name)
			}
		})
	}
}
