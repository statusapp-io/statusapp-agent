package checks

import "encoding/json"

// Monitor represents a monitor received from the API poll endpoint.
type Monitor struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	URL                string   `json:"url"`
	Type               string   `json:"type"`
	Interval           int      `json:"interval"`
	Timeout            int      `json:"timeout"`
	ExpectedStatusCode *int     `json:"expectedStatusCode"`
	HTTPMethod         *string  `json:"httpMethod"`
	RequestHeaders     *string  `json:"requestHeaders"`
	RequestBody        *string  `json:"requestBody"`
	Port               *int     `json:"port"`
	DNSRecord          *string  `json:"dnsRecord"`
	ExpectedIP         *string  `json:"expectedIp"`
	SSLCertExpiryDays  *int     `json:"sslCertExpiryDays"`
	GraphQLQuery       *string  `json:"graphqlQuery"`
	KeywordList        []string `json:"keywordList"`
	KeywordMatchType   *string  `json:"keywordMatchType"`
	KeywordCaseSensitive *bool  `json:"keywordCaseSensitive"`
	Assertions         json.RawMessage `json:"assertions"`
	AssertionsEnabled  *bool   `json:"assertionsEnabled"`
}

// Result represents the outcome of executing a check.
type Result struct {
	Status       string `json:"status"`
	ResponseTime *int   `json:"responseTime,omitempty"`
	StatusCode   *int   `json:"statusCode,omitempty"`
	Error        string `json:"error,omitempty"`
}

// Run executes the appropriate check based on monitor type.
func Run(m Monitor) Result {
	switch m.Type {
	case "HTTP", "HTTPS", "WEBSITE", "API", "GRAPHQL", "KEYWORD":
		return RunHTTP(m)
	case "PORT":
		return RunTCP(m)
	case "PING":
		return RunPing(m)
	case "DNS":
		return RunDNS(m)
	case "SSL_CERT":
		return RunSSL(m)
	default:
		return Result{Status: "DOWN", Error: "unsupported monitor type: " + m.Type}
	}
}
