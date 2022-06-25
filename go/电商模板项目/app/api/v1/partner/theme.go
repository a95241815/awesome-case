package partner

import (
	"fmt"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/util/gconv"
	"github.com/sz-sailing/gflib/library/response"
	"github.com/sz-sailing/gflib/library/slog"
	"github.com/sz-sailing/gflib/library/sredis"
	"go/types"
	"net/url"
	string2 "seller-theme/app/library/string"
	"seller-theme/app/library/structs"
	"seller-theme/app/library/surl"
	"seller-theme/app/service/category"
	"seller-theme/app/service/partner"
	"seller-theme/app/service/shop"
	"seller-theme/app/service/store"
	"seller-theme/app/service/theme"
)

type Theme struct{}

const versionPattern = `^(([0-9]|([1-9]([0-9]*))).){2}([0-9]|([1-9]([0-9]*)))([-](([0-9A-Za-z]|([1-9A-Za-z]([0-9A-Za-z]*)))[.]){0,}([0-9A-Za-z]|([1-9A-Za-z]([0-9A-Za-z]*)))){0,1}([+](([0-9A-Za-z]{1,})[.]){0,}([0-9A-Za-z]{1,})){0,1}$`

type GetPartnerThemeReq struct {
	Id int64 `v:"required|min:1#id必传|id必须大于0"`
}

// GetPartnerTheme 获取开发者模板详情
func (t Theme) GetPartnerThemeDetail(r *ghttp.Request) {
	ctx := slog.GetCtxVar(r)
	partnerId := partner.GetLoginID(r)
	if partnerId == 0 {
		response.JsonExit(r, -1, "没有登录", types.Struct{})
	}
	var data *GetPartnerThemeReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	apply, err := theme.GetThemeApply(data.Id, partnerId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -2, "获取模板失败")
	}
	if apply == nil {
		response.JsonExit(r, 0, "success", types.Struct{})
	}
	response.JsonExit(r, 0, "success", apply)
}

//PublishTheme 发布模板
func (t Theme) PublishTheme(r *ghttp.Request) {
	ctx := slog.GetCtxVar(r)
	partnerInfo := partner.GetClientInfo(r)
	if partnerInfo.Id == 0 {
		response.JsonExit(r, -1, "没有登录", types.Struct{})
	}
	var data *partner.PublishThemeReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	if !gregex.IsMatchString(`^[a-zA-Z]+$`, data.ThemeName) {
		response.JsonExit(r, -1, "标识只能输入英文", types.Struct{})
	}

	if !gregex.IsMatchString(versionPattern, data.Version) {
		response.JsonExit(r, -1, "版本号非法", types.Struct{})
	}
	//检查模板标识是否重复
	Result, err := partner.CheckThemeNameExists(ctx, data.ThemeName, data.Style, partnerInfo.Id)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error(), types.Struct{})
	}
	if Result {
		response.JsonExit(r, -1, "标识重复，请更换", types.Struct{})
	}
	shopName := data.ThemeName + "-" + data.Style
	password := string2.RandomString(6)
	re, err := shop.RegisterSeller(ctx, data, shopName, partnerInfo.Email, password, false)
	if err != nil {
		response.JsonExit(r, -1, gerror.Current(err).Error(), types.Struct{})
	}
	err = partner.AddTheme(ctx, data, partnerInfo.Id, shopName, partnerInfo.Email, password, re.TargetThemeId, re.ShopId, 0, string2.DistinctStringArray(data.CategoryIDs))
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "发布失败", types.Struct{})
	}
	var (
		admin = fmt.Sprintf("https://%s/admin/login", re.Domain)
		font  = fmt.Sprintf("https://%s", re.Domain)
	)

	//复制CSS文件
	err = partner.CopyCssToPartner(ctx, re)
	if err != nil {
		response.JsonExit(r, -4, gerror.Current(err).Error())
	}
	//复制其它文件
	err = partner.CopyThemeToPartner(ctx, re.OldShopId, re.ShopId, re.ShopThemeId, re.TargetThemeId, data.Version)
	if err != nil {
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	shop.SendThemeApplyNotice(data.AliasName, data.Desc, font, admin, partnerInfo.Email, password, false)

	response.JsonExit(r, 0, "发布成功", types.Struct{})
}

