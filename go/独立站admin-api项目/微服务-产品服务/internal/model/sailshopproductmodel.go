package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/mr"
	"github.com/tal-tech/go-zero/core/stores/redis"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tal-tech/go-zero/core/stores/cache"
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"github.com/tal-tech/go-zero/core/stores/sqlx"
	"github.com/tal-tech/go-zero/core/stringx"
	"github.com/tal-tech/go-zero/tools/goctl/model/sql/builderx"

	"github.com/grokify/html-strip-tags-go"
)

var (
	sailShopProductFieldNames          = builderx.RawFieldNames(&SailShopProduct{})
	sailShopProductRows                = strings.Join(sailShopProductFieldNames, ",")
	sailShopProductRowsExpectAutoSet   = strings.Join(stringx.Remove(sailShopProductFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	sailShopProductRowsWithPlaceHolder = strings.Join(stringx.Remove(sailShopProductFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"

	cacheSailShopProductIdPrefix            = "cache#sailShopProduct#id#"
	cacheSailShopProductShopIdHandlerPrefix = "cache#sailShopProduct#shopId#handler#"
)

type (
	SailShopProductModel interface {
		Insert(data SailShopProduct) (sql.Result, error)
		InsertProduct(data InsertProductData, redis *redis.Redis) (sql.Result, *RespImageData, error)
		FindOne(shopId int64, filter ...HandlerOption) (*SailShopProduct, error)
		FindList(options ListOptions, shopId int64) (*[]SailShopProduct, error)
		Count(shopId int64, options ListOptions) (int64, error)
		FindOneByShopIdHandler(shopId int64, handler string) (*SailShopProduct, error)
		Delete(shopId, productId int64, redis2 *redis.Redis) error
		Update(data InsertProductData, productId int64, redis2 *redis.Redis) (*RespImageData, error)
		UpdateDefaultImage(imageId int64, productId int64) error
	}

	defaultSailShopProductModel struct {
		sqlc.CachedConn
		table string
	}

	SailShopProduct struct {
		ImageTmpUrls       sql.NullString `db:"image_tmp_urls"`
		Scores             int64          `db:"scores"`            // 评分总数
		CountSales         int64          `db:"count_sales"`       // 销量
		YoutubeVideoPos    string         `db:"youtube_video_pos"` // youtube视频所在轮播图中位置
		DefaultImageId     int64          `db:"default_image_id"`  // 产品主图
		IsDel              int64          `db:"is_del"`            // 是否删除
		Status             int64          `db:"status"`            // 上架状态(1:上架;2:下架;)
		ProductStock       int64          `db:"product_stock"`     // 商品库存
		Sort               int64          `db:"sort"`              // 排序，越大越靠前
		Handler            string         `db:"handler"`           // 产品标题字符串索引
		ShopId             int64          `db:"shop_id"`           // 商店唯一ID
		DefaultSkuCode     string         `db:"default_sku_code"`  // 默认的SKU
		CountSkus          int64          `db:"count_skus"`        // SKU数量统计(不包括系统生成的默认sku)
		IsUseStock         int64          `db:"is_use_stock"`      // 是否启用库存
		SoldoutPolicy      sql.NullString `db:"soldout_policy"`    // 售罄策略(Y:无库存继续出售;N:无库存不能出售)
		Source             sql.NullString `db:"source"`            // 导入来源
		CreatedAt          time.Time      `db:"created_at"`        // 创建时间
		HandlerOrigin      string         `db:"handler_origin"`    // 产品标题字符串索引原始值
		Price              float64        `db:"price"`             // 商品价格
		IsShowComment      int64          `db:"is_show_comment"`
		SeoTitle           string         `db:"seo_title"`    // SEO标题
		UpdatedAt          time.Time      `db:"updated_at"`   // 更新时间
		BodyHtml           sql.NullString `db:"body_html"`    // 商品描述
		Weight             float64        `db:"weight"`       // 产品重量
		Comments           int64          `db:"comments"`     // 评论总数
		SeoDesc            string         `db:"seo_desc"`     // SEO描述
		Attribute          sql.NullString `db:"attribute"`    // 产品属性，json格式
		IsLogistics        int64          `db:"is_logistics"` // 是否需要物流运送
		SubTitle           string         `db:"sub_title"`    // 副标题
		Id                 int64          `db:"id"`
		Title              string         `db:"title"`     // 商品标题
		ImageIds           string         `db:"image_ids"` // 产品图片id集合，按逗号拼接
		DefaultImageTmpUrl string         `db:"default_image_tmp_url"`
		YoutubeVideoUrl    string         `db:"youtube_video_url"` // youtube视频地址
		CompareAtPrice     float64        `db:"compare_at_price"`  // 对比价格
		WeightUnit         string         `db:"weight_unit"`       // 产品重量单位
		PublishedAt        time.Time      `db:"published_at"`      // 发布时间
		IsRead             int64          `db:"is_read"`           // 产品导入 先判定是否已经读取 1读取 2未读取
	}

	ListOptions struct {
		Limit           int64
		Page            int64
		SinceId         int64
		CreatedAtMin    string
		CreatedAtMax    string
		UpdatedAtMin    string
		UpdatedAtMax    string
		PublishedAtMin  string
		PublishedAtMax  string
		PublishedStatus string
		Ids             string
		OrderBy         int64
		ProductId       int64
		Title           string
		Handlers        []string
		IsNewVersion    bool
	}

	InsertProductData struct {
		Id                int64
		Title             string
		Weight            float64
		WeightUnit        string
		Price             float64
		ComparePrice      float64
		SoldOutPolicy     string
		BodyHtml          string
		Status            string
		Sku               string
		ShopId            int64
		SeoTitle          string
		SeoDesc           string
		YoutubeVideoUrl   string
		YoutubeVideoPos   string
		RequireShipping   string
		IsUseStock        string
		Handler           string
		OriginHandler     string
		Tags              []string
		Attribute         string
		InventoryQuantity int64
		DefaultImage      ImageData
		Images            *[]ImageData
		Variants          *[]VariantData
	}

	ImageData struct {
		Id         int64
		FileKey    string
		ImageWidth int64
	}

	VariantData struct {
		Id                int64
		Price             float64
		ComparePrice      float64
		Weight            float64
		WeightUnit        string
		RequiresShipping  string
		InventoryQuantity int64
		Image             ImageData
		Spec              string
		Sort              int64
		SkuCode           string
		Title             string
		Options           string
		IsChecked         int64
	}

	RespImageData struct {
		ProductId       int64
		DefaultImageUrl string
		Images          []ImageItem
		Variants        []VariantItem
	}

	ImageItem struct {
		Url string
	}

	VariantItem struct {
		VariantId int64
		Url       string
	}

	UpdateProductData struct {
		Id            int64
		Title         string
		BodyHtml      string
		SeoDesc       string
		SeoTitle      string
		Status        string
		Price         float64
		ComparePrice  float64
		Weight        float64
		WeightUnit    string
		IsUseStock    string
		SoldoutPolicy string
		Variants      []VariantUpdateItem
	}

	ImageItemUpdate struct {
		Id  int64
		Url string
	}

	VariantUpdateItem struct {
		Id int64
		VariantItem
	}
)

func NewSailShopProductModel(conn sqlx.SqlConn, c cache.CacheConf) SailShopProductModel {
	return &defaultSailShopProductModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`sail_shop_product`",
	}
}

func GetAutoIncrementId(redis2 *redis.Redis) (int64, error) {
	id, err := redis2.Incr(AUTO_INCREMENT_KEY)
	return id, err
}

func (m *defaultSailShopProductModel) Insert(data SailShopProduct) (sql.Result, error) {
	sailShopProductShopIdHandlerKey := fmt.Sprintf("%s%v%v", cacheSailShopProductShopIdHandlerPrefix, data.ShopId, data.Handler)
	now := time.Now()
	data.CreatedAt, data.UpdatedAt = now, now
	data.Handler = data.Title
	data.HandlerOrigin = data.Title
	ret, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, sailShopProductRowsExpectAutoSet)
		return conn.Exec(query, data.ImageTmpUrls, data.Scores, data.CountSales, data.YoutubeVideoPos, data.DefaultImageId, data.IsDel, data.Status, data.ProductStock, data.Sort, data.Handler, data.ShopId, data.DefaultSkuCode, data.CountSkus, data.IsUseStock, data.SoldoutPolicy, data.Source, data.CreatedAt, data.HandlerOrigin, data.Price, data.IsShowComment, data.SeoTitle, data.UpdatedAt, data.BodyHtml, data.Weight, data.Comments, data.SeoDesc, data.Attribute, data.IsLogistics, data.SubTitle, data.Title, data.ImageIds, data.DefaultImageTmpUrl, data.YoutubeVideoUrl, data.CompareAtPrice, data.WeightUnit, data.PublishedAt, data.IsRead)
	}, sailShopProductShopIdHandlerKey)
	return ret, err
}

