# Terraform Provider for IPForge

A native [Terraform](https://www.terraform.io) provider for
[IPForge](https://github.com/betz-anthony/terraform-provider-ipforge) IP Address
Management. It manages subnets, addresses, VLANs, address allocations, DNS
records, and DHCP reservations through the IPForge `/api/v1` HTTP API.

This replaces the previous `Mastercard/restapi` workaround — instead of
hand-rolling raw REST bodies and JSON paths against IPForge, you get typed
resources, plan/diff support, import, and idempotent allocation.

## Requirements

- Terraform >= 1.0
- An IPForge deployment and an `ipfg_` API token (Settings → API Tokens)

## Quickstart

```hcl
terraform {
  required_providers {
    ipforge = {
      source = "betz-anthony/ipforge"
    }
  }
}

provider "ipforge" {
  url = "https://ipforge.example.com"
  # token is read from the IPFORGE_TOKEN environment variable
}
```

Provider configuration:

| Argument | Env var         | Description                                   |
| -------- | --------------- | --------------------------------------------- |
| `url`    | `IPFORGE_URL`   | IPForge base URL (e.g. `https://ipforge.example.com`). |
| `token`  | `IPFORGE_TOKEN` | `ipfg_` API token. Mark sensitive; prefer the env var. |

Export the token rather than committing it:

```bash
export IPFORGE_TOKEN="ipfg_xxxxxxxxxxxxxxxxxxxx"
export IPFORGE_URL="https://ipforge.example.com"
```

## Allocation → DNS record flow

The common pattern is to claim the next free IP in a subnet and then publish a
DNS record for it. `ipforge_allocation` is idempotent by hostname and releases
the address (and any provider-side registrations) on destroy.

```hcl
data "ipforge_subnet" "app" {
  cidr = "10.20.0.0/24"
}

resource "ipforge_allocation" "web01" {
  subnet_id    = data.ipforge_subnet.app.id
  hostname     = "web01"
  description  = "Web frontend"
  register_dns = false
}

resource "ipforge_dns_record" "web01" {
  zone        = "lab.example.com"
  name        = "web01.lab.example.com"
  record_type = "A"
  value       = ipforge_allocation.web01.address
  ttl         = 3600
}

output "web01_ip" {
  value = ipforge_allocation.web01.address
}
```

`ipforge_allocation` can also register DNS/DHCP in one step (`register_dns`,
`register_dhcp`, `dns_zone`) when you would rather IPForge own those records than
manage them as separate `ipforge_dns_record` / `ipforge_dhcp_reservation`
resources.

## Resources and data sources

| Type                        | Kind        | Purpose                                          |
| --------------------------- | ----------- | ------------------------------------------------ |
| `ipforge_subnet`            | resource    | Subnet (CIDR, name, VLAN, parent, description).   |
| `ipforge_address`           | resource    | A specific address within a subnet.               |
| `ipforge_allocation`        | resource    | Claim the next free address by hostname.          |
| `ipforge_vlan`              | resource    | VLAN definition.                                  |
| `ipforge_dns_record`        | resource    | DNS record in a zone (all attributes ForceNew).   |
| `ipforge_dhcp_reservation`  | resource    | Static DHCP reservation in a scope.               |
| `ipforge_subnet`            | data source | Look up a subnet by CIDR or name.                 |
| `ipforge_address`           | data source | Look up an address by IP.                         |

See [`examples/`](./examples) for per-resource usage and a full end-to-end
configuration.

## Development

```bash
go build ./...
go test ./...        # unit tests (httptest-based client tests)
```

Acceptance tests run against a live IPForge instance and are gated on `TF_ACC`:

```bash
TF_ACC=1 IPFORGE_URL=... IPFORGE_TOKEN=... go test ./internal/provider/ -v
```

Regenerate the provider documentation under `docs/` after schema changes:

```bash
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
tfplugindocs generate --provider-name ipforge
```

## License

See the repository for license details.
