package main

import (
	"flag"
	"fmt"
	"github.com/gogf/gf/os/genv"
	"google.golang.org/grpc/reflection"
	"os"

	"gitlab.jhongnet.com/mall/rpc-product-server/internal/config"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/server"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/svc"
	"gitlab.jhongnet.com/mall/rpc-product-server/product"

	"github.com/tal-tech/go-zero/core/conf"
	"github.com/tal-tech/go-zero/zrpc"
	"google.golang.org/grpc"

	"github.com/sz-sailing/gflib/library/sapollo"
)

var configFile = flag.String("f", "", "the config file")

func main() {
	flag.Parse()
	var ac = sapollo.Config{
		Appid:      "open_one", //开放平台所有的 Appid 统一为：open_one
		Cluster:    genv.Get("PLATFORM_NAME"),
		Namespaces: []string{"rpc_product_server.yaml"}, //只获取自己项目的命名空间
		Address:    genv.Get("APOLLO_ADDRESS"),
	}
	sapollo.Start(ac)

	if *configFile == "" {
		env := os.Getenv("ENV")
		switch env {
		case "LOCAL":
			*configFile = "etc/config-local.yaml"
		case "DEV":
			*configFile = "etc/config-dev.yaml"
		case "FAT":
			*configFile = "etc/config-fat.yaml"
		case "UAT":
			*configFile = "etc/config-uat.yaml"
		case "PRO":
			*configFile = "etc/config-pro.yaml"
		default:
			panic("need config file")
		}
	}

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	srv := server.NewProductRPCServer(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		product.RegisterProductRPCServer(grpcServer, srv)
		if os.Getenv("ENV") != "PRO" {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
