package services_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"snapp-task/db"
	. "snapp-task/services"
	"time"
)

var _ = Describe("UrlCheckerImpl", func() {
	var (
		mockDB      *db.DBMock
		urlChecker  UrlChecker
		err         error
		testPattern string
		testData    string
		isJson      bool
		statusCode  int
		testServer  *httptest.Server
		timeOut     time.Duration
	)

	BeforeEach(func() {
		mockDB = &db.DBMock{SaveDataFunc: func(url string, pattern string, data string) error {
			return nil
		}}
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(timeOut)
			w.WriteHeader(statusCode)
			if isJson {
				w.Header().Set("Content-Type", "application/json")
			}
			w.Write([]byte(testData))
		})
		testServer = httptest.NewServer(handler)
	})

	JustBeforeEach(func() {
		urlChecker = NewUrlCheckerImpl(testServer.URL, testPattern, mockDB)
		err = urlChecker.CheckData()
	})

	Context("when CheckData succeeds", func() {
		BeforeEach(func() {
			testPattern = "test_pattern"
			testData = "this_is_data_containing_test_pattern!"
			statusCode = http.StatusOK
			timeOut = time.Millisecond * 10
		})
		It("should save data to DB and not return an error", func() {
			Expect(err).To(BeNil())
			Expect(mockDB.SaveDataCalls()).To(HaveLen(1))
			call := mockDB.SaveDataCalls()[0]
			Expect(call.URL).To(Equal(testServer.URL))
			Expect(call.Pattern).To(Equal(testPattern))
			Expect(call.Data).To(Equal(testData))
		})
	})

	Context("when CheckData fails due to timeout", func() {
		BeforeEach(func() {
			testPattern = "test_pattern"
			statusCode = http.StatusOK
			timeOut = time.Millisecond * 1100
		})

		It("should return a timeout error", func() {
			Expect(err).To(MatchError("request timed out after 1s"))
		})
	})

	Context("When db fails to create", func() {
		BeforeEach(func() {
			testPattern = "test_pattern"
			testData = "this_is_data_containing_test_pattern!"
			statusCode = http.StatusOK
			timeOut = time.Millisecond * 10

			mockDB.SaveDataFunc = func(url string, pattern string, data string) error {
				return fmt.Errorf("insertion error")
			}
		})
		It("should return an error", func() {
			Expect(err).To(MatchError("insertion error"))
		})
	})
})
