package theme

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/util/gconv"
	"github.com/sz-sailing/gflib/library/slog"
	"seller-theme/app/dao"
)

func AddCustomizedPage(ctx context.Context, shopId int64, aliasName string, shopThemeId int64) error {

	if !gregex.IsMatchString(`^[\p{Han}-_a-zA-Z0-9]+$`, aliasName) {
		return errors.New("只能输入中文、数字、大小写字母及下划线")
	}
	theme, err := dao.SailShopTheme.Where("id=? AND is_del=0", shopThemeId).Fields("decoration").FindOne()
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	if theme == nil {
		return errors.New("未获取到装修配置")
	}

	var themeLists []*gjson.Json
	s := gjson.New(theme.Map()["decoration"])
	preset := s.GetArray("campaignConfig.preset")
	if preset != nil && len(preset) != 0 {
		campaignConfig := s.GetArray("campaignConfig.model")
		var campaignModel = make(map[string]interface{})
		var modelConfig = make(map[string]interface{})
		for _, campaign := range campaignConfig {
			item, _ := campaign.(map[string]interface{})
			if item["type"] != nil {
				campaignModel[gconv.String(item["type"])] = item
			}
		}
		var models = s.GetArray("campaignConfig.modelConfig.model")
		for _, model := range models {
			item, _ := model.(map[string]interface{})
			if item["type"] != nil {
				modelConfig[gconv.String(item["type"])] = model
			}
		}

		for _, v := range preset {
			if campaignModel[gconv.String(v)] != nil {
				themeLists = append(themeLists, gjson.New(campaignModel[gconv.String(v)]))
			} else if modelConfig[gconv.String(v)] != nil {
				themeLists = append(themeLists, gjson.New(modelConfig[gconv.String(v)]))
			}
		}

	}

	var configJsonStr interface{}
	if themeLists != nil {
		configJson, _ := json.Marshal(themeLists)
		configJsonStr = gconv.String(configJson)
	} else {
		configJsonStr = make([]string, 0)
	}

	shopThemePageCampaign, err := dao.SailShopThemePageCampaign.Where("shop_id=? AND shop_theme_id=? AND campaign_link=? AND is_del=0", shopId, shopThemeId, aliasName).Fields("campaign_link").FindOne()
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}

	var campaignLink = aliasName
	if shopThemePageCampaign != nil && shopThemePageCampaign.Map()["campaign_link"] != "" {
		campaignLink = campaignLink + RandomString(6)
		//拿通过随机数生成的链接再查一遍数据库，查到直接报错
		shopThemePageCampaign, err = dao.SailShopThemePageCampaign.Where("shop_id=? AND shop_theme_id=? AND campaign_link=? AND is_del=0", shopId, shopThemeId, campaignLink).Fields("campaign_link").FindOne()
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}
		if shopThemePageCampaign != nil {
			return errors.New("添加失败")
		}
	}

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx *gdb.TX) error {
		result, err := tx.Model("sail_shop_theme_page").
			Ctx(ctx).
			Insert(g.Map{
				"shop_page_id":  0,
				"shop_theme_id": shopThemeId,
				"config_name":   "campaign",
				"alias_name":    aliasName,
				"config_json":   configJsonStr,
				"shop_id":       shopId,
				"created_at":    gtime.Now(),
				"updated_at":    gtime.Now(),
			})

		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}
		shopThemePageId, err := result.LastInsertId()
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}

		result, err = tx.Model("sail_shop_theme_page_campaign").
			Ctx(ctx).
			Insert(g.Map{
				"shop_id":            shopId,
				"shop_theme_id":      shopThemeId,
				"shop_theme_page_id": shopThemePageId,
				"campaign_link":      campaignLink,
				"create_time":        gtime.Now(),
				"update_time":        gtime.Now(),
			})
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}
		return nil
	})
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	return nil
}

func DelCustomizedPage(ctx context.Context, shopId int64, shopThemePageId int64) error {
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx *gdb.TX) error {
		result, err := tx.Model("sail_shop_theme_page").
			Ctx(ctx).
			Update(map[string]int{"is_del": 1}, "id = ?", shopThemePageId)
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}
		affectedRows, err := result.RowsAffected()
		if affectedRows == 0 {
			return errors.New("删除失败")
		}
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}
		result, err = tx.Model("sail_shop_theme_page_campaign").
			Ctx(ctx).
			Update(map[string]int{"is_del": 1}, "shop_id=? AND shop_theme_page_id=?", shopId, shopThemePageId)
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}
		pageAffectedRows, err := result.RowsAffected()
		if pageAffectedRows == 0 {
			return errors.New("删除失败")
		}
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}
		return nil
	})
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	return nil
}
