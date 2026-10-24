# tailor

`tailor` — инструмент для сервисов, которые работают с контрактами из
[service-task-shemas](../service-task-shemas/README.md). Он генерирует типы для
нужных полей контракта и клиент чтения задач.

Утилита получает схему по `slug` и `version` и выбирает указанные поля,
включая унаследованные. Типы полей и их обязательность берутся из схемы.

Адрес сервиса задаётся переменной окружения `SERVICE_TASK_SCHEMAS_URL`,
а параметры генерации — в директиве `go:generate`:

```go
//go:generate tailor --slug standard-task --version 1 --fields title,priority,status --out internal/generated/task/types.go
```

Типы записываются в файл из `--out`; рядом создаются `client.go` и
`http_client.go`. Для `--out` нельзя использовать эти два имени.
Генерация завершается до начала записи файлов. Каждый файл заменяется атомарно.

Клиент передаёт закреплённый контракт `standard-task@1` в заголовке
`X-Schema` и возвращает только выбранные поля ответа:

```go
client, err := task.NewClient("http://localhost:8080")
if err != nil {
    return err
}

value, err := client.Get(ctx, taskID)
```

Для каждого контракта используй отдельный выходной пакет: имена `Client`,
`client.go` и `http_client.go` общие для всех сгенерированных клиентов.