func (m *defaultSailShopProductModel) InsertProduct(data InsertProductData, redis2 *redis.Redis) (sql.Result, *RespImageData, error) {
	var result sql.Result
	var respImageData RespImageData
	var lastProId int64
	lockKey := fmt.Sprintf("adminapi_lock_product_title_%d_%s", data.ShopId, data.Title)
	redisLock := redis.NewRedisLock(redis2, lockKey)
	redisLock.SetExpire(5)
	if ok, err := redisLock.Acquire(); !ok || err != nil {
		return nil, nil, errors.New(" product title repeated, too many request ")
	}
	defer func() {
		recover()
		redisLock.Release()
	}()

	productId, err := GetAutoIncrementId(redis2)
	if err != nil {
		logx.Error("获取商品自增id出错：", err)
		return nil, nil, errors.New(" internal server error ")
	}
	data.Id = productId
	respImageData.DefaultImageUrl = data.DefaultImage.FileKey
	if data.SeoDesc == "" {
		data.SeoDesc = m.GetDefaultSeoDesc(data.BodyHtml)
	}
	data.Handler = m.generateHandler(data.Title, data.ShopId)
	data.OriginHandler = m.generateOriginHandler(data.Title)

	if data.ShopId == 0 {
		logx.Error("缺少shop_id参数")
		return nil, nil, errors.New("shop_id 参数不能为空")
	}
	if data.Title == "" {
		logx.Error("缺少title参数")
		return nil, nil, errors.New("title参数不能为空")
	}
	logx.Error("variants", data.Variants)

	err = m.Transact(func(session sqlx.Session) error {
		resultProduct, err := StmtInsertProduct(session, data)
		if err != nil {
			logx.Error(err)
			return err
		}
		result = resultProduct
		lastProId, err = resultProduct.LastInsertId()
		if err != nil {
			logx.Error(err)
			return err
		}
		respImageData.ProductId = lastProId
		err = mr.Finish(func() error {
			if data.Variants != nil {
				if len(*data.Variants) != 0 {
					for v, variantData := range *data.Variants {
						variantData.Sort = int64(v + 1)
						resultVariant, err := StmtInsertVariant(session, variantData, lastProId, data.ShopId, redis2)
						if err != nil {
							logx.Error("新增子商品出错：", err)
							return errors.New("internal server error")
						}
						lastVariantId, err := resultVariant.LastInsertId()
						if err != nil {
							logx.Error("新增子商品出错：", err)
							return errors.New("internal server error")
						}
						tempVariantItem := VariantItem{}
						tempVariantItem.VariantId = lastVariantId
						tempVariantItem.Url = variantData.Image.FileKey
						respImageData.Variants = append(respImageData.Variants, tempVariantItem)
					}
				}
			} else {
				variantFileds := " shop_id, product_id, price,compare_at_price, weight, weight_unit, sort,  sku_code, is_checked, inventory_quantity, spec"
				variantValues := " ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?"
				variantInsertQuery := fmt.Sprintf(" insert into %s (%s) values (%s) ", "sail_shop_product_variant", variantFileds, variantValues)
				stmtVariant, err := session.Prepare(variantInsertQuery)
				if err != nil {
					logx.Error(err)
					return err
				}
				defer stmtVariant.Close()

				_, err = stmtVariant.Exec(data.ShopId, lastProId, data.Price, data.ComparePrice, data.Weight, data.WeightUnit, 1, data.Sku, 1, data.InventoryQuantity, "[]")
				if err != nil {
					logx.Errorf("insert product variant stmt exec: %s", err)
					return errors.New("internal server error")
				}
			}
			if len(data.Tags) != 0 {
				for _, tag := range data.Tags {
					tagResult, err := StmtQueryTag(session, data.ShopId, tag)
					switch err {
					case nil:
						_, err = StmtInsertTagProduct(session, data.ShopId, tagResult.Id, lastProId)
						if err != nil {
							logx.Error(err)
							return err
						}
					case sqlc.ErrNotFound:
						tagNewResult, err := StmtInsertTags(session, data.ShopId, tag)
						if err != nil {
							logx.Error(err)
							return err
						}
						newTagId, err := tagNewResult.LastInsertId()
						if err != nil {
							logx.Error(err)
							return err
						}
						_, err = StmtInsertTagProduct(session, data.ShopId, newTagId, lastProId)
						if err != nil {
							logx.Error(err)
							return err
						}
					default:
						logx.Error(err)
						return err
					}
				}
			}
			return nil
		}, func() error {
			detailData := SailShopProductDetail{
				ProductId: lastProId,
				BodyHtml:  data.BodyHtml,
			}
			detailModel := defaultSailShopProductDetailModel{
				CachedConn: m.CachedConn,
				table:      "sail_shop_product_detail",
			}
			_, err := detailModel.Insert(detailData, session)
			if err != nil {
				logx.Error("body_html新增失败:", err)
				return err
			}
			return nil
		})

		if err != nil {
			logx.Error(err)
			return err
		}
		if data.Images != nil {
			if len(*data.Images) != 0 {
				for _, imageData := range *data.Images {
					if imageData.FileKey == "" {
						logx.Error("商品图片添加出错")
						return errors.New("internal server error")
					}
					tempImageItem := ImageItem{}
					tempImageItem.Url = imageData.FileKey
					respImageData.Images = append(respImageData.Images, tempImageItem)
				}
			}
		}

		return nil

	})
	if err != nil {
		logx.Error(err)
		return nil, nil, err
	}
	err = m.AfterSave(data.ShopId, lastProId, ES_SYNC_EVENT_ADD, redis2)
	if err != nil {
		logx.Error(err)
	}
	return result, &respImageData, nil
}

