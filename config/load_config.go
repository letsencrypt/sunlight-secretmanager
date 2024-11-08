package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// Listen is the address to listen on, e.g. ":443".
	Listen string

	// ACME is the configuration for the ACME client. Optional. If missing,
	// Sunlight will listen for plain HTTP or h2c.
	ACME struct {
		// Email is the email address to use for ACME account registration.
		Email string

		// Host is the name for which autocert will obtain a certificate.
		Host string

		// Cache is the path to the autocert cache directory.
		Cache string

		// Directory is an ACME directory URL to request a certificate from.
		// Defaults to Let's Encrypt Production. Optional.
		Directory string
	}

	// Checkpoints, ETagS3, or DynamoDB store the latest checkpoint for each
	// log, with compare-and-swap semantics.
	//
	// Note that these are global as an extra safety measure: entries are keyed
	// by log ID (the hash of the public key), so even in case of
	// misconfiguration of the logs entries, even across different concurrent
	// instances of Sunlight, a log can't split.
	//
	// Only one of these can be set at the same time.

	// Checkpoints is the path to the SQLite file.
	//
	// The database must already exist to protect against accidental
	// misconfiguration. Create the table with:
	//
	//     $ sqlite3 checkpoints.db "CREATE TABLE checkpoints (logID BLOB PRIMARY KEY, body TEXT)"
	//
	Checkpoints string

	// ETagS3 is an S3-compatible object storage bucket that supports ETag on
	// both reads and writes, and If-Match on writes, such as Tigris.
	ETagS3 struct {
		// Region is the AWS region for the S3 bucket.
		Region string

		// Bucket is the name of the S3 bucket.
		Bucket string

		// Endpoint is the base URL the AWS SDK will use to connect to S3.
		Endpoint string
	}

	DynamoDB struct {
		// Region is the AWS region for the DynamoDB table.
		Region string

		// Table is the name of the DynamoDB table.
		//
		// The table must have a primary key named "logID" of type binary.
		Table string

		// Endpoint is the base URL the AWS SDK will use to connect to DynamoDB. Optional.
		Endpoint string
	}

	Logs []LogConfig
}

type LogConfig struct {
	// Name is the fully qualified log name for the checkpoint origin line, as a
	// schema-less URL. It doesn't need to be where the log is actually hosted,
	// but that's advisable.
	Name string

	// ShortName is the short name for the log, used as a metrics and logs label.
	ShortName string

	// Inception is the creation date of the log, as an RFC 3339 date.
	//
	// On the inception date, the log will be created if it doesn't exist. After
	// that date, a non-existing log will be a fatal error. This assumes it is
	// due to misconfiguration, and prevents accidental forks.
	Inception string

	// HTTPPrefix is the prefix for the HTTP endpoint of this log instance,
	// without trailing slash, but with a leading slash if not empty, and
	// without "/ct/v1" suffix.
	HTTPPrefix string

	// Roots is the path to the accepted roots as a PEM file.
	Roots string

	// Seed is the path to a file containing a secret seed from which the log's
	// private keys are derived. The whole file is used as HKDF input.
	//
	// To generate a new seed, run:
	//
	//   $ head -c 32 /dev/urandom > seed.bin
	//
	Seed string

	// PublicKey is the SubjectPublicKeyInfo for this log, base64 encoded.
	//
	// This is the same format as used in Google and Apple's log list JSON files.
	//
	// To generate this from a seed, run:
	//
	//   $ sunlight-keygen log.example/logA seed.bin
	//
	// If provided, the loaded private Key is required to match it. Optional.
	PublicKey string

	// Cache is the path to the SQLite deduplication cache file.
	Cache string

	// PoolSize is the maximum number of chains pending in the sequencing pool.
	// Since the pool is sequenced every second, it works as a qps limit. If the
	// pool is full, add-chain requests will be rejected with a 503. Zero means
	// no limit.
	PoolSize int

	// S3Region is the AWS region for the S3 bucket.
	S3Region string

	// S3Bucket is the name of the S3 bucket. This bucket must be dedicated to
	// this specific log instance.
	S3Bucket string

	// S3Endpoint is the base URL the AWS SDK will use to connect to S3. Optional.
	S3Endpoint string

	// S3KeyPrefix is a prefix on all keys written to S3. Optional.
	//
	// S3 doesn't have directories, but using a prefix ending in a "/" is
	// going to be treated like a directory in many tools using S3.
	S3KeyPrefix string

	// NotAfterStart is the start of the validity range for certificates
	// accepted by this log instance, as and RFC 3339 date.
	NotAfterStart string

	// NotAfterLimit is the end of the validity range (not included) for
	// certificates accepted by this log instance, as and RFC 3339 date.
	NotAfterLimit string
}

// Exported for use in main.go.
func LoadConfigFromYaml(configFile string) (map[string]string, error) {
	yml, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var sunlightConfig Config

	if err := yaml.Unmarshal(yml, &sunlightConfig); err != nil {
		return nil, err
	}

	logs := sunlightConfig.Logs
	nameSeedMap := make(map[string]string)

	for i := range logs {
		nameSeedMap[logs[i].Name] = logs[i].Seed
	}

	return nameSeedMap, nil
}
