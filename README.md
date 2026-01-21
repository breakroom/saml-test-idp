# SAML Test IDP

A self-contained SAML 2.0 Identity Provider for development and testing purposes. Designed to be used with a hardcoded list of users to simulate SAML authentication.

> [!CAUTION]
> Do not use this in production! It is designed purely for local development and testing.

## Features

- **Single Binary**: All HTML and CSS is embedded - no external files needed at runtime
- **Multiple Service Providers**: Configure multiple SPs, each with their own users and settings
- **Simple Configuration**: Single YAML config file for all settings
- **Custom User Attributes**: Define arbitrary attributes for each test user
- **No Passwords Required**: Simple dropdown UI to select a predefined user
- **IDP Metadata Endpoint**: Automatic metadata generation at `/metadata`

## Quick Start

### 1. Generate Certificates

```bash
make generate-certs
```

This creates self-signed certificates in the `certs/` directory.

### 2. Create Configuration

Copy the example configuration:

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml` to add your service providers and test users.

### 3. Build and Run

```bash
make build
make run
```

Or run directly:

```bash
./bin/saml-test-idp -config config.yaml
```

The IDP will start on `http://localhost:8080` by default.

## CLI Usage

```
Usage: saml-test-idp [options]

Options:
  -config string
        Path to YAML configuration file (default "config.yaml")
  -version
        Show version and exit
```

## Configuration

All configuration is done via the YAML config file. See `config.example.yaml` for a fully documented example.

### Example Configuration

```yaml
server:
  host: "localhost"
  port: 8080
  base_url: "http://localhost:8080"

idp:
  entity_id: "http://localhost:8080/metadata"
  certificate_path: "certs/idp.crt"
  private_key_path: "certs/idp.key"

service_providers:
  - entity_id: "https://myapp.example.com/saml/metadata"
    acs_url: "https://myapp.example.com/saml/acs"
    name_id_format: "email"  # email, persistent, transient, unspecified
    users:
      - name: "Alice Admin"
        name_id: "alice@example.com"
        attributes:
          email: "alice@example.com"
          firstName: "Alice"
          lastName: "Admin"
          groups:
            - "admins"
            - "users"
          role: "admin"
      
      - name: "Bob User"
        name_id: "bob@example.com"
        attributes:
          email: "bob@example.com"
          firstName: "Bob"
          lastName: "User"
          groups:
            - "users"
          role: "member"
```

### Configuration Reference

#### Server Settings

| Field | Description | Default |
|-------|-------------|---------|
| `server.host` | Host to bind to | `localhost` |
| `server.port` | Port to bind to | `8080` |
| `server.base_url` | Base URL for the IDP | `http://{host}:{port}` |

#### IDP Settings

| Field | Description |
|-------|-------------|
| `idp.entity_id` | Entity ID for the IDP (defaults to `{base_url}/metadata`) |
| `idp.certificate` | PEM-encoded certificate (inline) |
| `idp.certificate_path` | Path to PEM certificate file |
| `idp.private_key` | PEM-encoded private key (inline) |
| `idp.private_key_path` | Path to PEM private key file |

**Note:** Relative file paths (like `certs/idp.crt`) are resolved relative to the config file's directory, not the current working directory.

#### Service Provider Settings

| Field | Description |
|-------|-------------|
| `entity_id` | SP entity ID (required) |
| `acs_url` | Assertion Consumer Service URL |
| `metadata_file` | Path to SP metadata XML (alternative to `acs_url`) |
| `name_id_format` | Name ID format: `email`, `persistent`, `transient`, `unspecified` |
| `users` | List of test users for this SP |

#### User Settings

| Field | Description |
|-------|-------------|
| `name` | Display name shown in the login dropdown |
| `name_id` | Value used for the SAML NameID element |
| `attributes` | Arbitrary key-value attributes included in the assertion |

### Name ID Formats

| Config Value | SAML NameID Format |
|--------------|-------------------|
| `email` | `urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress` |
| `persistent` | `urn:oasis:names:tc:SAML:2.0:nameid-format:persistent` |
| `transient` | `urn:oasis:names:tc:SAML:2.0:nameid-format:transient` |
| `unspecified` | `urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified` |

## Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /metadata` | IDP metadata XML |
| `GET/POST /sso` | SSO endpoint (receives SAMLRequest from SP) |
| `GET/POST /login` | Login page with user selection |

## Integrating with Your Application

### 1. Get IDP Metadata

Point your application's SAML configuration to:

```
http://localhost:8080/metadata
```

### 2. Configure Your SP in the IDP

Add your application to `config.yaml`:

```yaml
service_providers:
  - entity_id: "your-app-entity-id"
    acs_url: "https://your-app.com/saml/callback"
    name_id_format: "email"
    users:
      - name: "Test User"
        name_id: "test@example.com"
        attributes:
          email: "test@example.com"
```

### 3. Test the Flow

1. Initiate SAML login from your application
2. You'll be redirected to the IDP login page
3. Select a user from the dropdown
4. Click "Sign In"
5. You'll be redirected back to your application with the SAML response

## Development

### Prerequisites

- Go 1.25 or later
- Make (optional, for convenience)

### Building

```bash
# Build the binary
make build

# Run tests
make test

# Run with test coverage
make test-coverage

# Format code
make fmt

# Clean build artifacts
make clean
```

## License

MIT License
