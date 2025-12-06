#!/bin/bash
# Save this as test-connectivity.sh in your project root
# Run with: bash test-connectivity.sh

set -e

ELASTIC_PASSWORD="${ELASTIC_PASSWORD:-changeme_secure_password_here}"

echo "=== Testing Elasticsearch Connectivity ==="
echo ""

echo "1. Testing Elasticsearch health (via service name)..."
# Try without auth first (security might be disabled)
if curl -fs http://elasticsearch:9200/_cluster/health > /dev/null 2>&1; then
    echo "✓ Elasticsearch is healthy (security disabled)"
    SECURITY_ENABLED=false
elif curl -fs -u "elastic:${ELASTIC_PASSWORD}" http://elasticsearch:9200/_cluster/health > /dev/null 2>&1; then
    echo "✓ Elasticsearch is healthy (security enabled)"
    SECURITY_ENABLED=true
else
    echo "✗ Elasticsearch connection failed"
    exit 1
fi

echo ""
if [ "$SECURITY_ENABLED" = "true" ]; then
    echo "2. Testing Elasticsearch authentication..."
    if curl -fs -u "elastic:${ELASTIC_PASSWORD}" http://elasticsearch:9200/_security/_authenticate > /dev/null 2>&1; then
        echo "✓ Authentication successful"
    else
        echo "✗ Authentication failed"
        exit 1
    fi

    echo ""
    echo "3. Testing kibana_system user..."
    if curl -fs -u "kibana_system:${ELASTIC_PASSWORD}" http://elasticsearch:9200/_security/_authenticate > /dev/null 2>&1; then
        echo "✓ kibana_system user works"
    else
        echo "✗ kibana_system authentication failed"
        exit 1
    fi
else
    echo "2. Skipping authentication tests (security is disabled)"
    echo ""
    echo "3. Skipping kibana_system test (security is disabled)"
fi

echo ""
echo "=== Testing Kibana Connectivity ==="
echo ""

echo "4. Testing Kibana status..."
if curl -fs http://kibana:5601/api/status | grep -q "available"; then
    echo "✓ Kibana is available"
else
    echo "⚠ Kibana may not be fully ready yet"
fi

echo ""
echo "=== All Tests Complete ==="
echo ""
echo "Connection URLs from within the container:"
echo "  - Elasticsearch: http://elasticsearch:9200"
echo "  - Kibana: http://kibana:5601"
echo ""
echo "Port forwarding from your host machine:"
echo "  - Elasticsearch: http://localhost:9200"
echo "  - Kibana: http://localhost:5601"
echo "  - GoCrawl: http://localhost:8080"
