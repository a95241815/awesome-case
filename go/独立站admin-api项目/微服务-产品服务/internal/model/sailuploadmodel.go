package model

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/tal-tech/go-zero/core/logx"
	"strconv"
	"strings"
	"time"

	"github.com/tal-tech/go-zero/core/stores/cache"
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"github.com/tal-tech/go-zero/core/stores/sqlx"
	"github.com/tal-tech/go-zero/core/stringx"
	"github.com/tal-tech/go-zero/tools/goctl/model/sql/builderx"
)

var (
	sailUploadFieldNames          = builderx.RawFieldNames(&SailUpload{})
	sailUploadRows                = strings.Join(sailUploadFieldNames, ",")
	sailUploadRowsExpectAutoSet   = strings.Join(stringx.Remove(sailUploadFieldNames, "`id`", "`create_at`", "`update_at`"), ",")
	sailUploadRowsWithPlaceHolder = strings.Join(stringx.Remove(sailUploadFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"

	cacheSailUploadIdPrefix = "cache#sailUpload#id#"
)

type (
	SailUploadModel interface {
		InsertProductImage(data SailUpload, productId int64) (sql.Result, error)
		InsertDefaultImg(data SailUpload, productId int64) (sql.Result, error)
		InsertVariantImg(data SailUpload, productId int64, variantId int64) (sql.Result, error)
		FindOne(id int64) (*SailUpload, error)
		Count(shopId, productId int64) (int64, error)
		FindList(shopId int64) (*[]SailUpload, error)
		Update(data SailUpload) error
		Delete(id int64) error
	}

	defaultSailUploadModel struct {
		sqlc.CachedConn
		table string
	}

	SailUpload struct {
		ShopId     int64     `db:"shop_id"`    // 商店唯一ID
		FileKey2   string    `db:"file_key2"`  // Oss FileKey 900
		IsDel      int64     `db:"is_del"`     // 是否删除
		UpdatedAt  time.Time `db:"updated_at"` // 更新时间
		Id         int64     `db:"id"`
		FileMd5    string    `db:"file_md5"`    // 文件md5
		FileKey    string    `db:"file_key"`    // Oss FileKey
		FileKey1   string    `db:"file_key1"`   // Oss FileKey 750
		FileKey3   string    `db:"file_key3"`   // Oss FileKey 1080
		ImageWidth int64     `db:"image_width"` // 图片实际宽度
		CreatedAt  time.Time `db:"created_at"`  // 创建时间
	}
)

func NewSailUploadModel(conn sqlx.SqlConn, c cache.CacheConf) SailUploadModel {
	return &defaultSailUploadModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`sail_upload`",
	}
}

func (m *defaultSailUploadModel) Insert(data SailUpload, productId int64) (sql.Result, error) {
	var res sql.Result
	if data.ShopId == 0 {
		return nil, errors.New("缺少shop_id参数")
	}

	now := time.Now()
	data.UpdatedAt, data.CreatedAt = now, now
	//fieldsArr := []string{"`shop_id`", "``"}
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, sailUploadRowsExpectAutoSet)
	//ret, err := m.ExecNoCache(query, data.ShopId, data.FileKey2, data.IsDel, data.UpdatedAt, data.FileMd5, data.FileKey, data.FileKey1, data.FileKey3, data.ImageWidth, data.CreatedAt)

	//var insertsql = `insert into sail_upload(uid, username, mobilephone) values (?, ?, ?)`

	err := m.Transact(func(session sqlx.Session) error {
		stmt, err := session.Prepare(query)
		if err != nil {
			return err
		}
		defer stmt.Close()

		// 返回任何错误都会回滚事务
		res, err = stmt.Exec(data.ShopId, data.FileKey2, data.IsDel, data.UpdatedAt, data.FileMd5, data.FileKey, data.FileKey1, data.FileKey3, data.ImageWidth, data.CreatedAt)
		if err != nil {
			logx.Errorf("insert userinfo stmt exec: %s", err)
			return err
		}
		lastId, err := res.LastInsertId()

		productModel := &defaultSailShopProductModel{
			CachedConn: m.CachedConn,
			table:      "sail_shop_product",
		}
		productResp, err := productModel.FindOne(data.ShopId, WithId(productId))
		switch err {
		case nil:

		case sqlc.ErrNotFound:
			logx.Error()
			return ErrNotFound
		default:
			return err
		}
		ids := strings.Split(productResp.ImageIds, ",")
		flag := false
		idStr := strconv.Itoa(int(lastId))
		for _, id := range ids {
			if id == idStr {
				flag = true
			}
		}
		if !flag {
			var finalIds string
			if productResp.ImageIds == "" {
				finalIds = idStr
			} else {
				finalIds = productResp.ImageIds + "," + idStr
			}
			var updateSql = "update sail_shop_product set image_ids = ? where id = ?"
			stmt2, err := session.Prepare(updateSql)
			if err != nil {
				return err
			}
			defer stmt2.Close()
			_, err = stmt2.Exec(finalIds, productId)
			if err != nil {
				logx.Error(err)
				return err
			}
			sailShopProductIdKey := fmt.Sprintf("%s%v", cacheSailShopProductIdPrefix, productId)
			err = m.CachedConn.DelCache(sailShopProductIdKey)
			if err != nil {
				logx.Error(err)
				return err
			}
		}

		return nil
	})
	return res, err
}

