# Config File

If using flags that require config options, `s3scanner` will search for `config.yml` in:

* (current directory)
* `/etc/s3scanner/`
* `$HOME/.s3scanner/`

```yaml
# Required by -db
db:
  uri: "postgresql://user:pass@db.host.name:5432/schema_name"

# Required by -mq
mq:
  queue_name: "aws"
  uri: "amqp://user:pass@localhost:5672"

# providers.custom required by `-provider custom`
#   address_style - Addressing style used by endpoints.
#     type: string
#     values: "path" or "vhost"
#   endpoint_format - Format of endpoint URLs. Should contain '$REGION' as placeholder for region name
#     type: string
#   insecure - Ignore SSL errors
#     type: boolean
# regions must contain at least one option
providers:
  custom: 
    address_style: "path"
    endpoint_format: "https://$REGION.vultrobjects.com"
    insecure: false
    regions:
      - "ewr1"
```

When `s3scanner` parses the config file, it will take the `endpoint_format` and replace `$REGION` for all `regions` listed to create a list of endpoint URLs.
