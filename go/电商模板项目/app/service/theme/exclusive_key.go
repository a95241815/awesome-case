package theme

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/util/gconv"
	"github.com/sz-sailing/gflib/library/slog"
	"github.com/sz-sailing/gflib/library/smongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

// GetOutput 获取专属模板密钥列表
type GetOutput struct {
	Count int64  `json:"count"`
	List  []list `json:"list"`
}
type list struct {
	AppID      int64  `json:"app_id"`
	CreateTime string `json:"create_time"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Scope      int64  `json:"scope"`
}

func GetKeyLists(ctx context.Context, shopId int64, page int64, limit int64) (GetOutput, error) {
	//获取专属模板密钥列表总数
	count, err := g.DB().Table("sail_application").Where("user_id = ? and type = 10 and is_del = 0", shopId).FindCount()

	var lists = GetOutput{
		Count: gconv.Int64(count),
		List:  []list{},
	}

	if err != nil {
		slog.Init(ctx).Error(err)
		return lists, err
	}

	if count < 1 {
		return lists, nil
	}

	if page <= 1 {
		page = 1
	}
	page = (page - 1) * limit

	//分页获取列表信息
	application, err := g.DB().Table("sail_application sa").LeftJoin("sail_application_auth_theme saat", "sa.app_id = saat.app_id").Fields("sa.app_id, sa.scope, sa.ctime, saat.name, saat.email").Where("sa.user_id = ? and sa.type = 10 and sa.is_del = 0", shopId).Offset(int(page)).Limit(int(limit)).Order("ctime DESC").FindAll()
	if err != nil {
		slog.Init(ctx).Error(err)
		return lists, err
	}

	if len(application) == 0 {
		return lists, nil
	}

	//组装输出专属模板密钥列表数据
	for _, value := range application {
		scope := 0
		if gconv.String(value["scope"]) != "" {
			scope = 1
		}

		time := strings.Fields(strings.TrimSpace(gconv.String(value["ctime"])))

		var t = list{}
		t.AppID = gconv.Int64(value["app_id"])
		t.Name = gconv.String(value["name"])
		t.Email = gconv.String(value["email"])
		t.CreateTime = gconv.String(time[0])
		t.Scope = int64(scope)

		lists.List = append(lists.List, t)
	}

	return lists, nil
}

// AddKey 添加专属模板密钥
func AddKey(ctx context.Context, shopId int64, name string, email string, scope int64) error {
	var appId int64

	var stringScope string
	if scope == 1 {
		stringScope = "write_themes"
	}

	// 查询店铺信息
	shop, err := g.DB().Table("sail_shop").Value("domain", "id", shopId)
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}
	if shop == nil {
		err := errors.New("店铺信息不存在")
		slog.Init(ctx).Error(err)
		return err
	}
	if shop.String() == "" {
		err := errors.New("店铺信息不存在")
		slog.Init(ctx).Error(err)
		return err
	}

	//开启事务
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx *gdb.TX) error {
		//添加电商应用表
		result, err := tx.Model("sail_application").Ctx(ctx).Insert(g.Map{
			"name":          "开放平台模版应用",
			"client_id":     RandomString(40),
			"client_secret": RandomString(40),
			"access_token":  RandomString(40),
			"user_id":       shopId,
			"scope":         stringScope,
			"type":          10,
			"status":        1,
			"is_del":        0,
			"domain":        shop.String(),
		})

		if result == nil {
			slog.Init(ctx).Error(err)
			return err
		}

		appId, err = result.LastInsertId()
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}

		//添加开放平台模版基础信息表
		result, err = tx.Model("sail_application_auth_theme").Ctx(ctx).Insert(g.Map{
			"app_id": appId,
			"name":   name,
			"email":  email,
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

// UpdateKey 修改专属模板密钥
func UpdateKey(ctx context.Context, shopId int64, appId int64, name string, email string, scope int64) error {
	var stringScope string
	if scope == 1 {
		stringScope = "write_themes"
	}

	//查询密钥信息
	application, err := g.DB().Table("sail_application sa").LeftJoin("sail_application_auth_theme saat", "sa.app_id = saat.app_id").Fields("sa.client_secret, saat.name").Where("sa.app_id = ? and sa.user_id = ? and sa.type = 10 and sa.is_del = 0", appId, shopId).FindOne()
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}

	if application["client_secret"].String() == "" || application["name"].String() == "" {
		err := errors.New("密钥数据不存在")
		slog.Init(ctx).Error(err)
		return err
	}
	//开启事务
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx *gdb.TX) error {
		//修改电商应用表
		_, err := tx.Model("sail_application").Ctx(ctx).Update(g.Map{
			"scope": stringScope,
		}, "app_id = ? and user_id = ?", appId, shopId)
		if err != nil {
			slog.Init(ctx).Error(err)
			return err
		}

		//修改开放平台模版基础信息表
		_, err = tx.Model("sail_application_auth_theme").Ctx(ctx).Update(g.Map{
			"name":  name,
			"email": email,
		}, "app_id = ?", appId)
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

// DelKey 删除专属模板密钥
func DelKey(ctx context.Context, shopId int64, appId int64) error {
	//查询密钥信息
	application, err := g.DB().Table("sail_application").
		Fields("app_id, is_del").
		Where("app_id = ? and user_id = ? and type = 10", appId, shopId).
		FindOne()
	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}

	if application == nil {
		err := errors.New("密钥数据不存在")
		slog.Init(ctx).Error(err)
		return err
	}

	if application["is_del"].Int() == 1 {
		err := errors.New("密钥数据已删除")
		slog.Init(ctx).Error(err)
		return err
	}

	_, err = g.DB().Table("sail_application").Update(g.Map{
		"is_del": 1,
	}, "app_id = ? and user_id = ?", appId, shopId)

	if err != nil {
		slog.Init(ctx).Error(err)
		return err
	}

	return nil
}

// GetKey 获取专属模板密钥
func GetKey(ctx context.Context, shopId int64, appId int64) (interface{}, error) {
	var secretKey string
	var result interface{}

	//查询密钥信息
	application, err := g.DB().Table("sail_application").
		Fields("app_id, client_secret").
		Where("app_id = ? and user_id = ? and type = 10 and is_del = 0", appId, shopId).
		FindOne()
	if err != nil {
		slog.Init(ctx).Error(err)
		return result, err
	}

	secretKey = application["client_secret"].String()

	if secretKey == "" {
		slog.Init(ctx).Info(fmt.Sprintf("店铺ID %d 密钥ID %d 不存在", shopId, appId))
		return result, nil
	}

	return g.Map{"secret_key": secretKey}, nil
}

// GetKeyInfo 获取专属模板密钥详细信息
func GetKeyInfo(ctx context.Context, shopId int64, appId int64) (interface{}, error) {
	var info = make(map[string]interface{}, 0)
	application, err := g.DB().Table("sail_application sa").
		LeftJoin("sail_application_auth_theme saat", "sa.app_id = saat.app_id").
		Fields("sa.app_id, sa.scope, sa.ctime, saat.name, saat.email").
		Where("sa.app_id = ? and sa.user_id = ? and sa.type = 10 and sa.is_del = 0", appId, shopId).
		FindOne()

	if err != nil {
		slog.Init(ctx).Error(err)
		return info, err
	}

	if len(application) == 0 {
		return info, err
	}

	scope := 0
	if gconv.String(application["scope"]) != "" {
		scope = 1
	}

	info = g.Map{
		"name":  gconv.String(application["name"]),
		"email": gconv.String(application["email"]),
		"scope": scope,
	}

	return info, nil
}

// LogsResponse 获取专属模板密钥操作日志
type LogsResponse struct {
	Count int64            `json:"count"`
	List  []*listsResponse `json:"list"`
}

type listsResponse struct {
	YearMonthDay      string `json:"year_month_day"`
	MinutesAndSeconds string `json:"minutes_and_seconds"`
	Operation         string `json:"operation"`
	ThemeName         string `json:"theme_name"`
}

type listsInit struct {
	Time        string `json:"time"`
	OperationId int64  `json:"operation_id"`
	ThemeId     int64  `json:"theme_id"`
}

func GetKeyLogs(ctx context.Context, shopId int64, appId int64, page int64, limit int64) (LogsResponse, error) {
	var logs = LogsResponse{
		Count: 0,
		List:  []*listsResponse{},
	}

	//查询密钥信息
	application, err := g.DB().Table("sail_application").Fields("app_id, client_secret").Where("app_id = ? and user_id = ? and type = 10 and is_del = 0", appId, shopId).FindOne()
	if err != nil {
		slog.Init(ctx).Error(err)
		return logs, err
	}

	if application == nil {
		slog.Init(ctx).Info("密钥数据不存在")
		return logs, nil
	}

	if page <= 1 {
		page = 1
	}
	page = (page - 1) * limit

	//从mongodb获取专属模板密钥操作日志
	result, err := GetKeyLogsMongo(ctx, shopId, application["client_secret"].String(), page, limit)
	if err != nil {
		slog.Init(ctx).Error(err)
		return logs, err
	}

	//组装日志输出总数数据
	logs.Count = result.Count

	//初始化日志列表数据
	var themeIds []int64
	var lists []listsInit
	for _, value := range result.List {
		lists = append(lists, listsInit{
			Time:        gconv.String(value.Time),
			OperationId: gconv.Int64(value.Operation),
			ThemeId:     gconv.Int64(value.ThemeId),
		})

		themeIds = append(themeIds, gconv.Int64(value.ThemeId))
	}

	//获取模板名称
	names, err := g.DB().Table("sail_shop_theme").Fields("id, theme_name").Where("id in (?)", themeIds).FindAll()

	//组装模板Id和名称对应关系
	themeNames := map[int64]string{}
	for _, value := range names {
		themeNames[gconv.Int64(value["id"])] = gconv.String(value["theme_name"])
	}

	//组装日志输出列表数据
	var Operation = [...]string{
		1: "获取了模板",
		2: "发布了模板",
		3: "激活了密钥",
	}
	for _, value := range lists {
		themeName := ""
		if value.ThemeId > 0 && value.OperationId != 3 {
			themeName = themeNames[value.ThemeId]
		}

		time := strings.Fields(strings.TrimSpace(gconv.String(value.Time)))
		logs.List = append(logs.List, &listsResponse{
			YearMonthDay:      gconv.String(time[0]),
			MinutesAndSeconds: gconv.String(time[1]),
			Operation:         gconv.String(Operation[value.OperationId]),
			ThemeName:         gconv.String(themeName),
		})

	}

	return logs, nil
}

// MongoOutput 从mongodb获取专属模板密钥操作日志
type MongoOutput struct {
	Count int64        `json:"count"`
	List  []mongoLists `json:"list"`
}

type mongoLists struct {
	Time      string `json:"time"`
	Operation int64  `bson:"operation" json:"operation"`
	ThemeId   int64  `bson:"theme_id" json:"theme_id"`
}

func GetKeyLogsMongo(ctx context.Context, shopId int64, secret string, offset int64, limit int64) (MongoOutput, error) {
	//链接mongoDB
	coll := smongodb.Conn().GetColl("operation_log")

	var lists = MongoOutput{
		Count: 0,
		List:  []mongoLists{},
	}

	//计算日志列表总数
	count, err := coll.CountDocuments(
		context.TODO(),
		bson.D{
			{"shop_id", gconv.Int64(shopId)},
			{"secret", gconv.String(secret)},
		},
		options.Count(),
	)

	if err != nil {
		slog.Init(ctx).Error(err)
		return lists, err
	}

	if count < 1 {
		return lists, nil
	}

	lists.Count = count

	//分页获取列表数据
	opts := options.Find().SetSkip(offset).SetLimit(limit).SetSort(g.Map{"time": -1}).SetProjection(bson.M{"_id": 0})
	cursor, err := coll.Find(
		context.TODO(),
		bson.D{
			{"shop_id", gconv.Int64(shopId)},
			{"secret", gconv.String(secret)},
		},
		opts,
	)

	if cursor == nil {
		slog.Init(ctx).Error(err)
		return lists, err
	}

	if err = cursor.All(context.TODO(), &lists.List); err != nil {
		slog.Init(ctx).Error(err)
		return lists, err
	}

	return lists, nil
}

// RandomString 生成随机数
func RandomString(n int) string {
	randBytes := make([]byte, n/2)
	_, _ = rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}
