# uploadcert_tencentcloud

## Overview

`uploadcert_tencentcloud` is an event handler module for Caddy that automatically uploads newly obtained certificates to Tencent Cloud SSL service and supports automatic certificate updates. The module supports whitelist and blacklist rules for certificates, allowing flexible control over which certificates should be uploaded.

## Features

- Listens for Caddy's `cert_obtained` event.
- Automatically uploads new certificates to Tencent Cloud SSL.
- Automatically updates existing certificates.
- Supports whitelist (`AllowList`) and blacklist (`BlockList`) rules for certificate uploads.
- Logs the upload and update processes.

## Usage

### Caddyfile Configuration Example

```caddyfile
{
    events {
        on cert_obtained upload_cert_tencentcloud {
            secret_id {$TENCENTCLOUD_SECRET_ID}
            secret_key {$TENCENTCLOUD_SECRET_KEY}
            allow_list example.com www.example.com
            block_list test.example.com
            try_delete_old_cert
        }
    }
}
```

- `secret_id` and `secret_key`: Tencent Cloud API credentials.
- `allow_list`：Only upload certificates for the domains in this list (optional).
- `block_list`：Do not upload certificates for the domains in this list (optional).
- `try_delete_old_cert`：Try to delete the old certificate when the certificate is updated (optional).

## Development & Contribution

Contributions through issues and pull requests are welcome to improve this module.

## License

MIT License