type ThemeListReq struct {
	Domain    string `v:"required#域名不能为空"`
	Password  string `v:"required#密钥不能为空"`
	PartnerId int64  `v:"required|min:1#合作者id不能为空|合作者id必须大于0"`
	ThemeId   int64
}

// ThemeList 获取开发者已上架模板列表
func (t Theme) GetPartnerStoreThemeList(r *ghttp.Request) {
	ctx := slog.GetCtxVar(r)
	partnerId := partner.GetLoginID(r)
	if partnerId == 0 {
		response.JsonExit(r, -1, "没有登录", types.Struct{})
	}
	themes, err := partner.GetThemeList(ctx, partnerId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error(), types.Struct{})
	}
	response.JsonExit(r, 0, "success", themes)
}

type DelPartnerThemeReq struct {
	Id int64 `v:"required#模板id不能为空"`
}

func (t Theme) DelPartnerTheme(r *ghttp.Request) {
	ctx := slog.GetCtxVar(r)
	partnerId := partner.GetLoginID(r)
	if partnerId == 0 {
		response.JsonExit(r, -1, "没有登录", types.Struct{})
	}
	var data *DelPartnerThemeReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error(), types.Struct{})
	}
	err := partner.DelTheme(ctx, data.Id, partnerId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error(), types.Struct{})
	}
	response.JsonExit(r, 0, "删除成功", types.Struct{})
}

type SendPreviewShopReq struct {
	ApplyId int64 `v:"required|min:1#申请id必填|申请id必须填写"`
	//Email string `v:"required#邮箱必填"`
}

func (t Theme) SendPreviewShop(r *ghttp.Request) {
	partnerInfo := partner.GetClientInfo(r)
	if partnerInfo.Id == 0 {
		response.JsonExit(r, -1, "没有登录", types.Struct{})
	}
	ctx := slog.GetCtxVar(r)
	var data *SendPreviewShopReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	isUpdate := false
	re, err := partner.GetPreviewShopInfo(ctx, partnerInfo.Id, partnerInfo.Email, data.ApplyId)
	if err != nil {
		slog.Init(ctx).Error(err)
	}
	if re == nil {
		response.JsonExit(r, -1, "模板信息不存在")
	}
	if re.HasOlder {
		isUpdate = true
	}
	domain := fmt.Sprintf("%s.%s", re.PreviewShopName, surl.Hostname())
	var (
		admin = fmt.Sprintf("https://%s/admin/login", domain)
		font  = fmt.Sprintf("https://%s", domain)
	)
	shop.SendThemeApplyNotice(re.AliasName, re.Desc, font, admin, partnerInfo.Email, re.Password, isUpdate)
	response.JsonExit(r, 0, "邮件发送成功", types.Struct{})
}

type GetPartnerThemeListReq struct {
	AliasName string
}

// 获取开发者所有模板列表
func (t Theme) GetPartnerThemeList(r *ghttp.Request) {
	ctx := slog.GetCtxVar(r)
	partnerId := partner.GetLoginID(r)
	var emptyRes = make([]interface{}, 0)
	if partnerId == 0 {
		response.JsonExit(r, -1, "没有登录", emptyRes)
	}
	var data *GetPartnerThemeListReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error(), emptyRes)
	}

	themeList, err := partner.GetPartnerThemeList(ctx, partnerId, data.AliasName)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, 0, "success", emptyRes)
	}
	if themeList == nil {
		response.JsonExit(r, 0, "success", emptyRes)
	}

	response.JsonExit(r, 0, "success", themeList)
}

type SearchThemeByAliasNameReq struct {
	AliasName string `v:"required#模板名称必填"`
}

// SearchThemeByAliasName 搜索模板 废弃
func (t Theme) SearchThemeByAliasName(r *ghttp.Request) {
	partnerId := partner.GetLoginID(r)
	if partnerId == 0 {
		response.JsonExit(r, -1, "没有登录", types.Struct{})
	}
	ctx := slog.GetCtxVar(r)
	var data *SearchThemeByAliasNameReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	themeList, err := partner.SearchThemeByAliasName(ctx, partnerId, data.AliasName)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error(), types.Struct{})
	}
	response.JsonExit(r, 0, "success", themeList)
}

