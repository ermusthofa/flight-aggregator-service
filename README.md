# Flight Search & Aggregation System

## Overview

This project is a simple flight search service that aggregates data from multiple airline providers, normalizes the responses, and returns a unified result.

The goal is to simulate some real-world problems like:

* different API response formats
* inconsistent timezones
* partial failures from providers
* performance considerations

---

## Features

* Aggregate flights from multiple providers:

  * AirAsia
  * Garuda Indonesia
  * Lion Air
  * Batik Air
* Normalize different response formats into one structure
* Search by origin, destination, and date
* Filtering (price, stops, etc.)
* Sorting (price, duration, time)
* Basic "best value" scoring

---

## How to Run

```bash
go mod tidy
go run cmd/main.go
```

Server will run on:

```
http://localhost:8080
```

---

## Example Request

```bash
curl -X POST http://localhost:8080/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "cabinClass": "economy"
  }'
```

---

## Project Structure

```
/cmd
  main.go

/internal
  /domain        -> core entities
  /usecase       -> business logic (aggregation)
  /provider      -> provider implementations (AirAsia, etc.)
  /cache         -> cache layer
  /mapper        -> domain -> response
  /dto           -> response models
  /transport     -> HTTP handler
```

---

## Design Notes

### Concurrency

All providers are called in parallel using goroutines.
There is a timeout to make sure slow providers don’t block the whole request.

---

### Retry

Each provider call has retry with exponential backoff.
This is mainly to handle unstable providers (like AirAsia in the mock).

---

### Caching

* In-memory cache with short TTL (5 seconds)
* Cache stores **raw aggregated results**

Filtering and sorting are applied after fetching from cache.

Cached data is treated as immutable (cloned on read) to avoid unexpected mutation across requests.

---

### Partial Results

The system still returns results even if some providers fail.
These results are also cached (short TTL), so the system stays responsive.

---

### Time Handling

Each provider uses different time formats and timezones.

All times are converted to `time.Time` and duration is calculated using:

```
arrival.Sub(departure)
```

This avoids relying on inconsistent provider fields.

---

### Normalization

Each provider has its own adapter that converts raw data into a unified domain model.

This helps keep the rest of the system clean.

---

## Tradeoffs / Notes

* Cache is short-lived to avoid stale pricing
* No persistence layer (in-memory only)
* Some fields are simplified (e.g. baggage mapping)

---

## Closing

This was built mainly to focus on backend logic and system design rather than completeness.

---
