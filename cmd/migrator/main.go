// cmd/migrator/main.go
package main

import (
	"errors"
	"flag"
	"fmt"

	// Библиотека для миграций
	"github.com/golang-migrate/migrate/v4"
	// Драйвер для выполнения миграций SQLite 3
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	// Драйвер для получения миграций из файлов
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Для работы с SQLite его можно скачать:
// go get github.com/mattn/go-sqlite3@v1.14.16

// Вспомогательная утилита миграции БД
// У выбранной нами библиотеки для миграций следующий формат нейминга миграций:
// <number>_<title>.<direction>.sql, где:
// number 		— используется для определения порядка применения миграций, они выполняются по возрастанию номеров.
//
//	Тут должно быть любое целое число — это может быть порядковый номер, timestamp и т.п. Я буду именовать по порядку: 1, 2, 3 и т.д.
//
// title 		— игнорируется библиотекой, он нужен только для людей, чтобы проще было ориентироваться в списке миграций
// direction 	— значение up или down. Файлы с параметром up в имени обновляют схему до новой версии, down — откатывают изменения.
// запуск утилиты:
// go run ./cmd/migrator --storage-path=./storage/sso.db --migrations-path=./migrations

// Если всё прошло хорошо, у вас должен появиться файл БД (./storage/sso.db) с актуальной схемой.
func main() {
	var storagePath, migrationsPath, migrationsTable string

	//получаем необходимые значения флагов запуска (аргументы командной строки)
	//путь до файла БД
	// Его достаточно, т.к. мы используем SQLite, други креды не нужны
	flag.StringVar(&storagePath, "storage-path", "", "path to storage")

	//путь до папки с миграциями
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migration")

	// Таблица, в которой будет храниться информация о миграциях. Она нужна
	// для того, чтобы понимать, какие миграции уже применены, а какие нет.
	// Дефолтное значение - 'migrations'.
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")

	flag.Parse() //выполняем парсинг флагов

	//валидация параметров
	if storagePath == "" {
		// Простейший способ обработки ошибки ;)
		// При необходимости, можете выбрать более подходящий вариант.
		// Меня паника пока устраивает, поскольку это вспомогательная утилита.
		panic("storage-path is required")
	}

	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	// Создаем объект мигратора, передав креды нашей БД
	// здесь вынесли нейминг таблицы для миграций в отдельный флаг
	// с помощью параметра ?x-migrations-table=%s.
	// Обычно это не обязательно, но я буду хранить отдельный набор миграций для тестов,
	// и информация о них будет храниться в отдельной таблице. Но об этом в разделе про тестирование.
	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", storagePath, migrationsTable),
	)
	if err != nil {
		panic(err)
	}

	// Выполняем миграции до последней версии
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}

		panic(err)
	}

}
