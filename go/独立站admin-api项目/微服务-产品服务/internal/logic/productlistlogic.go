package logic

import (
	"context"
	"errors"
	"github.com/tal-tech/go-zero/core/mr"
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/model"
	"sort"
	"strconv"
	"strings"
	"time"

	"gitlab.jhongnet.com/mall/rpc-product-server/internal/svc"
	"gitlab.jhongnet.com/mall/rpc-product-server/product"

	"github.com/tal-tech/go-zero/core/logx"
)

type ProductListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProductListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductListLogic {
	return &ProductListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ProductListLogic) ProductList(in *product.ProductListRequest) (*product.ProductListResponse, error) {
	l.Info(" ProductList Called ")
	if in.Fields == "" {
		in.Fields = "id,title,price,compare_price,weight,weight_unit,shop_url,product_url,body_html,seo_title,seo_desc,published_at,is_use_stock,soldout_policy,handler,status,created_at,updated_at,images,default_image,variants"
	}
	if in.Ids != "" && in.Limit == 0 {
		idSet := strings.Split(in.Ids, ",")
		tempIdSet := make([]string, 0)
		for _, s := range idSet {
			s = strings.TrimSpace(s)
			if s != "" && s != "0" {
				tempIdSet = append(tempIdSet, s)
			}
		}
		in.Ids = strings.Join(tempIdSet, ",")
		if len(tempIdSet) > 0 {
			in.Limit = int64(len(tempIdSet))
		}
	}
	var handle []string
	if in.Handlers != "" {
		tempHandle := strings.Split(in.Handlers, ",")
		for _, s := range tempHandle {
			if s != "" {
				handle = append(handle, s)
			}
		}
	}
	options := model.ListOptions{
		Limit:           in.Limit,
		Page:            in.Page,
		SinceId:         in.SinceId,
		CreatedAtMin:    in.CreatedAtMin,
		CreatedAtMax:    in.CreatedAtMax,
		UpdatedAtMin:    in.UpdatedAtMin,
		UpdatedAtMax:    in.UpdatedAtMax,
		PublishedAtMin:  in.PublishedAtMin,
		PublishedAtMax:  in.PublishedAtMax,
		PublishedStatus: in.PublishedStatus,
		Ids:             in.Ids,
		Title:           in.Title,
		Handlers:        handle,
		IsNewVersion:    in.IsNewVersion,
	}
	l.Error("options:", options)
	if in.ShopId == 0 {
		l.Error("shop_id 参数不能为空")
		return nil, errors.New(" argument shop_id is needed ")
	}
	resp, err := l.svcCtx.ReadModel.FindList(options, in.ShopId)
	switch err {
	case nil:

	case sqlc.ErrNotFound:
		return nil, sqlc.ErrNotFound
	default:
		l.Error(err)
		return nil, err
	}
	if resp == nil {
		l.Error("查询数据为空")
		return nil, errors.New("查询列表数据为空")
	}

	l.Error("列表条数", len(*resp))

	//productSet := make([]*product.Product, 0)

	r, err := mr.MapReduce(func(source chan<- interface{}) {
		for _, valPro := range *resp {
			source <- valPro
		}
	}, func(item interface{}, writer mr.Writer, cancel func(error)) {
		valPro := item.(model.SailShopProduct)
		resp, err := l.svcCtx.ReadModel.FindOne(in.ShopId, model.WithId(valPro.Id))
		switch err {
		case nil:

		case sqlc.ErrNotFound:
			l.Error("商品记录不存在,product_id:", valPro.Id)
			return
		default:
			l.Error(err)
			cancel(errors.New("internal server error"))
		}

		var productDetail product.Product
		var defaultImage product.ProductImage
		var images []*product.ProductImage
		var variants []*product.ProductVariant
		var bodyHtml string

		err = mr.Finish(func() error {
			if strings.Contains(in.Fields, "body_html") {
				respDetail, err := l.svcCtx.ReadProductDetailModel.FindOne(valPro.Id)
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
							ProductId:  valPro.Id,
							Sort:       0,
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
				variantResp, err := l.svcCtx.ReadVariantModel.FindList(in.ShopId, valPro.Id)
				//_, err = l.svcCtx.ReadVariantModel.FindList(in.ShopId, in.ProductId)
				switch err {
				case nil:
					if variantResp != nil {
						for _, valVar := range *variantResp {
							inventoryPolicy := "N"
							if valPro.IsUseStock == 1 {
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
					//return errors.New("internal server error")
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
								ProductId:  valPro.Id,
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
			cancel(errors.New("internal server error"))
		}

		var youtubeVideoPos int
		if resp.YoutubeVideoPos != "" {
			youtubeVideoPos, err = strconv.Atoi(resp.YoutubeVideoPos)
			if err != nil {
				l.Error(err)
				cancel(errors.New("internal server error"))
			}
		}

		productDetail = product.Product{
			ProductTitle:    resp.Title,
			ProductId:       valPro.Id,
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
		//tempImage := &product.Image{}
		if err != nil {
			l.Error(err)
			cancel(err)
		}

		writer.Write(&productDetail)

	}, func(pipe <-chan interface{}, writer mr.Writer, cancel func(error)) {
		var set []*product.Product
		for p := range pipe {
			set = append(set, p.(*product.Product))
		}
		writer.Write(set)
	})
	if err != nil {
		l.Error("查询产品列表出错: %v", err)
		return nil, err
	}
	result := r.([]*product.Product)
	sort.Slice(result, func(i, j int) bool {
		return result[i].ProductId > result[j].ProductId
	})

	return &product.ProductListResponse{Products: result}, nil
}

func (l *ProductListLogic) GetImage() error {
	return errors.New("123")
}

func (l *ProductListLogic) GetFunc(index, imageId, productId int64, imageChan chan *product.ProductImage) func() error {
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
				//return errors.New("internal server error")
			}
		}
		logx.Error(err)
		return err
	}
}
