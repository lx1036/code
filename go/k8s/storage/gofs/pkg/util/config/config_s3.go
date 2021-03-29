package config

import "time"

const (
	Region              = "region"
	DisableSSL          = "disableSSL"
	S3ForcePathStyle    = "S3ForcePathStyle"
	HTTPTimeout         = "HTTPTimeout"
	AccessKey           = "AccessKey"
	SecretKey           = "SecretKey"
	NoParallelMultipart = "NoParallelMultipart"
)

type S3Config struct {
	// Common Backend Config
	Region           string
	Endpoint         string
	AccessKey        string
	SecretKey        string
	Version          string
	DisableSSL       bool
	S3ForcePathStyle bool
	HTTPTimeout      time.Duration

	NoParallelMultipart bool
}

func ParseS3Config(cfg *Config) (*S3Config, error) {
	s3cfg := new(S3Config)

	s3cfg.Region = cfg.GetStringWithDefault(Region, "unused")
	s3cfg.DisableSSL = cfg.GetBoolWithDefault(DisableSSL, true)
	s3cfg.AccessKey = cfg.GetString(AccessKey)
	s3cfg.SecretKey = cfg.GetString(SecretKey)
	s3cfg.S3ForcePathStyle = cfg.GetBoolWithDefault(S3ForcePathStyle, true)
	s3cfg.HTTPTimeout = cfg.GetDuration(HTTPTimeout)

	s3cfg.NoParallelMultipart = cfg.GetBool(NoParallelMultipart)

	return s3cfg, nil
}
