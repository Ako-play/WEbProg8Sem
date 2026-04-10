package config

import (
	"os"

	"github.com/joho/godotenv"
)

// LoadDotenv подгружает переменные из первого найденного файла .env (не перезаписывает уже заданные в ОС).
// Порядок: текущая директория, родитель (запуск из backend/), корень репозитория на уровень выше.
func LoadDotenv() {
	paths := []string{".env", "../.env", "../../.env"}
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		_ = godotenv.Load(path)
		return
	}
}
