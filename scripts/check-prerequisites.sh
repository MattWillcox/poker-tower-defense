#!/bin/bash

# This script checks if the deployment prerequisites are met

echo "Checking deployment prerequisites..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    echo "   https://docs.docker.com/get-docker/"
    exit 1
else
    echo "✅ Docker is installed."
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    echo "   https://docs.docker.com/compose/install/"
    exit 1
else
    echo "✅ Docker Compose is installed."
fi

# Check if .env file exists
if [ ! -f .env ]; then
    echo "❌ .env file not found. Please create a .env file with your environment variables."
    echo "   You can copy the .env.example file and modify it."
    exit 1
else
    echo "✅ .env file exists."
fi

# Check if game-client.html exists
if [ ! -f game-client.html ]; then
    echo "⚠️ game-client.html not found. A placeholder page will be created for GitHub Pages."
else
    echo "✅ game-client.html exists."
fi

# Check if the required directories exist
mkdir -p .github/nginx
mkdir -p scripts
mkdir -p docs

echo "✅ Required directories exist."

echo "All prerequisites are met! You can now deploy your game."
echo ""
echo "To deploy to GitHub Pages and Digital Ocean:"
echo "1. Push your changes to GitHub"
echo "2. Set up the required secrets in your GitHub repository"
echo "3. The GitHub Actions workflows will automatically deploy your game"
echo ""
echo "For more information, see the deployment guide: .github/DEPLOYMENT.md" 