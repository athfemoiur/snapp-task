package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"snapp-task/db"
	"strings"
	"time"
)

//go:generate moq -out=mocked_checker.go . UrlChecker
type UrlChecker interface {
	CheckData() error
}

type UrlCheckerFactory func(url, pattern string, db db.DB) UrlChecker

type UrlCheckerImpl struct {
	Url     string
	Pattern string
	Db      db.DB
}

func NewUrlCheckerImpl(url, pattern string, db db.DB) UrlChecker {
	return &UrlCheckerImpl{Url: url, Pattern: pattern, Db: db}
}

func (uc *UrlCheckerImpl) CheckData() error {
	resultChan := make(chan error)
	timeout := 1 * time.Second

	go func() {
		resultChan <- uc.checkData()
	}()

	select {
	case err := <-resultChan:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("request timed out after %v", timeout)
	}
}

func (uc *UrlCheckerImpl) checkData() error {
	resp, err := http.Get(uc.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch data from URL: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	contentType := resp.Header.Get("Content-Type")
	matchedData, err := uc.findMatch(body, contentType)
	if err != nil {
		return err
	}
	if matchedData == "" {
		return nil
	}
	return uc.Db.SaveData(uc.Url, uc.Pattern, matchedData)
}

func (uc *UrlCheckerImpl) findMatch(content []byte, contentType string) (string, error) {
	isRegex := uc.isRegexPattern(uc.Pattern)
	var regex *regexp.Regexp
	var err error

	if isRegex {
		regex, err = regexp.Compile(uc.Pattern)
		if err != nil {
			return "", fmt.Errorf("invalid regex pattern: %v", err)
		}
	}

	matched := false
	var matchedData string

	if strings.Contains(contentType, "application/json") {
		var jsonData interface{}
		err = json.Unmarshal(content, &jsonData)
		if err != nil {
			return "", fmt.Errorf("failed to parse JSON response: %v", err)
		}

		matched, matchedData = uc.matchFoundInJSON(jsonData, uc.Pattern, regex)
	} else {
		responseString := string(content)
		if isRegex {
			if regex.MatchString(responseString) {
				matched = true
				matchedData = responseString
			}
		} else {
			if strings.Contains(responseString, uc.Pattern) {
				matched = true
				matchedData = responseString
			}
		}
	}
	if matched {
		return matchedData, nil
	}
	return "", nil
}

func (uc *UrlCheckerImpl) isRegexPattern(pattern string) bool {
	return strings.HasPrefix(pattern, "^") || strings.HasSuffix(pattern, "$") || strings.Contains(pattern, ".*")
}

func (uc *UrlCheckerImpl) matchFoundInJSON(data interface{}, pattern string, regex *regexp.Regexp) (bool, string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for _, value := range v {
			if matched, matchedData := uc.matchFoundInJSON(value, pattern, regex); matched {
				return true, matchedData
			}
		}
	case []interface{}:
		for _, item := range v {
			if matched, matchedData := uc.matchFoundInJSON(item, pattern, regex); matched {
				return true, matchedData
			}
		}
	case string:
		if regex != nil {
			if regex.MatchString(v) {
				return true, v
			}
		} else {
			if strings.Contains(v, pattern) {
				return true, v
			}
		}
	}
	return false, ""
}
