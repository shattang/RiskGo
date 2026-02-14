.PHONY: setup build-base build up down test test-unit test-integration clean clean-all

# Detect docker compose command
DOCKER_COMPOSE := docker compose

setup:
	./setup.sh

# ONLY RUN THIS ONCE - Compiles QuantLib into a base image (10 mins)
build-base:
	docker build -t riskgo-base -f cpp_engine/Dockerfile.base .

build:
	$(DOCKER_COMPOSE) build

# Now fast because it uses riskgo-base
up:
	$(DOCKER_COMPOSE) up --build

down:
	$(DOCKER_COMPOSE) down

# Safe clean - keeps your base image
clean:
	@echo "Cleaning up artifacts (keeping base image)..."
	$(DOCKER_COMPOSE) down --volumes --remove-orphans || true
	rm -rf go_gateway/proto/*.pb.go
	rm -rf cpp_engine/build/

# Aggressive clean - removes everything including compiled base
clean-all: clean
	@echo "Removing base image and vendor code..."
	docker rmi riskgo-base || true
	rm -rf vendor/

test: test-unit test-integration

test-unit:
	cd go_gateway && go test ./...

test-integration:
	@echo "Testing API endpoint..."
	curl -X POST http://localhost:3000/api/analyze_portfolio \
		-H "Content-Type: application/json" \
		-d '{"positions": [{"ticker": "AAPL", "quantity": 100, "beta": 1.0, "legs": [{"type": "CALL", "strike": 180, "expiry": "2026-06-19"}]}], "scenario_range": [-0.10, 0, 0.10]}'
