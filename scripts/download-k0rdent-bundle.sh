#!/bin/bash
# scripts/download-k0rdent-bundle.sh
#
# Downloads k0rdent enterprise airgap bundle and verifies signature
#
# Usage: ./download-k0rdent-bundle.sh <VERSION> <OUTPUT_DIR>
#
# Example:
#   ./scripts/download-k0rdent-bundle.sh 1.2.2 /tmp/bundles

set -e

VERSION="${1:-}"
OUTPUT_DIR="${2:-}"

if [ -z "$VERSION" ]; then
    echo "Error: VERSION argument is required"
    echo "Usage: $0 <VERSION> <OUTPUT_DIR>"
    echo "Example: $0 1.2.2 /tmp/bundles"
    exit 1
fi

if [ -z "$OUTPUT_DIR" ]; then
    OUTPUT_DIR="./build/airgap/bundle"
fi

BUNDLE_BASE_URL="https://get.mirantis.com/k0rdent-enterprise/${VERSION}"
BUNDLE_FILE="airgap-bundle-${VERSION}.tar.gz"
SIG_FILE="${BUNDLE_FILE}.sig"

# Create output directory
echo "Creating output directory: ${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

# Download bundle
echo "Downloading k0rdent enterprise bundle..."
echo "URL: ${BUNDLE_BASE_URL}/${BUNDLE_FILE}"
curl -L -o "${OUTPUT_DIR}/${BUNDLE_FILE}" "${BUNDLE_BASE_URL}/${BUNDLE_FILE}"

# Download signature
echo "Downloading signature..."
echo "URL: ${BUNDLE_BASE_URL}/${SIG_FILE}"
curl -L -o "${OUTPUT_DIR}/${SIG_FILE}" "${BUNDLE_BASE_URL}/${SIG_FILE}"

# Verify signature (if cosign is available)
if command -v cosign &> /dev/null; then
    echo "Verifying signature with cosign..."
    cosign verify-blob \
        --key https://get.mirantis.com/k0rdent-enterprise/cosign.pub \
        --signature "${OUTPUT_DIR}/${SIG_FILE}" \
        "${OUTPUT_DIR}/${BUNDLE_FILE}"
    echo "✓ Signature verified"
else
    echo "⚠ Warning: cosign not found, skipping signature verification"
    echo "  Install cosign to verify bundle authenticity:"
    echo "  https://docs.sigstore.dev/cosign/installation/"
fi

# Extract bundle
EXTRACT_DIR="${OUTPUT_DIR}/extracted"
echo "Extracting bundle to: ${EXTRACT_DIR}"
mkdir -p "${EXTRACT_DIR}"
tar -xzf "${OUTPUT_DIR}/${BUNDLE_FILE}" -C "${EXTRACT_DIR}"

echo ""
echo "✓ Bundle downloaded and extracted successfully"
echo "  Bundle: ${OUTPUT_DIR}/${BUNDLE_FILE}"
echo "  Extracted: ${EXTRACT_DIR}"
echo ""
echo "Next step: Convert to k0s bundle format"
echo "  ./scripts/convert-to-k0s-bundle.sh ${EXTRACT_DIR} ${OUTPUT_DIR}"