func (m *defaultSailUploadModel) InsertProductImage(data SailUpload, productId int64) (sql.Result, error) {
	var res sql.Result
	if data.ShopId == 0 {
		return nil, errors.New("缺少shop_id参数")
	}
	now := time.Now()
	data.UpdatedAt, data.CreatedAt = now, now
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, sailUploadRowsExpectAutoSet)
	//ret, err := m.ExecNoCache(query, data.ShopId, data.FileKey2, data.IsDel, data.UpdatedAt, data.FileMd5, data.FileKey, data.FileKey1, data.FileKey3, data.ImageWidth, data.CreatedAt)

	//var insertsql = `insert into sail_upload(uid, username, mobilephone) values (?, ?, ?)`

	err := m.Transact(func(session sqlx.Session) error {
		stmt, err := session.Prepare(query)
		if err != nil {
			return err
		}
		defer stmt.Close()

		// 返回任何错误都会回滚事务
		res, err = stmt.Exec(data.ShopId, data.FileKey2, data.IsDel, data.UpdatedAt, data.FileMd5, data.FileKey, data.FileKey1, data.FileKey3, data.ImageWidth, data.CreatedAt)
		if err != nil {
			logx.Errorf("insert userinfo stmt exec: %s", err)
			return err
		}
		lastId, err := res.LastInsertId()

		productModel := &defaultSailShopProductModel{
			CachedConn: m.CachedConn,
			table:      "sail_shop_product",
		}
		productResp, err := productModel.FindOne(data.ShopId, WithId(productId))
		switch err {
		case nil:

		case sqlc.ErrNotFound:
			logx.Error()
			return ErrNotFound
		default:
			return err
		}
		ids := strings.Split(productResp.ImageIds, ",")
		flag := false
		idStr := strconv.Itoa(int(lastId))
		for _, id := range ids {
			if id == idStr {
				flag = true
			}
		}
		if !flag {
			var finalIds string
			if productResp.ImageIds == "" {
				finalIds = idStr
			} else {
				finalIds = productResp.ImageIds + "," + idStr
			}
			var updateSql = "update sail_shop_product set image_ids = ? where id = ?"
			stmt2, err := session.Prepare(updateSql)
			if err != nil {
				return err
			}
			defer stmt2.Close()
			_, err = stmt2.Exec(finalIds, productId)
			if err != nil {
				logx.Error(err)
				return err
			}
			sailShopProductIdKey := fmt.Sprintf("%s%v", cacheSailShopProductIdPrefix, productId)
			err = m.CachedConn.DelCache(sailShopProductIdKey)
			if err != nil {
				logx.Error(err)
				return err
			}
		}

		return nil
	})
	return res, err
}