func (m *defaultSailShopProductModel) UpdateDefaultImage(imageId int64, productId int64) error {
	if imageId == 0 {
		return errors.New("image id is not set")
	}
	if productId == 0 {
		return errors.New("product id is not set")
	}
	query := fmt.Sprintf("update %s set  `default_image_id` = ?  where `id` = ? and `is_del` = 0 ", m.table)
	_, err := m.ExecNoCache(query, imageId, productId)
	return err
}

func (m *defaultSailShopProductModel) FindOne(shopId int64, filter ...HandlerOption) (*SailShopProduct, error) {
	argsQuery, extraWhere, err := SetQueryStr(filter)
	if err != nil {
		logx.Error(err)
		return nil, err
	}
	args := make([]interface{}, 0)
	args = append(args, shopId)
	args = append(args, argsQuery...)

	var resp SailShopProduct
	query := fmt.Sprintf("select %s from %s where `shop_id` = ? %s and is_del = 0 limit 1", sailShopProductRows, m.table, extraWhere)

	err = m.QueryRowNoCache(&resp, query, args...)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultSailShopProductModel) Count(shopId int64, options ListOptions) (int64, error) {
	var resp int64
	args := make([]interface{}, 0)
	args = append(args, shopId)
	//err := m.QueryRow(&resp, sailShopProductIdKey, func(conn sqlx.SqlConn, v interface{}) error {
	//	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", sailShopProductRows, m.table)
	//	return conn.QueryRow(v, query, id, shopId)
	//})

	query := fmt.Sprintf("select count(*) from %s where `shop_id` = ? and is_del = 0 ", m.table)
	if options.SinceId > 0 {
		query += " and id > ? "
		args = append(args, options.SinceId)
	}
	if options.Ids != "" {
		idSet := strings.Split(options.Ids, ",")
		if len(idSet) == 1 {
			idNum, err := strconv.Atoi(idSet[0])
			if err != nil {
				logx.Error("err:", errors.New(" field ids in wrong format "))
				return 0, errors.New(" field ids in wrong format ")
			}
			args = append(args, idNum)
			query += " and id in (?) "
		}
		if len(idSet) > 1 {
			for k, id := range idSet {
				idNum, err := strconv.Atoi(id)
				if err != nil {
					logx.Error("err:", errors.New(" field ids in wrong format "))
					return 0, errors.New(" field ids in wrong format ")
				}
				if k == 0 {
					args = append(args, idNum)
					query += " and id in ( ? "
				}
				if k > 0 {
					args = append(args, idNum)
					query += " , ? "
				}
				if k == len(idSet)-1 {
					//args = append(args, idNum)
					query += " ) "
					//argsQuery = append(argsQuery, idNum)
				}
			}
		}

	}
	if options.CreatedAtMin != "" {
		args = append(args, options.CreatedAtMin)
		query += " and created_at >= ? "
	}
	if options.CreatedAtMax != "" {
		args = append(args, options.CreatedAtMax)
		query += " and created_at <= ? "
	}
	if options.UpdatedAtMin != "" {
		args = append(args, options.UpdatedAtMin)
		query += " and updated_at >= ? "
	}
	if options.UpdatedAtMax != "" {
		args = append(args, options.UpdatedAtMax)
		query += " and updated_at <= ? "
	}
	if options.PublishedAtMax != "" {
		args = append(args, options.PublishedAtMax)
		query += " and published_at <= ? "
	}
	if options.PublishedAtMin != "" {
		args = append(args, options.PublishedAtMin)
		query += " and published_at >= ? "
	}
	if options.PublishedStatus == "published" {
		args = append(args, 1)
		query += " and status = ? "
	} else if options.PublishedStatus == "unpublished" {
		args = append(args, 2)
		query += " and status = ? "
	}

	if options.Title != "" {
		args = append(args, options.Title+"%")
		query += " and title like ? "
	}
	if len(options.Handlers) != 0 {
		if len(options.Handlers) == 1 {
			args = append(args, options.Handlers[0])
			query += " and handler in (?) "
		}
		if len(options.Handlers) > 1 {
			for k, id := range options.Handlers {

				if k == 0 {
					args = append(args, id)
					query += " and handler in ( ? "
				}
				if k > 0 {
					args = append(args, id)
					query += " , ? "
				}
				if k == len(options.Handlers)-1 {
					//args = append(args, idNum)
					query += " ) "
					//argsQuery = append(argsQuery, idNum)
				}
			}
		}

	}
	err := m.QueryRowNoCache(&resp, query, args...)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return 0, ErrNotFound
	default:
		return 0, err
	}
}

