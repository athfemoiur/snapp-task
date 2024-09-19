package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	Conn *sql.DB
}

func NewSQLiteDB(dataSourceName string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS matches (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        url TEXT,
        pattern TEXT,
        data TEXT
    );`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}
	return &SQLiteDB{Conn: db}, nil
}

func (db *SQLiteDB) SaveData(url, pattern, data string) error {
	query := "INSERT INTO matches (url, pattern, data) VALUES (?, ?, ?)"
	_, err := db.Conn.Exec(query, url, pattern, data)
	return err
}

func (db *SQLiteDB) Close() error {
	return db.Conn.Close()
}
