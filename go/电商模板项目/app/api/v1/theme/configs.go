package theme

import (
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/i18n/gi18n"
	"github.com/gogf/gf/net/ghttp"
	"github.com/sz-sailing/gflib/library/response"
	"github.com/sz-sailing/gflib/library/slog"
	"seller-theme/app/service/theme"
)

type Configs struct{}

type SetSettingReq struct {
	ThemeConfig string `v:"required#theme_config不能为空"`
	ShopThemeId int64  `v:"required|min:1#模板ID不能为空|模板id必须大于0"`
	Font        string
	Color       string
}

// SetSetting 设置模板配置
func (c Configs) SetSetting(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *SetSettingReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "保存模板配置失败")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	//先判断模板存在不存在
	themeInfo, err := theme.GetThemeOne(ctx, shopId, data.ShopThemeId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -2, "保存模板配置失败")
	}
	if themeInfo == nil || themeInfo.Id == 0 {
		response.JsonExit(r, -3, "保存模板配置失败")
	}
	err = theme.SetSetting(ctx, themeInfo.Id, data.ThemeConfig, data.Font, data.Color)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -4, "保存模板配置失败")
	}
	response.JsonExit(r, 0, "success", []int{})
}

type UpdatePageConfigsReq struct {
	PageConfigs string `v:"required#page_configs不能为空"`
	ThemeConfig struct {
		ConfigJson string `json:"config_json"`
	}
	SellerShopThemeId int64 `v:"required|min:1#模板ID不能为空|模板id必须大于0"`
	IsPublish         string
}

// UpdatePageConfigs 保存模板页面配置
func (c Configs) UpdatePageConfigs(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *UpdatePageConfigsReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "保存模板配置失败")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	lang := shopInfo.GetString("admin_language")
	shopId := shopInfo.GetInt64("id")
	//先判断模板存在不存在
	themeInfo, err := theme.GetThemeOne(ctx, shopId, data.SellerShopThemeId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -2, "保存模板配置失败")
	}
	if themeInfo == nil || themeInfo.Id == 0 {
		response.JsonExit(r, -3, "保存模板配置失败")
	}
	err = theme.UpdatePageConfigs(ctx, shopInfo.GetInt64("id"), data.PageConfigs, data.ThemeConfig.ConfigJson, data.SellerShopThemeId, data.IsPublish, lang)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -4, "保存模板配置失败")
	}
	response.JsonExit(r, 0, "success", []int{})
}

type GetPageConfigsReq struct {
	SellerShopThemeId int64 `v:"required|min:1#模板ID不能为空|模板id必须大于0"`
}

// GetPageConfigs 获取模板的所有配置
func (c Configs) GetPageConfigs(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *GetPageConfigsReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "获取模板失败")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	ctx = gi18n.WithLanguage(ctx, shopInfo.GetString("admin_language"))
	shopId := shopInfo.GetInt64("id")
	//先判断模板存在不存在
	themeInfo, err := theme.GetThemeOne(ctx, shopId, data.SellerShopThemeId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -2, gi18n.T(ctx, "900014"))
	}
	if themeInfo == nil || themeInfo.Id == 0 {
		response.JsonExit(r, -3, gi18n.T(ctx, "900014"))
	}
	re, err := theme.GetConfigs(ctx, shopId, data.SellerShopThemeId, shopInfo)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -4, "获取模板配置失败")
	}
	response.JsonExit(r, 0, "success", re)
}

type CustomizedReq struct {
	SellerShopThemeId int64  `v:"required|min:1#模板ID不能为空|模板id必须大于0"`
	AliasName         string `v:"required#自定义页面名称不能为空"`
}

// AddCustomizedPage 添加自定义页面
func (c Configs) AddCustomizedPage(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *CustomizedReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	err := theme.AddCustomizedPage(ctx, shopId, data.AliasName, data.SellerShopThemeId)
	if err != nil {
		response.JsonExit(r, -1, err.Error())
	}
	response.JsonExit(r, 0, "创建自定义页面成功")
}

type DelCustomizedReq struct {
	ShopThemePageId int64 `v:"required|min:1#店铺装修配置ID不能为空|店铺装修配置ID不合法"`
}

// DelCustomizedPage 删除自定义页面
func (c Configs) DelCustomizedPage(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *DelCustomizedReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "参数错误")
	}
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	err := theme.DelCustomizedPage(ctx, shopId, data.ShopThemePageId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, err.Error())
	}
	response.JsonExit(r, 0, "删除自定义页面成功")
}
