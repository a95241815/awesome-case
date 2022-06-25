package theme

import (
	"context"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"github.com/sz-sailing/gflib/library/response"
	"github.com/sz-sailing/gflib/library/slog"
	"github.com/sz-sailing/gflib/library/sredis"
	"net/url"
	"seller-theme/app/service/shop"
	"seller-theme/app/service/theme"
)

type UpDown struct{}

type ZipResult struct {
	Key     string `v:"required#key不能为空"`
	ThemeId int64  `v:"min:1#模板id必须大于0"`
}

//Result 前端上传模板文件到 S3 后，请求该接口获取处理结果
func (ud UpDown) Result(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *ZipResult
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -4, "参数错误")
	}

	env := gstr.ToUpper(genv.Get("ENV"))
	domain := r.GetHeader("Origin")
	if domain == "" {
		domain = r.GetHeader("origin")
	}
	if domain == "" {
		response.JsonExit(r, -4, "找不到店铺")
	}
	shopUrl, _ := url.Parse(domain)
	shopInfo, err := shop.GetShopByDomain(ctx, shopUrl.Host)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -4, "找不到店铺")
	}
	var ssid int64
	if shopInfo != nil {
		ssid = shopInfo.Id
	} else {
		if env != "DEV" {
			response.JsonExit(r, -4, "找不到店铺")
		} else {
			sid := r.GetHeader("Shopid")
			if sid == "" {
				sid = r.GetHeader("shopid")
			}
			if sid == "" {
				response.JsonExit(r, -4, "找不到店铺")
			}
			ssid = gconv.Int64(sid)
		}
	}
	themeId := data.ThemeId
	//如果传了模板id，说明是更新模板
	if themeId != 0 {
		_, err = theme.IsExclusive(data.ThemeId)
		if err != nil {
			response.JsonExit(r, -4, err.Error())
		}
	}

	redisClient := sredis.ClusterClient()
	var cacheKey string
	if themeId != 0 {
		cacheKey = g.Cfg().GetString("cache_key.theme_zip") + gconv.String(ssid) + ":" + gconv.String(themeId) + ":" + data.Key
	} else {
		cacheKey = g.Cfg().GetString("cache_key.theme_zip") + gconv.String(ssid) + ":" + data.Key
	}

	//查看处理情况
	dealStatus, err := redisClient.HGetAll(context.TODO(), cacheKey).Result()
	if err != nil {
		slog.Init(ctx).Redis().Error(err)
		response.JsonExit(r, -4, "文件上传失败")
	}
	//如果还没有开始，则将任务丢到集合里面去等待处理
	if len(dealStatus) == 0 {
		redisClient.SAdd(context.TODO(), g.Cfg().GetString("cache_key.theme_zip_set")+genv.Get("ENV"), gconv.String(map[string]string{"shop_id": gconv.String(ssid), "theme_id": gconv.String(themeId), "key": data.Key}))
		response.JsonExit(r, -2, "正在处理中。。。")
	}
	//已经处理成功
	if dealStatus["success"] != "" {
		response.JsonExit(r, 0, "处理成功", map[string]interface{}{"theme_id": dealStatus["theme_id"]})
	}
	//处理失败
	if dealStatus["fail"] != "" {
		response.JsonExit(r, -4, dealStatus["fail"])
	}
	response.JsonExit(r, -2, "处理中。。。")
}
