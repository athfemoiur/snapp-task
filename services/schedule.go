package services

import (
	_ "github.com/mattn/go-sqlite3"
	"log"
	"snapp-task/db"
	"time"
)

//go:generate moq -out=mocked_scheduler.go . CheckScheduler
type CheckScheduler interface {
	ScheduleCheck()
}

type SchedulerFactory func(url, pattern string, interval time.Duration, db db.DB) CheckScheduler

type CheckSchedulerImpl struct {
	Url               string
	Pattern           string
	Interval          time.Duration
	Db                db.DB
	UrlCheckerFactory UrlCheckerFactory
}

func NewCheckSchedulerImpl(url, pattern string, interval time.Duration, db db.DB, urlCheckerFactory UrlCheckerFactory) CheckScheduler {
	return &CheckSchedulerImpl{Url: url, Pattern: pattern, Interval: interval, Db: db, UrlCheckerFactory: urlCheckerFactory}
}

func (cs CheckSchedulerImpl) ScheduleCheck() {
	ticker := time.NewTicker(cs.Interval)
	defer ticker.Stop()

	errorChan := make(chan error)

	for {
		select {
		case <-ticker.C:
			checker := cs.UrlCheckerFactory(cs.Url, cs.Pattern, cs.Db)
			go func() {
				if err := checker.CheckData(); err != nil {
					errorChan <- err
				}
			}()

		case err := <-errorChan:
			log.Printf("Error encountered: %v\n", err)
		}
	}
}
