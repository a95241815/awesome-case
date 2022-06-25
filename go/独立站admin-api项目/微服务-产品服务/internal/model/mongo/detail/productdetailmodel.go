package detail

import (
	"context"
	"github.com/tal-tech/go-zero/core/logx"

	"github.com/globalsign/mgo/bson"
	cachec "github.com/tal-tech/go-zero/core/stores/cache"
	"github.com/tal-tech/go-zero/core/stores/mongoc"
)

type ProductDetailModel interface {
	Insert(ctx context.Context, data *ProductDetail) error
	FindOneByProductId(ctx context.Context, productId string) (*ProductDetail, error)
	UpdateByProductId(ctx context.Context, data *ProductDetail) error
	Delete(ctx context.Context, id string) error
}

type defaultProductDetailModel struct {
	*mongoc.Model
}

func NewProductDetailModel(url, collection string, c cachec.CacheConf) ProductDetailModel {
	return &defaultProductDetailModel{
		Model: mongoc.MustNewModel(url, collection, c),
	}
}

func (m *defaultProductDetailModel) Insert(ctx context.Context, data *ProductDetail) error {
	if !data.ID.Valid() {
		data.ID = bson.NewObjectId()
	}

	session, err := m.TakeSession()
	if err != nil {
		return err
	}

	defer m.PutSession(session)
	return m.GetCollection(session).Insert(data)
}

func (m *defaultProductDetailModel) FindOneByProductId(ctx context.Context, productId string) (*ProductDetail, error) {
	session, err := m.TakeSession()
	if err != nil {
		return nil, err
	}

	defer m.PutSession(session)
	var data ProductDetail

	err = m.GetCollection(session).FindOneNoCache(&data, bson.M{"product_id": productId})
	switch err {
	case nil:
		return &data, nil
	case mongoc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultProductDetailModel) UpdateByProductId(ctx context.Context, data *ProductDetail) error {
	session, err := m.TakeSession()
	if err != nil {
		logx.Error(err)
		return err
	}

	defer m.PutSession(session)

	return m.GetCollection(session).UpdateNoCache(bson.M{"product_id": data.ProductID}, bson.M{"$set": bson.M{"body_html": data.BodyHtml}})
}

func (m *defaultProductDetailModel) Delete(ctx context.Context, id string) error {
	session, err := m.TakeSession()
	if err != nil {
		return err
	}

	defer m.PutSession(session)

	return m.GetCollection(session).RemoveIdNoCache(bson.ObjectIdHex(id))
}
