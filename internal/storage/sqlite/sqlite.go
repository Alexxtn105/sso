// internal/storage/sqlite/sqlite.go

package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"grpc-service-ref/internal/domain/models"
	"grpc-service-ref/internal/storage"

	"github.com/mattn/go-sqlite3"
)

// создаём файл, в котором опишем тип Storage и конструктор для него:
type Storage struct {
	db *sql.DB
}

// Конструктор для хранилища
func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	// Указываем путь до файла БД
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// Для хранилища нужно реализовать три метода: SaveUser(), User(), App()
func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.sqlite.SaveUser"

	// запрос на добавление пользователя
	stmt, err := s.db.Prepare("INSERT INTO users(email, pass_hash) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
	}

	// Выполняем запрос, передав параметры
	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error

		// Небольшое кунг-фу для выявления ошибки ErrConstraintUnique
		// Суть этой конструкции в том, чтобы выявить ошибку нарушения констрэинта уникальности по email,
		// другими словами — когда мы пытаемся добавить в таблицу запись с параметром email,
		// который уже есть в таблице. Если мы её выявляем, то наружу нужно вернуть ошибку storage.ErrUserExists,
		// которую мы подготовили заранее.
		// Это нужно для того, чтобы вне зависимости от используемой БД всегда можно было определить
		// попытку добавления дубликата имеющегося пользователя.
		// Нам это понадобится в других слоях, чтобы отреагировать на подобный случай правильным образом.
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Получаем ID созданной записи
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// User returns user by email.
func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.sqlite.User"

	stmt, err := s.db.Prepare("SELECT id, email, pass_hash FROM users WHERE email = ?")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, email)

	var user models.User
	// Здесь мы аналогично определяем ошибку, но на этот раз нас интересует sql.ErrNoRows,
	// она означает что мы не смогли найти соответствующую запись.
	// В этом случае мы вернём наружу storage.ErrUserNotFound
	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// App returns app by id.
func (s *Storage) App(ctx context.Context, id int) (models.App, error) {
	const op = "storage.sqlite.App"

	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, id)

	var app models.App

	// Как и в предыдущих случаях, в случае отсутствия записи (sql.ErrNoRows),
	// возвращаем наружу storage.ErrAppNotFound.
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}

		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}

// Added by Alexx - Закрытие БД (аналогично GracefulStop для gRPC)
func (s *Storage) Close() error {
	const op = "storage.sqlite.Close"
	//s.Log.With(slog.String("op", op)).Info("stopping gRPC server", slog.Int("port", a.port))
	err := s.db.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
