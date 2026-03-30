# План рефакторинга legacy кода (ErrNilOrder)

## Шаги:
- [x] 1. Отрефакторить backend/internal/pkg/kafka/handlers/create_order_handler.go (удалить ErrNilOrder, исправить опечатку в ErrCreateOrder, локализовать ошибки).

- [x] 2. Отрефакторить backend/internal/pkg/kafka/consumer.go (удалить глобальные ошибки, упростить handleMessageError).
- [x] 3. Проверить компиляцию: go build ./backend/cmd/app ./backend/cmd/consumer (OK).
- [ ] 4. Завершить задачу.

Текущее: Рефакторинг завершён, компиляция OK.
