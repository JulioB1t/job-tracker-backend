package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

const maskedValue = "[masked]"

func logRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !shouldLogRequestBody(r) {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("request body read failed method=%s path=%s error=%v", r.Method, r.URL.Path, err)
			next.ServeHTTP(w, r)
			return
		}
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(body))

		log.Printf("request body method=%s path=%s body=%s", r.Method, r.URL.Path, maskRequestBody(body))
		next.ServeHTTP(w, r)
	})
}

func shouldLogRequestBody(r *http.Request) bool {
	if r.Body == nil || r.Body == http.NoBody {
		return false
	}

	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return false
	}
}

func maskRequestBody(body []byte) string {
	if len(bytes.TrimSpace(body)) == 0 {
		return ""
	}

	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "[unparseable JSON body]"
	}

	masked := maskJSONValue(payload, "")
	encoded, err := json.Marshal(masked)
	if err != nil {
		return "[masked JSON marshal failed]"
	}

	return string(encoded)
}

func maskJSONValue(value any, parentKey string) any {
	switch typed := value.(type) {
	case map[string]any:
		if strings.EqualFold(parentKey, "jobDescription") {
			return typed
		}

		masked := make(map[string]any, len(typed))
		for key, nestedValue := range typed {
			if shouldMaskField(key) {
				masked[key] = maskedValue
				continue
			}

			masked[key] = maskJSONValue(nestedValue, key)
		}
		return masked
	case []any:
		masked := make([]any, len(typed))
		for i, nestedValue := range typed {
			masked[i] = maskJSONValue(nestedValue, parentKey)
		}
		return masked
	default:
		return value
	}
}

func shouldMaskField(key string) bool {
	normalized := strings.ToLower(key)

	switch normalized {
	case "note", "notes", "email", "phone", "address", "location", "joburl", "url", "linkedinurl":
		return true
	default:
		return strings.Contains(normalized, "email") ||
			strings.Contains(normalized, "phone") ||
			strings.Contains(normalized, "address") ||
			strings.Contains(normalized, "token") ||
			strings.Contains(normalized, "secret") ||
			strings.Contains(normalized, "password")
	}
}
