package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"snapp-task/api"
	"snapp-task/db"
	"snapp-task/services"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("APIServer", func() {
	var (
		server         *api.APIServer
		recorder       *httptest.ResponseRecorder
		requestPayload api.RequestMessage
		mockScheduler  *services.CheckSchedulerMock
	)

	BeforeEach(func() {
		mockScheduler = &services.CheckSchedulerMock{
			ScheduleCheckFunc: func() {},
		}
		schedulerFactory := func(url, pattern string, interval time.Duration, db db.DB) services.CheckScheduler {
			return mockScheduler
		}
		mockDB := &db.DBMock{SaveDataFunc: func(url string, pattern string, data string) error {
			return nil
		}}
		server = api.NewAPIServer(":8080", mockDB, schedulerFactory)
		recorder = httptest.NewRecorder()
	})

	JustBeforeEach(func() {
		payload, _ := json.Marshal(requestPayload)
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		handler := http.HandlerFunc(server.HandleRequest)
		handler.ServeHTTP(recorder, req)
	})

	Describe("HandleRequest", func() {
		Context("when the request is valid", func() {
			BeforeEach(func() {
				requestPayload = api.RequestMessage{URL: "https://www.google.com", Pattern: "test", Interval: 1}
			})
			It("should return status 200", func() {
				Expect(recorder.Code).To(Equal(http.StatusOK))
			})
			It("should call scheduler", func() {
				Eventually(func() int {
					return len(mockScheduler.ScheduleCheckCalls())
				}, 500*time.Millisecond, 100*time.Millisecond).Should(Equal(1))
			})
		})
		Context("when the url is not valid", func() {
			BeforeEach(func() {
				requestPayload = api.RequestMessage{URL: "httpssss://www.google.com", Pattern: "test", Interval: 1}
			})
			It("should return status 400", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
		Context("when the pattern is not valid", func() {
			BeforeEach(func() {
				requestPayload = api.RequestMessage{URL: "https://www.google.com", Pattern: "", Interval: 1}
			})
			It("should return status 400", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
		Context("when the interval is not valid", func() {
			BeforeEach(func() {
				requestPayload = api.RequestMessage{URL: "https://www.google.com", Pattern: "test", Interval: 0}
			})
			It("should return status 400", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})
})