type PublishThemeReq struct {
	ShopThemeId   int64  `v:"required|min:1#店铺模板id不能为空|店铺模板id必须大于0"`
	ThemeName     string `v:"required#模板标识不能为空"`
	AliasName     string `v:"required#模板名称不能为空"`
	Version       string `v:"required#版本号不能为空"`
	StyleName     string `v:"required#风格名称不能为空"`
	Style         string `v:"required#风格标识不能为空"`
	Desc          string `v:"required#模板描述不能为空"`
	CompanyType   int
	CompanyId     int
	Phone         string
	Source        int
	SourceThemeId int64
	UpdateDesc    string
}

// UpdatePartnerStoreTheme 更新模板版本（已上架）
func (t Theme) UpdatePartnerStoreTheme(r *ghttp.Request) {
	partnerInfo := partner.GetClientInfo(r)
	if partnerInfo.Id == 0 {
		response.JsonExit(r, -1, "没有登录", types.Struct{})
	}
	ctx := slog.GetCtxVar(r)
	var data *structs.UpdatePartnerStoreThemeReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error(), types.Struct{})
	}
	if !gregex.IsMatchString(versionPattern, data.Version) {
		response.JsonExit(r, -1, "版本号非法", types.Struct{})
	}
	//迭代后 前端直接传shop_theme_id，不需要再查一遍 todo
	one, err := partner.GetStoreTheme(ctx, partnerInfo.Id, data)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}

	themeData := &partner.PublishThemeReq{}
	themeData.SourceThemeId = data.SourceThemeId
	themeData.ShopThemeId = one.ShopThemeId
	themeData.AliasName = data.AliasName
	themeData.StyleName = data.StyleName
	themeData.Version = data.Version
	themeData.Desc = data.Desc
	themeData.UpdateDesc = data.UpdateDesc

	themeData.ThemeName = one.ThemeName
	themeData.Style = one.Style

	shopName := one.ThemeName + "-" + one.Style
	//更新模板时生成新店铺名，暂定规则：shopName + 随机数
	shopName = shopName + "-" + string2.RandomString(6)
	password := string2.RandomString(6)
	re, err := shop.RegisterSeller(ctx, themeData, shopName, partnerInfo.Email, password, true)
	if err != nil {
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	err = partner.AddTheme(ctx, themeData, partnerInfo.Id, shopName,
		partnerInfo.Email, password, re.TargetThemeId, re.ShopId, 3, string2.DistinctStringArray(data.CategoryIDs))
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, "发布失败", types.Struct{})
	}

	var (
		admin = fmt.Sprintf("https://%s/admin/login", re.Domain)
		font  = fmt.Sprintf("https://%s", re.Domain)
	)

	//复制CSS文件
	err = partner.CopyCssToPartner(ctx, re)
	if err != nil {
		response.JsonExit(r, -4, gerror.Current(err).Error())
	}
	//复制其它文件
	err = partner.CopyThemeToPartner(ctx, re.OldShopId, re.ShopId, re.ShopThemeId, re.TargetThemeId, data.Version)
	if err != nil {
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}

	shop.SendThemeApplyNotice(one.AliasName, data.Desc, font, admin, partnerInfo.Email, password, true)
	response.JsonExit(r, 0, "更新成功", types.Struct{})
}

type ThemeApplyToStoreReq struct {
	ApplyId int64 `v:"required|min:1#申请id必填|申请id不能为空"`
}

