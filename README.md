# dynamic-dns-route53

*A little golang application to register a DNS record with Route53 for a machine behind NAT on a dynamic IP. Handy for running a home server.*

## Usage

This application sends a DNS query for OpenDNS's `myip.opendns.com` subdomain, which will respond with the A Record for your current public IP (an alternative method could be to use a HTTP query to https://diagnostic.opendns.com/myip).

The application will cache the result in a cache file in `/tmp` and compare it between runs. If the file changes (or is missing), it'll then attempt to upsert an A Record to Route53 for the configured domain and the new IP address.

The `dynamic-dns-route53` application should respect all the typical ways your AWS credentials can be configured (ex. files at `~/.aws/credentials`, IAM role, etc). The application will neeed the `route53:ChangeResourceRecordSets` permission.


```
$ ./dynamic-dns-route53 -help
Usage of ./dynamic-dns-route53:
  -name string
        domain name
  -path string
        path to cache file (default "/tmp/.dynamic-dns-route53.cache")
  -server string
        resolver (default "resolver1.opendns.com")
  -target string
        lookup target (default "myip.opendns.com")
  -ttl int
        TTL in seconds (default 60)
  -zoneId string
        Route53 zone ID
```