func (m *defaultSailShopProductModel) FindList(options ListOptions, shopId int64) (*[]SailShopProduct, error) {
	args := make([]interface{}, 0)
	args = append(args, shopId)

	argsQuery := make([]interface{}, 0)
	argsQuery = append(argsQuery, shopId)

	var count int64
	logx.Info(count)
	queryCount := fmt.Sprintf("select count(*) from %s where `shop_id` = ? and is_del = 0 ", m.table)
	query := fmt.Sprintf("select %s from %s where `shop_id` = ? and is_del = 0 ", sailShopProductRows, m.table)
	logx.Info("options:", options)
	if options.SinceId > 0 {
		queryCount += " and id > ? "
		args = append(args, options.SinceId)
		query += " and id > ? "
		argsQuery = append(argsQuery, options.SinceId)
	}
	if options.Ids != "" {
		idSet := strings.Split(options.Ids, ",")
		if len(idSet) == 1 {
			idNum, err := strconv.Atoi(idSet[0])
			if err != nil {
				logx.Error("err:", errors.New(" field ids in wrong format "))
				return nil, errors.New(" field ids in wrong format ")
			}
			queryCount += " and id in (?) "
			args = append(args, idNum)
			query += " and id in (?) "
			argsQuery = append(argsQuery, idNum)
		}
		if len(idSet) > 1 {
			for k, id := range idSet {
				idNum, err := strconv.Atoi(id)
				if err != nil {
					logx.Error("err:", errors.New(" field ids in wrong format "))
					return nil, errors.New(" field ids in wrong format ")
				}
				if k == 0 {
					queryCount += " and id in ( ? "
					args = append(args, idNum)
					query += " and id in ( ? "
					argsQuery = append(argsQuery, idNum)
				}
				if k > 0 {
					queryCount += " , ? "
					args = append(args, idNum)
					query += " , ? "
					argsQuery = append(argsQuery, idNum)
				}
				if k == len(idSet)-1 {
					queryCount += " ) "
					//args = append(args, idNum)
					query += " ) "
					//argsQuery = append(argsQuery, idNum)
				}
			}
		}

	}
	if options.CreatedAtMin != "" {
		queryCount += " and created_at >= ? "
		args = append(args, options.CreatedAtMin)
		query += " and created_at >= ? "
		argsQuery = append(argsQuery, options.CreatedAtMin)
	}
	if options.CreatedAtMax != "" {
		queryCount += " and created_at <= ? "
		args = append(args, options.CreatedAtMax)
		query += " and created_at <= ? "
		argsQuery = append(argsQuery, options.CreatedAtMax)
	}
	if options.UpdatedAtMin != "" {
		queryCount += " and updated_at >= ? "
		args = append(args, options.UpdatedAtMin)
		query += " and updated_at >= ? "
		argsQuery = append(argsQuery, options.UpdatedAtMin)
	}
	if options.UpdatedAtMax != "" {
		queryCount += " and updated_at <= ? "
		args = append(args, options.UpdatedAtMax)
		query += " and updated_at <= ? "
		argsQuery = append(argsQuery, options.UpdatedAtMax)
	}
	if options.PublishedAtMax != "" {
		queryCount += " and published_at <= ? "
		args = append(args, options.PublishedAtMax)
		query += " and published_at <= ? "
		argsQuery = append(argsQuery, options.PublishedAtMax)
	}
	if options.PublishedAtMin != "" {
		queryCount += " and published_at >= ? "
		args = append(args, options.PublishedAtMin)
		query += " and published_at >= ? "
		argsQuery = append(argsQuery, options.PublishedAtMin)
	}
	if options.PublishedStatus == "published" {
		queryCount += " and status = ? "
		args = append(args, 1)
		query += " and status = ? "
		argsQuery = append(argsQuery, 1)
	} else if options.PublishedStatus == "unpublished" {
		queryCount += " and status = ? "
		args = append(args, 2)
		query += " and status = ? "
		argsQuery = append(argsQuery, 2)
	}
	if options.Title != "" {
		queryCount += " and title like ? "
		args = append(args, options.Title+"%")
		query += " and title like ? "
		argsQuery = append(argsQuery, options.Title+"%")
	}
	if len(options.Handlers) != 0 {
		if len(options.Handlers) == 1 {
			queryCount += " and handler in (?) "
			args = append(args, options.Handlers[0])
			query += " and handler in (?) "
			argsQuery = append(argsQuery, options.Handlers[0])
		}
		if len(options.Handlers) > 1 {
			for k, id := range options.Handlers {

				if k == 0 {
					queryCount += " and handler in ( ? "
					args = append(args, id)
					query += " and handler in ( ? "
					argsQuery = append(argsQuery, id)
				}
				if k > 0 {
					queryCount += " , ? "
					args = append(args, id)
					query += " , ? "
					argsQuery = append(argsQuery, id)
				}
				if k == len(options.Handlers)-1 {
					queryCount += " ) "
					//args = append(args, idNum)
					query += " ) "
					//argsQuery = append(argsQuery, idNum)
				}
			}
		}

	}

	query += " order by created_at desc"

	if options.Limit > 0 {
		if options.IsNewVersion == true {
			if options.Limit >= 250 {
				options.Limit = 250
			}
		} else {
			if options.Limit >= 100 {
				options.Limit = 100
			}
		}

		query += "  limit ? "
		argsQuery = append(argsQuery, options.Limit)
	} else {

		query += "  limit ? "
		if options.IsNewVersion == true {
			argsQuery = append(argsQuery, 50)
		} else {
			argsQuery = append(argsQuery, 20)
		}

	}
	logx.Info(query, argsQuery)

	err := m.QueryRowNoCache(&count, queryCount, args...)
	if err != nil {
		logx.Error("查询产品数量出错:", err)

	}
	totalPages := int64(math.Ceil(float64(count) / float64(options.Limit)))
	if options.Page <= 0 {
		options.Page = 1
	}
	logx.Info("totalPage:", totalPages)
	logx.Info("options.page:", options.Page)
	query += " offset  ? "
	argsQuery = append(argsQuery, options.Limit*(options.Page-1))

	//sailShopProductIdKey := fmt.Sprintf("%s%v", cacheSailShopProductIdPrefix, id)

	var resp []SailShopProduct
	//err := m.QueryRow(&resp, sailShopProductIdKey, func(conn sqlx.SqlConn, v interface{}) error {
	//	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", sailShopProductRows, m.table)
	//	return conn.QueryRow(v, query, id, shopId)
	//})

	err = m.QueryRowsNoCache(&resp, query, argsQuery...)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		logx.Error("err:", err)
		return nil, ErrNotFound
	default:
		logx.Error("err:", err)
		return nil, err
	}
}