// ThemeApplyToStore 模板申请上架
func (t Theme) ThemeApplyToStore(r *ghttp.Request) {
	partnerId := partner.GetLoginID(r)
	if partnerId == 0 {
		response.JsonExit(r, -1, "没有登录", types.Struct{})
	}
	ctx := slog.GetCtxVar(r)
	var data *ThemeApplyToStoreReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error(), types.Struct{})
	}
	applyThemeInfo, err := partner.FindApplyThemeById(ctx, partnerId, data.ApplyId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	newer, err := partner.IsNewerVersionOnStore(ctx, applyThemeInfo.ThemeName, applyThemeInfo.Style, applyThemeInfo.Version)
	if err != nil {
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	if newer {
		response.JsonExit(r, -1, "存在更新的已上架版，当前版本失效")
	}

	newPassword, err := partner.ThemeApplyToStore(ctx, partnerId, data.ApplyId, applyThemeInfo.PublishThemeId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	//发送邮箱
	applyThemeInfo.Password = newPassword
	partner.SendThemeApplyNoticeToPlatform(ctx, applyThemeInfo)
	response.JsonExit(r, 0, "success", types.Struct{})
}

type HandlerThemeApplyReq struct {
	ApplyId       int64 `v:"required|min:1#申请id必填|申请id不能为空"`
	OperateStatus int64 `v:"required|min:1#请选择申请或拒绝|状态必选"`
}

// HandlerThemeApply 处理上架  废弃 废弃 废弃
func (t Theme) HandlerThemeApply(r *ghttp.Request) {
	ctx := slog.GetCtxVar(r)
	var data *HandlerThemeApplyReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error(), types.Struct{})
	}
	re, err := partner.FindApplyThemeById(ctx, 0, data.ApplyId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error(), types.Struct{})
	}
	if data.OperateStatus == 1 {
		err = partner.UploadToStore(ctx, re)
		if err != nil {
			slog.Init(ctx).Error(err)
			response.JsonExit(r, -1, gerror.Current(err).Error(), types.Struct{})
		}
		response.JsonExit(r, 0, "success", types.Struct{})
	}
	err = partner.DenyUploadToStore(ctx, data.ApplyId, false)
	if err != nil {
		response.JsonExit(r, -1, gerror.Current(err).Error(), types.Struct{})
	}
	response.JsonExit(r, 0, "success", types.Struct{})
}

func (t Theme) AllowThemeApply(r *ghttp.Request) {
	ctx := slog.GetCtxVar(r)
	applyId := r.GetInt64("id")
	aliasName := r.GetString("aliasName")
	styleName := r.GetString("styleName")
	desc := r.GetString("desc")
	version := r.GetString("version")
	email := r.GetString("email")
	token := r.GetString("token")
	stime := url.QueryEscape(r.GetString("stime"))

	//验证链接是否有效
	err := partner.CheckRoute(r, ctx)
	if err != nil {
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	data := g.Map{
		"status": gconv.Int(1),
	}
	res, err := partner.ChangeThemeStoreStatus(ctx, applyId, 1, token, stime)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}

	slog.Init(ctx).Info("AllowPublish theme ", data, res)

	routerTokenKey := fmt.Sprintf("PartnerTheme:routeTokenKey:%s", r.GetString("routeToken"))
	_, _ = sredis.Client().Del(ctx, routerTokenKey).Result()

	//告知开发者，应用审核通过
	shop.SendThemeAuditNotice(aliasName, styleName, desc, version, email, true)
	response.JsonExit(r, 0, "处理成功")
}

