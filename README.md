# libdns-omglol

[![Go Reference](https://pkg.go.dev/badge/github.com/folone/libdns-omglol.svg)](https://pkg.go.dev/github.com/folone/libdns-omglol)
[![CI](https://github.com/folone/libdns-omglol/actions/workflows/ci.yml/badge.svg)](https://github.com/folone/libdns-omglol/actions/workflows/ci.yml)

[omg.lol](https://omg.lol) DNS provider for [libdns](https://github.com/libdns/libdns).

This package implements the libdns interfaces against the [omg.lol DNS API](https://api.omg.lol), enabling programmatic management of DNS records for omg.lol addresses. It is the provider used by [`caddy-dns-omglol`](https://github.com/folone/caddy-dns-omglol) to automate ACME DNS-01 certificate challenges with Caddy.

## Install

```bash
go get github.com/folone/libdns-omglol
```

## Usage

```go
import "github.com/folone/libdns-omglol"

provider := &omglol.Provider{
    APIKey:  "your-omglol-api-key",
    Address: "yourname", // your omg.lol handle, e.g. "yourname" for yourname.omg.lol
}
```

### Implemented interfaces

| Interface | Method |
|-----------|--------|
| `libdns.RecordGetter` | `GetRecords` |
| `libdns.RecordAppender` | `AppendRecords` |
| `libdns.RecordSetter` | `SetRecords` |
| `libdns.RecordDeleter` | `DeleteRecords` |

## omg.lol DNS API

| Method | Endpoint | Used by |
|--------|----------|---------|
| `GET` | `/address/{address}/dns` | `GetRecords`, `SetRecords`, `DeleteRecords` |
| `POST` | `/address/{address}/dns` | `AppendRecords`, `SetRecords` |
| `PATCH` | `/address/{address}/dns/{id}` | `SetRecords` |
| `DELETE` | `/address/{address}/dns/{id}` | `DeleteRecords` |

Authentication uses a Bearer token (`Authorization: Bearer <api_key>`). You can find your API key in your [omg.lol account settings](https://home.omg.lol/account).

## Name conventions

omg.lol stores DNS record names as `<label>.<address>` (e.g. `_acme-challenge.yourname`). This package translates transparently between that format and the zone-relative names used by libdns (e.g. `_acme-challenge`).

## Related

- [`caddy-dns-omglol`](https://github.com/folone/caddy-dns-omglol) — Caddy module that uses this provider for ACME DNS-01 challenges
- [libdns](https://github.com/libdns/libdns) — the interface this package implements
- [omg.lol API docs](https://api.omg.lol)