func (m *defaultSailUploadModel) InsertDefaultImg(data SailUpload, productId int64) (sql.Result, error) {
	var res sql.Result
	if data.ShopId == 0 {
		return nil, errors.New("缺少shop_id参数")
	}
	now := time.Now()
	data.UpdatedAt, data.CreatedAt = now, now
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, sailUploadRowsExpectAutoSet)
	//ret, err := m.ExecNoCache(query, data.ShopId, data.FileKey2, data.IsDel, data.UpdatedAt, data.FileMd5, data.FileKey, data.FileKey1, data.FileKey3, data.ImageWidth, data.CreatedAt)

	//var insertsql = `insert into sail_upload(uid, username, mobilephone) values (?, ?, ?)`

	err := m.Transact(func(session sqlx.Session) error {
		stmt, err := session.Prepare(query)
		if err != nil {
			return err
		}
		defer stmt.Close()

		// 返回任何错误都会回滚事务
		res, err = stmt.Exec(data.ShopId, data.FileKey2, data.IsDel, data.UpdatedAt, data.FileMd5, data.FileKey, data.FileKey1, data.FileKey3, data.ImageWidth, data.CreatedAt)
		if err != nil {
			logx.Errorf("insert userinfo stmt exec: %s", err)
			return err
		}
		lastId, err := res.LastInsertId()

		productModel := &defaultSailShopProductModel{
			CachedConn: m.CachedConn,
			table:      "sail_shop_product",
		}
		productResp, err := productModel.FindOne(data.ShopId, WithId(productId))
		switch err {
		case nil:

		case sqlc.ErrNotFound:
			logx.Error()
			return ErrNotFound
		default:
			return err
		}
		ids := strings.Split(productResp.ImageIds, ",")
		flag := false
		idStr := strconv.Itoa(int(lastId))
		for _, id := range ids {
			if id == idStr {
				flag = true
			}
		}
		if !flag {
			var finalIds string
			if productResp.ImageIds == "" {
				finalIds = idStr
			} else {
				finalIds = productResp.ImageIds + "," + idStr
			}

			var updateSql = "update sail_shop_product set image_ids = ?, default_image_id = ? where id = ?"
			stmt2, err := session.Prepare(updateSql)
			if err != nil {
				return err
			}
			defer stmt2.Close()
			_, err = stmt2.Exec(finalIds, lastId, productId)
			if err != nil {
				logx.Error(err)
				return err
			}
			sailShopProductIdKey := fmt.Sprintf("%s%v", cacheSailShopProductIdPrefix, productId)
			err = m.CachedConn.DelCache(sailShopProductIdKey)
			if err != nil {
				logx.Error(err)
				return err
			}
		}

		return nil
	})
	return res, err
}

func (m *defaultSailUploadModel) InsertVariantImg(data SailUpload, productId int64, variantId int64) (sql.Result, error) {
	var res sql.Result
	if data.ShopId == 0 {
		return nil, errors.New("缺少shop_id参数")
	}
	now := time.Now()
	data.UpdatedAt, data.CreatedAt = now, now
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, sailUploadRowsExpectAutoSet)
	//ret, err := m.ExecNoCache(query, data.ShopId, data.FileKey2, data.IsDel, data.UpdatedAt, data.FileMd5, data.FileKey, data.FileKey1, data.FileKey3, data.ImageWidth, data.CreatedAt)

	//var insertsql = `insert into sail_upload(uid, username, mobilephone) values (?, ?, ?)`

	err := m.Transact(func(session sqlx.Session) error {
		stmt, err := session.Prepare(query)
		if err != nil {
			return err
		}
		defer stmt.Close()

		// 返回任何错误都会回滚事务
		res, err = stmt.Exec(data.ShopId, data.FileKey2, data.IsDel, data.UpdatedAt, data.FileMd5, data.FileKey, data.FileKey1, data.FileKey3, data.ImageWidth, data.CreatedAt)
		if err != nil {
			logx.Errorf("insert userinfo stmt exec: %s", err)
			return err
		}
		lastId, err := res.LastInsertId()

		productModel := &defaultSailShopProductModel{
			CachedConn: m.CachedConn,
			table:      "sail_shop_product",
		}
		productResp, err := productModel.FindOne(data.ShopId, WithId(productId))
		switch err {
		case nil:

		case sqlc.ErrNotFound:
			logx.Error()
			return ErrNotFound
		default:
			return err
		}
		ids := strings.Split(productResp.ImageIds, ",")
		flag := false
		idStr := strconv.Itoa(int(lastId))
		for _, id := range ids {
			if id == idStr {
				flag = true
			}
		}
		if !flag {
			var finalIds string
			if productResp.ImageIds == "" {
				finalIds = idStr
			} else {
				finalIds = productResp.ImageIds + "," + idStr
			}
			var updateSql = "update sail_shop_product set image_ids = ? where id = ?"
			stmt2, err := session.Prepare(updateSql)
			if err != nil {
				return err
			}
			defer stmt2.Close()
			_, err = stmt2.Exec(finalIds, productId)
			if err != nil {
				logx.Error(err)
				return err
			}
			sailShopProductIdKey := fmt.Sprintf("%s%v", cacheSailShopProductIdPrefix, productId)
			err = m.CachedConn.DelCache(sailShopProductIdKey)
			if err != nil {
				logx.Error(err)
				return err
			}
		}

		var updateVariantSql = "update sail_shop_product_variant set image_id = ?, is_set_default_img = 1 where id = ?"
		stmt3, err := session.Prepare(updateVariantSql)
		if err != nil {
			logx.Error(err)
			return err
		}
		defer stmt3.Close()
		_, err = stmt3.Exec(lastId, variantId)
		if err != nil {
			logx.Error(err)
			return err
		}

		return nil
	})
	return res, err
}

