package logic

import (
	"context"
	"errors"
	"github.com/tal-tech/go-zero/core/mr"
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/model"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/svc"
	"gitlab.jhongnet.com/mall/rpc-product-server/product"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tal-tech/go-zero/core/logx"
)

type ProductDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductDetailLogic {
	return &ProductDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ProductDetailLogic) ProductDetail(in *product.ProductDetailRequest) (*product.ProductDetailResponse, error) {
	l.Info("ProductDetail Called")
	if in.Fields == "" {
		in.Fields = "id,title,price,compare_price,weight,weight_unit,shop_url,product_url,body_html,seo_title,seo_desc,published_at,is_use_stock,soldout_policy,handler,status,created_at,updated_at,images,default_image,variants"
	}
	options := make([]model.HandlerOption, 0)
	l.Info("arguments:", in.ProductHandler, in.ProductId)
	if in.ProductHandler == "" && in.ProductId == 0 {
		l.Error("缺少查询参数")
		return nil, errors.New("product_id or product_handler is missing")
	}
	if in.ShopId == 0 {
		l.Error("缺少店铺id参数")
		return nil, errors.New("shop_id is missing")
	}
	if in.ProductId != 0 {
		options = append(options, model.WithId(in.ProductId))
	}
	if in.ProductHandler != "" {
		options = append(options, model.WithHandler(in.ProductHandler))
	}

	resp, err := l.svcCtx.ReadModel.FindOne(in.ShopId, options...)
	switch err {
	case nil:

	case sqlc.ErrNotFound:
		l.Error("商品记录不存在")
		return nil, errors.New("product record not found")
	default:
		l.Error(err)
		return nil, errors.New("internal server error")
	}

	var productDetail product.Product
	var defaultImage product.ProductImage
	var images []*product.ProductImage
	var variants []*product.ProductVariant
	var bodyHtml string

	err = mr.Finish(func() error {
		if strings.Contains(in.Fields, "body_html") {
			respDetail, err := l.svcCtx.ReadProductDetailModel.FindOne(in.ProductId)
			switch err {
			case nil:
				bodyHtml = respDetail.BodyHtml
			case sqlc.ErrNotFound:
				logx.Error("body_html not found")
			default:
				l.Error(err)
			}
		}

		return nil
	}, func() (err error) {
		if strings.Contains(in.Fields, "default_image") {
			if resp.DefaultImageId != 0 {
				respDefaultImage, err := l.svcCtx.ReadImageModel.FindOne(resp.DefaultImageId)
				switch err {
				case nil:
					defaultImage = product.ProductImage{
						Id:         respDefaultImage.Id,
						ProductId:  resp.Id,
						Width:      respDefaultImage.ImageWidth,
						Url:        respDefaultImage.FileKey,
						VariantIds: nil,
						FileKey:    respDefaultImage.FileKey,
						CreatedAt:  respDefaultImage.CreatedAt.Local().Format(time.RFC3339),
						UpdatedAt:  respDefaultImage.UpdatedAt.Local().Format(time.RFC3339),
					}
				case sqlc.ErrNotFound:
					l.Error("图片记录不存在")
					//return sqlc.ErrNotFound
				default:
					l.Error(err)
					return errors.New("internal server error")
				}
			}
		}

		return nil
	}, func() (err error) {
		if strings.Contains(in.Fields, "variants") {
			variantResp, err := l.svcCtx.ReadVariantModel.FindList(in.ShopId, resp.Id)
			switch err {
			case nil:
				if variantResp != nil {
					//var inventoryQuantity string
					for _, valVar := range *variantResp {
						inventoryPolicy := "N"
						if resp.IsUseStock == 1 {
							inventoryPolicy = "Y"
						}
						imageUrl := ""
						if valVar.ImageId != 0 {
							respImage, err := l.svcCtx.ReadImageModel.FindOne(valVar.ImageId)
							switch err {
							case nil:
								imageUrl = respImage.FileKey
							case sqlc.ErrNotFound:
							default:

							}
						}
						tempVariant := &product.ProductVariant{
							Id:                valVar.Id,
							Sku:               valVar.SkuCode,
							Title:             valVar.Title,
							Price:             valVar.Price,
							ComparePrice:      valVar.CompareAtPrice,
							Spec:              valVar.Spec.String,
							Weight:            valVar.Weight,
							WeightUnit:        valVar.WeightUnit,
							RequiresShipping:  valVar.RequiresShipping,
							ImageId:           valVar.ImageId,
							CreatedAt:         valVar.CreatedAt.Local().Format(time.RFC3339),
							UpdatedAt:         valVar.UpdatedAt.Local().Format(time.RFC3339),
							Sort:              valVar.Sort,
							InventoryQuantity: valVar.InventoryQuantity,
							InventoryPolicy:   inventoryPolicy,
							IsShow:            valVar.IsShow,
							ImageUrl:          imageUrl,
						}
						variants = append(variants, tempVariant)
					}
				}
			case sqlc.ErrNotFound:
				l.Error("子商品记录不存在")
				//return sqlc.ErrNotFound
			default:
				l.Error(err)
				return errors.New("internal server error")
			}
		}

		return nil
	}, func() (err error) {
		if strings.Contains(in.Fields, "images") {
			imageSet := strings.Split(resp.ImageIds, ",")
			imageMap := map[string]int{}
			for k, v := range imageSet {
				imageMap[v] = k
			}
			resultImage, err := mr.MapReduce(func(source chan<- interface{}) {
				for _, valPro := range imageSet {
					source <- valPro
				}
			}, func(item interface{}, writer mr.Writer, cancel func(error)) {
				uid := item.(string)
				convId, err := strconv.Atoi(uid)
				if err != nil {
					l.Error(err)
					return
				}
				if convId != 0 {
					respDefaultImage, err := l.svcCtx.ReadImageModel.FindOne(int64(convId))
					switch err {
					case nil:
						tempImage := product.ProductImage{
							Id:         respDefaultImage.Id,
							ProductId:  resp.Id,
							Sort:       0,
							Width:      respDefaultImage.ImageWidth,
							Url:        respDefaultImage.FileKey,
							VariantIds: nil,
							FileKey:    respDefaultImage.FileKey,
							CreatedAt:  respDefaultImage.CreatedAt.Local().Format(time.RFC3339),
							UpdatedAt:  respDefaultImage.UpdatedAt.Local().Format(time.RFC3339),
						}
						writer.Write(&tempImage)
					case sqlc.ErrNotFound:
						l.Error("图片记录不存在")
						//return sqlc.ErrNotFound
					default:
						l.Error(err)
						//return errors.New("internal server error")
					}
				}
			}, func(pipe <-chan interface{}, writer mr.Writer, cancel func(error)) {
				var uids []*product.ProductImage
				for p := range pipe {
					uids = append(uids, p.(*product.ProductImage))
				}
				writer.Write(uids)
			})
			if err != nil {
				logx.Error(err)
				return err
			}

			images = resultImage.([]*product.ProductImage)
			sort.Slice(images, func(i, j int) bool {
				return imageMap[strconv.Itoa(int(images[i].Id))] < imageMap[strconv.Itoa(int(images[j].Id))]
			})
		}

		return nil
	})

	if err != nil {
		l.Error(err)
		return nil, errors.New("internal server error")
	}
	var youtubeVideoPos int
	if resp.YoutubeVideoPos != "" {
		youtubeVideoPos, err = strconv.Atoi(resp.YoutubeVideoPos)
		if err != nil {
			l.Error(err)
			return nil, errors.New("internal server error")
		}
	}

	productDetail = product.Product{
		ProductTitle:    resp.Title,
		ProductId:       resp.Id,
		Price:           resp.Price,
		CompareAtPrice:  resp.CompareAtPrice,
		Weight:          resp.Weight,
		WeightUnit:      resp.WeightUnit,
		ShopUrl:         "",
		PreviewUrl:      "/products/" + resp.Handler,
		BodyHtml:        bodyHtml,
		SeoTitle:        resp.SeoTitle,
		SeoDesc:         resp.SeoDesc,
		PublishedAt:     resp.PublishedAt.Local().Format(time.RFC3339),
		IsUseStock:      resp.IsUseStock,
		SoldoutPolicy:   resp.SoldoutPolicy.String,
		Handle:          resp.Handler,
		Status:          resp.Status,
		CreatedAt:       resp.CreatedAt.Local().Format(time.RFC3339),
		UpdatedAt:       resp.UpdatedAt.Local().Format(time.RFC3339),
		Images:          images,
		DefaultImage:    &defaultImage,
		Variants:        variants,
		SubTitle:        resp.SubTitle,
		Attribute:       resp.Attribute.String,
		Comments:        resp.Comments,
		IsShowComment:   resp.IsShowComment,
		Scores:          resp.Scores,
		CountSkus:       resp.CountSkus,
		IsRead:          resp.IsRead,
		CountSales:      resp.CountSales,
		YoutubeVideoPos: int64(youtubeVideoPos),
		YoutubeVideoUrl: resp.YoutubeVideoUrl,
	}

	return &product.ProductDetailResponse{Product: &productDetail}, nil
}

