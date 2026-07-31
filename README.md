![Foster Falls Hillshare from Digital Elevation Model](docs/assets/foster_falls_hillshade.png)
> Foster Falls Hillshade from Digital Elevation Model

# Rockfall Hazard Analysis Platform

Talus aims to mitigate one of the top causes of death / injury in outdoor rock climbing (rockfall) by providing climbers with concrete rockfall risk scores for their local wall. This is done via an end-to-end pipeline from raw DEM ingestion to route risk visualization. Freeze-thaw risk windows combined with GPU-identified source zones above a route create actionable information for mitigating rockfall hazards.

---

## Tech Stack
 
| Layer | Technology |
|---|---|
| Microservices | Go (5 services) |
| GPU kernels | CUDA |
| Database | PostgreSQL + PostGIS |
| Containerization | Docker, Docker Compose |
| Frontend | Leaflet.js, vanilla JS |
| CI | GitHub Actions |
| DEM source | USGS 3DEP 1/3 arc-second (~10m/cell) |
| Weather | Open-Meteo API |

---

## GPU Benchmark
 
Kernels run on a GTX 1060 3GB against the Foster Falls, TN. CPU baseline is single-threaded sequential execution of the same computation.
 
| Kernel | GPU | CPU | Speedup | Cells |
|---|---|---|---|---|
| Sobel slope/aspect | 569ms | 6,041ms | **10.6×** | 116.9M |
| Plan/profile curvature | 565ms | 1,180ms | **2.1×** | 116.9M |
| Terrain Ruggedness Index | 318ms | 1,633ms | **5.1×** | 116.9M |

---

## What It Does

1. Ingests a USGS 3DEP GeoTIFF DEM and USGS NGMDB geology polygons for a climbing area
2. Runs GPU-accelerated terrain analysis to identify rockfall source
3. Fetches NOAA HRRR temperature forecasts and computes freeze-thaw risk windows
4. Scores route risk using a proximity-based formula
5. Fires webhook alerts when risk exceeds a configured threshold
6. Displays results on a Leaflet.js map dashboard with source zone overlay, route overlay, freeze-thaw timeline, and GPU benchmark table

---

## Architecture

Four Go microservices communicating via REST, one standalone CUDA terrain binary invoked as a subprocess, and a shared PostgreSQL + PostGIS database.

```
USGS 3DEP ──► S1 Ingestion (Go) ──► S2 Terrain Preprocessing (Go)
                                              │
                                       CUDA Binary (C/CUDA)
                                       Sobel · Curvature · TRI
                                              │
NOAA NOMADS ──► S4 Hazard Analysis (Go) ◄────┘
                        │
                        ▼
               S5 API Gateway (Go) ──► Browser / Mobile Client
                        │
               PostgreSQL + PostGIS
```

---

## Services

| Service | Language | Responsibility |
|---|---|---|
| S1 — Ingestion | Go | GeoTIFF parse, GPX ingest, geology ingest |
| S2 — Terrain | Go + CUDA subprocess | GPU terrain derivatives, source zone detection |
| S4 — Hazard | Go | Proximity risk scoring, freeze-thaw compute, webhook alerts |
| S5 — Gateway | Go + Leaflet.js | REST API, static dashboard |
| PostgreSQL + PostGIS | SQL | Shared persistent store, spatial queries |

---

## Prerequisites

- Docker and Docker Compose
- Go 1.25+
- CUDA 12.2 toolkit

---

## Setup

```bash
git clone https://github.com/ethan-howes/talus
cd talus
cp .env.example .env
# fill in POSTGRES_PASSWORD and CUDA_TERRAIN_BINARY_PATH
docker compose -f deployments/docker-compose.yml --env-file .env up --build
```
---

## Data Sources

| Dataset | Source | Use |
|---|---|---|
| 3DEP DEM (1/3 arc-second) | [USGS National Map](https://apps.nationalmap.gov/downloader/) | Primary terrain input |
| National Geologic Map | [USGS NGMDB](https://ngmdb.usgs.gov/) | Rock type and simulation parameters |
| HRRR Forecasts | [NOAA NOMADS](https://nomads.ncep.noaa.gov/) | Freeze-thaw temperature forecasts |
| GPX Routes | Personal / AllTrails | Route geometries for risk intersection |

---
