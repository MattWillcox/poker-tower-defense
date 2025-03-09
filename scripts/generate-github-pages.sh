#!/bin/bash

# This script generates a GitHub Pages version of the game client

# Create docs directory if it doesn't exist
mkdir -p docs

# Get the backend URL from the command line argument or use a default
BACKEND_URL=${1:-"your-droplet-ip-or-domain.com"}

# Copy the game-client.html content
cat game-client.html > docs/game-client-content.html

# Create the index.html file with the template
cat > docs/index.html << EOL
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Poker Tower Defense</title>
    <style>
$(grep -n '<style>' game-client.html | cut -d: -f2- | sed 's/<style>//' | sed 's/<\/style>//')
    </style>
    <script>
        // This script will replace the backend URLs when the page loads
        window.backendUrl = "${BACKEND_URL}";
        
        // Function to replace localhost URLs with the backend URL
        function replaceBackendUrls() {
            // Find all script tags that might contain WebSocket or API URLs
            const scripts = document.querySelectorAll('script:not([src])');
            scripts.forEach(script => {
                if (script.innerHTML.includes('localhost:3000')) {
                    script.innerHTML = script.innerHTML
                        .replace(/ws:\/\/localhost:3000\/ws/g, 'wss://' + window.backendUrl + '/ws')
                        .replace(/http:\/\/localhost:3000\/api/g, 'https://' + window.backendUrl + '/api');
                }
            });
        }
        
        // Execute after the page has loaded
        document.addEventListener('DOMContentLoaded', replaceBackendUrls);
    </script>
</head>
<body>
$(grep -v -e '<!DOCTYPE' -e '<html' -e '<head>' -e '<meta' -e '<title>' -e '<style>' -e '</style>' -e '</head>' -e '<body>' -e '</body>' -e '</html>' game-client.html)
</body>
</html>
EOL

# Replace any remaining localhost:3000 references
sed -i "s|ws://localhost:3000|wss://${BACKEND_URL}|g" docs/index.html
sed -i "s|http://localhost:3000|https://${BACKEND_URL}|g" docs/index.html

echo "GitHub Pages version generated in the docs directory with backend URL: ${BACKEND_URL}"
echo "Your game will be available at https://your-username.github.io/your-repo-name/"
echo "The client will connect to your backend using secure WebSocket (WSS) and HTTPS." 