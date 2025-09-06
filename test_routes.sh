#!/bin/bash

# Start server in background
./staticsend --port 3003 --db ./data/test3.db > server.log 2>&1 &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Test routes
echo "Testing routes on port 3003..."
echo ""

# Test root route (should show landing page)
echo "1. Testing GET / (should be landing page):"
curl -s -o /dev/null -w "HTTP %{http_code}" http://localhost:3003/
echo ""

# Test login route
echo "2. Testing GET /login (should be login page):"
curl -s -o /dev/null -w "HTTP %{http_code}" http://localhost:3003/login
echo ""

# Test register route  
echo "3. Testing GET /register (should be register page):"
curl -s -o /dev/null -w "HTTP %{http_code}" http://localhost:3003/register
echo ""

# Test dashboard without auth (should redirect to login)
echo "4. Testing GET /dashboard without auth (should redirect):"
curl -s -o /dev/null -w "HTTP %{http_code}" http://localhost:3003/dashboard
echo ""

# Test health endpoint
echo "5. Testing GET /health:"
curl -s http://localhost:3003/health
echo ""

# Stop server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "Test completed!"