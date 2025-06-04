#!/bin/bash

echo "🧪 Testing Fowergram GraphQL API"
echo "==============================="

BASE_URL="http://localhost:8000"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Function to make GraphQL requests
graphql_request() {
    local query=$1
    local token=$2
    
    if [ -z "$token" ]; then
        curl -s -X POST $BASE_URL/graphql \
            -H "Content-Type: application/json" \
            -d "{\"query\": \"$query\"}" | jq '.'
    else
        curl -s -X POST $BASE_URL/graphql \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $token" \
            -d "{\"query\": \"$query\"}" | jq '.'
    fi
}

# 1. Login to get token
echo -e "\n${GREEN}1. Logging in...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/api/auth/signin \
    -H "Content-Type: application/json" \
    -d '{"email": "demo@fowergram.com", "password": "demopass123"}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.accessToken')
echo "Access Token: ${TOKEN:0:50}..."

# 2. Test Me Query
echo -e "\n${GREEN}2. Testing Me Query...${NC}"
graphql_request 'query { me { id email username } }' "$TOKEN"

# 3. Test Create Post Mutation
echo -e "\n${GREEN}3. Testing Create Post Mutation...${NC}"
graphql_request 'mutation { createPost(input: { content: "Test post content", mediaUrls: ["https://example.com/image.jpg"] }) { id content createdAt } }' "$TOKEN"

# 4. Test Get Posts Query
echo -e "\n${GREEN}4. Testing Get Posts Query...${NC}"
graphql_request 'query { posts(first: 10) { edges { node { id content createdAt author { username } } } } }' "$TOKEN"

# 5. Test Like Post Mutation
echo -e "\n${GREEN}5. Testing Like Post Mutation...${NC}"
# Note: Replace POST_ID with an actual post ID from the previous query
graphql_request 'mutation { likePost(postId: "POST_ID") { id likesCount } }' "$TOKEN"

# 6. Test Follow User Mutation
echo -e "\n${GREEN}6. Testing Follow User Mutation...${NC}"
# Note: Replace USER_ID with an actual user ID
graphql_request 'mutation { followUser(userId: "USER_ID") { id followersCount } }' "$TOKEN"

# 7. Test Get User Profile Query
echo -e "\n${GREEN}7. Testing Get User Profile Query...${NC}"
# Note: Replace USERNAME with an actual username
graphql_request 'query { user(username: "USERNAME") { id username bio followersCount followingCount postsCount } }' "$TOKEN"

# 8. Test Search Users Query
echo -e "\n${GREEN}8. Testing Search Users Query...${NC}"
graphql_request 'query { searchUsers(query: "demo", first: 10) { edges { node { id username } } } }' "$TOKEN"

# 9. Test Get Feed Query
echo -e "\n${GREEN}9. Testing Get Feed Query...${NC}"
graphql_request 'query { feed(first: 10) { edges { node { id content author { username } } } } }' "$TOKEN"

# 10. Test without authentication (should fail)
echo -e "\n${GREEN}10. Testing without authentication (should fail)...${NC}"
graphql_request 'query { me { id email } }'

echo -e "\n${GREEN}✅ GraphQL API Test Complete!${NC}" 