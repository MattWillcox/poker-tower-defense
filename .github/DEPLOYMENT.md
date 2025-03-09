# Deployment Guide for Realtime Game Backend

This guide explains how to set up GitHub Actions to automatically deploy your game to a Digital Ocean droplet.

## Prerequisites

1. A Digital Ocean droplet with Ubuntu installed
2. Docker and Docker Compose installed on the droplet
3. Nginx installed on the droplet
4. A GitHub repository with your game code

## Setting Up GitHub Secrets

You need to add the following secrets to your GitHub repository:

1. Go to your GitHub repository
2. Click on "Settings" > "Secrets and variables" > "Actions"
3. Add the following secrets:

### SSH_PRIVATE_KEY

This is your private SSH key that allows GitHub Actions to connect to your Digital Ocean droplet.

To generate a new SSH key pair:
```bash
ssh-keygen -t rsa -b 4096 -f ~/.ssh/github_actions
```

Add the public key to your Digital Ocean droplet:
```bash
cat ~/.ssh/github_actions.pub | ssh user@your-droplet-ip "cat >> ~/.ssh/authorized_keys"
```

Add the private key content as the `SSH_PRIVATE_KEY` secret:
```bash
cat ~/.ssh/github_actions
```

### SSH_KNOWN_HOSTS

This contains the SSH fingerprint of your Digital Ocean droplet to prevent man-in-the-middle attacks.

Get the fingerprint:
```bash
ssh-keyscan -H your-droplet-ip
```

Add the output as the `SSH_KNOWN_HOSTS` secret.

### SSH_USER

The username you use to log in to your Digital Ocean droplet (usually `root` or a user with sudo privileges).

### SSH_HOST

The IP address or hostname of your Digital Ocean droplet.

### DOMAIN_NAME

The domain name that will be used to access your game. If you don't have a domain, you can use the droplet's IP address.

## Preparing Your Droplet

1. Install Docker and Docker Compose:
```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.18.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

2. Install Nginx:
```bash
sudo apt update
sudo apt install -y nginx
```

3. Enable Nginx to start on boot:
```bash
sudo systemctl enable nginx
```

## Deployment Architecture

This project uses a split deployment architecture:

1. **Frontend (Game Client)**: Hosted on GitHub Pages
2. **Backend (Game Server)**: Hosted on your Digital Ocean droplet

This separation provides several benefits:
- Free hosting for static content on GitHub Pages
- Reduced load on your Digital Ocean droplet
- Automatic HTTPS for the client via GitHub Pages
- Simplified deployment process

### Deployment Workflow

When you push changes to your repository:

1. The `github-pages.yml` workflow generates and deploys the client to GitHub Pages
2. The `deploy.yml` workflow deploys the backend to your Digital Ocean droplet

Both workflows run independently, ensuring that your client and server are always in sync.

### Accessing Your Game

Your game will be accessible at:
- **Client**: `https://your-username.github.io/your-repo-name/`
- **Backend**: `http://your-droplet-ip-or-domain/`

The client will automatically connect to your backend using the URL specified in the `BACKEND_URL` secret.

## How the Deployment Works

When you push to the main branch, GitHub Actions will:

1. Build the Docker images
2. Transfer the necessary files to your Digital Ocean droplet
3. Load the Docker images on the droplet
4. Start the containers using Docker Compose
5. Set up Nginx to serve the game client and proxy WebSocket connections

## Testing the Deployment

After the GitHub Actions workflow completes successfully, you can access your game at:

```
http://your-domain-or-ip/
```

## Troubleshooting

If you encounter issues with the deployment:

