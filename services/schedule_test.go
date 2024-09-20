package services_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"snapp-task/db"
	. "snapp-task/services"
	"time"
)

var _ = Describe("CheckSchedulerImpl", func() {
	var (
		mockDB         *db.DBMock
		scheduler      *CheckSchedulerImpl
		mockedChecker  *UrlCheckerMock
		checkerFactory UrlCheckerFactory
		testChan       chan struct{}
		testInterval   time.Duration
	)

	BeforeEach(func() {
		testInterval = 250 * time.Millisecond
		mockDB = &db.DBMock{SaveDataFunc: func(url string, pattern string, data string) error {
			return nil
		}}
		testChan = make(chan struct{})
		mockedChecker = &UrlCheckerMock{
			CheckDataFunc: func() error {
				defer func() {
					select {
					case testChan <- struct{}{}:
					default:
					}
				}()
				return nil
			},
		}
		checkerFactory = func(url, pattern string, db db.DB) UrlChecker {
			return mockedChecker
		}
	})

	Describe("ScheduleCheck", func() {
		It("should call CheckData on the UrlChecker", func() {
			scheduler = &CheckSchedulerImpl{
				Url:               "https://example.com",
				Pattern:           "test_pattern",
				Interval:          testInterval,
				Db:                mockDB,
				UrlCheckerFactory: checkerFactory,
			}

			go scheduler.ScheduleCheck()

			select {
			case <-testChan:
				Expect(mockedChecker.CheckDataCalls()).To(HaveLen(1))
			case <-time.After(testInterval + time.Millisecond):
				Fail("CheckData was not called on time")
			}
		})
	})
})
