package main_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"time"

	"snapp-task"
)

var _ = Describe("UrlCheckerImpl", func() {
	var (
		mockDB     *main.DBMock
		urlChecker main.UrlChecker
		err        error
	)

	BeforeEach(func() {
		mockDB = &main.DBMock{SaveDataFunc: func(url string, pattern string, data string) error {
			return nil
		}}
	})

	JustBeforeEach(func() {
		err = urlChecker.CheckData()
	})

	Context("when CheckData succeeds", func() {
		BeforeEach(func() {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"key": "value"}`))
			})
			server := httptest.NewServer(handler)
			urlChecker = main.NewUrlCheckerImpl(server.URL, "value", mockDB)
		})

		It("should save data to DB and not return an error", func() {
			Expect(err).To(BeNil())
			Expect(mockDB.SaveDataCalls()).To(HaveLen(1))
			call := mockDB.SaveDataCalls()[0]
			Expect(call.URL).NotTo(BeNil())
			Expect(call.Pattern).To(Equal("value"))
			Expect(call.Data).To(ContainSubstring("value"))
		})
	})

	Context("when CheckData fails due to timeout", func() {
		BeforeEach(func() {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(10 * time.Second)
				w.WriteHeader(http.StatusOK)
			})
			server := httptest.NewServer(handler)
			urlChecker = main.NewUrlCheckerImpl(server.URL, "value", mockDB)
		})

		It("should return a timeout error", func() {
			Expect(err).To(MatchError("request timed out after 5s"))
		})
	})

	Context("When db fails to create", func() {
		BeforeEach(func() {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"key": "value"}`))
			})
			server := httptest.NewServer(handler)
			urlChecker = main.NewUrlCheckerImpl(server.URL, "value", mockDB)
			mockDB.SaveDataFunc = func(url string, pattern string, data string) error {
				return fmt.Errorf("insertion error")
			}
		})
		It("should return an error", func() {
			Expect(err).To(MatchError("failed to insert data into DB: insertion error"))
		})
	})
})
