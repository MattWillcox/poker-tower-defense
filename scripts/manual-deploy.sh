#!/bin/bash

# This script performs a manual deployment of the game to a Digital Ocean droplet
# Usage: ./manual-deploy.sh <ssh_user> <ssh_host> <domain_name> [email]

# Check if required arguments are provided
if [ $# -lt 3 ]; then
    echo "Usage: $0 <ssh_user> <ssh_host> <domain_name> [email]"
    echo "Example: $0 root 123.456.789.123 example.com admin@example.com"
    exit 1
fi

SSH_USER=$1
SSH_HOST=$2
DOMAIN_NAME=$3
EMAIL=${4:-"admin@example.com"}

echo "Deploying to $SSH_HOST as $SSH_USER with domain $DOMAIN_NAME"

# Build the Docker image
echo "Building Docker image..."
docker-compose build

# Get the image ID
BACKEND_IMAGE_ID=$(docker images -q realtime-game-backend-backend)

if [ -z "$BACKEND_IMAGE_ID" ]; then
    echo "Error: Backend image not found. Trying alternative method..."
    BACKEND_IMAGE_ID=$(docker images | grep backend | awk '{print $3}' | head -n 1)
    
    if [ -z "$BACKEND_IMAGE_ID" ]; then
        echo "Could not find any backend image. Exiting."
        exit 1
    fi
fi

echo "Found backend image with ID: $BACKEND_IMAGE_ID"

# Save the image
echo "Saving Docker image..."
docker save $BACKEND_IMAGE_ID > backend.tar

# Create deployment directory on the server
echo "Creating deployment directory on the server..."
ssh $SSH_USER@$SSH_HOST "mkdir -p ~/realtime-game-backend"

# Transfer files
echo "Transferring files to the server..."
scp docker-compose.yml $SSH_USER@$SSH_HOST:~/realtime-game-backend/

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating default .env file..."
    echo "DB_HOST=postgres" > .env
    echo "DB_PORT=5432" >> .env
    echo "DB_USER=postgres" >> .env
    echo "DB_PASSWORD=postgres" >> .env
    echo "DB_NAME=gamedb" >> .env
    echo "REDIS_HOST=redis" >> .env
    echo "REDIS_PORT=6379" >> .env
fi

scp .env $SSH_USER@$SSH_HOST:~/realtime-game-backend/
scp backend.tar $SSH_USER@$SSH_HOST:~/realtime-game-backend/
scp -r .github/nginx $SSH_USER@$SSH_HOST:~/realtime-game-backend/
scp scripts/setup-ssl.sh $SSH_USER@$SSH_HOST:~/realtime-game-backend/
ssh $SSH_USER@$SSH_HOST "chmod +x ~/realtime-game-backend/setup-ssl.sh"

# Deploy on the server
echo "Deploying on the server..."
ssh $SSH_USER@$SSH_HOST "cd ~/realtime-game-backend && \
docker load < backend.tar && \
BACKEND_IMAGE_ID=\$(docker images | grep -v REPOSITORY | head -n 1 | awk '{print \$3}') && \
docker tag \$BACKEND_IMAGE_ID realtime-game-backend-backend:latest && \
docker-compose down && \
docker-compose up -d && \
if [ ! -f /etc/nginx/sites-available/game ]; then \
  sudo cp ~/realtime-game-backend/nginx/game.conf /etc/nginx/sites-available/game && \
  sudo sed -i 's/DOMAIN_NAME/$DOMAIN_NAME/g' /etc/nginx/sites-available/game && \
  sudo ln -sf /etc/nginx/sites-available/game /etc/nginx/sites-enabled/ && \
  sudo nginx -t && \
  sudo systemctl reload nginx; \
fi && \
if [ ! -d /etc/letsencrypt/live/$DOMAIN_NAME ]; then \
  echo 'Setting up SSL/TLS with Let\'s Encrypt...' && \
  ./setup-ssl.sh $DOMAIN_NAME $EMAIL; \
fi"

echo "Deployment completed!"
echo "Your game should now be accessible at:"
echo "  - Backend: https://$DOMAIN_NAME"
echo "  - GitHub Pages: https://your-username.github.io/your-repo-name/" 