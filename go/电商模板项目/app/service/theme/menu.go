package theme

import (
	"context"
	"errors"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/i18n/gi18n"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"github.com/sz-sailing/gflib/library/slog"
	"seller-theme/app/dao"
	"seller-theme/app/library/sbrowser"
	"seller-theme/app/model"
	"seller-theme/app/service/shop"
)

// OriginMenu 菜单
type OriginMenu struct {
	Title         string      `json:"title"`
	Keyword       string      `json:"keyword"`
	Handler       string      `json:"handler"`
	Url           string      `json:"url"`
	RelUrl        string      `json:"rel_url"`
	Type          int64       `json:"type"`
	MenuId        int64       `json:"menuId"`
	OriginalTitle string      `json:"originalTitle"`
	Id            interface{} `json:"id"` //兼容老数据，这里是string类型，不改为int
	Image         string      `json:"image"`
}

// SingleMenu 单个菜单的结构
type SingleMenu struct {
	Name         string       `json:"name"`
	Children     []SingleMenu `json:"children"`
	Href         string       `json:"href"`
	Origin       OriginMenu   `json:"origin"`
	Keyword      string       `json:"keyword"`
	Handler      string       `json:"handler"`
	Uid          int64        `json:"uid"`
	FilePreviews string       `json:"file_previews"`
	Id           interface{}  `json:"id"`
}

// GetMenus 获取菜单
func GetMenus(ctx context.Context, shopId int64) ([]map[string]interface{}, error) {
	menus, err := dao.SailShopMenu.
		Fields([]string{"id", "code", "title", "nav_menu_json"}).
		Order("updated_at desc").
		Where("shop_id = ? and is_del = 0", shopId).
		FindAll()
	if err != nil {
		slog.Init(ctx).Error(err)
		return nil, err
	}
	if menus == nil {
		return nil, nil
	}
	var res []map[string]interface{}
	for _, temp := range menus {
		var menu model.SailShopMenu
		err := temp.Struct(&menu)
		if err != nil {
			slog.Init(ctx).Error(err)
			return nil, err
		}
		var item = map[string]interface{}{
			"id":            menu.Id,
			"code":          menu.Code,
			"title":         menu.Title,
			"nav_menu_json": gjson.New(menu.NavMenuJson),
		}
		res = append(res, item)
	}
	return res, nil
}

func GetMenu(ctx context.Context, shopId int64, menuId int64) (*model.SailShopMenu, error) {
	temp, err := dao.SailShopMenu.FindOne("id = ? and shop_id = ? and is_del = 0", menuId, shopId)
	if err != nil {
		slog.Init(ctx).Error(err)
		return nil, err
	}
	var menu *model.SailShopMenu
	err = temp.Struct(&menu)
	if err != nil {
		slog.Init(ctx).Error(err)
		return nil, err
	}
	return menu, nil
}

// SaveNavMenu 保存菜单
func SaveNavMenu(ctx context.Context, shopId int64, title string, menus string, code string) error {
	count, err := dao.SailShopMenu.Where("shop_id = ? and code = ? and is_del = 0", shopId, code).FindCount()
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	if count > 0 {
		return errors.New("菜单 code 已经存在，code 为：" + code)
	}
	data := map[string]string{
		"nav_menu_json": gjson.New(menus).MustToJsonString(),
		"title":         title,
		"code":          code,
		"shop_id":       gconv.String(shopId),
		"created_at":    gtime.New().Format("Y-m-d H:i:s"),
		"updated_at":    gtime.New().Format("Y-m-d H:i:s"),
	}
	_, err = dao.SailShopMenu.Insert(data)
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	return nil
}

