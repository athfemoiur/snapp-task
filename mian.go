package main

import (
	"snapp-task/api"
	"snapp-task/db"
	"snapp-task/services"
	"time"
)

const serverAddress = ":8080"
const dbPath = "data.db"

func main() {
	sqliteDB, err := db.NewSQLiteDB(dbPath)
	if err != nil {
		panic(err)
	}
	defer sqliteDB.Close()
	checkerFactory := func(url, pattern string, db db.DB) services.UrlChecker {
		return services.NewUrlCheckerImpl(url, pattern, db)
	}
	schedulerFactory := func(url, pattern string, interval time.Duration, db db.DB) services.CheckScheduler {
		return services.NewCheckSchedulerImpl(url, pattern, interval, db, checkerFactory)
	}
	server := api.NewAPIServer(serverAddress, sqliteDB, schedulerFactory)
	if runErr := server.Run(); runErr != nil {
		panic(runErr)
	}
}
