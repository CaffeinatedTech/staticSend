#!/bin/bash

# Test script for settings functionality

# Test 1: Check if registration is enabled by default
curl -s http://localhost:8081/auth/register -X POST -d "email=test@example.com&password=password" | grep -q "HX-Redirect" && echo "✓ Registration works when enabled" || echo "✗ Registration failed when enabled"

# Test 2: Disable registration
sqlite3 ./data/test_settings.db "UPDATE app_settings SET value='false' WHERE key='registration_enabled'"

# Test 3: Check if registration is now disabled
curl -s http://localhost:8081/auth/register -X POST -d "email=test2@example.com&password=password" | grep -q "registration is currently disabled" && echo "✓ Registration correctly disabled" || echo "✗ Registration not properly disabled"

# Test 4: Re-enable registration
sqlite3 ./data/test_settings.db "UPDATE app_settings SET value='true' WHERE key='registration_enabled'"

# Test 5: Verify registration works again
curl -s http://localhost:8081/auth/register -X POST -d "email=test3@example.com&password=password" | grep -q "HX-Redirect" && echo "✓ Registration works after re-enabling" || echo "✗ Registration failed after re-enabling"

echo "Settings functionality test completed!"