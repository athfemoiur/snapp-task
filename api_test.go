package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"snapp-task"
)

var _ = Describe("APIServer", func() {
	var (
		server         *main.APIServer
		recorder       *httptest.ResponseRecorder
		requestPayload main.RequestMessage
		mockScheduler  *main.CheckSchedulerMock
	)

	BeforeEach(func() {
		mockScheduler = &main.CheckSchedulerMock{
			ScheduleCheckFunc: func() {},
		}
		schedulerFactory := func(url, pattern string, interval time.Duration, db main.DB) main.CheckScheduler {
			return mockScheduler
		}
		mockDB := &main.DBMock{SaveDataFunc: func(url string, pattern string, data string) error {
			return nil
		}}
		server = main.NewAPIServer(":8080", mockDB, schedulerFactory)
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
				requestPayload = main.RequestMessage{URL: "https://www.google.com", Pattern: "test", Interval: 1}
			})
			It("should return status 200", func() {
				Expect(recorder.Code).To(Equal(http.StatusOK))
			})
		})
		Context("when the url is not valid", func() {
			BeforeEach(func() {
				requestPayload = main.RequestMessage{URL: "httpssss://www.google.com", Pattern: "test", Interval: 1}
			})
			It("should return status 400", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
		Context("when the pattern is not valid", func() {
			BeforeEach(func() {
				requestPayload = main.RequestMessage{URL: "https://www.google.com", Pattern: "", Interval: 1}
			})
			It("should return status 400", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
		Context("when the interval is not valid", func() {
			BeforeEach(func() {
				requestPayload = main.RequestMessage{URL: "https://www.google.com", Pattern: "test", Interval: 0}
			})
			It("should return status 400", func() {
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})
})
