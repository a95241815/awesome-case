package theme

import (
	"context"
	"errors"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/i18n/gi18n"
	"github.com/gogf/gf/util/gconv"
	"github.com/sz-sailing/gflib/library/slog"
	"github.com/sz-sailing/gflib/library/sredis"
	"reflect"
	"seller-theme/app/dao"
	"seller-theme/app/library/sbrowser"
	"seller-theme/app/library/sencrypt"
	"seller-theme/app/model"
	"seller-theme/app/service/shop"
	"seller-theme/app/service/static_file"
)

type ThemeConfigs struct {
	PageConfigs      interface{} `json:"page_configs"`
	ThemeConfig      interface{} `json:"theme_config"`
	DecorationConfig interface{} `json:"decoration_config"`
	ThemeSns         interface{} `json:"theme_sns"`
	ThemeTag         string      `json:"theme_tag"`
	ThemeFont        interface{} `json:"theme_font"`
	ThemeColor       interface{} `json:"theme_color"`
}

func GetConfigs(ctx context.Context, shopId int64, themeId int64, uinfo *gjson.Json) (map[string]interface{}, error) {
	tempShopTheme, err := dao.SailShopTheme.FindOne("id = ? and shop_id = ?", themeId, shopId)
	if err != nil {
		slog.Init(ctx).Error(err)
		return nil, err
	}
	lang := uinfo.GetString("admin_language")
	ctx = gi18n.WithLanguage(ctx, lang)
	if tempShopTheme == nil {
		return nil, errors.New(gi18n.T(ctx, "900014"))
	}
	var shopTheme model.SailShopTheme
	err = tempShopTheme.Struct(&shopTheme)
	if err != nil {
		return nil, err
	}
	temp := make(map[string]interface{})
	tempDecoration := make(map[string]interface{})
	tempSns := make(map[string]interface{})
	tempFont := make(map[string]interface{})
	tempColor := make(map[string]interface{})

	//这个是由map组成的切片
	shopThemeSetting := gconv.MapsDeep(shopTheme.Setting)
	//for i, m := range shopThemeSetting {
	//	shopThemeSetting[i] = transLanguageData(m, lang, 1, 0)
	//}
	temp = map[string]interface{}{
		"config_name":   "theme_config",
		"shop_theme_id": themeId,
		"config_json":   gjson.New(shopThemeSetting).MustToJsonString(),
	}

	var decorationJson *gjson.Json
	if shopTheme.Decoration != "" {
		t := gconv.Map(shopTheme.Decoration)
		mc := gconv.Map(t["modelConfig"])
		mcm := gconv.MapsDeep(mc["model"])
		//for i, m := range mcm {
		//	mcm[i] = transLanguageData(m, lang, 1, 0)
		//}
		mc["model"] = mcm
		t["modelConfig"] = mc
		decorationJson = gjson.New(t)
		tempDecoration = map[string]interface{}{
			"config_name":   "decoration_config",
			"shop_theme_id": themeId,
			"config_json":   decorationJson.MustToJsonString(),
		}
	} else {
		tempDecoration = map[string]interface{}{
			"config_name":   "decoration_config",
			"shop_theme_id": themeId,
			"config_json":   []string{},
		}
	}

	tempSns = map[string]interface{}{
		"config_name":   "theme_sns",
		"shop_theme_id": themeId,
		"config_json":   shopTheme.Sns,
	}
	tempFont = map[string]interface{}{
		"config_name":   "decoration_config",
		"shop_theme_id": themeId,
		"config_json":   shopTheme.Font,
	}
	tempColor = map[string]interface{}{
		"config_name":   "decoration_config",
		"shop_theme_id": themeId,
		"config_json":   shopTheme.Color,
	}

	tempSailTheme, err := dao.SailTheme.FindOne("theme_name = ? and style = ?", shopTheme.SourceThemeName, shopTheme.ThemeStyle)
	if err != nil {
		slog.Init(ctx).Error(err)
		return nil, err
	}
	var sailTheme *model.SailTheme
	err = tempSailTheme.Struct(&sailTheme)
	if err != nil {
		slog.Init(ctx).Error(err)
		return nil, err
	}
	if sailTheme != nil {
		tempFont["default_config_json"] = sailTheme.Font
		tempColor["default_config_json"] = sailTheme.Color
	} else {
		tempFont["default_config_json"] = ""
		tempColor["default_config_json"] = ""
	}
	pageConfigs := []*model.SailShopThemePage{}
	tempInitialPageConfigs, err := dao.SailShopThemePage.
		Where("shop_id = ? and shop_theme_id = ? and config_name != 'campaign' and is_del = 0", shopId, themeId).
		FindAll()
	if err != nil {
		slog.Init(ctx).Error(err)
		return nil, err
	}
	var InitialPageConfigs []*model.SailShopThemePage
	err = tempInitialPageConfigs.Structs(&InitialPageConfigs)
	if err != nil {
		return nil, err
	}
	tempCampaignPageConfigs, err := dao.SailShopThemePage.
		Where("shop_id = ? and shop_theme_id = ? and config_name = 'campaign' and is_del = 0", shopId, themeId).
		Order("id desc").
		FindAll()
	if err != nil {
		slog.Init(ctx).Error(err)
		return nil, err
	}
	var campaignPageConfigs []*model.SailShopThemePage
	err = tempCampaignPageConfigs.Structs(&campaignPageConfigs)
	if err != nil {
		slog.Init(ctx).Error(err)
		return nil, err
	}
	pageConfigs = append(pageConfigs, InitialPageConfigs...)
	pageConfigs = append(pageConfigs, campaignPageConfigs...)
	if shopTheme.SourceThemeId == 3 {
		for _, pageConfig := range pageConfigs {
			//a := gi18n.GetContent(ctx, pageConfig.ConfigName)
			//if a != "" {
			//	pageConfigs[i].AliasName = a
			//}
			if pageConfig.ConfigName != "index" {
				continue
			}
			configJsons := gconv.MapDeep(pageConfig.ConfigJson)
			for s, i2 := range configJsons {
				ii2 := gconv.MapDeep(i2)
				//n := gi18n.GetContent(ctx, gconv.String(ii2["type"]))
				//if n != "" {
				//	ii2["name"] = n
				//}
				if ii2["type"] != "banner" {
					configJsons[s] = ii2
					continue
				}
				if _, ok := ii2["images_config"]; ok {
					if len(gconv.MapDeep(ii2["images_config"])) > 0 {
						configJsons[s] = ii2
						continue
					}
				}
				var imagesConfigTemp = map[string]interface{}{
					"same_size_first": true,
					"hight":           0,
					"width":           0,
				}
				ii2["images_config"] = imagesConfigTemp
				configJsons[s] = ii2
				break
			}
		}
	}

	var campaignList = map[int64]string{}
	if campaignPageConfigs != nil {
		var camIds []int64
		for _, config := range campaignPageConfigs {
			camIds = append(camIds, config.Id)
		}
		tempCampaignPages, err := dao.SailShopThemePageCampaign.
			Where("shop_id = ? and shop_theme_id = ? and shop_theme_page_id in (?)", shopId, themeId, camIds).
			Fields([]string{"shop_theme_page_id", "campaign_link"}).
			FindAll()
		if err != nil {
			slog.Init(ctx).Error(err)
			return nil, err
		}
		var campaignPages []*model.SailShopThemePageCampaign
		err = tempCampaignPages.Structs(&campaignPages)
		if err != nil {
			return nil, err
		}
		if tempCampaignPages != nil {
			for _, page := range campaignPages {
				campaignList[page.ShopThemePageId] = page.CampaignLink
			}
		}
	}
	pageConfigs = proPageConfigs(ctx, shopId, pageConfigs)
	var tempPageConfigs = []interface{}{}
	for _, config := range pageConfigs {
		tc := gjson.NewWithTag(config, "orm")
		var handler string
		https, _ := shop.GetInternalDomain(ctx, shopId)
		if config.ConfigName == "campaign" {
			handler = campaignList[config.Id]
			_ = tc.Set("url", https+sbrowser.Preview("campaign_buyer", themeId, handler))
		}
		_ = tc.Set("preview_url", https+sbrowser.Preview(config.ConfigName, themeId, handler))
		tempPageConfigs = append(tempPageConfigs, tc)
	}
	var re = map[string]interface{}{
		"page_configs":      tempPageConfigs,
		"theme_config":      temp,
		"decoration_config": tempDecoration,
		"theme_sns":         tempSns,
		"theme_tag":         shopTheme.SourceThemeName,
		"theme_font":        tempFont,
		"theme_color":       tempColor,
	}
	return re, nil
}

