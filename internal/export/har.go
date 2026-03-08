package export

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// HAR represents an HTTP Archive 1.2 document.
type HAR struct {
	Log harLog `json:"log"`
}

type harLog struct {
	Version string     `json:"version"`
	Creator harCreator `json:"creator"`
	Entries []harEntry `json:"entries"`
}

type harCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type harEntry struct {
	StartedDateTime string    `json:"startedDateTime"`
	Time            float64   `json:"time"` // ms
	Request         harReq    `json:"request"`
	Response        harResp   `json:"response"`
	Timings         harTiming `json:"timings"`
	Comment         string    `json:"comment,omitempty"`
}

type harReq struct {
	Method      string       `json:"method"`
	URL         string       `json:"url"`
	HTTPVersion string       `json:"httpVersion"`
	Headers     []harHeader  `json:"headers"`
	QueryString []harKV      `json:"queryString"`
	PostData    *harPostData `json:"postData,omitempty"`
	BodySize    int          `json:"bodySize"`
	HeadersSize int          `json:"headersSize"`
}

type harResp struct {
	Status      int         `json:"status"`
	StatusText  string      `json:"statusText"`
	HTTPVersion string      `json:"httpVersion"`
	Headers     []harHeader `json:"headers"`
	Content     harContent  `json:"content"`
	RedirectURL string      `json:"redirectURL"`
	BodySize    int         `json:"bodySize"`
	HeadersSize int         `json:"headersSize"`
}

type harHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type harKV struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type harPostData struct {
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

type harContent struct {
	Size     int    `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
}

type harTiming struct {
	Send    float64 `json:"send"`
	Wait    float64 `json:"wait"`
	Receive float64 `json:"receive"`
}

// sensitiveHARHeaders are masked in HAR output (lowercase).
var sensitiveHARHeaders = map[string]bool{
	"authorization":   true,
	"x-api-key":       true,
	"x-goog-api-key":  true,
	"api-key":         true,
}

// maskValue truncates a sensitive header value to a safe preview.
func maskValue(v string) string {
	if len(v) <= 8 {
		return "***"
	}
	// Show first 4 chars and last 4 chars.
	return v[:4] + "..." + v[len(v)-4:]
}

// ToHAR converts a slice of probe Requests to an HAR 1.2 document.
// Sensitive headers are masked. The version string identifies the probe build.
func ToHAR(requests []*store.Request, version string) *HAR {
	entries := make([]harEntry, 0, len(requests))
	for _, r := range requests {
		entries = append(entries, reqToHAREntry(r))
	}
	return &HAR{
		Log: harLog{
			Version: "1.2",
			Creator: harCreator{Name: "probe", Version: version},
			Entries: entries,
		},
	}
}

func reqToHAREntry(r *store.Request) harEntry {
	latencyMS := float64(r.Latency.Milliseconds())

	// Request headers
	reqHeaders := headersToHAR(r.RequestHeaders, true)

	// Request body
	var postData *harPostData
	if len(r.RequestBody) > 0 {
		postData = &harPostData{
			MimeType: "application/json",
			Text:     string(r.RequestBody),
		}
	}

	// Response headers
	respHeaders := headersToHAR(r.ResponseHeaders, false)

	// Response content
	respText := string(r.ResponseBody)
	respSize := len(r.ResponseBody)

	// Determine status text
	statusText := httpStatusText(r.StatusCode)

	entry := harEntry{
		StartedDateTime: r.StartedAt.UTC().Format(time.RFC3339Nano),
		Time:            latencyMS,
		Request: harReq{
			Method:      r.Method,
			URL:         r.URL,
			HTTPVersion: "HTTP/1.1",
			Headers:     reqHeaders,
			QueryString: []harKV{},
			PostData:    postData,
			BodySize:    len(r.RequestBody),
			HeadersSize: -1,
		},
		Response: harResp{
			Status:      r.StatusCode,
			StatusText:  statusText,
			HTTPVersion: "HTTP/1.1",
			Headers:     respHeaders,
			Content: harContent{
				Size:     respSize,
				MimeType: contentType(r.ResponseHeaders),
				Text:     respText,
			},
			RedirectURL: "",
			BodySize:    respSize,
			HeadersSize: -1,
		},
		Timings: harTiming{
			Send:    0,
			Wait:    latencyMS,
			Receive: 0,
		},
	}

	// Add TTFT as a comment if available.
	if r.TTFT > 0 {
		entry.Comment = "ttft=" + r.TTFT.String()
	}

	return entry
}

func headersToHAR(headers map[string]string, maskSensitive bool) []harHeader {
	result := make([]harHeader, 0, len(headers))
	for k, v := range headers {
		if maskSensitive && sensitiveHARHeaders[strings.ToLower(k)] {
			v = maskValue(v)
		}
		result = append(result, harHeader{Name: k, Value: v})
	}
	return result
}

func contentType(headers map[string]string) string {
	for k, v := range headers {
		if strings.ToLower(k) == "content-type" {
			return v
		}
	}
	return "application/json"
}

func httpStatusText(code int) string {
	texts := map[int]string{
		200: "OK", 201: "Created", 204: "No Content",
		400: "Bad Request", 401: "Unauthorized", 403: "Forbidden",
		404: "Not Found", 429: "Too Many Requests",
		500: "Internal Server Error", 502: "Bad Gateway", 503: "Service Unavailable",
	}
	if t, ok := texts[code]; ok {
		return t
	}
	return ""
}

// HARToJSON serialises an HAR document to indented JSON.
func HARToJSON(h *HAR) ([]byte, error) {
	return json.MarshalIndent(h, "", "  ")
}
