output "dns_app_id" {
  value = "${citrixitm_dns_app.website.id}"
}

output "dns_app_version" {
  value = "${citrixitm_dns_app.website.version}"
}

output "dns_app_cname" {
  value = "${citrixitm_dns_app.website.cname}"
}

output "platform_id" {
  value = citrixitm_platform.tolga_test.id
}
