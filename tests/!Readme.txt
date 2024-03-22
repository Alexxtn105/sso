 Функциональные тесты мы положим в отдельную папку — tests/. 
 Структура будет выглядеть так:

./tests
├── auth_register_login_test.go ... Тест-кейсы для проверки Login 
├── some_other_case_test.go ....... Какие-то другие кейсы
├── migrations .................... Миграции для тестов
│   └── 1_init_apps.up.sql
└── suite
    └── suite.go .................. Подготовка всего необходимого для тестов

Подробнее:

tests               — храним сами файлы тестов с конкретными кейсами 
                      (у них обязательно должен быть суффикс _test.go)
tests/migrations    — дополнительные миграции, которые нужны только для тестов. 
                      Обычно я их использую для инициализации БД самыми необходимыми 
                      данными. Например, здесь мы напишем миграцию для добавления 
                      тестового приложения в таблицу apps.
tests/suite         — здесь мы будем готовить всё, что необходимо каждому тесту. 
                      Например, соединение с БД, создание gRPC-клиента и др.

Выполнять миграции из папки tests/migrations будем утилитой миграции ./cmd/migrator, но с другими параметрами:
go run ./cmd/migrator --storage-path=./storage/sso.db --migrations-path=./tests/migrations --migrations-table=migrations_test

То есть, мы используем для них отдельную таблицу migrations_test, чтобы эти тестовые миграции были независимы от основных. Также указываем, соответственно, другой путь до файлов миграций.

---------------------------------------------------------
Запуск тестов

Если у вас свежая чистая БД, то первым делом прогоняем основные миграции:
go run ./cmd/migrator --storage-path=./storage/sso.db --migrations-path=./migrations

Затем тестовые миграции:
go run ./cmd/migrator --storage-path=./storage/sso.db --migrations-path=./tests/migrations --migrations-table=migratТеперь запускаем приложение:
go run ./cmd/sso --config=./config/local_tests.yaml

Я обычно использую для тестов отдельный конфиг-файл, поэтому здесь указано соответствующее имя: local_tests.yaml. Но делать так не обязательно.

И наконец можем запустить сами тесты, указав где они находятся:
go test ./tests -count=1 -v

Параметр -count=1 — стандартный способ запустить тесты с игнорированием кэша, а -v добавить больше подробностей в вывод теста.

Если вы всё сделали правильно, вы должны увидеть примерно вот такой результат:
=== RUN   TestRegisterLogin_Login_HappyPath
=== PAUSE TestRegisterLogin_Login_HappyPath
=== CONT  TestRegisterLogin_Login_HappyPath
--- PASS: TestRegisterLogin_Login_HappyPath (0.15s)
PASS
ok      grpc-service-ref/tests  0.663s