func GetLevelOne(ctx context.Context, shopId int64, code string, child bool) ([]interface{}, error) {
	var collections = []interface{}{}
	temp, err := dao.SailShopMenu.Where("shop_id = ? and code = ? and is_del = 0", shopId, code).FindOne()
	if err != nil {
		slog.Init(ctx).Error(err)
		return nil, err
	}
	var menu *model.SailShopMenu
	err = temp.Struct(&menu)
	if err != nil {
		slog.Init(ctx).Error(err)
		return collections, err
	}
	if menu == nil {
		slog.Init(ctx).Info("没有菜单数据")
		return collections, nil
	}
	var menuMap []SingleMenu
	err = gconv.Structs(menu.NavMenuJson, &menuMap)
	if err != nil {
		slog.Init(ctx).Error(err)
		return collections, err
	}
	//menuMap = HandlerMenuJson(shopId, menuMap)
	if !child {
		for _, i := range menuMap {
			collections = append(collections, i.Name)
		}
	} else {
		for _, i := range menuMap {
			if len(i.Children) > 0 {
				collections = append(collections, i.Name)
			}
		}
	}
	return collections, nil
}

func HandlerMenuJson(ctx context.Context, shopId int64, menus []SingleMenu) []SingleMenu {
	for i, menu := range menus {
		if menu.Keyword == "product-lists" && menu.Handler != "" {
			category, err := dao.SailShopCategory.
				Where("shop_id = ? and status = 1 and is_del = 0 and handler = ?", shopId, menu.Handler).
				FindOne()
			if err != nil {
				slog.Init(ctx).Error(err)
				menus = append(menus[1:], menus[:i+1]...)
				continue
			}
			if category == nil {
				menus = append(menus[1:], menus[:i+1]...)
				continue
			}
			menus[i].Name = gconv.String(category.Map()["title"])
		}
		if menu.Keyword == "product-detail" && menu.Handler != "" {
			product, err := dao.SailShopProduct.
				Where("shop_id = ? and is_del = 0 and handler = ?", shopId, menu.Handler).
				FindOne()
			if err != nil {
				slog.Init(ctx).Error(err)
				menus = append(menus[1:], menus[:i+1]...)
				continue
			}
			if product == nil {
				menus = append(menus[1:], menus[:i+1]...)
				continue
			}
			menus[i].Name = gconv.String(product.Map()["title"])
		}
		if menu.Keyword == "page" && menu.Handler != "" {
			page, err := dao.SailShopPage.
				Where("shop_id = ? and is_del = 0 and handler = ?", shopId, menu.Handler).
				FindOne()
			if err != nil {
				slog.Init(ctx).Error(err)
				menus = append(menus[1:], menus[:i+1]...)
				continue
			}
			if page == nil {
				menus = append(menus[1:], menus[:i+1]...)
				continue
			}
			menus[i].Name = gconv.String(page.Map()["title"])
		}
		if len(menu.Children) > 0 {
			menus[i].Children = HandlerMenuJson(ctx, shopId, menu.Children)
		}
	}
	return menus
}

// GetMenuNav 菜单栏编码
func GetMenuNav(menus []SingleMenu) []SingleMenu {
	var ii int64
	for i, menu := range menus {
		menus[i].Uid = ii
		ii++
		if menu.Children != nil {
			menus[i].Children = GetMenuNav(menu.Children)
		}
	}
	return menus
}

// DeleteMenu 删除菜单
func DeleteMenu(ctx context.Context, shopId int64, menuId int64) error {
	_, err := dao.SailShopMenu.
		Where("id = ? and shop_id = ? and is_del = 0", menuId, shopId).
		Data(map[string]int64{"is_del": 1}).
		Update()
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	return nil
}

