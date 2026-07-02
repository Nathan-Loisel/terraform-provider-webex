# Method 1: Static access token
# Set via WEBEX_TOKEN environment variable or directly:
provider "webex" {
  token = "your-access-token"
}

# Method 2: OAuth refresh (recommended for Service Apps)
# Set via WEBEX_CLIENT_ID, WEBEX_CLIENT_SECRET, WEBEX_REFRESH_TOKEN
# environment variables or directly:
# provider "webex" {
#   client_id     = "your-client-id"
#   client_secret = "your-client-secret"
#   refresh_token = "your-refresh-token"
# }
