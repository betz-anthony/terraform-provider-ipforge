terraform {
  required_providers {
    ipforge = {
      source = "betz-anthony/ipforge"
    }
  }
}

# url and token may also be supplied via the IPFORGE_URL / IPFORGE_TOKEN
# environment variables (recommended for the token).
provider "ipforge" {
  url   = "https://ipforge.example.com"
  token = var.ipforge_token # ipfg_... API token
}