func UpdateNavMenu(ctx context.Context, shopId int64, title string, menus string, menuId int64, code string, lang string) error {
	ctx = gi18n.WithLanguage(ctx, lang)
	menu, err := dao.SailShopMenu.Where("shop_id = ? and code = ? and is_del = 0", shopId, code).FindOne()
	if err != nil {
		slog.Init(ctx).Error(err)
		return errors.New(gi18n.T(ctx, "000000"))
	}
	if menu == nil {
		return errors.New(gi18n.T(ctx, "900000"))
	}
	var mi = gconv.Int64(menu.Map()["id"])
	if mi != menuId {
		return errors.New(gi18n.T(ctx, "900003"))
	}
	_, err = dao.SailShopMenu.
		Where("id = ?", mi).
		Data(map[string]string{"nav_menu_json": menus, "title": title}).
		Update()
	if err != nil {
		slog.Init(ctx).Error(err)
		return errors.New(gi18n.T(ctx, "900004"))
	}
	return nil
}

func GetNavItems(ctx context.Context, lang string) ([]OriginMenu, error) {
	var res = []OriginMenu{}
	if lang != "zh-CN" {
		en := g.Cfg().Get("menu-nav-items-en")
		err := gconv.Structs(en, &res)
		if err != nil {
			slog.Init(ctx).Error(err)
			return nil, err
		}
	} else {
		cn := g.Cfg().Get("menu-nav-items-zh-CN")
		err := gconv.Structs(cn, &res)
		if err != nil {
			slog.Init(ctx).Error(err)
			return nil, err
		}
	}
	return res, nil
}

func GetNavPages(ctx context.Context, shopId int64) ([]OriginMenu, error) {
	var list = []OriginMenu{}
	pages, err := dao.SailShopPage.
		Where("shop_id = ? and is_del = 0 and status = 1", shopId).
		Fields([]string{"id", "handler", "title"}).
		FindAll()
	if err != nil {
		slog.Init(ctx).Error(err)
		return list, err
	}
	domain, err := shop.GetPreviewDomain(ctx, shopId)
	if err != nil {
		slog.Init(ctx).Error(err)
		return list, err
	}
	for _, temp := range pages {
		var page model.SailShopPage
		err := temp.Struct(&page)
		if err != nil {
			slog.Init(ctx).Error(err)
			return nil, err
		}
		var t OriginMenu
		t.Id = gconv.String(page.Id)
		t.Title = page.Title
		t.Url = domain + sbrowser.Buyer("page", page.Handler)
		t.Handler = page.Handler
		t.Keyword = "page"
		t.RelUrl = sbrowser.Buyer("page", page.Handler)
		list = append(list, t)
	}
	return list, nil
}

func GetNavPolicy(ctx context.Context, shopId int64) ([]OriginMenu, error) {
	var list = []OriginMenu{}
	pages, err := dao.SailShopSyspageSetting.
		Where("shop_id = ? and item = 'policy'", shopId).
		Fields([]string{"id", "handler", "title"}).
		FindAll()
	if err != nil {
		slog.Init(ctx).Error(err)
		return list, err
	}
	domain, err := shop.GetPreviewDomain(ctx, shopId)
	if err != nil {
		slog.Init(ctx).Error(err)
		return list, err
	}
	for _, temp := range pages {
		var page model.SailShopPage
		err := temp.Struct(&page)
		if err != nil {
			slog.Init(ctx).Error(err)
			return nil, err
		}
		var t OriginMenu
		t.Id = gconv.String(page.Id)
		t.Title = page.Title
		t.Url = domain + sbrowser.Buyer("policy", page.Handler)
		t.Handler = page.Handler
		t.Keyword = "policy"
		t.RelUrl = sbrowser.Buyer("policy", page.Handler)
		list = append(list, t)
	}
	return list, nil
}

func MenuMarketingDel(ctx context.Context, shopId int64) error {
	_, err := dao.SailShopMenuMarketing.Where("shop_id = ?", shopId).Delete()
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	return nil
}