func (m *defaultSailShopProductModel) FindOneByShopIdHandler(shopId int64, handler string) (*SailShopProduct, error) {
	sailShopProductShopIdHandlerKey := fmt.Sprintf("%s%v%v", cacheSailShopProductShopIdHandlerPrefix, shopId, handler)
	var resp SailShopProduct
	err := m.QueryRowIndex(&resp, sailShopProductShopIdHandlerKey, m.formatPrimary, func(conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
		query := fmt.Sprintf("select %s from %s where `shop_id` = ? and `handler` = ? limit 1", sailShopProductRows, m.table)
		if err := conn.QueryRow(&resp, query, shopId, handler); err != nil {
			return nil, err
		}
		return resp.Id, nil
	}, m.queryPrimary)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultSailShopProductModel) Update(data InsertProductData, productId int64, redis2 *redis.Redis) (*RespImageData, error) {
	if data.Title != "" {
		lockKey := fmt.Sprintf("adminapi_lock_product_title_%d_%s", data.ShopId, data.Title)
		redisLock := redis.NewRedisLock(redis2, lockKey)
		redisLock.SetExpire(5)
		if ok, err := redisLock.Acquire(); !ok || err != nil {
			return nil, errors.New(" product title repeated, too many request ")
		}
		defer func() {
			recover()
			redisLock.Release()
		}()
	}
	var respImageData RespImageData
	respImageData.DefaultImageUrl = data.DefaultImage.FileKey
	handlerOrigin := m.generateOriginHandler(data.Title)
	data.OriginHandler = m.generateHandler(data.Title, data.ShopId)
	err := m.Transact(func(session sqlx.Session) error {
		respVariants, err := StmtQueryVariants(session, data.ShopId, productId)
		switch err {
		case nil:
		case sqlc.ErrNotFound:

		default:
			logx.Error(err)
			return err
		}
		imageSet := make([]int64, 0)
		if data.Images != nil {
			for _, i2 := range *data.Images {
				if i2.Id != 0 {
					imageSet = append(imageSet, i2.Id)
				} else {
					if i2.FileKey == "" {
						logx.Error("商品图片添加出错")
						return errors.New("internal server error")
					}
					tempImageItem := ImageItem{}
					tempImageItem.Url = i2.FileKey
					respImageData.Images = append(respImageData.Images, tempImageItem)
				}
			}
		}

		err = StmtUpdateProductImages(session, data.ShopId, productId, imageSet)
		if err != nil {
			logx.Error(err)
			return err
		}

		err = StmtUpdateProduct(session, data, productId, handlerOrigin, 0)
		if err != nil {
			logx.Error(err)
			return err
		}
		if data.Variants != nil {
			if len(*data.Variants) != 0 {
				oldVariants := map[int64]string{}
				for i, variantData := range *data.Variants {
					//data.Title = m.
					variantData.Sort = int64(i + 1)
					var variantId int64
					if variantData.Id != 0 {
						if _, ok := oldVariants[variantData.Id]; !ok {
							oldVariants[variantData.Id] = "1"
						}
						variantId = variantData.Id
						err = StmtUpdateVariant(session, variantData, productId, data.ShopId)
						if err != nil {
							logx.Error(err)
							return err
						}
					} else {
						result, err := StmtInsertVariant(session, variantData, productId, data.ShopId, redis2)
						if err != nil {
							logx.Error(err)
							return err
						}
						lastVariantId, err := result.LastInsertId()
						if err != nil {
							logx.Error(err)
							return err
						}
						variantId = lastVariantId
					}

					tempVariantItem := VariantItem{}
					tempVariantItem.VariantId = variantId
					tempVariantItem.Url = variantData.Image.FileKey
					respImageData.Variants = append(respImageData.Variants, tempVariantItem)
				}
				for _, variant := range *respVariants {
					if _, ok := oldVariants[variant.Id]; !ok {
						variantModel := defaultSailShopProductVariantModel{
							CachedConn: m.CachedConn,
							table:      "sail_shop_product_variant",
						}
						err := variantModel.DeleteVariant(session, data.ShopId, productId, variant.Id)
						if err != nil {
							logx.Error(err)
							return err
						}
					}
				}
			}
		}
		if data.BodyHtml != "shop_zero_string" {
			detailData := SailShopProductDetail{
				ProductId: productId,
				BodyHtml:  data.BodyHtml,
			}
			detailModel := defaultSailShopProductDetailModel{
				CachedConn: m.CachedConn,
				table:      "sail_shop_product_detail",
			}
			err := detailModel.Update(detailData, session)
			if err != nil {
				logx.Error("body_html新增失败:", err)
				return err
			}
		}

		//if data.BodyHtml != "" {
		//	mongoData := detail.ProductDetail{
		//		ProductID: strconv.Itoa(int(productId)),
		//		BodyHtml:  data.BodyHtml,
		//	}
		//	errCollection := collection.UpdateByProductId(context.Background(), &mongoData)
		//	if errCollection != nil {
		//		logx.Error("body_html新增失败:", err, ",product_id:", productId)
		//		return err
		//	}
		//}
		if data.Tags != nil {
			err := m.StmtDeleteProductItems(data.ShopId, productId, "sail_shop_tags_product", session)
			if err != nil {
				logx.Error(err)
			} else {
				if len(data.Tags) != 0 {
					for _, tag := range data.Tags {
						tagResult, err := StmtQueryTag(session, data.ShopId, tag)
						switch err {
						case nil:
							_, err = StmtInsertTagProduct(session, data.ShopId, tagResult.Id, productId)
							if err != nil {
								logx.Error(err)
								return err
							}
						case sqlc.ErrNotFound:
							tagNewResult, err := StmtInsertTags(session, data.ShopId, tag)
							if err != nil {
								logx.Error(err)
								return err
							}
							newTagId, err := tagNewResult.LastInsertId()
							if err != nil {
								logx.Error(err)
								return err
							}
							_, err = StmtInsertTagProduct(session, data.ShopId, newTagId, productId)
							if err != nil {
								logx.Error(err)
								return err
							}
						default:
							logx.Error(err)
							return err
						}
					}
				}
			}

		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	err = m.AfterSave(data.ShopId, productId, ES_SYNC_EVENT_UPDATE, redis2)
	if err != nil {
		logx.Error(err)
	}
	return &respImageData, nil
}

func (m *defaultSailShopProductModel) Delete(shopId, productId int64, redis2 *redis.Redis) error {
	productInfo, err := m.FindOne(shopId, WithId(productId))
	if err != nil {
		return err
	}

	if productInfo.Status == 1 {
		logx.Error("only unpublished product can be deleted")
		return errors.New("only unpublished product can be deleted")
	}
	err = m.Transact(func(session sqlx.Session) error {
		err := m.StmtDeleteProductItems(shopId, productId, m.table, session)
		if err != nil {

			logx.Error(err)

			return err
		}
		tagsModel := defaultSailShopTagsProductModel{
			CachedConn: m.CachedConn,
			table:      "sail_shop_tags_product",
		}
		tagsProductInfo, err := tagsModel.FindList(shopId, productId)
		switch err {
		case nil:
			errTags := m.StmtDeleteProductItems(shopId, productId, "sail_shop_tags_product", session)
			if errTags != nil {
				logx.Error(errTags)
			}
			type DeleteTag struct {
				ShopId int64 `json:"shop_id"`
				TagId  int64 `json:"tag_id"`
			}
			if len(*tagsProductInfo) != 0 {
				for _, product := range *tagsProductInfo {
					tempDeleteTag := DeleteTag{
						ShopId: shopId,
						TagId:  product.TagId,
					}
					tagProductByte, err := json.Marshal(tempDeleteTag)
					if err != nil {
						logx.Error(err)
					} else {
						_, err := redis2.Lpush(CLEAR_TAGS_QUEUE, string(tagProductByte))
						if err != nil {
							logx.Error(err)
						}
					}
				}
			}
		case sqlc.ErrNotFound:

		default:

		}

		return nil
	})
	if err != nil {
		logx.Error(err)
		return err
	}
	return nil
}

func (m *defaultSailShopProductModel) StmtDeleteProductItems(shopId, productId int64, tableName string, session sqlx.Session) error {
	args := make([]interface{}, 0)
	args = append(args, shopId, productId)
	idStr := "`product_id`"
	if tableName == "`sail_shop_product`" {
		idStr = "`id`"
	}
	query := fmt.Sprintf("update %s set is_del = 1 where `shop_id` = ? and %s = ? and `is_del` = 0  ", tableName, idStr)

	stmt, err := session.Prepare(query)
	if err != nil {
		logx.Error(err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(args...)
	if err != nil {
		logx.Error("删除商品出错：", err)
		return errors.New("internal server error")
	}
	return nil
}

func (m *defaultSailShopProductModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheSailShopProductIdPrefix, primary)
}

func (m *defaultSailShopProductModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", sailShopProductRows, m.table)
	return conn.QueryRow(v, query, primary)
}

func (m *defaultSailShopProductModel) generateOriginHandler(title string) string {
	title = strings.ToLower(title)
	filters := []string{",", "!", "/", " ", "?", "%", "#", "&", "=", "$", "*", "+", "[", "]", "(", ")", "{", "}", `"`, "'", ",", "，"}
	replace := "-"
	for _, filter := range filters {
		title = strings.Replace(title, filter, replace, -1)
	}
	reg := regexp.MustCompile(`/\s+/`)
	title = reg.ReplaceAllString(title, replace)
	return title
}

func (m *defaultSailShopProductModel) generateHandler(title string, shopId int64) string {
	origin := m.generateOriginHandler(title)
	handler := origin
	var count int64
	var productResp SailShopProduct
	query := fmt.Sprintf("select count(*) from %s where `shop_id` = ? and handler_origin = ? ", m.table)
	err := m.QueryRowNoCache(&count, query, shopId, origin)
	switch err {
	case nil:
		if count > 0 {
			handler = origin + "-" + strconv.Itoa(int(count))
		}
		productQuery := fmt.Sprintf("select %s from %s where `shop_id` = ? and handler = ? and is_del = 0 order by id desc limit 1 ", sailShopProductRows, m.table)
		err := m.QueryRowNoCache(&productResp, productQuery, shopId, handler)
		switch err {
		case nil:
			sourceStringSet := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
				"a", "b", "c", "d", "e", "f",
				"g", "h", "i", "j", "k", "l",
				"m", "n", "o", "p", "q", "r",
				"s", "t", "u", "v", "w", "x",
				"y", "z"}
			num := productResp.Id
			code := ""
			for i := num; i > 0; i = int64(math.Floor(float64(i) / float64(36))) {
				index := i % 36
				code = sourceStringSet[index] + code
			}
			if len(code) < 5 {
				for k := 5 - len(code); k > 0; k-- {
					code = "0" + code
				}
			}
			handler = origin + "-" + code

			return handler
		case sqlc.ErrNotFound:
			return handler
		default:
			logx.Error("查询商品信息失败：", err)
			return handler
		}

	case sqlc.ErrNotFound:
		return handler
	default:
		logx.Error("查询商品数量信息失败：", err)
		return handler
	}
}

func (m *defaultSailShopProductModel) FilterSeoDesc(seoDesc string) string {
	filters := []string{"&nbsp;", "  ", "\n"}
	replace := " "
	for i := 0; i < len(filters); i++ {
		seoDesc = strings.Replace(seoDesc, filters[i], replace, -1)
	}
	reg := regexp.MustCompile(`/[^\x{4e00}-\x{9fa5}^0-9^A-Z^a-z]+/u`)
	seoDesc = reg.ReplaceAllString(seoDesc, replace)
	return seoDesc
}

func (m *defaultSailShopProductModel) FilterSeoTitle(seoTitle string) string {
	filters := []string{"&nbsp;", " ", ",", "，", "--", "---", "----"}
	replace := "-"
	for i := 0; i < len(filters); i++ {
		seoTitle = strings.Replace(seoTitle, filters[i], replace, -1)
	}
	if string(seoTitle[len(seoTitle)-1]) == "-" {
		seoTitle = seoTitle[:len(seoTitle)-1]
	}
	return seoTitle
}

func (m *defaultSailShopProductModel) GetDefaultSeoDesc(bodyHtml string) string {
	bodyHtml = strip.StripTags(bodyHtml)
	seoDesc := m.FilterSeoDesc(bodyHtml)
	seoDesc = SubString(seoDesc, 0, 320)
	return seoDesc
}

func SubString(str string, begin, length int) string {
	//fmt.Println("Substring =", str)
	rs := []rune(str)
	lth := len(rs)
	//fmt.Printf("begin=%d, end=%d, lth=%d\n", begin, length, lth)
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length

	if end > lth {
		end = lth
	}
	//fmt.Printf("begin=%d, end=%d, lth=%d\n", begin, length, lth)
	return string(rs[begin:end])
}

func StmtInsertProduct(session sqlx.Session, data InsertProductData) (sql.Result, error) {
	var result sql.Result
	if data.Id == 0 {
		logx.Error("缺少id参数")
		return nil, errors.New("internal server error")
	}
	if data.ShopId == 0 {
		logx.Error("缺少shop_id参数")
		return nil, errors.New("shop_id 参数不能为空")
	}
	if data.Title == "" {
		logx.Error("缺少title参数")
		return nil, errors.New("title参数不能为空")
	}
	isUseStock := 0
	if data.IsUseStock == "Y" {
		isUseStock = 1
	}
	//h := md5.New()
	//h.Write([]byte(data.BodyHtml))
	//strMd5 := hex.EncodeToString(h.Sum(nil))

	fields := " id, shop_id, title,weight, weight_unit, price,compare_at_price, status, default_sku_code, seo_title, seo_desc, is_logistics, is_use_stock, handler, handler_origin, default_image_id, attribute, source, youtube_video_url, youtube_video_pos, product_stock"
	values := " ?, ?, ?, ?, ?, ?, ?, ?, ?, ? , ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?"
	if data.SoldOutPolicy == "Y" || data.SoldOutPolicy == "N" {
		fields += " , soldout_policy"
		values += " , ? "
	}
	status := 2
	requiresShipping := 0
	if data.RequireShipping == "Y" {
		requiresShipping = 1
	}
	if data.Status == "published" {
		fields += " , published_at "
		values += " , ? "
		status = 1
	}

	productInsertQuery := fmt.Sprintf("insert into %s (%s) values (%s) ", "sail_shop_product", fields, values)
	stmt, err := session.Prepare(productInsertQuery)
	if err != nil {
		logx.Error(err)
		return nil, err
	}
	defer stmt.Close()
	// 返回任何错误都会回滚事务
	args := make([]interface{}, 0)
	//" shop_id, title,weight, weight_unit, price,compare_at_price,soldout_policy, status, default_sku_code, seo_title, seo_desc, requires_shipping, is_use_stock, handler, handler_origin "

	args = append(args, data.Id, data.ShopId, data.Title, data.Weight, data.WeightUnit, data.Price, data.ComparePrice, status, data.Sku, data.SeoTitle, data.SeoDesc, requiresShipping, isUseStock, data.Handler, data.OriginHandler, 0, data.Attribute, "adminapi", data.YoutubeVideoUrl, data.YoutubeVideoPos, data.InventoryQuantity)
	if data.SoldOutPolicy == "Y" || data.SoldOutPolicy == "N" {
		args = append(args, data.SoldOutPolicy)
	}
	if data.Status == "published" {
		args = append(args, time.Now())
	}
	result, err = stmt.Exec(args...)
	if err != nil {
		logx.Errorf("insert product stmt exec: %s", err)
		return nil, err
	}
	return result, err
}

func StmtUpdateProduct(session sqlx.Session, data InsertProductData, productId int64, handlerOrigin string, defaultImageId int64) error {
	if data.ShopId == 0 {
		logx.Error("缺少shop_id参数")
		return errors.New("shop_id 参数不能为空")
	}
	updateDataArr := make([]string, 0)
	args := make([]interface{}, 0)
	if data.Title != "" {
		updateDataArr = append(updateDataArr, "title = ?")
		args = append(args, data.Title)
	}
	if data.Weight != 0 {
		updateDataArr = append(updateDataArr, "weight = ?")
		args = append(args, data.Weight)
	}
	if data.WeightUnit != "shop_zero_string" {
		updateDataArr = append(updateDataArr, "weight_unit = ?")
		args = append(args, data.WeightUnit)
	}
	if data.Price != 0 {
		updateDataArr = append(updateDataArr, "price = ?")
		args = append(args, data.Price)
	}
	if data.ComparePrice != 0 {
		updateDataArr = append(updateDataArr, "compare_at_price = ?")
		args = append(args, data.ComparePrice)
	}
	if data.InventoryQuantity > 0 {
		updateDataArr = append(updateDataArr, "product_stock = ?")
		args = append(args, data.InventoryQuantity)
	}
	if data.Status != "" {
		if data.Status == "published" {
			updateDataArr = append(updateDataArr, "status = ?")
			args = append(args, 1)
			updateDataArr = append(updateDataArr, "published_at = ?")
			args = append(args, time.Now())
		} else if data.Status == "unpublished" {
			updateDataArr = append(updateDataArr, "status = ?")
			args = append(args, 2)
			updateDataArr = append(updateDataArr, "published_at = ?")
			args = append(args, time.Time{})
		}
	}
	if data.Sku != "" {
		updateDataArr = append(updateDataArr, "default_sku_code = ?")
		args = append(args, data.Sku)
	}
	if data.SeoTitle != "" {
		updateDataArr = append(updateDataArr, "seo_title = ?")
		args = append(args, data.SeoTitle)
	}
	if data.SeoDesc != "" {
		updateDataArr = append(updateDataArr, "seo_desc = ?")
		args = append(args, data.SeoDesc)
	}
	if data.RequireShipping != "" {
		if data.RequireShipping == "Y" {
			updateDataArr = append(updateDataArr, "is_logistics = ?")
			args = append(args, 1)
		} else if data.RequireShipping == "N" {
			updateDataArr = append(updateDataArr, "is_logistics = ?")
			args = append(args, 0)
		}
	}
	if data.IsUseStock != "" {
		if data.IsUseStock == "Y" {
			updateDataArr = append(updateDataArr, "is_use_stock = ?")
			args = append(args, 1)
		} else if data.IsUseStock == "N" {
			updateDataArr = append(updateDataArr, "is_use_stock = ?")
			args = append(args, 0)
		}
	}
	//if data.BodyHtml != "shop_zero_string" {
	//	//h := md5.New()
	//	//h.Write([]byte(data.BodyHtml))
	//	//strMd5 := hex.EncodeToString(h.Sum(nil))
	//	updateDataArr = append(updateDataArr, "body_html = ?")
	//	args = append(args, data.BodyHtml)
	//}
	if data.Handler != "" {
		updateDataArr = append(updateDataArr, "handler = ?")
		args = append(args, data.Handler)
	}
	if data.YoutubeVideoPos != "-999999" {
		updateDataArr = append(updateDataArr, "youtube_video_pos = ?")
		args = append(args, data.YoutubeVideoPos)
	}
	if data.YoutubeVideoUrl != "shop_zero_string" {
		updateDataArr = append(updateDataArr, "youtube_video_url = ?")
		args = append(args, data.YoutubeVideoUrl)
	}
	updateDataArr = append(updateDataArr, "handler_origin = ?")
	args = append(args, handlerOrigin)
	updateDataArr = append(updateDataArr, "default_image_id = ?")
	args = append(args, defaultImageId)
	if data.Attribute != "" {
		updateDataArr = append(updateDataArr, "attribute = ?")
		args = append(args, data.Attribute)
	}
	updateDataArr = append(updateDataArr, "source = ?")
	args = append(args, "adminapi")
	if data.SoldOutPolicy == "Y" || data.SoldOutPolicy == "N" {
		updateDataArr = append(updateDataArr, "soldout_policy = ?")
		args = append(args, data.SoldOutPolicy)
	}

	args = append(args, data.ShopId, productId)

	updateStr := strings.Join(updateDataArr, ",")
	productInsertQuery := fmt.Sprintf("update %s set %s where `shop_id` = ? and `id` = ? and `is_del` = 0 ", "sail_shop_product", updateStr)
	logx.Info("productUpdateSql:", productInsertQuery)
	stmt, err := session.Prepare(productInsertQuery)
	if err != nil {
		logx.Error(err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(args...)
	if err != nil {
		logx.Errorf("update product stmt exec: %s", err)
		return err
	}
	return err
}

func StmtInsertVariant(session sqlx.Session, variantData VariantData, productId, shopId int64, redis2 *redis.Redis) (sql.Result, error) {
	isShow := 1
	variantId, err := GetAutoIncrementId(redis2)
	if err != nil {
		logx.Error("获取子商品自增id出错：", err)
		return nil, errors.New(" internal server error ")
	}
	variantData.Id = variantId
	variantFields := " `id`, `shop_id`, `product_id`, `price`,`compare_at_price`, `weight`, `weight_unit`, `requires_shipping`,`inventory_quantity`, `image_id`, `spec`, `sort`, `sku_code`, `title`,`options`, `is_checked` , `is_show` "
	variantValues := " ?, ?, ?, ?, ?, ?, ?, ?, ?, ? , ?, ?, ?, ?, ?, ?, ? "
	variantInsertQuery := fmt.Sprintf("insert into %s (%s) values (%s) ", "sail_shop_product_variant", variantFields, variantValues)
	var requireShipping int64
	if variantData.RequiresShipping == "Y" {
		requireShipping = 1
	} else {
		requireShipping = 0
	}
	if variantData.Spec == "" {
		variantData.Spec = "[]"
	}
	if variantData.Price == 0 {
		logx.Error("子商品缺少price参数")
		return nil, errors.New("variants item need price param ")
	}
	stmt1, err := session.Prepare(variantInsertQuery)
	if err != nil {
		logx.Error(err)
		return nil, err
	}
	defer stmt1.Close()
	//"shop_id, product_id, price,compare_at_price, weight, weight_unit, requires_shipping,inventory_quantity, image_id, spec, sort, sku_code, title,options, is_checked "
	resultVariant, err := stmt1.Exec(variantData.Id, shopId, productId, variantData.Price, variantData.ComparePrice, variantData.Weight, variantData.WeightUnit, requireShipping, variantData.InventoryQuantity, 0, variantData.Spec, variantData.Sort, variantData.SkuCode, variantData.Title, variantData.Options, variantData.IsChecked, isShow)
	if err != nil {
		logx.Error("新增子商品出错：", err)
		return nil, errors.New("internal server error")
	}
	return resultVariant, nil
}

func StmtUpdateVariant(session sqlx.Session, data VariantData, productId, shopId int64) error {
	//isShow := 1
	updateDataArr := make([]string, 0)
	args := make([]interface{}, 0)
	if data.Weight != 0 {
		updateDataArr = append(updateDataArr, "weight = ?")
		args = append(args, data.Weight)
	}
	if data.WeightUnit != "" {
		updateDataArr = append(updateDataArr, "weight_unit = ?")
		args = append(args, data.WeightUnit)
	}
	if data.Price != 0 {
		updateDataArr = append(updateDataArr, "price = ?")
		args = append(args, data.Price)
	}
	if data.ComparePrice != 0 {
		updateDataArr = append(updateDataArr, "compare_at_price = ?")
		args = append(args, data.ComparePrice)
	}
	if data.SkuCode != "" {
		updateDataArr = append(updateDataArr, "sku_code = ?")
		args = append(args, data.SkuCode)
	}

	if data.RequiresShipping != "" {
		if data.RequiresShipping == "Y" {
			updateDataArr = append(updateDataArr, "requires_shipping = ?")
			args = append(args, 1)
		} else if data.RequiresShipping == "N" {
			updateDataArr = append(updateDataArr, "requires_shipping = ?")
			args = append(args, 0)
		}
	}
	if data.Spec != "" {
		updateDataArr = append(updateDataArr, "spec = ?")
		args = append(args, data.Spec)
	}
	if data.Title != "" {
		updateDataArr = append(updateDataArr, "title = ?")
		args = append(args, data.Title)
	}
	if data.InventoryQuantity != 0 {
		updateDataArr = append(updateDataArr, "inventory_quantity = ?")
		args = append(args, data.InventoryQuantity)
	}
	if data.IsChecked != 0 {
		updateDataArr = append(updateDataArr, "is_checked = ?")
		args = append(args, data.IsChecked)
	}
	if data.Sort != 0 {
		updateDataArr = append(updateDataArr, "sort = ?")
		args = append(args, data.Sort)
	}
	if data.Options != "" {
		updateDataArr = append(updateDataArr, "options = ?")
		args = append(args, data.Options)
	}
	args = append(args, shopId, productId, data.Id)
	updateDataStr := strings.Join(updateDataArr, ",")

	//variantFields := "`shop_id`, `product_id`, `price`,`compare_at_price`, `weight`, `weight_unit`, `requires_shipping`,`inventory_quantity`, `image_id`, `spec`, `sort`, `sku_code`, `title`,`options`, `is_checked` , `is_show` "

	variantInsertQuery := fmt.Sprintf("update %s set %s where `shop_id` = ? and `product_id` = ? and `id` = ? and `is_del` = 0", "sail_shop_product_variant", updateDataStr)

	stmt1, err := session.Prepare(variantInsertQuery)
	if err != nil {
		logx.Error(err)
		return err
	}
	defer stmt1.Close()
	//"shop_id, product_id, price,compare_at_price, weight, weight_unit, requires_shipping,inventory_quantity, image_id, spec, sort, sku_code, title,options, is_checked "
	_, err = stmt1.Exec(args...)
	if err != nil {
		logx.Error("新增子商品出错：", err)
		return errors.New("internal server error")
	}
	return nil
}

func StmtUpdateTags(session sqlx.Session, shopId int64, tag string) (sql.Result, error) {
	fields := []string{"`shop_id`", "`name`"}
	args := make([]interface{}, 0)
	args = append(args, shopId, tag)
	result, err := StmtInsert(session, "sail_shop_tags", fields, args)
	if err != nil {
		logx.Error("新增tags失败：", err)
		return nil, err
	}
	return result, nil
}

func StmtUpdateTagProduct(session sqlx.Session, shopId, tagId, productId int64) (sql.Result, error) {
	fields := []string{"`shop_id`", "`tag_id`", "`product_id`"}
	args := make([]interface{}, 0)
	args = append(args, shopId, tagId, productId)
	result, err := StmtInsert(session, "sail_shop_tags_product", fields, args)
	if err != nil {
		logx.Error("新增tags失败：", err)
		return nil, err
	}
	return result, nil
}

func StmtUpdateProductImages(session sqlx.Session, shopId, productId int64, imageIds []int64) error {
	imageIdsSet := make([]string, 0)
	for _, id := range imageIds {
		imageIdsSet = append(imageIdsSet, strconv.Itoa(int(id)))
	}
	imageIdsStr := strings.Join(imageIdsSet, ",")
	query := fmt.Sprintf("update %s set `image_ids` = ? where `shop_id` = ? and `id` = ? and is_del = 0 ", "sail_shop_product")
	args := make([]interface{}, 0)
	args = append(args, imageIdsStr, shopId, productId)
	stmt1, err := session.Prepare(query)
	if err != nil {
		logx.Error(err)
		return err
	}
	defer stmt1.Close()
	_, err = stmt1.Exec(args...)
	if err != nil {
		logx.Error("新增tags失败：", err)
		return err
	}
	return nil
}

func StmtInsertTags(session sqlx.Session, shopId int64, tag string) (sql.Result, error) {
	fields := []string{"`shop_id`", "`name`"}
	args := make([]interface{}, 0)
	args = append(args, shopId, tag)
	result, err := StmtInsert(session, "sail_shop_tags", fields, args)
	if err != nil {
		logx.Error("新增tags失败：", err)
		return nil, err
	}
	return result, nil
}

func StmtDeleteTagsProduct(session sqlx.Session, shopId, productId int64) error {
	args := make([]interface{}, 0)
	args = append(args, shopId, productId)
	query := fmt.Sprintf("delete from %s where `shop_id` = ? and `product_id` = ? ", "sail_shop_tags_product")

	stmt1, err := session.Prepare(query)
	if err != nil {
		logx.Error(err)
		return err
	}
	defer stmt1.Close()
	//"shop_id, product_id, price,compare_at_price, weight, weight_unit, requires_shipping,inventory_quantity, image_id, spec, sort, sku_code, title,options, is_checked "
	_, err = stmt1.Exec(args...)
	if err != nil {
		logx.Error("删除商品标签出错：", err)
		return errors.New("internal server error")
	}
	return nil
}

func StmtInsertTagProduct(session sqlx.Session, shopId, tagId, productId int64) (sql.Result, error) {
	var respTagsProduct SailShopTagsProduct
	queryEx := fmt.Sprintf("select %s from %s where `shop_id` = ? and `tag_id`= ? and `product_id` = ? ", sailShopTagsProductRows, "sail_shop_tags_product")
	stmt, err := session.Prepare(queryEx)
	if err != nil {
		logx.Error(err)
		return nil, err
	}
	err = stmt.QueryRow(&respTagsProduct, shopId, tagId, productId)
	switch err {
	case nil:
		return nil, nil
	case sqlc.ErrNotFound:

	default:
		logx.Error(err)
		return nil, err
	}
	fields := []string{"`shop_id`", "`tag_id`", "`product_id`"}
	args := make([]interface{}, 0)
	args = append(args, shopId, tagId, productId)
	result, err := StmtInsert(session, "sail_shop_tags_product", fields, args)
	if err != nil {
		logx.Error("新增tags失败：", err)
		return nil, err
	}
	return result, nil
}

func StmtInsert(session sqlx.Session, tableName string, fields []string, args []interface{}) (sql.Result, error) {
	lengthFields := len(fields)
	lengthArgs := len(args)
	if lengthFields != lengthArgs {
		return nil, errors.New("args count and fields count do not match")
	}
	fieldsStr := strings.Join(fields, ",")
	valuesStr := ""
	for i := 0; i < lengthFields; i++ {
		if i == 0 {
			valuesStr += "?"
		} else {
			valuesStr += ", ?"
		}
	}
	insertSql := fmt.Sprintf("insert into %s (%s) values (%s) ", tableName, fieldsStr, valuesStr)
	stmt, err := session.Prepare(insertSql)
	if err != nil {
		logx.Error("添加标签失败：", err)
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(args...)
	return result, err
}
func StmtQueryTag(session sqlx.Session, shopId int64, tag string) (*SailShopTags, error) {
	var resp SailShopTags
	query := fmt.Sprintf("select %s from %s where `shop_id` = ? and `name` = ? and `is_del` = 0 ", sailShopTagsRows, "sail_shop_tags")
	stmt, err := session.Prepare(query)
	if err != nil {
		logx.Error("查询标签失败：", err)
		return nil, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(&resp, shopId, tag)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func StmtQueryVariants(session sqlx.Session, shopId, productId int64) (*[]SailShopProductVariant, error) {
	var resp []SailShopProductVariant
	query := fmt.Sprintf("select %s from %s where `shop_id` = ? and `product_id` = ? and `is_del` = 0 ", sailShopProductVariantRows, "sail_shop_product_variant")
	stmt, err := session.Prepare(query)
	if err != nil {
		logx.Error("查询子商品失败：", err)
		return nil, err
	}
	defer stmt.Close()
	err = stmt.QueryRows(&resp, shopId, productId)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultSailShopProductModel) AfterSave(shopId, productId int64, event string, redis2 *redis.Redis) error {
	temp := EsQueue{
		ShopId:    shopId,
		ProductId: productId,
		Event:     event,
	}
	tempByte, err := json.Marshal(temp)
	if err != nil {
		logx.Error("添加到es队列失败:", err)
		return errors.New("internal server error")
	}
	_, err = redis2.Lpush(PRODUCT_ES_SYNC_QUEUE, string(tempByte))
	if err != nil {
		logx.Error("添加到es队列失败:", err)
		return errors.New("internal server error")
	}
	return nil
}

type EsQueue struct {
	ShopId    int64  `json:"shop_id"`
	ProductId int64  `json:"product_id"`
	Event     string `json:"event"`
}

const PRODUCT_ES_SYNC_QUEUE = "es:sync:product:queue"
const CLEAR_TAGS_QUEUE = "sail:shop:product:tags:clear"

const ES_SYNC_EVENT_ADD = "add"
const ES_SYNC_EVENT_UPDATE = "update"
const AUTO_INCREMENT_KEY = "jh_auto_increment"
