package theme

import (
	"context"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/i18n/gi18n"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/util/gconv"
	"github.com/sz-sailing/gflib/library/response"
	"github.com/sz-sailing/gflib/library/slog"
	"seller-theme/app/library/sbrowser"
	"seller-theme/app/model"
	"seller-theme/app/service/shop"
	"seller-theme/app/service/static_file"
	"seller-theme/app/service/store"
	"seller-theme/app/service/theme"
)

type SellerShop struct{}

// GetSellerThemes 获取卖家所有模板
func (s SellerShop) GetSellerThemes(r *ghttp.Request) {
	ctx := r.GetCtx()
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	//获取店铺后台的语言
	ctx = gi18n.WithLanguage(ctx, shopInfo.GetString("admin_language"))
	shopId := shopInfo.GetInt64("id")
	themes, err := theme.GetThemeList(ctx, shopId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -2, "获取模板列表失败")
	}
	if len(themes) == 0 {
		response.JsonExit(r, -3, gi18n.T(ctx, "900014"))
	}
	publishThemeId := shopInfo.GetInt64("publish_theme_id")
	var imageIds = make([]int64, 0)
	for _, shopTheme := range themes {
		if shopTheme.PcImageId > 0 {
			imageIds = append(imageIds, gconv.Int64(shopTheme.PcImageId))
		}
		if shopTheme.MobileImageId > 0 {
			imageIds = append(imageIds, gconv.Int64(shopTheme.MobileImageId))
		}
	}

	images, err := static_file.GetPreviewUrlByFileIds(ctx, shopId, imageIds...)
	//要返回的数据结构
	type ToRet struct {
		model.SailShopTheme
		PcImageUrl     string `orm:"pc_image_url" json:"PcImageUrl"`
		MobileImageUrl string `orm:"mobile_image_url" json:"MobileImageUrl"`
		PreviewUrl     string `orm:"preview_url" json:"PreviewUrl"`
	}
	var themeLists = []*gjson.Json{}
	var themeUse *gjson.Json
	var upTheme *gjson.Json
	for _, shopTheme := range themes {
		var themeOne ToRet
		themeOne.SailShopTheme = *shopTheme
		if images[gconv.Int64(shopTheme.PcImageId)] != "" {
			themeOne.PcImageUrl = images[gconv.Int64(shopTheme.PcImageId)]
		}
		if images[gconv.Int64(shopTheme.MobileImageId)] != "" {
			themeOne.MobileImageUrl = images[gconv.Int64(shopTheme.MobileImageId)]
		}
		//themeOne.ThemeName = gi18n.T(ctx, shopTheme.ThemeName)
		themeOne.ThemeName = shopTheme.ThemeName
		if shopTheme.Status == 3 {
			previewDomain, err := shop.GetPreviewDomain(ctx, shopId)
			if err != nil {
				slog.Init(ctx).Error(err)
				previewDomain = ""
			}
			themeOne.PreviewUrl = previewDomain + sbrowser.Buyer("index", "")
			themeUse = gjson.NewWithTag(themeOne, "orm")
			_ = themeUse.Set("created_at", themeOne.CreatedAt.Format("Y-m-d H:i:s"))
			_ = themeUse.Set("updated_at", themeOne.UpdatedAt.Format("Y-m-d H:i:s"))
		} else {
			previewDomain, err := shop.GetInternalDomain(ctx, shopId)
			if err != nil {
				slog.Init(ctx).Error(err)
				previewDomain = ""
			}
			themeOne.PreviewUrl = previewDomain + sbrowser.Preview("index", shopTheme.Id, "")
			themeOneJ := gjson.NewWithTag(themeOne, "orm")
			_ = themeOneJ.Set("created_at", themeOne.CreatedAt.Format("Y-m-d H:i:s"))
			_ = themeOneJ.Set("updated_at", themeOne.UpdatedAt.Format("Y-m-d H:i:s"))
			themeLists = append(themeLists, themeOneJ)
		}
		if shopTheme.Id == publishThemeId {
			upTheme = gjson.NewWithTag(themeOne, "orm")
			_ = upTheme.Set("created_at", themeOne.CreatedAt.Format("Y-m-d H:i:s"))
			_ = upTheme.Set("updated_at", themeOne.UpdatedAt.Format("Y-m-d H:i:s"))
		}
	}
	if themeUse == nil {
		themeUse = upTheme
	}
	var re = map[string]interface{}{
		"list": themeLists,
		"used": themeUse,
	}
	response.JsonExit(r, 0, "success", re)
}

type GetThemeFromStoreReq struct {
	ShopThemeId int64 `v:"required#模板ID不能为空"`
}

