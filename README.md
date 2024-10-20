# DockerGo

# Запуск
1. make env
2. sudo make up || sudo make compose

# Описание?

## Аутентификация

Клиент обращается к системе X, в которой он авторизован, получает ticket. Система X хранит у себя пару значений "тикет"-"айпи". Клиент делает запрос на подключение к вебсокет серверу ws://localhost:6060/ws?ticket=ticket. Вебсокет сервер обращается к системе X проверяет валидность ticket и сравнивает айпи.

## Сообщения

Сообщения от сервера для клиентов получаются из очереди RabbitMQ output в виде
```
{
    mode: [
        all // all connected
        touser // userid requered
        group // message for specific connected group - @ToDo
    ]
    userid: string // optional, required if mode is touser
    message: string // data for websocket client
}
```
## Примеры
```
{
    "mode": "touser",
    "userid": "1",
    "message": "test 1"
}
```
```
{
    "mode": "all",
    "message": "test 2"
}
```
