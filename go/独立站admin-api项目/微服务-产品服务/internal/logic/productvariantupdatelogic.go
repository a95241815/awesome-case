package logic

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/model"
	"path"
	"reflect"
	"time"

	"gitlab.jhongnet.com/mall/rpc-product-server/internal/svc"
	"gitlab.jhongnet.com/mall/rpc-product-server/product"

	"github.com/tal-tech/go-zero/core/logx"
)

type ProductVariantUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProductVariantUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductVariantUpdateLogic {
	return &ProductVariantUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ProductVariantUpdateLogic) ProductVariantUpdate(in *product.ProductVariantUpdateRequest) (*product.ProductVariantUpdateResponse, error) {
	var url string
	var width int64
	var err error

	productInfo, err := l.svcCtx.ReadModel.FindOne(in.ShopId, model.WithId(in.ProductId))
	switch err {
	case nil:
		//if productInfo.CountSkus > 125-1 {
		//	l.Error(" max size of product variants is 125 ")
		//	return &product.ProductVariantUpdateResponse{}, errors.New(" max size of product variants is 125 ")
		//}
		if productInfo.IsUseStock == 1 && productInfo.SoldoutPolicy.String == "Y" && in.Variant.InventoryQuantity == 0 {
			l.Error("variant.inventory_quantity not set")
			return &product.ProductVariantUpdateResponse{}, errors.New("variant.inventory_quantity not set")
		}
	case sqlc.ErrNotFound:
		l.Error(" product record not found ")
		return &product.ProductVariantUpdateResponse{}, errors.New(" product record not found")
	default:
		l.Error("query product info error:", err)
		return &product.ProductVariantUpdateResponse{}, errors.New(" internal server error ")
	}
	variants, err := l.svcCtx.ReadVariantModel.FindList(in.ShopId, in.ProductId)
	switch err {
	case nil:
		specKeysMap := map[string]string{}
		for _, variant := range *variants {
			temp := map[string]string{}
			err := json.Unmarshal([]byte(variant.Spec.String), &temp)
			if err != nil {
				l.Error("unmarshal error:", err)
				return &product.ProductVariantUpdateResponse{}, errors.New(" internal server error")
			}
			for k, _ := range temp {
				specKeysMap[k] = "1"
			}
		}
		specMap := map[string]string{}
		err := json.Unmarshal([]byte(in.Variant.Spec), &specMap)
		if err != nil {
			l.Error("unmarshal error:", err)
			return &product.ProductVariantUpdateResponse{}, errors.New(" internal server error")
		}
		for _, i2 := range *variants {
			if i2.Id != in.VariantId {
				temp := map[string]string{}
				err := json.Unmarshal([]byte(i2.Spec.String), &temp)
				if err != nil {
					l.Error("unmarshal error:", err)
					return &product.ProductVariantUpdateResponse{}, errors.New(" internal server error")
				}
				if reflect.DeepEqual(temp, specMap) {
					l.Error("error:", err)
					return &product.ProductVariantUpdateResponse{}, errors.New("field spec repeated")
				}
			}
		}
		if len(specMap) != len(specKeysMap) {
			l.Error("error:", "field spec invalid")
			return &product.ProductVariantUpdateResponse{}, errors.New("field spec invalid")
		}
		for i, _ := range specMap {
			if _, ok := specKeysMap[i]; !ok {
				l.Error("error:", "field spec.name invalid")
				return &product.ProductVariantUpdateResponse{}, errors.New("field spec.name invalid")
			}
		}
	case sqlc.ErrNotFound:

	default:
		l.Error(err)
		return &product.ProductVariantUpdateResponse{}, err
	}

	if in.Variant.ImageUrl != "" {
		imageAddLogic := NewProductImageAddLogic(l.ctx, l.svcCtx)
		if l.svcCtx.ProjectENV == "xshoppy" {
			url, width, err = imageAddLogic.UploadImage(in.Variant.ImageUrl)
		} else {
			url, width, err = imageAddLogic.UploadImageEmy(in.Variant.ImageUrl)
			url = "uploader/" + path.Base(url)
		}
		if err != nil {
			l.Error("???????????????????????????", err)
			return &product.ProductVariantUpdateResponse{}, errors.New("internal server error")
		}
	}

	addReq := model.VariantData{
		Id:                in.VariantId,
		Price:             in.Variant.Price,
		ComparePrice:      in.Variant.ComparePrice,
		Weight:            in.Variant.Weight,
		WeightUnit:        in.Variant.WeightUnit,
		RequiresShipping:  in.Variant.RequiresShipping,
		InventoryQuantity: in.Variant.InventoryQuantity,
		Image: model.ImageData{
			FileKey:    url,
			ImageWidth: width,
		},
		Spec:      in.Variant.Spec,
		Sort:      in.Variant.Sort,
		SkuCode:   in.Variant.Sku,
		Title:     in.Variant.Title,
		Options:   in.Variant.Options,
		IsChecked: in.Variant.IsChecked,
	}

	//addData.Spec = sql.NullString(in.Variant.Spec)

	err = l.svcCtx.WriteVariantModel.Update(addReq, in.ShopId, in.ProductId, in.VariantId)
	if err != nil {
		l.Error(err)
		return &product.ProductVariantUpdateResponse{}, err
	}

	imageUrl := ""

	respVariant, err := l.svcCtx.WriteVariantModel.FindOne(in.ShopId, in.VariantId)
	switch err {
	case nil:
		if respVariant.ImageId != 0 {
			respImage, err := l.svcCtx.ReadImageModel.FindOne(respVariant.ImageId)
			switch err {
			case nil:
				imageUrl = respImage.FileKey
			case sqlc.ErrNotFound:
			default:
				l.Error(err)
				return nil, err
			}
		}
	case sqlc.ErrNotFound:
		l.Error("?????????????????????")
		return nil, errors.New("?????????????????????")
	default:
		return nil, err
	}
	inventoryPolicy := "N"
	if productInfo.IsUseStock == 1 {
		inventoryPolicy = "Y"
	}

	res := product.ProductVariantUpdateResponse{Variant: &product.ProductVariant{
		Id:                respVariant.Id,
		ProductId:         respVariant.ProductId,
		Sku:               respVariant.SkuCode,
		Title:             respVariant.Title,
		Price:             respVariant.Price,
		ComparePrice:      respVariant.CompareAtPrice,
		Spec:              respVariant.Spec.String,
		Weight:            respVariant.Weight,
		WeightUnit:        respVariant.WeightUnit,
		RequiresShipping:  respVariant.RequiresShipping,
		ImageId:           respVariant.ImageId,
		CreatedAt:         respVariant.CreatedAt.Local().Format(time.RFC3339),
		UpdatedAt:         respVariant.UpdatedAt.Local().Format(time.RFC3339),
		Sort:              respVariant.Sort,
		InventoryQuantity: respVariant.InventoryQuantity,
		InventoryPolicy:   inventoryPolicy,
		IsShow:            respVariant.IsShow,
		ImageUrl:          imageUrl,
	}}
	return &res, nil
}