func (l *ProductDetailLogic) GetImage() error {
	return errors.New("123")
}

func (l *ProductDetailLogic) GetFunc(index, imageId, productId int64, imageChan chan *product.ProductImage) func() error {
	return func() error {
		var err error
		if imageId != 0 {
			respDefaultImage, err := l.svcCtx.ReadImageModel.FindOne(imageId)
			switch err {
			case nil:
				tempImage := &product.ProductImage{
					Id:         respDefaultImage.Id,
					ProductId:  productId,
					Sort:       index,
					Width:      respDefaultImage.ImageWidth,
					Url:        respDefaultImage.FileKey,
					VariantIds: nil,
					FileKey:    respDefaultImage.FileKey,
					CreatedAt:  respDefaultImage.CreatedAt.Local().Format(time.RFC3339),
					UpdatedAt:  respDefaultImage.UpdatedAt.Local().Format(time.RFC3339),
				}
				imageChan <- tempImage
			case sqlc.ErrNotFound:
				l.Error("图片记录不存在")
				//return sqlc.ErrNotFound
			default:
				l.Error(err)
				return errors.New("internal server error")
			}
		}
		return err
	}
}

func GetProductUrl(category, handler string) string {
	productUrlMap := map[string]string{
		"index":          "",
		"product-detail": "/collections/all-category/products/" + handler,
		"category":       "/categories/",
		"product-lists":  "/products?handler=" + handler,
		"cart":           "/cart",
		"order":          "/orders",
		"result":         "/order/result",
		"error":          "/error",
		"condition":      "/condition",
		"search":         "/search",
		"page":           "/page/custom/" + handler,
		"unpaid":         "/preview/email/unpaid",
	}
	return productUrlMap[category]
}
