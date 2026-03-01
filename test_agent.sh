#!/bin/bash

echo "🧪 Testing Multi-Agent Workflow"
echo "================================"

curl -s -X POST http://localhost:8080/api/agent/workflow \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Explain quantum computing in simple terms"
  }' | jq


echo ""
echo "✅ Test complete"
