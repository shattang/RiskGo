# RiskGo 🚀

A high-performance, Dockerized microservice architecture for Options Risk Analysis. RiskGo combines the speed of **C++ (QuantLib)** with the flexibility of **Go (Fiber)** to provide real-time portfolio Greeks and scenario-based stress testing.

---

## 🏗 Architecture

RiskGo is split into two specialized services that communicate over high-speed gRPC:

```text
[ User / CLI ] --(REST/JSON)--> [ Go Gateway ] --(gRPC)--> [ C++ Engine ]
                                     |                         |
                             (Yahoo Finance API)          (QuantLib 1.32)
```

- **Service A (C++ Engine):** The "Muscle". Uses QuantLib 1.32 to perform precise analytic pricing of European options (Delta, Gamma, Theta).
- **Service B (Go Gateway):** The "Orchestrator". Handles REST requests, manages real-time market data fetching from Yahoo Finance, and aggregates results.

---

## 🚀 Getting Started

### Prerequisites
- **Docker & Docker Compose** (Required)
- **Python 3.x** (For the interactive CLI)
- **Make** (Optional, but recommended)

### 1. Initial Setup
RiskGo uses a **Base Image Strategy** to cache the heavy QuantLib compilation. You only need to build the "Factory" once.

```bash
make setup       # Download & vendorize QuantLib source
make build-base  # Compile the base image (approx. 10 mins)
```
*Note: If you modify the library source in `vendor/`, run `make build-base` again to update the cache.*

### 2. Launch the Services
Start the Go and C++ services in the background:
```bash
make up
```
The gateway will be available at `http://localhost:3000`.

### 3. Run Your First Analysis
Use the interactive Python CLI to analyze a position:
```bash
pip install requests
python3 risk_cli.py
```

---

## 🧪 Testing & Maintenance

### Automated Tests
```bash
make test-unit         # Run internal Go logic tests (cache, provider)
make test-integration  # Run a full end-to-end API scenario check
```

### Cleanup
- `make clean`: Removes app containers but **keeps your 10-minute base image safe**.
- `make clean-all`: Wipes everything, including the base image and vendorized source.

---

## 📡 API Reference

### `POST /api/analyze_portfolio`
Calculates Greeks and PnL shocks for a multi-leg portfolio.

**Request Schema:**
- `positions`: Array of ticker positions.
- `scenario_range`: Array of spot price shocks (e.g., `-0.10` for -10%).
- `volatility` (Optional): Global override for annual volatility (e.g., `0.3`). Defaults to 30% if not provided by provider.

**Example Request:**
```json
{
  "positions": [
    {
      "ticker": "AAPL",
      "quantity": 100,
      "beta": 1.0,
      "legs": [{"type": "CALL", "strike": 180, "expiry": "2026-06-19"}]
    }
  ],
  "scenario_range": [-0.10, 0, 0.10],
  "volatility": 0.25
}
```

**Example Response:**
```json
{
  "AAPL": {
    "-0.10": {
      "pnl": -2473.16,
      "delta": 97.90,
      "gamma": 0.05,
      "theta": 217.27
    },
    "0.00": {
      "pnl": 0,
      "delta": 97.90,
      "gamma": 0.05,
      "theta": 217.27
    },
    "0.10": {
      "pnl": 2513.51,
      "delta": 97.90,
      "gamma": 0.05,
      "theta": 217.27
    }
  }
}
```

---

## 📦 Deployment & Distribution

RiskGo supports multi-platform builds (Linux, Windows/WSL2, macOS) using Docker Buildx.

### Building for Production
```bash
# Build universal images for Intel and Apple Silicon
docker buildx build --platform linux/amd64,linux/arm64 -t your-repo/risk-cpp-engine:latest -f cpp_engine/Dockerfile --push .
docker buildx build --platform linux/amd64,linux/arm64 -t your-repo/risk-go-gateway:latest -f go_gateway/Dockerfile --push .
```

End-users can run the app without source code using a simple `docker-compose.yaml` pointing to these pre-built images.
