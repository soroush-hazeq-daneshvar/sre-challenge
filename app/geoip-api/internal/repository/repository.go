package repository

import (
	"context"
	"time"
)

type GeoIPRecord struct {
	IP          string
	Country     string
	CountryCode string
	City        string
	Region      string
	Latitude    float64
	Longitude   float64
	Provider    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Repository interface {
	GetByIP(ctx context.Context, ip string) (*GeoIPRecord, error)
	Upsert(ctx context.Context, record *GeoIPRecord) error
	Ping(ctx context.Context) error
	Close()
	Migrate(ctx context.Context) error
}
