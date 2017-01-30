package main

import (
	"errors"

	"github.com/najeira/conv"
	"github.com/pelletier/go-toml"
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
	config.LogLevel = conv.String(rootTree.Get("log_level"), "error")

	tailTree := rootTree.Get("tail").(*toml.TomlTree)
	config.LogLevelTail = conv.String(tailTree.Get("log_level"), config.LogLevel)
	config.File = conv.String(tailTree.Get("file"), "")
	config.Tail = conv.String(tailTree.Get("tail"), "")

	s3Tree := rootTree.Get("s3").(*toml.TomlTree)
	config.LogLevelS3 = conv.String(s3Tree.Get("log_level"), config.LogLevel)
	config.Key = conv.String(s3Tree.Get("key"), "")
	config.Secret = conv.String(s3Tree.Get("secret"), "")
	config.Region = conv.String(s3Tree.Get("region"), "")
	config.Bucket = conv.String(s3Tree.Get("bucket"), "")
	config.Path = conv.String(s3Tree.Get("path"), "")
	config.Hostname = conv.Bool(s3Tree.Get("hostname"), false)
	config.PublicRead = conv.Bool(s3Tree.Get("public_read"), false)
	config.ReducedRedundancy = conv.Bool(s3Tree.Get("reduced_redundancy"), false)
	config.TimeFormat = conv.String(s3Tree.Get("time_format"), DefaultTimeFormat)
	config.BufferSize = int(conv.Int(s3Tree.Get("buffer_size"), DefaultBufferSize))
	config.FlushInterval = conv.Int(s3Tree.Get("flush_interval"), DefaultFlushInterval)

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
