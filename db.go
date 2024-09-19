package main

//go:generate moq -out=mocked_db.go . DB
type DB interface {
	SaveData(url, pattern, data string) error
}
