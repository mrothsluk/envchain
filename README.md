# envchain

> CLI tool to manage layered environment variable configs across dev, staging, and prod with secret redaction.

---

## Installation

```bash
go install github.com/yourusername/envchain@latest
```

Or download a prebuilt binary from the [releases page](https://github.com/yourusername/envchain/releases).

---

## Usage

Define your environment configs in layered YAML files:

```
envs/
  base.yaml
  dev.yaml
  staging.yaml
  prod.yaml
```

Load and export variables for a target environment:

```bash
# Export merged env vars for staging
envchain load --env staging

# Run a command with the resolved environment
envchain run --env prod -- ./myapp serve

# Print config with secrets redacted
envchain show --env dev --redact
```

envchain merges `base.yaml` with the target environment file, with environment-specific values taking precedence. Secrets matching common patterns (e.g. keys containing `SECRET`, `TOKEN`, `PASSWORD`) are automatically redacted in output.

---

## Configuration

```yaml
# envs/base.yaml
APP_PORT: 8080
LOG_LEVEL: info
DB_HOST: localhost

# envs/prod.yaml
LOG_LEVEL: warn
DB_HOST: prod-db.internal
DB_PASSWORD: s3cr3t  # redacted in output
```

---

## Contributing

Pull requests are welcome. Please open an issue first to discuss any significant changes.

---

## License

[MIT](LICENSE) © 2024 yourusername