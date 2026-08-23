package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/sre-challenge/geoip-api/internal/metrics"
	"github.com/sre-challenge/geoip-api/internal/service"
)

type GeoIPHandler struct {
	service *service.GeoIPService
}

func NewGeoIPHandler(svc *service.GeoIPService) *GeoIPHandler {
	return &GeoIPHandler{service: svc}
}

func (h *GeoIPHandler) GetCountry(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	endpoint := "/country"

	ip := r.URL.Query().Get("ip")
	if ip == "" {
		writeError(w, http.StatusBadRequest, "missing required query parameter: ip")
		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, endpoint, strconv.Itoa(http.StatusBadRequest)).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, endpoint).Observe(time.Since(start).Seconds())
		return
	}

	result, err := h.service.GetCountry(r.Context(), ip)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, endpoint, strconv.Itoa(http.StatusBadRequest)).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, endpoint).Observe(time.Since(start).Seconds())
		return
	}

	writeJSON(w, http.StatusOK, result)
	metrics.HTTPRequestsTotal.WithLabelValues(r.Method, endpoint, strconv.Itoa(http.StatusOK)).Inc()
	metrics.HTTPRequestDuration.WithLabelValues(r.Method, endpoint).Observe(time.Since(start).Seconds())
}

func (h *GeoIPHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (h *GeoIPHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Health(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
