package theme

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"github.com/sz-sailing/gflib/library/slog"
	"github.com/sz-sailing/gflib/library/sredis"
	"seller-theme/app/dao"
	"seller-theme/app/model"
	"time"
)

type ThemeOne struct {
	Id              int64       `orm:"id,primary"        json:"id"`                //
	SourceThemeId   int64       `orm:"source_theme_id"   json:"source_theme_id"`   // 来源商城模板ID
	SourceThemeName string      `orm:"source_theme_name" json:"source_theme_name"` // 模板标识
	ThemeName       string      `orm:"theme_name"        json:"theme_name"`        // 商店模板名称
	ThemeStyle      string      `orm:"theme_style"       json:"theme_style"`       // 商店模版风格
	ShopId          int64       `orm:"shop_id"           json:"shop_id"`           // 商店唯一ID
	IsDel           int         `orm:"is_del"            json:"is_del"`            // 0表示禁用 1表示启用
	IsOwned         int         `orm:"is_owned"          json:"is_owned"`          // 是否私有的模板，用户上传的模板、用户新建的模板、用户基于模板商店的模板修改过的模板，都是私有模板，只有从模板商店获取，没有修改过模板文件的，才是公开模板。私有模板（1），公开模板（0）
	Status          int         `orm:"status"            json:"status"`            // 1表示禁用 2表示正常 3表示发布中
	UpdatedAt       *gtime.Time `orm:"updated_at"        json:"updated_at"`        // 更新时间
}

// GetThemeOne 获取指定的模板信息
func GetThemeOne(ctx context.Context, shopId int64, themeId int64) (*model.SailShopTheme, error) {
	var re *model.SailShopTheme
	key := g.Cfg().GetString("cache_key.theme_info") + gconv.String(shopId) + ":" + gconv.String(themeId)
	cache, err := sredis.ClusterClient().Get(context.Background(), key).Result()
	if err != nil && err != redis.Nil {
		g.Log().Async(true).Cat("redis").Error(err)
	}
	if err == redis.Nil || cache == "" {
		one, err := g.DB("default").Table("sail_shop_theme").FindOne("id=? AND shop_id=? AND is_del=0", themeId, shopId)
		if err = one.Struct(&re); err != nil {
			g.Log().Async(true).Cat("redis").Error(err)
			return nil, err
		}
		_, err = sredis.ClusterClient().SetEX(context.Background(), key, gconv.String(one), 60*time.Second).Result()
		if err != nil {
			g.Log().Async(true).Cat("redis").Error(err)
		}
		return re, nil
	}
	_ = gconv.Struct(cache, &re)
	return re, nil
}

// GetThemeList 获取店铺安装的所有模板
func GetThemeList(ctx context.Context, shopId int64) ([]*model.SailShopTheme, error) {
	temp, err := dao.SailShopTheme.Where("shop_id = ? and is_del = 0", shopId).Order("updated_at desc").FindAll()
	if err != nil {
		slog.Init(ctx).Error(err)
		return nil, err
	}
	var sailShopTheme []*model.SailShopTheme
	err = temp.Structs(&sailShopTheme)
	if err != nil {
		slog.Init(ctx).Error(err)
		return nil, err
	}
	return sailShopTheme, err
}

// DelTheme 删除一套模板，硬删除
func DelTheme(ctx context.Context, shopId int64, themeId int64) error {
	//开启事务
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx *gdb.TX) error {
		_, err := tx.Model("sail_shop_theme").Ctx(ctx).Delete("shop_id = ? and id = ?", shopId, themeId)
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}
		_, err = tx.Model("sail_shop_theme_page").Ctx(ctx).Delete("shop_id = ? and shop_theme_id = ?", shopId, themeId)
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}
		_, err = tx.Model("sail_shop_theme_page_campaign").Ctx(ctx).Delete("shop_id = ? and shop_theme_id = ?", shopId, themeId)
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// RenameTheme 模板重命名
func RenameTheme(ctx context.Context, shopId int64, themeId int64, themeName string) error {
	data := map[string]string{
		"theme_name": themeName,
	}
	_, err := dao.SailShopTheme.Where("id = ? and shop_id = ?", themeId, shopId).Update(data)
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	return nil
}

// UpTheme 发布模板
func UpTheme(ctx context.Context, shopId int64, themeId int64) error {
	//开启事务
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx *gdb.TX) error {
		_, err := tx.Model("sail_shop_theme").Ctx(ctx).Update(map[string]int{"status": 2}, "shop_id = ? and status = 3", shopId)
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}
		_, err = tx.Model("sail_shop").Ctx(ctx).Update(map[string]int64{"publish_theme_id": themeId}, "id = ? ", shopId)
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}
		_, err = tx.Model("sail_shop_theme").Ctx(ctx).Update(map[string]int{"status": 3}, "id = ? and shop_id = ?", themeId, shopId)
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// IsExclusive 根据模板id判断是否专属模板
func IsExclusive(themeId int64) (bool, error) {
	isExclusive, err := dao.SailShopTheme.FindValue("is_owned", "id = ?", themeId)
	if err != nil {
		return false, err
	}
	if isExclusive == nil {
		return false, errors.New("模板不存在")
	}
	if isExclusive.Val() == 0 {
		return false, errors.New("不是专属模板")
	}
	return true, nil
}

func GetThemeApply(id int64, partnerId uint64) (*gjson.Json, error) {
	temp, err := dao.SailThemeApply.FindOne("id = ? and partner_id = ?", id, partnerId)
	if err != nil {
		return nil, err
	}
	if temp == nil {
		return nil, nil
	}
	ThemeApplyInfo := gjson.NewWithTag(temp, "orm")
	return ThemeApplyInfo, nil
}