//递归处理多语言替换
func transLanguageData(data map[string]interface{}, lang string, d int, i int) map[string]interface{} {
	ctx := gi18n.WithLanguage(context.TODO(), lang)
	//如果给定的map里面有type和name字段，则直接替换
	if _, ok := data["type"]; ok {
		if _, ok := data["name"]; ok {
			trans := gi18n.GetContent(ctx, gconv.String(data["type"]))
			if trans != "" {
				data["name"] = trans
			}
		}
	}
	//如果需要递归
	if i < d {
		for s, i2 := range data {
			reflectType := reflect.TypeOf(i2)
			//第一层遍历，如果是map则直接递归
			if reflectType.Kind() == reflect.Map {
				data[s] = transLanguageData(gconv.MapDeep(i2), lang, d, i+1)
			}
			//如果是数组或者切片，则再入一层
			if reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Array {
				if i+1 < d {
					ii2 := gconv.MapsDeep(i2)
					for i3, i4 := range ii2 {
						reflectType := reflect.TypeOf(i4)
						if reflectType.Kind() == reflect.Map {
							ii2[i3] = transLanguageData(gconv.MapDeep(i4), lang, d, i+2)
						}
					}
					data[s] = ii2
				}
			}
		}
	}
	return data
}

// UpdatePageConfigs 保存配置
func UpdatePageConfigs(ctx context.Context, shopId int64, pageConfigJson string, themeConfig string, sellerShopThemeId int64, isPublish string, lang string) error {
	ctx = gi18n.WithLanguage(ctx, lang)
	tempSailShopThemePages, err := dao.SailShopThemePage.
		Where("shop_id = ? and shop_theme_id = ?", shopId, sellerShopThemeId).
		FindAll()
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	if tempSailShopThemePages == nil {
		return errors.New(gi18n.T(ctx, "900012"))
	}
	var sailShopThemePages []model.SailShopThemePage
	err = tempSailShopThemePages.Structs(&sailShopThemePages)
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	pageConfigMaps := gconv.MapsDeep(pageConfigJson)
	//开启事务
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx *gdb.TX) error {
		for _, page := range sailShopThemePages {
			for i2, configMap := range pageConfigMaps {
				configMap = gconv.Map(configMap)
				if page.Id == gconv.Int64(configMap["id"]) {
					_, err := tx.Model("sail_shop_theme_page").Ctx(ctx).
						Where("id = ?", page.Id).
						Data(map[string]string{"config_json": sencrypt.FilterScriptJson(ctx, gconv.String(configMap["config_json"]))}).
						Update()
					if err != nil {
						slog.Init(ctx).Error(err)
						return err
					}
					pageConfigMaps = append(pageConfigMaps[1:], pageConfigMaps[:i2+1]...)
					break
				}
			}
		}
		_, err := tx.Model("sail_shop_theme").Ctx(ctx).
			Where("id = ?", sellerShopThemeId).
			Data(map[string]interface{}{
				"setting": sencrypt.FilterScriptJson(ctx, themeConfig),
				"is_up":   0}).
			Update()
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}
		if isPublish != "" {
			_, err := tx.Model("sail_shop").Ctx(ctx).
				Where("id = ?", shopId).
				Data("publish_theme_id = ?", sellerShopThemeId).
				Update()
			if err != nil {
				slog.Init(ctx).Error(err)
				return err
			}
			_, err = tx.Model("sail_shop_theme").Ctx(ctx).
				Where("shop_id = ? and status = 3", shopId).
				Data("status = 2").
				Update()
			if err != nil {
				slog.Init(ctx).Error(err)
				return err
			}
			_, err = tx.Model("sail_shop_theme").Ctx(ctx).
				Where("id = ?", sellerShopThemeId).
				Data("status = 3").
				Update()
			if err != nil {
				slog.Init(ctx).Error(err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	//将截图加入队列
	screenItem := gconv.String(shopId) + "-" + gconv.String(sellerShopThemeId)
	go sredis.ClusterClient("buyer_default").LPush(context.Background(), "screen", screenItem)
	return nil
}

// SetSetting 修改模板配置
func SetSetting(ctx context.Context, themeId int64, sns string, font string, color string) error {
	data := map[string]string{
		"sns":   sencrypt.FilterHtmlJson(ctx, sns),
		"font":  sencrypt.FilterHtmlJson(ctx, font),
		"color": sencrypt.FilterHtmlJson(ctx, color),
	}
	_, err := dao.SailShopTheme.Where("id = ?", themeId).Data(data).Update()
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	return nil
}

func proPageConfigs(ctx context.Context, shopId int64, pageConfigs []*model.SailShopThemePage) []*model.SailShopThemePage {
	//ctx := gi18n.WithLanguage(context.TODO(), lang)
	for key, pageConfig := range pageConfigs {
		//pageConfig := p
		pageName := pageConfig.ConfigName
		//t := gi18n.GetContent(ctx, pageName)
		//if t != "" {
		//	pageConfig.AliasName = t
		//}
		pageJsons := gconv.Maps(pageConfig.ConfigJson)
		for s, i := range pageJsons {
			ii := gconv.Map(i)
			//t = gi18n.GetContent(ctx, gconv.String(ii["type"]))
			//if t != "" {
			//	ii["name"] = t
			//}
			pageJsons[s] = ii
		}
		switch pageName {
		case "index":
			for s, i := range pageJsons {
				ij := gconv.Map(i)
				switch ij["type"] {
				case "category":
				case "category01":
					pageJsons[s] = resetIndexCategoryConfig(ctx, shopId, ij)
					break
				case "collection":
				case "collection01":
					pageJsons[s] = resetIndexCollectionConfig(ctx, shopId, ij)
					break
				}
			}
			break
		}
		pageConfig.ConfigJson = gconv.String(pageJsons)
		pageConfigs[key] = pageConfig
	}
	return pageConfigs
}

func resetIndexCategoryConfig(ctx context.Context, shopId int64, pageConfig map[string]interface{}) map[string]interface{} {
	source := gconv.String(pageConfig["source"])
	tempSailShopCategory, err := dao.SailShopCategory.
		Fields([]string{"id", "title", "image_id"}).
		Where("id in (?) and shop_id = ? and status = 1 and is_del = 0", source, shopId).
		Order("FIELD(`id`,?) ASC", source).
		FindAll()
	if err != nil {
		slog.Init(ctx).Error(err)
		pageConfig["list"] = []map[string]string{}
		return pageConfig
	}
	if tempSailShopCategory == nil {
		pageConfig["list"] = []map[string]string{}
		return pageConfig
	}
	var imageId []int64
	var sailShopCategory []model.SailShopCategory
	err = tempSailShopCategory.Structs(&sailShopCategory)
	if err != nil {
		slog.Init(ctx).Error(err)
		return pageConfig
	}
	for _, category := range sailShopCategory {
		imageId = append(imageId, category.ImageId)
	}
	imageUrls, err := static_file.GetFilesByIds(ctx, shopId, imageId...)
	if err != nil {
		slog.Init(ctx).Error(err)
		pageConfig["list"] = []map[string]string{}
		return pageConfig
	}
	if len(imageUrls) == 0 {
		pageConfig["list"] = []map[string]string{}
		return pageConfig
	}
	var imageItem []map[string]interface{}
	for _, category := range sailShopCategory {
		var ii = map[string]interface{}{}
		ii["id"] = category.Id
		ii["title"] = category.Title
		ii["image_id"] = category.ImageId
		for i, url := range imageUrls {
			if category.ImageId == gconv.Int64(url["file_id"]) {
				ii["image_url"] = url["file_preview"]
				imageUrls = append(imageUrls[:i], imageUrls[i+1:]...)
			}
		}
		imageItem = append(imageItem, ii)
	}
	pageConfig["list"] = imageItem
	return pageConfig
}

func resetIndexCollectionConfig(ctx context.Context, shopId int64, pageConfig map[string]interface{}) map[string]interface{} {
	categoryId := gconv.Int64(pageConfig["source"])
	if categoryId == 0 {
		return pageConfig
	}
	tempSailShopCategory, err := dao.SailShopCategory.
		Fields([]string{"id", "title", "image_id"}).
		Where("id = ? and shop_id = ? and status = 1 and is_del = 0", categoryId, shopId).
		FindOne()
	if err != nil {
		slog.Init(ctx).Error(err)
		return pageConfig
	}
	if tempSailShopCategory == nil {
		return pageConfig
	}
	var sailShopCategory model.SailShopCategory
	err = tempSailShopCategory.Struct(&sailShopCategory)
	if err != nil {
		slog.Init(ctx).Error(err)
		return pageConfig
	}
	imageUrls, err := static_file.GetFilesByIds(ctx, shopId, categoryId)
	if err != nil {
		slog.Init(ctx).Error(err)
		return pageConfig
	}
	if len(imageUrls) == 0 {
		return pageConfig
	}
	pageConfig["source_title"] = sailShopCategory.Title
	pageConfig["source_image"] = imageUrls[0]["file_id"]
	return pageConfig
}

// GetCampaignList 获取店铺装修自定义页面信息
func GetCampaignList(ctx context.Context, shopId int64, themeId int64, ThemePageIds []int64) map[int64]string {
	tempCampaignList, err := dao.SailShopThemePageCampaign.
		Fields("shop_theme_page_id", "campaign_link").
		Where("shop_id = ? and shop_theme_id = ? and shop_theme_page_id in (?)", shopId, themeId, ThemePageIds).
		FindAll()
	if err != nil {
		slog.Init(ctx).Error(err)
		return nil
	}
	var campaignList []model.SailShopThemePageCampaign
	err = tempCampaignList.Structs(&campaignList)
	if err != nil {
		slog.Init(ctx).Error(err)
		return nil
	}
	var re = map[int64]string{}
	for _, campaign := range campaignList {
		re[campaign.ShopThemePageId] = campaign.CampaignLink
	}
	return re
}
