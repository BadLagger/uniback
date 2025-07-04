# Проект банковского приложения для обработки запросов #

В рамках проекта реализована REST API для банковского сервиса со следующими функциями:

- Регистрация пользователей с проверкой уникальности.
- Аутентификация пользователей.
- Создание банковских счетов и управление ими.
- Операции с картами: генерация, просмотр, оплата.
- Переводы между счетами и пополнение баланса.
- Реализовано шифрование данных.
- Кредитные операции: оформление кредита, график платежей (todo).
- Аналитика финансовых операций (todo).
- Интеграция с внешними сервисами(todo):
- Центральный банк РФ — для определения ключевой ставки(todo)
- SMTP — для отправки уведомлений по электронной почте(todo)

# Стэк #

- Язык: Go 1.23+.
- Работа с БД: PostgreSQL + lib/pq.
- Аутентификация: JWT (golang-jwt/jwt/v5).
- Шифрование: bcrypt, HMAC-SHA256, PGP.

# REST API #

POST /register - регистрация новых пользователей
```
{
    "username": "FirstOne",
    "password": "123user",
    "email": "mycool@mail.com",
    "phone": "+79993332255"
}
```

POST /login - аутентификация
```
{
    "username": "FirstOne",
    "password": "123user"
}
```

GET /accounts - полчение списка всех счетов пользователя

POST /accounts/new - создание нового счёта
```
{
    "account_type": "credit"
}
```

POST /accounts/deposit - пополнение счёта
```
{
    "account_number": "40881010875173177486",
    "amount": 100.00
}
```
POST /accounts/withdrawal - списание со счёта
```
{
    "account_number": "40881010875173177486",
    "amount": 100.00
}
```
POST /accounts/transfer - перевод между счетами
```
{
    "source_account_number": "40881010875173177486",
    "destination_account_number": "40881025286971573351",
    "amount": 100.00
}
```
POST /cards/new - новая карта с привязкой к счёту
```
{
    "account_number": "40881066752914644069"
}
```

# Шифрование #

В процессе аутентификации генерируется JWT, который необходимо отсылать во всех запросах кроме registration и login.

Пароли пользователей шифруется с использованием bcrytp и хранятся в БД.

При создании новых карт все данные карт шифруются с помощью PGP и сохраняются в БД. Если при запуске приложения ключи не существуют то они автоматически будут сгенерированы.

# Конфиги #

Конфиги задаются с помощью переменных окружения. Ознакомиться со списком можно в файле util/config.go

# БД #

Для запуска приложения требуется развёрнутый PostgreSQL сервер. Для доступа к БД требуется указать соответсвующие переменные окружения. Все таблицы будут автоматически созданы с помощью файлов миграций.

# Тестирование #

1. Нужно для начала запустить сервер PostgreSQL и узнать порт и адрес (например, адрес 192.168.0.33 порт 9997)
2. Создать БД, пользователя и пароль под которым приложение будет подключаться к БД. (например, имя БД bank, пользователь user1, пароль 333)
3. Перейти в папку с проектом и скомпилировать приложение
```
go build ./
```
4. Установите нужные параметры приложения:
 Для Linux 
```
DB_HOST=192.168.0.33 DB_PORT=9997 DB_NAME=bank DB_USERNAME=user1 DB_PASSWORD=333
```
Для Windows
```
setx DB_HOST "192.168.0.33"
setx DB_PORT "9997"
setx DB_NAME "bank"
setx DB_USERNAME "user1"
setx DB_PASSWORD "333"
```
_Примечание: В Windows возможно после установки переменых придётся перезапустить командную строку_

5. Запустите POSTMAN и импортируйте json с запросами из папки postman в корне проекта.
6. Запустите скомпилированное приложение из командрой строки. В случае успешного запуска оно последнее что вы увидите в консоли: "2025/07/02 20:16:54 [INFO]: Try to start server..."
7. После этого можно с помощью POSTMAN отсылать запросы. В консоли можно наблюдать лог (по умолчанию уровень логирования DEBUG). Можно изменять перед компиляцией. Это первая строка в main().

# Выход #

Чтобы выйти из приложения надо просто нажать Ctrl+C (или каким-то иным способом отправить  SIGTERM). При этом приложение аккуратно закроется.

