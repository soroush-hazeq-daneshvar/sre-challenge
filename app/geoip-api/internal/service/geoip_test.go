package service

import (
	"context"
	"testing"

	"github.com/sre-challenge/geoip-api/internal/repository"
)

type mockRepo struct {
	records map[string]*repository.GeoIPRecord
}

func (m *mockRepo) GetByIP(_ context.Context, ip string) (*repository.GeoIPRecord, error) {
	if r, ok := m.records[ip]; ok {
		return r, nil
	}
	return nil, errNoRows
}

func (m *mockRepo) Upsert(_ context.Context, record *repository.GeoIPRecord) error {
	m.records[record.IP] = record
	return nil
}

func (m *mockRepo) Ping(_ context.Context) error  { return nil }
func (m *mockRepo) Close()                          {}
func (m *mockRepo) Migrate(_ context.Context) error { return nil }

var errNoRows = &mockError{msg: "no rows"}

type mockError struct{ msg string }

func (e *mockError) Error() string { return e.msg }

type mockProvider struct {
	record *repository.GeoIPRecord
}

func (m *mockProvider) Lookup(_ context.Context, ip string) (*repository.GeoIPRecord, error) {
	return m.record, nil
}

func TestGetCountryFromCache(t *testing.T) {
	repo := &mockRepo{
		records: map[string]*repository.GeoIPRecord{
			"8.8.8.8": {
				IP:          "8.8.8.8",
				Country:     "United States",
				CountryCode: "US",
			},
		},
	}
	svc := NewGeoIPService(repo, &mockProvider{})

	result, err := svc.GetCountry(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.CacheHit {
		t.Error("expected cache hit")
	}
	if result.Country != "United States" {
		t.Errorf("expected United States, got %s", result.Country)
	}
}

func TestGetCountryInvalidIP(t *testing.T) {
	svc := NewGeoIPService(&mockRepo{records: map[string]*repository.GeoIPRecord{}}, &mockProvider{})

	_, err := svc.GetCountry(context.Background(), "not-an-ip")
	if err == nil {
		t.Fatal("expected error for invalid IP")
	}
}

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		ip    string
		valid bool
	}{
		{"8.8.8.8", true},
		{"192.168.1.1", true},
		{"256.1.1.1", false},
		{"not-an-ip", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := isValidIP(tt.ip); got != tt.valid {
			t.Errorf("isValidIP(%q) = %v, want %v", tt.ip, got, tt.valid)
		}
	}
}
