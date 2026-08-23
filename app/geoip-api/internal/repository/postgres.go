package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(databaseURL string) (*PostgresRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &PostgresRepository{pool: pool}, nil
}

func (r *PostgresRepository) Migrate(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS geoip_cache (
		ip VARCHAR(45) PRIMARY KEY,
		country VARCHAR(100) NOT NULL,
		country_code VARCHAR(10) NOT NULL,
		city VARCHAR(100),
		region VARCHAR(100),
		latitude DOUBLE PRECISION,
		longitude DOUBLE PRECISION,
		provider VARCHAR(50) NOT NULL DEFAULT 'ipapi',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_geoip_cache_country_code ON geoip_cache(country_code);
	CREATE INDEX IF NOT EXISTS idx_geoip_cache_updated_at ON geoip_cache(updated_at);
	`
	_, err := r.pool.Exec(ctx, query)
	return err
}

func (r *PostgresRepository) GetByIP(ctx context.Context, ip string) (*GeoIPRecord, error) {
	query := `
	SELECT ip, country, country_code, city, region, latitude, longitude, provider, created_at, updated_at
	FROM geoip_cache WHERE ip = $1
	`
	var record GeoIPRecord
	err := r.pool.QueryRow(ctx, query, ip).Scan(
		&record.IP, &record.Country, &record.CountryCode, &record.City,
		&record.Region, &record.Latitude, &record.Longitude, &record.Provider,
		&record.CreatedAt, &record.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *PostgresRepository) Upsert(ctx context.Context, record *GeoIPRecord) error {
	query := `
	INSERT INTO geoip_cache (ip, country, country_code, city, region, latitude, longitude, provider, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	ON CONFLICT (ip) DO UPDATE SET
		country = EXCLUDED.country,
		country_code = EXCLUDED.country_code,
		city = EXCLUDED.city,
		region = EXCLUDED.region,
		latitude = EXCLUDED.latitude,
		longitude = EXCLUDED.longitude,
		provider = EXCLUDED.provider,
		updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, query,
		record.IP, record.Country, record.CountryCode, record.City,
		record.Region, record.Latitude, record.Longitude, record.Provider,
	)
	return err
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *PostgresRepository) Close() {
	r.pool.Close()
}
