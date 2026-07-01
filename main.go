package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var version = "dev"

type appInfo struct {
	Service       string `json:"service"`
	Version       string `json:"version"`
	AppVersionEnv string `json:"app_version_env"`
	PodName       string `json:"pod_name"`
	PodNamespace  string `json:"pod_namespace"`
	NodeName      string `json:"node_name"`
	Region        string `json:"region"`
	ScanMode      string `json:"scan_mode"`
}

type scanRequest struct {
	PackageCode string `json:"package_code"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /readyz", readyz)
	mux.HandleFunc("GET /info", info)
	mux.HandleFunc("GET /config-check", configCheck)
	mux.HandleFunc("GET /secret-check", secretCheck)
	mux.HandleFunc("POST /scan-test", scanTest)
	mux.HandleFunc("/", notFound)

	addr := ":" + value("PORT", "8080")
	log.Printf("scanner-platform version=%s listening on %s", version, addr)
	if err := http.ListenAndServe(addr, logging(mux)); err != nil {
		log.Fatal(err)
	}
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "scanner-platform",
	})
}

func readyz(w http.ResponseWriter, _ *http.Request) {
	required := []string{
		"SP_FEATURE_FLAG",
		"SP_REGION",
		"SP_SCAN_MODE",
		"SP_API_TOKEN",
		"SP_DB_PASSWORD",
		"APP_VERSION",
		"POD_NAME",
		"POD_NAMESPACE",
		"NODE_NAME",
	}

	missing := make([]string, 0)
	for _, key := range required {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}

	status := http.StatusOK
	state := "ready"
	if len(missing) > 0 {
		status = http.StatusServiceUnavailable
		state = "not_ready"
	}

	writeJSON(w, status, map[string]any{
		"status":  state,
		"missing": missing,
	})
}

func info(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, appInfo{
		Service:       "scanner-platform",
		Version:       version,
		AppVersionEnv: value("APP_VERSION", "unknown"),
		PodName:       value("POD_NAME", "unknown"),
		PodNamespace:  value("POD_NAMESPACE", "unknown"),
		NodeName:      value("NODE_NAME", "unknown"),
		Region:        value("SP_REGION", "unknown"),
		ScanMode:      value("SP_SCAN_MODE", "unknown"),
	})
}

func configCheck(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"feature_flag": value("SP_FEATURE_FLAG", ""),
		"region":       value("SP_REGION", ""),
		"scan_mode":    value("SP_SCAN_MODE", ""),
	})
}

func secretCheck(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"api_token_present":   strings.TrimSpace(os.Getenv("SP_API_TOKEN")) != "",
		"db_password_present": strings.TrimSpace(os.Getenv("SP_DB_PASSWORD")) != "",
		"api_token":           mask(os.Getenv("SP_API_TOKEN")),
		"db_password":         mask(os.Getenv("SP_DB_PASSWORD")),
	})
}

func scanTest(w http.ResponseWriter, r *http.Request) {
	var req scanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"status": "rejected",
			"error":  "invalid_json",
		})
		return
	}

	req.PackageCode = strings.TrimSpace(req.PackageCode)
	if req.PackageCode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"status": "rejected",
			"error":  "package_code_required",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":       "accepted",
		"package_code": req.PackageCode,
		"scan_mode":    value("SP_SCAN_MODE", "unknown"),
		"processed_by": value("POD_NAME", "unknown"),
		"node":         value("NODE_NAME", "unknown"),
		"processed_at": time.Now().UTC().Format(time.RFC3339),
	})
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotFound, map[string]string{
		"status": "not_found",
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write response: %v", err)
	}
}

func value(key, fallback string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return fallback
}

func mask(secret string) string {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return ""
	}
	return "********"
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
