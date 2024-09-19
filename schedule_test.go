package main_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"snapp-task"
	"time"
)

var _ = Describe("CheckSchedulerImpl", func() {
	var (
		mockDB    *main.DBMock
		scheduler *main.CheckSchedulerImpl
	)

	BeforeEach(func() {
		mockDB = &main.DBMock{SaveDataFunc: func(url string, pattern string, data string) error {
			return nil
		}}

		mockChecker := &main.UrlCheckerMock{
			CheckDataFunc: func() error {
				return nil
			},
		}

		schedulerFactory := func(url, pattern string, db main.DB) main.UrlChecker {
			return mockChecker
		}

		scheduler = &main.CheckSchedulerImpl{
			Url:               "https://example.com",
			Pattern:           "test_pattern",
			Interval:          1 * time.Second,
			Db:                mockDB,
			UrlCheckerFactory: schedulerFactory,
		}
	})

	Describe("ScheduleCheck", func() {
		It("should call CheckData on the UrlChecker", func() {
			done := make(chan struct{})

			mockChecker := &main.UrlCheckerMock{
				CheckDataFunc: func() error {
					defer func() {
						select {
						case done <- struct{}{}:
						default:
						}
					}()
					return nil
				},
			}

			schedulerFactory := func(url, pattern string, db main.DB) main.UrlChecker {
				return mockChecker
			}
			scheduler = &main.CheckSchedulerImpl{
				Url:               "https://example.com",
				Pattern:           "test_pattern",
				Interval:          500 * time.Millisecond,
				Db:                mockDB,
				UrlCheckerFactory: schedulerFactory,
			}

			go scheduler.ScheduleCheck()

			select {
			case <-done:
				Expect(mockChecker.CheckDataCalls()).To(HaveLen(1))
			case <-time.After(501 * time.Millisecond):
				Fail("CheckData was not called on time")
			}
		})
	})
})
