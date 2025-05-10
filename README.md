# SubPub - Реализация шаблона Publisher-Subscriber на Go

![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)
[![Лицензия](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Alexandr-Fox/subpub)](https://goreportcard.com/report/github.com/Alexandr-Fox/subpub)
[![Coverage Status](https://coveralls.io/repos/github/Alexandr-Fox/subpub/badge.svg)](https://coveralls.io/github/Alexandr-Fox/subpub)

SubPub - это легковесная, потокобезопасная реализация шаблона Publisher-Subscriber на языке Go, предназначенная для эффективного распределения сообщений между компонентами системы.

## Особенности

- 🚀 **Асинхронная доставка** - Подписчики не блокируют издателей
- 🔒 **Потокобезопасность** - Безопасное использование из множества горутин
- ⏱️ **Грациозное завершение** - Поддержка контекста для управления shutdown
- 🏎 **Высокая производительность** - Оптимизировано для минимальных задержек
- � **Простой API** - Легкая интеграция в существующие проекты
- 🧪 **Полное тестирование** - Комплексное покрытие тестами

## Установка

```bash
go get github.com/Alexandr-Fox/subpub
```

## Быстрый старт

```go
package main

import (
	"context"
	"fmt"
	"github.com/Alexandr-Fox/subpub"
	"time"
)

func main() {
	// Создаем экземпляр SubPub
	bus := subpub.NewSubPub()

	// Подписываемся на тему
	sub, err := bus.Subscribe("новости", func(msg interface{}) {
		fmt.Printf("Получены новости: %v\n", msg)
	})
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	// Публикуем сообщение
	err = bus.Publish("новости", "Срочно: Вышел Go 2.0!")
	if err != nil {
		panic(err)
	}

	// Даем время на обработку
	time.Sleep(100 * time.Millisecond)

	// Грациозное завершение
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	bus.Close(ctx)
}
```

## Справочник API

### `NewSubPub() SubPub`

Создает новый экземпляр SubPub.

### `Subscribe(тема string, обработчик MessageHandler) (Subscription, error)`

Подписывается на сообщения указанной темы. Функция-обработчик будет вызываться для каждого опубликованного сообщения.

### `Publish(тема string, сообщение interface{}) error`

Публикует сообщение для всех подписчиков указанной темы.

### `Close(ctx context.Context) error`

Грациозно завершает работу, ожидая обработки всех сообщений или отмены контекста.

### `Unsubscribe()`

Отменяет подписку.

## Примеры использования

- Событийно-ориентированная архитектура
- Общение между микросервисами
- Системы реального времени
- Распределение логов
- Взаимодействие слабосвязанных компонентов

## Лицензия

Проект распространяется под лицензией MIT - подробности в файле [LICENSE](LICENSE).

---

<p align="center">
  <i>✨ Эффективный обмен сообщениями для Go-приложений ✨</i>
</p>