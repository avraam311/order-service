## Обзор

Сервис заказов для хранения и получения заказа по id.
Стэк: Golang, Gin, Docker, PostgreSQL, React.

---

## Запуск

### 1. Сканируйте репозиторий

```bash
git clone https://github.com/avraam311/order-service
cd order-service
````

### 2. Создайте `.env` файл

Скопируйте шаблонный файл .env.example и заполните его нужными данными:

```bash
cp .env.example .env
```

Редактируйте `.env` как нужно.

---

### 3. Запуск проекта через docker-compose

Чтобы собрать и запустить проект

```bash
make up
```

---

### 4 Доступ к странице

Доступ к странице frontend:

```
http://localhost:3000
```

---------

## Примечания

* Убедитесь, что Docker и docker-compose установлены на вашей ос
* `.env` файл настроен правильно.
* Backend слушает на порту `8080`.
* Frontend слушает на порту `3000`.
* Напишите "make down", чтобы остановить работу системы
* Напишите в терминале "make producer", нажмите enter а затем введите свое сообщения в формате json, чтобы отправить его в кафку