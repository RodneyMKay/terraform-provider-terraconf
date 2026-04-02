provider "terraconf" {
}

data "terraconf_yaml" "input_data" {
  input_glob  = "./dnsrecords/**/*.{yaml,yml}"
  schema_file = "./schema/record.schema.json"
}

resource "cloudflare_dns_record" "records" {
  for_each = {
    for resource in data.terraconf_yaml.input_data.output
    : resource.id => resource
  }

  type = "A"
  tags = ["source:generated"]

  zone_id = each.value.zone_id
  name    = each.value.name
  ttl     = each.value.ttl
  content = each.value.target_ipv4
}
