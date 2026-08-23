package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/sre-challenge/geoip-api/internal/metrics"
	"github.com/sre-challenge/geoip-api/internal/provider"
	"github.com/sre-challenge/geoip-api/internal/repository"
)

var ipRegex = regexp.MustCompile(`^(?:\d{1,3}\.){3}\d{1,3}$|^(?:[0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}$`)

type GeoIPService struct {
	repo     repository.Repository
	provider provider.Provider
}

func NewGeoIPService(repo repository.Repository, provider provider.Provider) *GeoIPService {
	return &GeoIPService{repo: repo, provider: provider}
}

type CountryResponse struct {
	IP          string  `json:"ip"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	City        string  `json:"city,omitempty"`
	Region      string  `json:"region,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	CacheHit    bool    `json:"cache_hit"`
	Source      string  `json:"source"`
}

func (s *GeoIPService) GetCountry(ctx context.Context, ip string) (*CountryResponse, error) {
	if !isValidIP(ip) {
		return nil, fmt.Errorf("invalid ip address: %s", ip)
	}

	record, err := s.repo.GetByIP(ctx, ip)
	if err == nil {
		metrics.CacheHitsTotal.Inc()
		return toResponse(record, true, "cache"), nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("cache lookup: %w", err)
	}

	metrics.CacheMissesTotal.Inc()

	record, err = s.provider.Lookup(ctx, ip)
	if err != nil {
		return nil, fmt.Errorf("provider lookup: %w", err)
	}

	if err := s.repo.Upsert(ctx, record); err != nil {
		return nil, fmt.Errorf("cache upsert: %w", err)
	}

	return toResponse(record, false, "external"), nil
}

func (s *GeoIPService) Health(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func isValidIP(ip string) bool {
	if !ipRegex.MatchString(ip) {
		return false
	}
	return net.ParseIP(ip) != nil
}

func toResponse(record *repository.GeoIPRecord, cacheHit bool, source string) *CountryResponse {
	return &CountryResponse{
		IP:          record.IP,
		Country:     record.Country,
		CountryCode: record.CountryCode,
		City:        record.City,
		Region:      record.Region,
		Latitude:    record.Latitude,
		Longitude:   record.Longitude,
		CacheHit:    cacheHit,
		Source:      source,
	}
}