// GetThemeFromStore 从模板商城获取一套模板
func (s SellerShop) GetThemeFromStore(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *GetThemeFromStoreReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "获取模板失败")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	//先判断模板存在不存在
	themeInfo, err := store.GetThemeOne(ctx, data.ShopThemeId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -2, "获取模板失败")
	}
	if themeInfo == nil || themeInfo.Id == 0 {
		response.JsonExit(r, -3, "获取模板失败")
	}
	newThemeId, err := store.GetThemeFromStore(ctx, shopId, data.ShopThemeId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -4, gerror.Current(err).Error())
	}
	response.JsonExit(r, 0, "success", map[string]int64{"theme_id": newThemeId})
}

// QrCode 店铺的二维码
func (s SellerShop) QrCode(r *ghttp.Request) {
	ctx := r.GetCtx()
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	previewDomain, err := shop.GetPreviewDomain(ctx, shopId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "获取二维码失败")
	}
	previewUrl := previewDomain + sbrowser.Buyer("index", "")
	response.JsonExit(r, 0, "success", map[string]string{"qr_code": previewUrl})
}

type UpThemeReq struct {
	ThemeId int64 `v:"required#模板ID不能为空"`
}

// UpTheme 发布模板
func (s SellerShop) UpTheme(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *UpThemeReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "发布模板失败")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	//先判断模板存在不存在
	themeInfo, err := theme.GetThemeOne(ctx, shopId, data.ThemeId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -3, "发布模板失败")
	}
	if themeInfo == nil || themeInfo.Id == 0 {
		response.JsonExit(r, -2, "发布模板失败")
	}
	err = theme.UpTheme(ctx, shopId, data.ThemeId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -4, "重命名失败")
	}
	response.JsonExit(r, 0, "success", make([]string, 0))
}

type RenameThemeReq struct {
	ThemeId   int64  `v:"required#模板ID不能为空"`
	ThemeName string `v:"required#模板名称不能为空"`
}

// RenameTheme 模板重命名
func (s SellerShop) RenameTheme(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *RenameThemeReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "重命名失败")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	//先判断模板存在不存在
	themeInfo, err := theme.GetThemeOne(ctx, shopId, data.ThemeId)
	if themeInfo == nil || themeInfo.Id == 0 {
		response.JsonExit(r, -2, "重命名失败")
	}
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -3, "重命名失败")
	}
	err = theme.RenameTheme(ctx, shopId, data.ThemeId, data.ThemeName)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -4, "重命名失败")
	}
	response.JsonExit(r, 0, "success", make([]string, 0))
}

type DelThemeReq struct {
	ThemeId int64 `v:"required#模板ID不能为空"`
}

// DelTheme 删除模板
func (s SellerShop) DelTheme(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *DelThemeReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "删除失败")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	//先判断模板存在不存在
	themeInfo, err := theme.GetThemeOne(ctx, shopId, data.ThemeId)
	if themeInfo == nil || themeInfo.Id == 0 {
		response.JsonExit(r, -2, "删除失败")
	}
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -3, "删除失败")
	}
	err = theme.DelTheme(ctx, shopId, data.ThemeId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -4, "删除失败")
	}
	response.JsonExit(r, 0, "success", make([]string, 0))
}

type CopyReq struct {
	ShopThemeId int64 `v:"required#模板ID不能为空"`
}

// CopyTheme 对现有模板创建副本
func (s SellerShop) CopyTheme(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *CopyReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "创建副本失败")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	//先判断模板存在不存在
	themeInfo, err := theme.GetThemeOne(ctx, shopId, data.ShopThemeId)
	if themeInfo == nil || themeInfo.Id == 0 {
		response.JsonExit(r, -2, "创建副本失败")
	}

	//复制模板在mysql的配置库信息
	targetThemeId, err := theme.CopySetting(ctx, shopId, data.ShopThemeId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -3, "创建副本失败")
	}
	//如果是公有模板，直接返回复制成功
	if themeInfo.IsOwned == 0 {
		response.JsonExit(r, 0, "success", make([]string, 0))
	}
	//复制CSS文件
	err = theme.CopyCss(ctx, shopId, data.ShopThemeId, targetThemeId)
	if err != nil {
		slog.Init(ctx).Error(err)
		//要删除 targetThemeId
		defer func(ctx context.Context, shopId int64, themeId int64) {
			err := theme.DelTheme(ctx, shopId, themeId)
			if err != nil {
				slog.Init(ctx).Error(err)
			}
		}(ctx, shopId, targetThemeId)
		response.JsonExit(r, -4, "创建副本失败")
	}
	//复制其他文件
	err = theme.CopyFileFromMongo(ctx, shopId, data.ShopThemeId, targetThemeId)
	if err != nil {
		slog.Init(ctx).Error(err)
		//要删除 targetThemeId
		defer func(ctx context.Context, shopId int64, themeId int64) {
			err := theme.DelTheme(ctx, shopId, themeId)
			if err != nil {
				slog.Init(ctx).Error(err)
			}
		}(ctx, shopId, targetThemeId)
		response.JsonExit(r, -5, "创建副本失败")
	}
	response.JsonExit(r, 0, "success", make([]string, 0))
}
