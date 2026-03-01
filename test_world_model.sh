#!/bin/bash

echo "🧪 Testing World Model"
echo "======================"
echo ""

# Test 1: Store a fact
echo "1️⃣ Storing a fact..."
curl -X POST http://localhost:8080/api/world-model/store \
  -H "Content-Type: application/json" \
  -d '{
    "key": "alice workplace",
    "value": "ACME Corporation",
    "source": "user",
    "session_id": "test-1"
  }'
echo -e "\n"

# Test 2: Store another fact
echo "2️⃣ Storing another fact..."
curl -X POST http://localhost:8080/api/world-model/store \
  -H "Content-Type: application/json" \
  -d '{
    "key": "alice role",
    "value": "Software Engineer",
    "source": "user",
    "session_id": "test-1"
  }'
echo -e "\n"

# Test 3: List all facts
echo "3️⃣ Listing all facts..."
curl http://localhost:8080/api/world-model/facts
echo -e "\n"

# Test 4: Search for facts
echo "4️⃣ Searching for 'alice'..."
curl "http://localhost:8080/api/world-model/search?q=alice"
echo -e "\n"

# Test 5: Run workflow that uses world model
echo "5️⃣ Running workflow that should use stored facts..."
curl -X POST http://localhost:8080/api/agent/workflow \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Where does Alice work and what is her role?"
  }'
echo -e "\n"

echo "✅ Tests complete!"
