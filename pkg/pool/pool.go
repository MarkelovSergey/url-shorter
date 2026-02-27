// Package pool предоставляет generic-пул объектов для типов с методом Reset().
package pool

import "sync"

// Resetter — интерфейс, который должны реализовывать типы для хранения в Pool.
// Метод Reset() должен сбрасывать состояние объекта к начальным значениям.
type Resetter interface {
	Reset()
}

// Pool — generic-обёртка над sync.Pool, хранящая объекты, реализующие Resetter.
// Перед возвратом объекта в пул вызывается метод Reset() для очистки его состояния.
type Pool[T Resetter] struct {
	pool sync.Pool
	new  func() T
}

// New создаёт и возвращает указатель на новый Pool.
// Параметр newFunc — функция-конструктор, создающая новые экземпляры типа T.
func New[T Resetter](newFunc func() T) *Pool[T] {
	p := &Pool[T]{
		new: newFunc,
	}

	p.pool.New = func() any {
		return newFunc()
	}

	return p
}

// Get извлекает объект из пула.
// Если пул пуст, создаётся новый объект с помощью функции-конструктора.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put возвращает объект в пул после вызова его метода Reset().
// Это гарантирует, что объект будет в чистом состоянии при следующем извлечении.
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}
