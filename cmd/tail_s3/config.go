package main

import (
	"errors"

	"github.com/najeira/toml"
)

type Config struct {
	// app
	LogLevel string

	// tail
	LogLevelTail string
	File         string
	Tail         string

	// S3
	LogLevelS3        string
	Key               string
	Secret            string
	Region            string
	Bucket            string
	Path              string
	Hostname          bool
	PublicRead        bool
	ReducedRedundancy bool
	TimeFormat        string
	BufferSize        int
	FlushInterval     int64
}

func LoadConfig(file string) (*Config, error) {
	rootTree, err := toml.LoadFile(file)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	config.LogLevel = rootTree.GetString("log_level", "error")

	tailTree := rootTree.GetTree("tail")
	config.LogLevelTail = tailTree.GetString("log_level", config.LogLevel)
	config.File = tailTree.GetString("file", "")
	config.Tail = tailTree.GetString("tail", "")

	s3Tree := rootTree.GetTree("s3")
	config.LogLevelS3 = s3Tree.GetString("log_level", config.LogLevel)
	config.Key = s3Tree.GetString("key", "")
	config.Secret = s3Tree.GetString("secret", "")
	config.Region = s3Tree.GetString("region", "")
	config.Bucket = s3Tree.GetString("bucket", "")
	config.Path = s3Tree.GetString("path", "")
	config.Hostname = s3Tree.GetBool("hostname", false)
	config.PublicRead = s3Tree.GetBool("public_read", false)
	config.ReducedRedundancy = s3Tree.GetBool("reduced_redundancy", false)
	config.TimeFormat = s3Tree.GetString("time_format", DefaultTimeFormat)
	config.BufferSize = int(s3Tree.GetInt("buffer_size", DefaultBufferSize))
	config.FlushInterval = s3Tree.GetInt("flush_interval", DefaultFlushInterval)

	if config.File == "" {
		return nil, errors.New("file is not configured")
	}
	if config.Region == "" {
		return nil, errors.New("region is not configured")
	}
	if config.Bucket == "" {
		return nil, errors.New("bucket is not configured")
	}
	return config, nil
}
