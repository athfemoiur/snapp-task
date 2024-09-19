package main

import "time"

const serverAddress = ":8080"
const dbPath = "data.db"

func main() {
	db, err := NewSQLiteDB(dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	checkerFactory := func(url, pattern string, db DB) UrlChecker {
		return NewUrlCheckerImpl(url, pattern, db)
	}
	schedulerFactory := func(url, pattern string, interval time.Duration, db DB) CheckScheduler {
		return NewCheckSchedulerImpl(url, pattern, interval, db, checkerFactory)
	}
	server := NewAPIServer(serverAddress, db, schedulerFactory)
	if runErr := server.Run(); runErr != nil {
		panic(runErr)
	}
}
