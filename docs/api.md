# API Documentation

## Base URL

```
Production: https://geoip.example.com
Staging:    https://geoip-staging.example.com
Local:      http://geoip.local
```

## Endpoints

### GET /country

Lookup country information for an IP address.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `ip` | string | Yes | IPv4 or IPv6 address |

**Example Request:**

```bash
curl "http://geoip.local/country?ip=8.8.8.8"
```

**Success Response (200):**

```json
{
  "ip": "8.8.8.8",
  "country": "United States",
  "country_code": "US",
  "city": "Mountain View",
  "region": "California",
  "latitude": 37.4056,
  "longitude": -122.0775,
  "cache_hit": false,
  "source": "external"
}
```

**Cached Response (200):**

```json
{
  "ip": "8.8.8.8",
  "country": "United States",
  "country_code": "US",
  "city": "Mountain View",
  "region": "California",
  "latitude": 37.4056,
  "longitude": -122.0775,
  "cache_hit": true,
  "source": "cache"
}
```

**Error Response (400):**

```json
{
  "error": "invalid ip address: not-an-ip"
}
```

**Error Response (400 - Provider failure):**

```json
{
  "error": "provider lookup: external api returned 429: rate limited"
}
```

---

### GET /health

Liveness probe endpoint. Returns 200 if the service is running.

```bash
curl http://geoip.local/health
```

```json
{"status": "healthy"}
```

---

### GET /ready

Readiness probe endpoint. Returns 200 if the service can serve requests (database connected).

```bash
curl http://geoip.local/ready
```

```json
{"status": "ready"}
```

**Unavailable (503):**

```json
{"error": "database not ready"}
```

---

### GET /metrics

Prometheus metrics endpoint.

```bash
curl http://geoip.local/metrics
```

**Available Metrics:**

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `geoip_http_requests_total` | Counter | method, endpoint, status | Total HTTP requests |
| `geoip_http_request_duration_seconds` | Histogram | method, endpoint | Request latency |
| `geoip_cache_hits_total` | Counter | - | Cache hits |
| `geoip_cache_misses_total` | Counter | - | Cache misses |
| `geoip_external_api_calls_total` | Counter | status | External API calls |
| `geoip_external_api_duration_seconds` | Histogram | - | External API latency |

---

## Rate Limits

- External provider (ipapi.co): 1000 requests/day (free tier)
- Cache significantly reduces external API calls
- No rate limit on the API itself (configure at Ingress level if needed)

## Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `ip` | string | Queried IP address |
| `country` | string | Full country name |
| `country_code` | string | ISO 3166-1 alpha-2 code |
| `city` | string | City name (if available) |
| `region` | string | Region/state name |
| `latitude` | float | Geographic latitude |
| `longitude` | float | Geographic longitude |
| `cache_hit` | boolean | Whether result came from cache |
| `source` | string | `cache` or `external` |
