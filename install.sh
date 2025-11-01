#!/bin/bash

# Code Quality Reviewer - Installation Script
# This script helps you set up the Code Quality Reviewer quickly

set -e

BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "   Code Quality Reviewer - Installation"
echo "   Powered by AWS Bedrock Claude Sonnet 4.5"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${NC}"
echo ""

# Check prerequisites
echo -e "${BLUE}📋 Checking prerequisites...${NC}"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    echo "Please install Docker: https://docs.docker.com/get-docker/"
    exit 1
fi
echo -e "${GREEN}✅ Docker found${NC}"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    echo "Please install Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi
echo -e "${GREEN}✅ Docker Compose found${NC}"

echo ""

# Create .env file
echo -e "${BLUE}🔧 Setting up configuration...${NC}"

if [ ! -f .env ]; then
    echo -e "${YELLOW}📝 Creating .env file from template...${NC}"
    cp .env.docker .env

    echo ""
    echo -e "${YELLOW}⚠️  Please edit .env file and add your credentials:${NC}"
    echo "   1. AWS_BEARER_TOKEN_BEDROCK - Your AWS Bedrock API key"
    echo "   2. BEDROCK_MODEL_ARN - Your Bedrock model ARN"
    echo "   3. API_KEY - A strong random API key"
    echo ""
    echo "   Example:"
    echo "   AWS_BEARER_TOKEN_BEDROCK=ABSKQmVkcm9ja0FQSUtleS0..."
    echo "   BEDROCK_MODEL_ARN=arn:aws:bedrock:ap-south-1:123456:inference-profile/..."
    echo "   API_KEY=$(openssl rand -hex 32 2>/dev/null || echo 'your-secure-random-key')"
    echo ""

    read -p "Press Enter after you've edited .env file..."
else
    echo -e "${GREEN}✅ .env file already exists${NC}"
fi

# Validate .env
if ! grep -q "AWS_BEARER_TOKEN_BEDROCK=.\+" .env; then
    echo -e "${RED}❌ AWS_BEARER_TOKEN_BEDROCK is not set in .env${NC}"
    echo "Please edit .env and add your AWS Bedrock API key"
    exit 1
fi

if ! grep -q "API_KEY=.\+" .env; then
    echo -e "${RED}❌ API_KEY is not set in .env${NC}"
    echo "Please edit .env and add a strong API key"
    exit 1
fi

echo -e "${GREEN}✅ Configuration validated${NC}"
echo ""

# Build and start services
echo -e "${BLUE}🐳 Building and starting Docker containers...${NC}"
echo "This may take a few minutes on first run..."
echo ""

docker-compose up -d --build

echo ""
echo -e "${GREEN}✅ Services started successfully!${NC}"
echo ""

# Wait for services to be healthy
echo -e "${BLUE}⏳ Waiting for services to be ready...${NC}"
sleep 10

# Check API health
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -s http://localhost:3000/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ API server is ready!${NC}"
        break
    fi

    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
        echo -e "${RED}❌ API server failed to start${NC}"
        echo "Check logs: docker-compose logs api"
        exit 1
    fi

    echo -n "."
    sleep 2
done

echo ""
echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}🎉 Installation Complete!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BLUE}📊 Services running:${NC}"
echo "   • API Server:  http://localhost:3000"
echo "   • Dashboard:   http://localhost:3001"
echo "   • Database:    localhost:3306"
echo ""
echo -e "${BLUE}🔑 Your API Key (from .env):${NC}"
API_KEY=$(grep "^API_KEY=" .env | cut -d '=' -f2)
echo "   $API_KEY"
echo ""
echo -e "${BLUE}📝 Next steps:${NC}"
echo ""
echo "1. Test the API:"
echo "   curl http://localhost:3000/health"
echo ""
echo "2. Add a repository to track:"
echo "   curl -X POST http://localhost:3000/api/organizations \\"
echo "     -H 'Authorization: Bearer $API_KEY' \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{"
echo "       \"name\": \"My Organization\","
echo "       \"github_org_name\": \"your-github-org\","
echo "       \"github_token\": \"ghp_your_github_token\","
echo "       \"repos\": [\"repo-name\"]"
echo "     }'"
echo ""
echo "3. View the dashboard:"
echo "   Open http://localhost:3001 in your browser"
echo ""
echo -e "${BLUE}📚 Useful commands:${NC}"
echo "   • View logs:        docker-compose logs -f"
echo "   • Stop services:    docker-compose stop"
echo "   • Restart services: docker-compose restart"
echo "   • Remove all:       docker-compose down -v"
echo ""
echo -e "${BLUE}📖 Documentation:${NC}"
echo "   • Setup Guide:  SETUP.md"
echo "   • Quick Start:  QUICKSTART.md"
echo "   • Security:     SECURITY.md"
echo ""
