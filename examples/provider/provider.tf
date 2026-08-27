terraform {
  required_providers {
    freshdesk = {
      source  = "slop-place/freshdesk"
      version = "~> 0.1"
    }
  }
}

# Credentials are read from FRESHDESK_DOMAIN and FRESHDESK_API_KEY when they
# are not given here.
provider "freshdesk" {
  domain  = "acme" # or "acme.freshdesk.com"
  api_key = var.freshdesk_api_key
}

variable "freshdesk_api_key" {
  type      = string
  sensitive = true
}
