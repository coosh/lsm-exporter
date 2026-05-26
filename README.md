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

## Configuration

| Environment variable | Default | Description |
|---|---|---|
| `LLAMA_SWAP_URL` | `http://localhost:8080` | Base URL of the llama-swap instance |
| `LISTEN_PORT` | `9090` | Port to serve `/metrics` on |
| `SCRAPE_TIMEOUT` | `5s` | HTTP timeout for upstream requests |

## Running

```sh
LLAMA_SWAP_URL=http://my-llama-swap-host:8080 ./lsm-exporter
```

Metrics are served at `http://localhost:9090/metrics`. A `/health` endpoint returns `OK`.

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
