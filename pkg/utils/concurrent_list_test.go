package utils

import (
	"sync"
	"testing"
)

func TestConcurrentList_BasicOperations(t *testing.T) {
	list := NewConcurrentList[string]()

	// 1. Проверяем работу с пустым списком
	if list.Len() != 0 {
		t.Errorf("Ожидалась длина 0, получили %d", list.Len())
	}
	if _, ok := list.Get(0); ok {
		t.Error("Get на пустом списке должен возвращать false")
	}

	// 2. Добавление и получение элементов
	list.Append("first")
	list.Append("second")

	if list.Len() != 2 {
		t.Errorf("Ожидалась длина 2, получили %d", list.Len())
	}

	if val, ok := list.Get(0); !ok || val != "first" {
		t.Errorf("Некорректное значение по индексу 0: %q", val)
	}

	if _, ok := list.Get(5); ok {
		t.Error("Get с выходом за границы должен возвращать false")
	}
	if _, ok := list.Get(-1); ok {
		t.Error("Get с отрицательным индексом должен возвращать false")
	}
}

func TestConcurrentList_SnapshotAndIterators(t *testing.T) {
	list := NewConcurrentList[int]()
	list.Append(10)
	list.Append(20)
	list.Append(30)

	// 1. Тест Snapshot
	snap := list.Snapshot()
	if len(snap) != 3 || snap[0] != 10 || snap[2] != 30 {
		t.Errorf("Некорректный Snapshot: %v", snap)
	}

	// 2. Тест итератора All()
	allCount := 0
	for idx, val := range list.All() {
		if idx == 0 && val != 10 {
			t.Errorf("Ожидалось 10, получили %d", val)
		}
		allCount++
	}
	if allCount != 3 {
		t.Errorf("Итератор All обошел %d элементов вместо 3", allCount)
	}

	// 3. Тест итератора Values() с досрочным выходом (break)
	breakCount := 0
	for val := range list.Values() {
		breakCount++
		if val == 20 {
			break // Проверяем, что yield корректно обрабатывает прерывание цикла
		}
	}
	if breakCount != 2 {
		t.Errorf("Ожидалось прерывание на 2-м элементе, обошли %d", breakCount)
	}
}

func TestConcurrentList_DataRace(t *testing.T) {
	list := NewConcurrentList[int]()
	var wg sync.WaitGroup

	workers := 10
	iterations := 500

	// Параллельные писатели
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				list.Append(workerID*iterations + j)
			}
		}(i)
	}

	// Параллельные читатели через итераторы
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				// Читаем через All() во время активной записи
				for _, val := range list.All() {
					_ = val
				}
				// Читаем через Values()
				for val := range list.Values() {
					_ = val
				}
			}
		}()
	}

	wg.Wait()

	expectedLen := workers * iterations
	if list.Len() != expectedLen {
		t.Errorf("Ожидаемая финальная длина %d, получили %d", expectedLen, list.Len())
	}
}
