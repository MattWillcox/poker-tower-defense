#!/bin/bash

# This script sets up SSL/TLS with Let's Encrypt on a Digital Ocean droplet
# Usage: ./setup-ssl.sh <domain_name> <email>

# Check if domain name and email are provided
if [ $# -lt 2 ]; then
    echo "Usage: $0 <domain_name> <email>"
    exit 1
fi

DOMAIN=$1
EMAIL=$2

echo "Setting up SSL/TLS for $DOMAIN with email $EMAIL"

# Install Certbot
echo "Installing Certbot..."
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx

# Create directory for ACME challenge
sudo mkdir -p /var/www/certbot

# Obtain SSL certificate
echo "Obtaining SSL certificate for $DOMAIN..."
sudo certbot certonly --webroot -w /var/www/certbot -d $DOMAIN --email $EMAIL --agree-tos --non-interactive

# Check if certificate was obtained successfully
if [ ! -d "/etc/letsencrypt/live/$DOMAIN" ]; then
    echo "Failed to obtain SSL certificate. Please check the domain name and try again."
    exit 1
fi

echo "SSL certificate obtained successfully!"

# Set up auto-renewal
echo "Setting up auto-renewal..."
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer

# Test auto-renewal
echo "Testing auto-renewal..."
sudo certbot renew --dry-run

echo "SSL/TLS setup completed successfully!"
echo "Your site should now be accessible at https://$DOMAIN"
echo ""
echo "To test WebSocket connections, use:"
echo "  wss://$DOMAIN/ws"
echo ""
echo "To test API connections, use:"
echo "  https://$DOMAIN/api"
echo ""
echo "Remember to update your GitHub Pages client to use WSS instead of WS!" 