1. Check the GitHub Actions logs for errors
2. SSH into your droplet and check the Docker logs:
```bash
docker-compose logs
```
3. Check the Nginx logs:
```bash
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

## Manual Deployment

If you need to deploy manually:

1. SSH into your droplet
2. Navigate to the deployment directory:
```bash
cd ~/realtime-game-backend
```
3. Pull the latest changes and restart the containers:
```bash
docker-compose down
docker-compose up -d
```
4. Copy the updated game client:
```bash
sudo cp game-client.html /var/www/html/game/index.html
```

## Using GitHub Pages for the Game Client

You can host the game client on GitHub Pages while running the backend on your Digital Ocean droplet. This approach has several advantages:

1. Free hosting for your static content
2. Automatic HTTPS
3. Global CDN for faster loading

### Setting Up GitHub Pages

1. Make sure your repository has GitHub Pages enabled:
   - Go to your repository on GitHub
   - Click on "Settings" > "Pages"
   - Under "Source", select the "gh-pages" branch
   - Click "Save"

2. Add the `BACKEND_URL` secret to your GitHub repository:
   - Go to your repository on GitHub
   - Click on "Settings" > "Secrets and variables" > "Actions"
   - Add a new secret named `BACKEND_URL` with the value of your Digital Ocean droplet's IP address or domain name

3. Push your changes to the main branch:
   - The GitHub Actions workflow will automatically generate the GitHub Pages version of your game client
   - The workflow will deploy the generated files to the gh-pages branch

4. Access your game:
   - Your game will be available at `https://your-username.github.io/your-repo-name/`

### Updating the Backend URL

If you need to update the backend URL:

1. Update the `BACKEND_URL` secret in your GitHub repository
2. Manually trigger the GitHub Pages workflow:
   - Go to your repository on GitHub
   - Click on "Actions" > "Deploy to GitHub Pages"
   - Click "Run workflow" > "Run workflow"

### CORS Configuration

When using GitHub Pages with your backend on a different domain, you need to configure CORS (Cross-Origin Resource Sharing) on your backend server. The Nginx configuration already includes the necessary headers, but you may need to update your Go server code to allow requests from your GitHub Pages domain.

Add the following headers to your API responses:

```go
w.Header().Set("Access-Control-Allow-Origin", "https://your-username.github.io")
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Origin, X-Requested-With")
```

Or to allow all origins during development:

```go
w.Header().Set("Access-Control-Allow-Origin", "*")
```

### WebSocket Security

When your GitHub Pages site uses HTTPS (which it does by default), your WebSocket connections must also use WSS (WebSocket Secure) to avoid mixed content issues. Our deployment setup automatically configures SSL/TLS on your Digital Ocean droplet using Let's Encrypt, which provides free SSL certificates.

#### How SSL/TLS is Set Up

The deployment workflow includes the following steps for SSL/TLS:

1. The Nginx configuration is set up to handle both HTTP and HTTPS traffic
2. HTTP traffic is automatically redirected to HTTPS
3. Let's Encrypt is used to obtain and manage SSL certificates
4. Certificates are automatically renewed before they expire

#### SSL/TLS Requirements

To set up SSL/TLS, you need:

1. A domain name pointing to your Digital Ocean droplet
2. The `DOMAIN_NAME` secret set in your GitHub repository
3. An email address for Let's Encrypt notifications (optional, set as `SSL_EMAIL` secret)

#### Adding the SSL_EMAIL Secret

1. Go to your repository on GitHub
2. Click on "Settings" > "Secrets and variables" > "Actions"
3. Add a new secret named `SSL_EMAIL` with your email address

#### Testing SSL/TLS

After deployment, you can test your secure connections:

- HTTPS: `https://your-domain.com/`
- WSS: `wss://your-domain.com/ws`

#### Troubleshooting SSL/TLS

If you encounter issues with SSL/TLS:

1. Check if the domain is correctly pointing to your Digital Ocean droplet
2. Verify that the Let's Encrypt certificates were obtained successfully:
   ```bash
   sudo certbot certificates
   ```
3. Check the Nginx error logs:
   ```bash
   sudo tail -f /var/log/nginx/error.log
   ```
4. Manually run the SSL setup script:
   ```bash
   cd ~/realtime-game-backend
   ./setup-ssl.sh your-domain.com your-email@example.com
   ``` 