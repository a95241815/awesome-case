package svc

import (
	"github.com/tal-tech/go-zero/core/stores/redis"
	"github.com/tal-tech/go-zero/core/stores/sqlx"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/config"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/model"
	"os"
)

type ServiceContext struct {
	Config                    config.Config
	ReadModel                 model.SailShopProductModel
	WriteModel                model.SailShopProductModel
	ReadImageModel            model.SailUploadModel
	WriteImageModel           model.SailUploadModel
	ReadVariantModel          model.SailShopProductVariantModel
	WriteVariantModel         model.SailShopProductVariantModel
	ReadCategoryModel         model.SailShopCategoryModel
	WriteCategoryModel        model.SailShopCategoryModel
	ReadCategoryProductModel  model.SailShopCategoryProductModel
	WriteCategoryProductModel model.SailShopCategoryProductModel
	ReadProductDetailModel    model.SailShopProductDetailModel
	WriteProductDetailModel   model.SailShopProductDetailModel
	ReadProductCommentsModel  model.SailProductCommentsModel
	WriteProductCommentsModel model.SailProductCommentsModel
	ImgCDN                    string
	StaticStorage             config.StaticStorage
	AliOss                    config.AliOss
	ProjectENV                string
	MongoDBName               string
	RedisClient               *redis.Redis
	RedisClientSaas           *redis.Redis
	ENV                       string
}

func NewServiceContext(c config.Config) *ServiceContext {
	redisClientPlatform := c.Cache[0].NewRedis()
	saasPass := c.Cache[1].Pass
	if os.Getenv("ENV") == "FAT" || os.Getenv("ENV") == "FAT" {
		saasPass = "Lp8^w#H$r43@"
	}
	redisClientSaas := redis.NewRedis(c.Cache[1].Host, c.Cache[1].Type, saasPass)
	return &ServiceContext{
		Config:                    c,
		ReadModel:                 model.NewSailShopProductModel(sqlx.NewMysql(c.ReadDataSource), c.Cache),
		WriteModel:                model.NewSailShopProductModel(sqlx.NewMysql(c.WriteDataSource), c.Cache),
		ReadImageModel:            model.NewSailUploadModel(sqlx.NewMysql(c.ReadDataSource), c.Cache),
		WriteImageModel:           model.NewSailUploadModel(sqlx.NewMysql(c.WriteDataSource), c.Cache),
		ReadVariantModel:          model.NewSailShopProductVariantModel(sqlx.NewMysql(c.ReadDataSource), c.Cache),
		WriteVariantModel:         model.NewSailShopProductVariantModel(sqlx.NewMysql(c.WriteDataSource), c.Cache),
		ReadCategoryModel:         model.NewSailShopCategoryModel(sqlx.NewMysql(c.ReadDataSource), c.Cache),
		WriteCategoryModel:        model.NewSailShopCategoryModel(sqlx.NewMysql(c.WriteDataSource), c.Cache),
		ReadCategoryProductModel:  model.NewSailShopCategoryProductModel(sqlx.NewMysql(c.ReadDataSource), c.Cache),
		WriteCategoryProductModel: model.NewSailShopCategoryProductModel(sqlx.NewMysql(c.WriteDataSource), c.Cache),
		ReadProductDetailModel:    model.NewSailShopProductDetailModel(sqlx.NewMysql(c.ReadDataSource), c.Cache),
		WriteProductDetailModel:   model.NewSailShopProductDetailModel(sqlx.NewMysql(c.WriteDataSource), c.Cache),
		ReadProductCommentsModel:  model.NewSailProductCommentsModel(sqlx.NewMysql(c.ReadDataSource), c.Cache),
		WriteProductCommentsModel: model.NewSailProductCommentsModel(sqlx.NewMysql(c.WriteDataSource), c.Cache),
		ImgCDN:                    c.ImgCDN,
		StaticStorage:             c.StaticStorage,
		AliOss:                    c.AliOss,
		ProjectENV:                c.ProjectENV,
		MongoDBName:               c.MongoDBName,
		RedisClient:               redisClientPlatform,
		RedisClientSaas:           redisClientSaas,
		ENV:                       c.ENV,
	}
}
