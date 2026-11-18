# sunlight-secretmanager

sunlight-secretmanager is a command-line tool to manage a
[Sunlight](https://sunlight.dev/) CT Log's seed material for its private key.

All CT logs have a private key which they use to create Signed Certificate
Timestamps (SCTs) and Signed Tree Heads (STHs). Sunlight does not take this
private key as input directly. Instead, its configuration requires a seed file
containing at least 32 bytes of random data, from which the log's ECDSA P-256
key (and ML-DSA-44 key in v0.9.0 and above) will be derived.

The purpose of sunlight-secretmanager is to authenticate to AWS Secrets Manager,
retrieve a stored seed, and write it to a file in a tmpfs. It knows what seed to
retrieve and where to write the output file by parsing the same config file
which configures the Sunlight log itself.

If it successfully retrieves a secret from AWS Secrets Manager but that secret
is empty, it will generate a new seed and save it back to AWS before proceeding.
This allows for seamless setup of new log shards simply by adding them to
Terraform.

## Usage

Sign in the AWS SDK so it populates your environment with the appropriate
values, and then:

```shell
$ sunlight-secretmanager -config /path/to/sunlight/config.yml
```