func (m *defaultSailUploadModel) FindOne(id int64) (*SailUpload, error) {
	//sailUploadIdKey := fmt.Sprintf("%s%v", cacheSailUploadIdPrefix, id)
	var resp SailUpload
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", sailUploadRows, m.table)
	//err := m.QueryRow(&resp, sailUploadIdKey, func(conn sqlx.SqlConn, v interface{}) error {
	//	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", sailUploadRows, m.table)
	//	return conn.QueryRow(v, query, id)
	//})
	err := m.QueryRowNoCache(&resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultSailUploadModel) FindList(shopId int64) (*[]SailUpload, error) {
	//sailUploadIdKey := fmt.Sprintf("%s%v", cacheSailUploadIdPrefix, id)
	var resp []SailUpload
	query := fmt.Sprintf("select %s from %s where `shop_id` = ? limit 10", sailUploadRows, m.table)
	//err := m.QueryRow(&resp, sailUploadIdKey, func(conn sqlx.SqlConn, v interface{}) error {
	//	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", sailUploadRows, m.table)
	//	return conn.QueryRow(v, query, id)
	//})
	err := m.QueryRowsNoCache(&resp, query, shopId)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultSailUploadModel) Count(shopId, productId int64) (int64, error) {
	var resp int64
	//err := m.QueryRow(&resp, sailUploadIdKey, func(conn sqlx.SqlConn, v interface{}) error {
	//	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", sailUploadRows, m.table)
	//	return conn.QueryRow(v, query, id)
	//})
	query := fmt.Sprintf("select count(*) from %s where `shop_id` = ? and `product_id` = ? ", m.table)
	err := m.QueryRowNoCache(&resp, query, shopId, productId)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return 0, ErrNotFound
	default:
		return 0, err
	}
}

func (m *defaultSailUploadModel) Update(data SailUpload) error {
	sailUploadIdKey := fmt.Sprintf("%s%v", cacheSailUploadIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, sailUploadRowsWithPlaceHolder)
		return conn.Exec(query, data.ShopId, data.FileKey2, data.IsDel, data.UpdatedAt, data.FileMd5, data.FileKey, data.FileKey1, data.FileKey3, data.ImageWidth, data.CreatedAt, data.Id)
	}, sailUploadIdKey)
	return err
}

func (m *defaultSailUploadModel) Delete(id int64) error {

	sailUploadIdKey := fmt.Sprintf("%s%v", cacheSailUploadIdPrefix, id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.Exec(query, id)
	}, sailUploadIdKey)
	return err
}

func (m *defaultSailUploadModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheSailUploadIdPrefix, primary)
}

func (m *defaultSailUploadModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", sailUploadRows, m.table)
	return conn.QueryRow(v, query, primary)
}
