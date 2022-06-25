package logic

import (
	"context"
	"errors"
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/model"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/svc"
	"gitlab.jhongnet.com/mall/rpc-product-server/product"

	"github.com/tal-tech/go-zero/core/logx"
)

type CategoryProductAddLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCategoryProductAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CategoryProductAddLogic {
	return &CategoryProductAddLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CategoryProductAddLogic) CategoryProductAdd(in *product.CategoryProductAddRequest) (*product.CategoryProductAddResponse, error) {
	data := model.SailShopCategoryProduct{
		ShopId:      in.ShopId,
		CategoryId:  in.CategoryId,
		ProductId:   in.ProductId,
		ProductSort: in.Sort,
	}
	_, err := l.svcCtx.ReadCategoryModel.FindOne(in.ShopId, model.WithId(in.CategoryId))
	switch err {
	case nil:

	case sqlc.ErrNotFound:
		l.Error("category record not found")
		return nil, errors.New("category record not found")
	default:
		l.Error(err)
		return nil, err
	}
	_, err = l.svcCtx.ReadModel.FindOne(in.ShopId, model.WithId(in.ProductId))
	switch err {
	case nil:

	case sqlc.ErrNotFound:
		l.Error("product record not found")
		return nil, errors.New("product record not found")
	default:
		l.Error(err)
		return nil, err
	}

	_, err = l.svcCtx.ReadCategoryProductModel.FindOneByShopIdCategoryIdProductId(in.ShopId, in.CategoryId, in.ProductId)
	switch err {
	case nil:
		l.Error("product record already exists")
		return nil, errors.New("product record already exists")
	case sqlc.ErrNotFound:

	default:
		l.Error(err)
		return nil, errors.New("internal server error")
	}

	result, err := l.svcCtx.WriteCategoryProductModel.Insert(data)
	if err != nil {
		l.Error("insert category_product failed:", err)
		return &product.CategoryProductAddResponse{}, err
	}
	lastId, err := result.LastInsertId()
	if err != nil {
		l.Error("insert category_product failed:", err)
		return &product.CategoryProductAddResponse{}, err
	}
	resp, err := l.svcCtx.WriteCategoryProductModel.FindOne(lastId)
	switch err {
	case nil:
		categoryProduct := product.CategoryProduct{
			ProductId:  resp.ProductId,
			CategoryId: resp.CategoryId,
			Sort:       resp.ProductSort,
			Id:         resp.Id,
		}
		return &product.CategoryProductAddResponse{CategoryProduct: &categoryProduct}, nil
	case sqlc.ErrNotFound:
		l.Error("insert category product error:", err)
		return &product.CategoryProductAddResponse{}, errors.New("insert category product error")
	default:
		l.Error("insert category product error:", err)
		return &product.CategoryProductAddResponse{}, err
	}
}
