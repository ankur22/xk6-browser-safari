#!/bin/bash

# Setup script for Redis cross-browser testing

set -e

echo "🚀 Setting up Redis for cross-browser visual comparison..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if docker-compose is available
if command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
else
    echo "❌ docker-compose is not available. Please install Docker Compose."
    exit 1
fi

# Start Redis
echo "📦 Starting Redis container..."
$COMPOSE_CMD up -d

# Wait for Redis to be ready
echo "⏳ Waiting for Redis to be ready..."
sleep 2

# Check if Redis is responding
if docker exec k6-redis redis-cli ping > /dev/null 2>&1; then
    echo "✅ Redis is running on localhost:6379"
else
    echo "❌ Redis failed to start"
    exit 1
fi

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Build k6 with extensions:"
echo "     xk6 build --with github.com/grafana/xk6-browser-safari=. --with github.com/grafana/xk6-redis"
echo ""
echo "  2. Run the visual comparison test:"
echo "     ./k6 run examples/visual-comparison-redis.js"
echo ""
echo "  3. Or run individual phases:"
echo "     ./k6 run --export safari examples/visual-comparison-redis.js"
echo "     ./k6 run --export chromium examples/visual-comparison-redis.js"
echo "     ./k6 run --export compare examples/visual-comparison-redis.js"
echo ""
echo "To stop Redis: $COMPOSE_CMD down"

