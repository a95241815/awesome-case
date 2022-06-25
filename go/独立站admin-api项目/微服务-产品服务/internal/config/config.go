package config

import (
	"github.com/tal-tech/go-zero/core/stores/cache"
	"github.com/tal-tech/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	ReadDataSource  string
	WriteDataSource string
	ImgCDN          string
	Cache           cache.CacheConf
	StaticStorage   StaticStorage
	AliOss          AliOss
	ProjectENV      string
	MongoLink       string `json:"MongoLink"`
	MongoDBName     string `json:"MongoDBName"`
	ENV             string
}

type StaticStorage struct {
	AccessKey    string
	AccessSecret string
	Bucket       string
	Region       string
}

type AliOss struct {
	Endpoint        string
	AccessKeyId     string
	AccessKeySecret string
	BucketName      string
}
