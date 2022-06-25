package theme

import (
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/util/gconv"
	"github.com/sz-sailing/gflib/library/response"
	"github.com/sz-sailing/gflib/library/slog"
	"seller-theme/app/library/sencrypt"
	"seller-theme/app/service/product"
	"seller-theme/app/service/theme"
)

type Menu struct{}

// MenuList 获取菜单列表
func (m Menu) MenuList(r *ghttp.Request) {
	ctx := r.GetCtx()
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	menus, err := theme.GetMenus(ctx, shopId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "菜单列表失败")
	}
	if menus != nil {
		response.JsonExit(r, 0, "success", menus)
	}
	response.JsonExit(r, 0, "success", []string{})
}

type MenuSaveReq struct {
	Menus string
	Title string
}

// MenuSave 保存菜单
func (m Menu) MenuSave(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *MenuSaveReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "获取样式、字体失败")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	code := sencrypt.GetRandomString(12)
	err := theme.SaveNavMenu(ctx, shopId, data.Title, data.Menus, code)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "菜单保存失败")
	}
	response.JsonExit(r, 0, "success", []string{})
}

type MenuGetLevelOneReq struct {
	Code  string `v:"required#模code不能为空"`
	Child bool
}

// MenuGetLevelOne 获取一级菜单
func (m Menu) MenuGetLevelOne(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *MenuGetLevelOneReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "参数错误")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	list, err := theme.GetLevelOne(ctx, shopId, data.Code, data.Child)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -2, "获取菜单失败")
	}
	response.JsonExit(r, 0, "success", map[string]interface{}{"list": list})
}

type MenuInfoReq struct {
	Id int64 `v:"required|min:1#菜单ID不能为空|menu_id必须大于0"`
}

func (m Menu) MenuInfo(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *MenuInfoReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "参数错误")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	menu, err := theme.GetMenu(ctx, shopId, data.Id)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -2, "出现错误")
	}
	if menu == nil {
		response.JsonExit(r, -5, "没有菜单数据")
	}
	res := gjson.NewWithTag(menu, "orm")
	if menu.NavMenuJson != "" {
		var navMenus []theme.SingleMenu
		err = gconv.Structs(menu.NavMenuJson, &navMenus)
		if err != nil {
			slog.Init(ctx).Error(err)
			response.JsonExit(r, -3, "出现错误")
		}
		err = res.Set("nav_menu_json", theme.GetMenuNav(navMenus))
		if err != nil {
			slog.Init(ctx).Error(err)
			response.JsonExit(r, -4, "出现错误")
		}
		_ = res.Set("created_at", menu.CreatedAt.Format("Y-m-d H:i:s"))
		_ = res.Set("updated_at", menu.UpdatedAt.Format("Y-m-d H:i:s"))
	}
	response.JsonExit(r, 0, "success", res)
}

type MenuDeleteReq struct {
	Id int64 `v:"required|min:1#菜单ID不能为空|id必须大于0"`
}

func (m Menu) MenuDelete(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *MenuDeleteReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "参数错误")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	err := theme.DeleteMenu(ctx, shopId, data.Id)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -2, "出现错误")
	}
	response.JsonExit(r, 0, "success", []string{})
}

type MenuUpdateReq struct {
	Id    int64  `v:"required|min:1#菜单ID不能为空|menu_id必须大于0"`
	Code  string `v:"required#code不能为空"`
	Menus string
	Title string
}

func (m Menu) MenuUpdate(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *MenuUpdateReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "参数错误")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	lang := shopInfo.GetString("admin_language")
	err := theme.UpdateNavMenu(ctx, shopId, data.Title, data.Menus, data.Id, data.Code, lang)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -2, err.Error())
	}
	response.JsonExit(r, 0, "success", []string{})
}

type MenuNavCategoriesReq struct {
	Type int64
}

func (m Menu) MenuNavCategories(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *MenuNavCategoriesReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "参数错误")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	list, err := product.GetCategoryListsForMenu(ctx, shopId, data.Type)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -2, "获取菜单失败")
	}
	response.JsonExit(r, 0, "success", map[string][]theme.OriginMenu{"list": list})
}

type MenuNavProductsReq struct {
	Page     int
	PageSize int
}

func (m Menu) MenuNavProducts(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *MenuNavProductsReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "参数错误")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	if data.Page <= 0 {
		data.Page = 1
	}
	if data.PageSize <= 0 {
		data.PageSize = 20
	}
	list, total, err := product.GetProductsForMenu(ctx, shopId, data.Page, data.PageSize)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -2, "获取菜单失败")
	}
	response.JsonExit(r, 0, "success", map[string]interface{}{"list": list, "total": total})
}

func (m Menu) MenuNavItems(r *ghttp.Request) {
	ctx := r.GetCtx()
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	lang := shopInfo.GetString("admin_language")
	list, err := theme.GetNavItems(ctx, lang)
	if err != nil {
		response.JsonExit(r, -2, "获取菜单失败")
	}
	response.JsonExit(r, 0, "success", list)
}

func (m Menu) MenuNavPages(r *ghttp.Request) {
	ctx := r.GetCtx()
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	list, err := theme.GetNavPages(ctx, shopId)
	if err != nil {
		response.JsonExit(r, -2, "获取菜单失败")
	}
	response.JsonExit(r, 0, "success", map[string][]theme.OriginMenu{"list": list})
}

func (m Menu) MenuNavPolicy(r *ghttp.Request) {
	ctx := r.GetCtx()
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	list, err := theme.GetNavPolicy(ctx, shopId)
	if err != nil {
		response.JsonExit(r, -2, "获取菜单失败")
	}
	response.JsonExit(r, 0, "success", map[string][]theme.OriginMenu{"list": list})
}

func (m Menu) MenuMarketingDel(r *ghttp.Request) {
	ctx := r.GetCtx()
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	err := theme.MenuMarketingDel(ctx, shopId)
	if err != nil {
		response.JsonExit(r, -2, "删除失败")
	}
	response.JsonExit(r, 0, "success", []string{})
}

func (m Menu) MenuMarketingList(r *ghttp.Request) {
	ctx := r.GetCtx()
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	list, err := theme.MenuMarketingList(ctx, shopId)
	if err != nil {
		response.JsonExit(r, -2, "删除失败")
	}
	response.JsonExit(r, 0, "success", list)
}

type MenuMarketingSaveReq struct {
	Menu string `v:"required#menu不能为空"`
}

func (m Menu) MenuMarketingSave(r *ghttp.Request) {
	ctx := r.GetCtx()
	var data *MenuMarketingSaveReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "参数错误")
	}
	//获取店铺信息，网关层面会将店铺信息放到头部的 Uinfo 字段
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	err := theme.MenuMarketingSave(ctx, shopId, data.Menu)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -2, "保存菜单失败")
	}
	response.JsonExit(r, 0, "success", "op success")
}
