#!/bin/bash

echo "🧪 Testing Fowergram JWT Authentication System"
echo "============================================="

BASE_URL="http://localhost:8000"

echo ""
echo "1. Testing Health Check..."
curl -s -X GET $BASE_URL/health | jq '.'

echo ""
echo "2. Testing User Registration..."
SIGNUP_RESPONSE=$(curl -s -X POST $BASE_URL/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "demo@fowergram.com", "password": "demopass123", "username": "demouser"}')
echo $SIGNUP_RESPONSE | jq '.'

echo ""
echo "3. Testing User Login..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/api/auth/signin \
  -H "Content-Type: application/json" \
  -d '{"email": "demo@fowergram.com", "password": "demopass123"}')
echo $LOGIN_RESPONSE | jq '.'

# Extract token
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.accessToken')
echo "Access Token: ${TOKEN:0:50}..."

echo ""
echo "4. Testing Protected Route /api/auth/me..."
curl -s -X GET $BASE_URL/api/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo ""
echo "5. Testing GraphQL Playground (development mode)..."
curl -s -X GET $BASE_URL/playground -I | head -1

echo ""
echo "6. Testing GraphQL Endpoint..."
curl -s -X POST $BASE_URL/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query": "query { __typename }"}' | jq '.'

echo ""
echo "7. Testing Protected Route without token (should fail)..."
curl -s -X GET $BASE_URL/api/auth/me | jq '.'

echo ""
echo "✅ Authentication System Test Complete!" 