package a

import (
	"log"
	"os"
)

// Тестируем использование panic
func TestPanic() {
	panic("test panic") // want "использование встроенной функции panic"
}

// Тестируем использование log.Fatal вне main
func TestLogFatal() {
	log.Fatal("fatal error") // want "вызов log.Fatal вне функции main пакета main"
}

// Тестируем использование os.Exit вне main
func TestOsExit() {
	os.Exit(1) // want "вызов os.Exit вне функции main пакета main"
}

// Тестируем комбинированное использование
func TestCombined() {
	if true {
		panic("test") // want "использование встроенной функции panic"
	}
	log.Fatal("error") // want "вызов log.Fatal вне функции main пакета main"
}

// Обычная функция без проблем
func TestNormal() {
	println("normal function")
}
