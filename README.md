# lsm-exporter

A Prometheus metrics exporter for [llama-swap](https://github.com/mostlygeek/llama-swap).

Scrapes llama-swap's own system and GPU metrics, and for each loaded model proxies the underlying llama-server's metrics endpoint — injecting a `model` label so all series are distinguishable in a single scrape.

## Metrics

**From llama-swap (`/metrics`):**

| Metric | Type | Description |
|---|---|---|
| `llamaswap_cpu_util_percent` | gauge | CPU utilisation per core |
| `llamaswap_memory_total_bytes` | gauge | Total system memory |
| `llamaswap_memory_used_bytes` | gauge | Used system memory |
| `llamaswap_memory_free_bytes` | gauge | Free system memory |
| `llamaswap_swap_total_bytes` | gauge | Total swap |
| `llamaswap_swap_used_bytes` | gauge | Used swap |
| `llamaswap_load_average` | gauge | Load average (1m / 5m / 15m) |
| `llamaswap_network_bytes_total` | counter | Network bytes by interface and direction |
| `llamaswap_gpu_temperature_celsius` | gauge | GPU temperature |
| `llamaswap_gpu_vram_temperature_celsius` | gauge | VRAM temperature |
| `llamaswap_gpu_util_percent` | gauge | GPU utilisation |
| `llamaswap_gpu_memory_util_percent` | gauge | GPU memory utilisation |
| `llamaswap_gpu_memory_used_bytes` | gauge | GPU memory used |
| `llamaswap_gpu_memory_total_bytes` | gauge | GPU memory total |
| `llamaswap_gpu_fan_speed_percent` | gauge | GPU fan speed |
| `llamaswap_gpu_power_draw_watts` | gauge | GPU power draw |

**From each loaded model (llama-server `/metrics`):**

All standard llama-server Prometheus metrics, with a `model` label injected. Only models currently in `ready` state are scraped — no models are loaded on demand.

## Usage

```
lsm-exporter [options]

  -a  listen address      (default: 0.0.0.0)
  -p  listen port         (default: 9090)
  -l  llama-swap base URL (default: http://localhost:8080)
  -t  scrape timeout (s)  (default: 5)
```

```sh
lsm-exporter -a 0.0.0.0 -p 9090 -l http://llama-swap.host:8080 -t 60
```

Metrics are served at `http://<listen-address>:<port>/metrics`. A `/health` endpoint returns `OK`.

## Docker image

A multi-arch image is published at `ghcr.io/coosh/lsm-exporter` for `linux/amd64` and `linux/arm64`.

### Environment variables

All CLI flags have equivalent environment variables:

| Variable | CLI flag | Default |
|---|---|---|
| `LSM_LISTEN_ADDR` | `-a` | `0.0.0.0` |
| `LSM_LISTEN_PORT` | `-p` | `9090` |
| `LLAMASWAP_URL` | `-l` | `http://localhost:8080` |
| `SCRAPE_TIMEOUT` | `-t` | `5` |

### Run directly

```sh
docker run -d \
  --name lsm-exporter \
  -p 9090:9090 \
  -e LLAMASWAP_URL=http://llama-swap.host:8080 \
  ghcr.io/coosh/lsm-exporter:latest
```

### Docker Compose

```yaml
services:
  lsm-exporter:
    image: ghcr.io/coosh/lsm-exporter:latest
    container_name: lsm-exporter
    restart: unless-stopped
    ports:
      - "9090:9090"
    environment:
      LLAMASWAP_URL: "http://llama-swap.host:8080"
```

Replace `llama-swap.host` with the actual hostname or IP of your llama-swap instance. If llama-swap is also running in Docker, use its service name instead.

## Building

```sh
go build -o lsm-exporter ./src/
```

Pre-built binaries for Linux (amd64/arm64), macOS (amd64/arm64), and Windows (amd64) are available on the [releases page](../../releases).

## Prometheus scrape config

```yaml
scrape_configs:
  - job_name: lsm-exporter
    static_configs:
      - targets: ['localhost:9090']
```
