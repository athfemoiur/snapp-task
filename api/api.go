package api

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"snapp-task/db"
	"snapp-task/services"
	"time"
)

type apiError struct {
	Error string `json:"error"`
}

func writeJson(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

type APIServer struct {
	addr             string
	db               db.DB
	schedulerFactory services.SchedulerFactory
}

func NewAPIServer(addr string, db db.DB, schedulerFactory services.SchedulerFactory) *APIServer {
	return &APIServer{
		addr:             addr,
		db:               db,
		schedulerFactory: schedulerFactory,
	}
}

func (s *APIServer) Run() error {
	router := http.NewServeMux()
	router.HandleFunc("POST /", s.HandleRequest)
	server := &http.Server{
		Addr:    s.addr,
		Handler: router,
	}
	log.Println("Starting server on ", s.addr)
	return server.ListenAndServe()
}

type RequestMessage struct {
	URL      string `json:"url"`
	Interval int    `json:"interval"`
	Pattern  string `json:"pattern"`
}

func validateURL(input string) bool {
	parsedURL, err := url.ParseRequestURI(input)
	if err != nil {
		return false
	}
	return parsedURL.Scheme == "http" || parsedURL.Scheme == "https"
}

func validatePattern(pattern string) bool {
	if pattern == "" {
		return false
	}
	if _, err := regexp.Compile(pattern); err != nil {
		return false
	}
	return true
}

func validateInterval(input int) bool {
	return input >= 1
}

func (s *APIServer) HandleRequest(writer http.ResponseWriter, request *http.Request) {
	var req RequestMessage
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		writeJson(writer, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}
	if !validateURL(req.URL) {
		writeJson(writer, http.StatusBadRequest, apiError{Error: "Invalid url"})
		return
	}
	if !validateInterval(req.Interval) {
		writeJson(writer, http.StatusBadRequest, apiError{Error: "Invalid interval"})
		return
	}
	if !validatePattern(req.Pattern) {
		writeJson(writer, http.StatusBadRequest, apiError{Error: "Invalid pattern"})
		return
	}
	scheduler := s.schedulerFactory(req.URL, req.Pattern, time.Duration(req.Interval)*time.Second, s.db)
	go scheduler.ScheduleCheck()
	writeJson(writer, http.StatusOK, struct{}{})
}
