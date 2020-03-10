provider "citrixitm" {
  client_id     = var.itm_client_id
  client_secret = var.itm_client_secret
}

resource "citrixitm_dns_app" "website" {
  name           = "Website"
  description    = "DNS routing for the website"
  app_data       = file("${path.module}/website.dns.js")
  fallback_cname = "origin.example.com"
}

resource "citrixitm_platform" "platform_1" {
  name            = "Platform 1"
  description     = "Platform 1 description"
  category        = "Delivery Networks"
  openmix_enabled = true

  radar = {
    probe_response_time_url = "https://example.com/cedexis/r20.gif"
    probe_availability_url  = "https://example.com/cedexis/r20.gif"
    probe_throughput_url    = "https://example.com/cedexis/r20-100KB.png"
    weight                  = 10
  }
}
