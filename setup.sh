#!/bin/bash
set -e

QUANTLIB_VERSION="1.41"
QUANTLIB_URL="https://github.com/lballabio/QuantLib/releases/download/v${QUANTLIB_VERSION}/QuantLib-${QUANTLIB_VERSION}.tar.gz"

echo "Checking for vendorized QuantLib..."
if [ ! -d "vendor/quantlib" ]; then
    echo "QuantLib not found in vendor/. Downloading version ${QUANTLIB_VERSION}..."
    mkdir -p vendor
    wget -q --show-progress $QUANTLIB_URL -O QuantLib.tar.gz
    tar -xzf QuantLib.tar.gz -C vendor/
    mv vendor/QuantLib-${QUANTLIB_VERSION} vendor/quantlib
    rm QuantLib.tar.gz
    echo "QuantLib vendorized successfully."
else
    echo "QuantLib already vendorized."
fi

echo "Setup complete. You can now run 'docker-compose up --build'."