// 拒绝上架
func (t Theme) RejectThemeApply(r *ghttp.Request) {
	ctx := slog.GetCtxVar(r)
	applyId := r.GetInt64("id")
	aliasName := r.GetString("aliasName")
	styleName := r.GetString("styleName")
	desc := r.GetString("desc")
	version := r.GetString("version")
	email := r.GetString("email")
	token := r.GetString("token")
	stime := url.QueryEscape(r.GetString("stime"))

	//验证链接是否有效
	err := partner.CheckRoute(r, ctx)
	if err != nil {
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	data := g.Map{
		"status": gconv.Int(2),
	}
	re, err := partner.FindApplyThemeById(ctx, 0, applyId)
	if err != nil {
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	if re == nil {
		response.JsonExit(r, -1, "模板信息不存在")
	}
	hasOlder, err := partner.IsOlderVersionOnStore(ctx, re.ThemeName, re.Style, re.Version)
	if err != nil {
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}

	err = partner.DenyUploadToStore(ctx, applyId, hasOlder, token, stime)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}

	slog.Init(ctx).Info("reject theme ", data)

	routerTokenKey := fmt.Sprintf("PartnerTheme:routeTokenKey:%s", r.GetString("routeToken"))
	_, _ = sredis.Client().Del(ctx, routerTokenKey).Result()

	//告知开发者，应用审核不通过
	shop.SendThemeAuditNotice(aliasName, styleName, desc, version, email, false)
	response.JsonExit(r, 0, "处理成功")
}

type GetStoreThemeInfoByIdReq struct {
	ThemeId int64 `v:"required|min:1#模板id必填|模板id必须填写"`
}

func (t Theme) GetStoreThemeInfoById(r *ghttp.Request) {
	ctx := slog.GetCtxVar(r)
	var data *GetStoreThemeInfoByIdReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	info, err := partner.GetStoreThemeInfoById(ctx, data.ThemeId)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	response.JsonExit(r, 0, "success", info)
}

type UpdateThemeVersionReq struct {
	StoreThemeId int64 `v:"required|min:1#商城模板id必填|商城模板id必须填写"`
	ShopThemeId  int64 `v:"required|min:1#店铺模板id必填|店铺模板id必须填写"`
}

// UpdateThemeVersion 商户更新开发者模板版本
func (t Theme) UpdateThemeVersion(r *ghttp.Request) {
	ctx := slog.GetCtxVar(r)
	var data *UpdateThemeVersionReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	shopInfo := gjson.New(r.GetHeader("Uinfo"))
	shopId := shopInfo.GetInt64("id")
	err := store.UpdateThemeVersion(ctx, data.StoreThemeId, data.ShopThemeId, shopId)
	if err != nil {
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	response.JsonExit(r, 0, "更新成功")
}

func (t Theme) EditPartnerThemeInfo(r *ghttp.Request) {
	partnerId := partner.GetLoginID(r)
	if partnerId == 0 {
		response.JsonExit(r, -1, "没有登录", types.Struct{})
	}
	ctx := slog.GetCtxVar(r)
	var data *structs.EditPartnerThemeInfoReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, 0, gerror.Current(err).Error())
	}

	up := make(map[string]interface{})
	if data.AliasName != "" {
		up["alias_name"] = data.AliasName
	}
	if data.StyleName != "" {
		up["style_name"] = data.StyleName
	}
	if data.Desc != "" {
		up["desc"] = data.Desc
	}
	if data.UpdateDesc != "" {
		up["update_desc"] = data.UpdateDesc
	}
	if len(up) == 0 {
		response.JsonExit(r, -1, "没有修改")
	}
	_, err := partner.ModifyPartnerThemeInfo(ctx, partnerId, data.Id, up)
	if err != nil {
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	//追加模板分类的修改
	err = category.SaveThemeCategory(ctx, int(data.Id), string2.DistinctStringArray(data.CategoryIDs), category.PIRATE_CATEGORY_TYPE)
	if err != nil {
		response.JsonExit(r, -1, gerror.Current(err).Error())
		return
	}
	response.JsonExit(r, 0, "success")
}

type UpdateOutStoreThemeReq struct {
	Id          int64 `v:"required|min:0#申请id必须填写|申请id必填"`
	AliasName   string
	StyleName   string
	UpdateDesc  string
	CategoryIDs []string `v:"required#分类ID不能为空" json:"category_ids"`
}

// 更新未上架模板信息
func (t Theme) UpdateOutStoreTheme(r *ghttp.Request) {
	partnerId := partner.GetLoginID(r)
	if partnerId == 0 {
		response.JsonExit(r, -1, "没有登录", types.Struct{})
	}
	ctx := slog.GetCtxVar(r)
	var data *UpdateOutStoreThemeReq
	if err := r.Parse(&data); err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	var up = make(map[string]interface{})
	if data.AliasName != "" {
		up["alias_name"] = data.AliasName
	}
	if data.StyleName != "" {
		up["style_name"] = data.StyleName
	}
	if data.UpdateDesc != "" {
		up["update_desc"] = data.UpdateDesc
	}
	if len(up) == 0 {
		response.JsonExit(r, -1, "没有修改")
	}
	//修改模板申请表信息
	themeApplyInfo, err := partner.ModifyPartnerThemeInfo(ctx, partnerId, data.Id, up)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}
	//获取申请模板信息
	//themeApplyInfo, err := partner.FindApplyThemeById(ctx, partnerId, data.Id)
	if err != nil {
		slog.Init(ctx).Error(err)
		response.JsonExit(r, -1, gerror.Current(err).Error())
	}

	//追加模板分类的修改
	err = category.SaveThemeCategory(ctx, int(data.Id), string2.DistinctStringArray(data.CategoryIDs), category.PIRATE_CATEGORY_TYPE)
	if err != nil {
		response.JsonExit(r, -1, gerror.Current(err).Error())
		return
	}

	response.JsonExit(r, 0, "ok")
}
