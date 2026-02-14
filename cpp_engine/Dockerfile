# Stage 1: Build the application using the pre-built base
FROM riskgo-base AS builder

WORKDIR /app
COPY cpp_engine/ .
COPY protos/ ../protos/
RUN mkdir -p build && cd build && \
    cmake .. && \
    make -j$(nproc)

# Stage 2: Runtime
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y \
    libboost-system-dev \
    libgrpc++1 \
    libprotobuf23 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/lib/libQuantLib.so.1* /usr/local/lib/
COPY --from=builder /app/build/risk_engine_server /usr/local/bin/
RUN ldconfig

EXPOSE 50051
CMD ["risk_engine_server"]
