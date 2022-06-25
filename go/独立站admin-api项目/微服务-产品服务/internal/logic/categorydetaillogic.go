package logic

import (
	"context"
	"errors"
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/model"
	"time"

	"gitlab.jhongnet.com/mall/rpc-product-server/internal/svc"
	"gitlab.jhongnet.com/mall/rpc-product-server/product"

	"github.com/tal-tech/go-zero/core/logx"
)

type CategoryDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCategoryDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CategoryDetailLogic {
	return &CategoryDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CategoryDetailLogic) CategoryDetail(in *product.CategoryDetailRequest) (*product.CategoryDetailResponse, error) {
	options := make([]model.HandlerOption, 0)
	if in.CategoryId != 0 {
		options = append(options, model.WithId(in.CategoryId))
	}
	if in.Handler != "" {
		options = append(options, model.WithHandler(in.Handler))
	}
	respItem, err := l.svcCtx.ReadCategoryModel.FindOne(in.ShopId, options...)
	switch err {
	case nil:
		productIds := ""

		count, err := l.svcCtx.ReadCategoryProductModel.Count(in.ShopId, respItem.Id)
		switch err {
		case nil:

		case sqlc.ErrNotFound:

		default:
			return &product.CategoryDetailResponse{}, err
		}
		item := product.Category{
			Id:              respItem.Id,
			Handler:         respItem.Handler,
			SeoTitle:        respItem.SeoTitle,
			SeoDesc:         respItem.SeoDesc,
			Title:           respItem.Title,
			BodyHtml:        respItem.BodyHtml,
			ProductIds:      productIds,
			CountProducts:   count,
			CreatedAt:       respItem.CreatedAt.Local().Format(time.RFC3339),
			UpdatedAt:       respItem.UpdatedAt.Local().Format(time.RFC3339),
			ProductSortType: respItem.ProductSortType,
		}

		respImage, err := l.svcCtx.ReadImageModel.FindOne(respItem.ImageId)
		switch err {
		case nil:
			item.ImageUrl = respImage.FileKey
		case sqlc.ErrNotFound:
			l.Error("分类图片不存在")
		default:
			l.Error("获取分类图片出错：", err)
		}

		return &product.CategoryDetailResponse{Category: &item}, nil
	case sqlc.ErrNotFound:
		l.Error("查询分类为空")
		return &product.CategoryDetailResponse{}, errors.New("category record not found")
	default:
		l.Error("查询产品图片出错：", err)
		return &product.CategoryDetailResponse{}, err
	}
}
