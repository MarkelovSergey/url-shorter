package main

import (
	"log"
	"os"
)

// В пакете main функция main может использовать log.Fatal и os.Exit
func main() {
	// Эти вызовы должны быть разрешены
	if false {
		log.Fatal("fatal error in main")
	}
	if false {
		os.Exit(1)
	}

	// Но panic все равно должен быть обнаружен
	if false {
		panic("panic in main") // want "использование встроенной функции panic"
	}
}

// Но в других функциях пакета main они запрещены
func helper() {
	log.Fatal("error") // want "вызов log.Fatal вне функции main пакета main"
	os.Exit(1)         // want "вызов os.Exit вне функции main пакета main"
	panic("test")      // want "использование встроенной функции panic"
}
