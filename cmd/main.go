package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Создаем контекст, который требуется для всех операций
	ctx := context.Background()

	// Создаем клиент для подключения к Redis
	// Адрес: localhost:6379 (стандартный порт)
	// Пароль: "" (пустая строка, так как вы его не устанавливали)
	// База данных: 0 (используется по умолчанию)
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",   // Адрес вашего Redis-сервера в WSL[citation:3][citation:7][citation:9]
		Password: "NaS16cyd93@go_7m", // Пароль не установлен[citation:1][citation:3]
		DB:       0,                  // Используем базу данных по умолчанию[citation:1][citation:3]
	})
	defer rdb.Close() // Закрываем соединение при завершении программы

	// Проверяем подключение командой PING
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Не удалось подключиться к Redis: %v", err)
	}
	fmt.Println("Успешное подключение:", pong) // Должно вывести "PONG"

	// Пример: запись и чтение данных
	// Записываем значение
	err = rdb.Set(ctx, "my_key", "my_value", 0).Err()
	if err != nil {
		log.Fatalf("Ошибка при записи: %v", err)
	}

	// Читаем значение
	val, err := rdb.Get(ctx, "my_key").Result()
	if err != nil {
		log.Fatalf("Ошибка при чтении: %v", err)
	}
	fmt.Println("Прочитано значение:", val) // Должно вывести "my_value"
}
