package pool

import (
	"sync"
	"testing"
)

// testStruct — тестовая структура, реализующая интерфейс Resetter
type testStruct struct {
	Value  int
	Name   string
	Slice  []int
	Map    map[string]int
	Active bool
}

func (t *testStruct) Reset() {
	if t == nil {
		return
	}
	t.Value = 0
	t.Name = ""
	t.Slice = t.Slice[:0]
	clear(t.Map)
	t.Active = false
}

func newTestStruct() *testStruct {
	return &testStruct{
		Slice: make([]int, 0, 10),
		Map:   make(map[string]int),
	}
}

func TestNew(t *testing.T) {
	p := New(newTestStruct)
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.new == nil {
		t.Fatal("new function is nil")
	}
}

func TestPool_Get(t *testing.T) {
	p := New(newTestStruct)

	obj := p.Get()
	if obj == nil {
		t.Fatal("Get() returned nil")
	}

	// Проверяем, что это новый объект с нулевыми значениями
	if obj.Value != 0 || obj.Name != "" || obj.Active != false {
		t.Error("Get() returned object with non-zero values")
	}
}

func TestPool_Put(t *testing.T) {
	p := New(newTestStruct)

	obj := p.Get()
	obj.Value = 42
	obj.Name = "test"
	obj.Slice = append(obj.Slice, 1, 2, 3)
	obj.Map["key"] = 100
	obj.Active = true

	p.Put(obj)

	// Получаем объект обратно (должен быть сброшен)
	obj2 := p.Get()

	// Примечание: sync.Pool не гарантирует возврат того же объекта,
	// но если это тот же объект, он должен быть сброшен
	if obj2.Value != 0 {
		t.Errorf("Value not reset, got %d", obj2.Value)
	}
	if obj2.Name != "" {
		t.Errorf("Name not reset, got %s", obj2.Name)
	}
	if obj2.Active != false {
		t.Error("Active not reset")
	}
}

func TestPool_ResetCalledOnPut(t *testing.T) {
	// Проверяем, что объект сбрасывается при вызове Put
	p := New(newTestStruct)
	obj := p.Get()
	obj.Value = 100
	obj.Name = "before reset"

	// Put должен вызвать Reset
	p.Put(obj)

	// После Put объект должен быть сброшен
	// Объект в пуле должен быть в начальном состоянии
	if obj.Value != 0 || obj.Name != "" {
		t.Error("Reset was not called on Put")
	}
}

func TestPool_ConcurrentAccess(t *testing.T) {
	p := New(newTestStruct)

	var wg sync.WaitGroup
	iterations := 100
	goroutines := 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				obj := p.Get()
				obj.Value = id*1000 + j
				obj.Name = "test"
				obj.Slice = append(obj.Slice, j)
				p.Put(obj)
			}
		}(i)
	}

	wg.Wait()
}

// Бенчмарк-тесты
func BenchmarkPool_GetPut(b *testing.B) {
	p := New(newTestStruct)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj := p.Get()
		obj.Value = i
		p.Put(obj)
	}
}

func BenchmarkPool_Parallel(b *testing.B) {
	p := New(newTestStruct)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := p.Get()
			obj.Value = 42
			obj.Name = "test"
			p.Put(obj)
		}
	})
}

// trackingResetter — структура с методом Reset для отслеживания вызовов
type trackingResetter struct {
	value int
}

func (tr *trackingResetter) Reset() {
	tr.value = 0
}

func TestPool_GenericConstraint(t *testing.T) {
	// Тест проверяет корректность работы generic-ограничения
	p := New(func() *trackingResetter {
		return &trackingResetter{value: 0}
	})

	obj := p.Get()
	obj.value = 42
	p.Put(obj)

	// Объект должен быть сброшен
	if obj.value != 0 {
		t.Error("trackingResetter not reset properly")
	}
}