func MenuMarketingList(ctx context.Context, shopId int64) ([]*gjson.Json, error) {
	var (
		menu1 = []*gjson.Json{}
		menu2 = map[int][]*gjson.Json{}
		menu3 = map[int][]*gjson.Json{}
	)
	menus, err := dao.SailShopMenuMarketing.
		Where("shop_id = ?", shopId).
		Order("level desc, sort asc").
		FindAll()
	if err != nil {
		slog.Init(ctx).Error(err)
		return menu1, err
	}
	if menus == nil {
		return menu1, nil
	}
	for _, temp := range menus {
		var menu model.SailShopMenuMarketing
		err := temp.Struct(&menu)
		if err != nil {
			slog.Init(ctx).Error(err)
			return nil, err
		}
		temp := gjson.NewWithTag(menu, "orm")
		var tempOrigin OriginMenu
		err = gconv.Struct(menu.Origin, &tempOrigin)
		if err != nil {
			slog.Init(ctx).Error(err)
			return nil, err
		}
		_ = temp.Set("origin", tempOrigin)
		_ = temp.Set("created_at", menu.CreatedAt.Format("Y-m-d H:i:s"))
		_ = temp.Set("updated_at", menu.UpdatedAt.Format("Y-m-d H:i:s"))
		if menu.Level == 3 {
			menu3[menu.Pcode] = append(menu3[menu.Pcode], temp)
		}
		if menu.Level == 2 {
			if children, ok := menu3[menu.Code]; ok {
				_ = temp.Set("children", children)
			}
			menu2[menu.Pcode] = append(menu2[menu.Pcode], temp)
		}
		if menu.Level == 1 {
			if children, ok := menu2[menu.Code]; ok {
				_ = temp.Set("children", children)
			}
			menu1 = append(menu1, temp)
		}
	}
	return menu1, nil
}

func MenuMarketingSave(ctx context.Context, shopId int64, menuData string) error {
	var menus []SingleMenu
	err := gconv.Structs(menuData, &menus)
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	var (
		itemUpdate = map[interface{}]map[string]interface{}{}
		itemInsert []map[string]interface{}
		itemExist  []interface{}
	)
	marketingFormatData(shopId, menus, 1, 0, itemUpdate, &itemInsert, &itemExist)
	//开启事务
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx *gdb.TX) error {
		if len(itemExist) > 0 {
			_, err := tx.Model("sail_shop_menu_marketing").Ctx(ctx).Delete("id not in (?)", itemExist)
			if err != nil {
				slog.Init(ctx).Error(err)
				return err
			}
		}
		if len(itemInsert) > 0 {
			_, err = tx.Model("sail_shop_menu_marketing").Ctx(ctx).Insert(itemInsert)
			if err != nil {
				slog.Init(ctx).Error(err)
				return err
			}
		}
		if len(itemUpdate) > 0 {
			for id, m := range itemUpdate {
				_, err = tx.Model("sail_shop_menu_marketing").Ctx(ctx).Where("id = ?", id).Update(m)
				if err != nil {
					slog.Init(ctx).Error(err)
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	return nil
}

func marketingFormatData(shopId int64, menus []SingleMenu, level int64, pcode int64,
	itemUpdate map[interface{}]map[string]interface{}, itemInsert *[]map[string]interface{}, itemExist *[]interface{}) {
	var i int64 = 1
	for k, menu := range menus {
		var item = map[string]interface{}{
			"shop_id":     shopId,
			"code":        i,
			"name":        menu.Name,
			"keyword":     menu.Keyword,
			"handler":     menu.Handler,
			"href":        menu.Href,
			"pcode":       pcode,
			"sort":        k + 1,
			"level":       level,
			"origin":      gjson.New(menu.Origin).MustToJsonString(),
			"category_id": menu.Origin.Id,
		}
		if gconv.Int(menu.Id) > 0 {
			itemUpdate[menu.Id] = item
			*itemExist = append(*itemExist, menu.Id)
		} else {
			*itemInsert = append(*itemInsert, item)
		}
		if menu.Children != nil {
			l := level + 1
			marketingFormatData(shopId, menu.Children, l, i, itemUpdate, itemInsert, itemExist)
		}
		i++
	}
}
