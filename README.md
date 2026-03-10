## REST сервис для агрегации данных о подписках пользователей на онлайн сервисы.

Сервис разворачивается в докер контейнере, пробрасывается на хост машину на порт `8090` (по умолчанию)
Вторым контейнером разворачивается postgres, пробрасывается на хост машину на порт `5432` (по умолчанию)

### Ручки:
- Create `localhost:8090/subscriptions` POST
- Read `localhost:8090/subscriptions/uuid` GET 
- Update `localhost:8090/subscriptions` PATH 
- Delete `localhost:8090/subscriptions/uuid` DELETE 
- List `localhost:8090/subscriptions` GET 

### Для подсчета суммарной стоимости:
- Cost `localhost:8090/cost` GET 

### Swagger UI
- [localhost:8090/swagger/index.html](http://localhost:8090/swagger/index.html)

### Развертывание:
```bash
make test-server-deploy
```
или
```bash
docker compose up
```

### Миграция:
Инит БД и заполнение тестовыми данными
```bash
make postgres-migrate-up
```

Очистка БД
```bash
make postgres-migrate-down
```
