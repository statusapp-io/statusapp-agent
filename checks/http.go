package checks

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RunHTTP performs an HTTP/HTTPS/WEBSITE/API/GRAPHQL check.
func RunHTTP(m Monitor) Result {
	method := "GET"
	if m.HTTPMethod != nil && *m.HTTPMethod != "" {
		method = *m.HTTPMethod
	}

	// For GraphQL, default to POST
	if m.Type == "GRAPHQL" && (m.HTTPMethod == nil || *m.HTTPMethod == "") {
		method = "POST"
	}

	var bodyReader io.Reader
	if m.RequestBody != nil && *m.RequestBody != "" {
		bodyReader = strings.NewReader(*m.RequestBody)
	}

	// Build GraphQL body if needed
	if m.Type == "GRAPHQL" && m.GraphQLQuery != nil && bodyReader == nil {
		gqlBody, _ := json.Marshal(map[string]interface{}{
			"query": *m.GraphQLQuery,
		})
		bodyReader = strings.NewReader(string(gqlBody))
	}

	req, err := http.NewRequest(method, m.URL, bodyReader)
	if err != nil {
		return Result{Status: "DOWN", Error: fmt.Sprintf("invalid request: %v", err)}
	}

	// Parse and set headers
	if m.RequestHeaders != nil && *m.RequestHeaders != "" {
		var headers map[string]string
		if json.Unmarshal([]byte(*m.RequestHeaders), &headers) == nil {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		}
	}

	if m.Type == "GRAPHQL" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	timeout := time.Duration(m.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	start := time.Now()
	resp, err := client.Do(req)
	responseTime := int(time.Since(start).Milliseconds())

	if err != nil {
		return Result{
			Status:       "DOWN",
			ResponseTime: &responseTime,
			Error:        fmt.Sprintf("request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	result := Result{
		Status:       "UP",
		ResponseTime: &responseTime,
		StatusCode:   &statusCode,
	}

	// Check expected status code
	if m.ExpectedStatusCode != nil && *m.ExpectedStatusCode > 0 {
		if statusCode != *m.ExpectedStatusCode {
			result.Status = "DOWN"
			result.Error = fmt.Sprintf("expected status %d, got %d", *m.ExpectedStatusCode, statusCode)
			return result
		}
	} else if statusCode >= 400 {
		result.Status = "DOWN"
		result.Error = fmt.Sprintf("HTTP %d", statusCode)
		return result
	}

	// Keyword check
	if m.Type == "KEYWORD" && m.KeywordList != nil && len(m.KeywordList) > 0 {
		body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
		if err != nil {
			result.Status = "DOWN"
			result.Error = fmt.Sprintf("failed to read body: %v", err)
			return result
		}
		bodyStr := string(body)
		caseSensitive := m.KeywordCaseSensitive != nil && *m.KeywordCaseSensitive
		matchType := "contains"
		if m.KeywordMatchType != nil {
			matchType = *m.KeywordMatchType
		}

		if !caseSensitive {
			bodyStr = strings.ToLower(bodyStr)
		}

		for _, kw := range m.KeywordList {
			search := kw
			if !caseSensitive {
				search = strings.ToLower(kw)
			}
			found := strings.Contains(bodyStr, search)
			if matchType == "not_contains" {
				if found {
					result.Status = "DOWN"
					result.Error = fmt.Sprintf("keyword '%s' found (should not be present)", kw)
					return result
				}
			} else {
				if !found {
					result.Status = "DOWN"
					result.Error = fmt.Sprintf("keyword '%s' not found", kw)
					return result
				}
			}
		}
	}

	return result
}
