package logic

import (
	"context"
	"encoding/json"
	"errors"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/mr"
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/model"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/svc"
	"gitlab.jhongnet.com/mall/rpc-product-server/product"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ProductAddLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProductAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductAddLogic {
	return &ProductAddLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ProductAddLogic) ProductAdd(in *product.ProductAddRequest) (*product.ProductAddResponse, error) {
	l.Info("shop_id:", in.ShopId)
	if in.Fields == "" {
		in.Fields = "id,title,price,compare_price,weight,weight_unit,shop_url,product_url,body_html,seo_title,seo_desc,published_at,is_use_stock,soldout_policy,handler,status,created_at,updated_at,images,default_image,variants"
	}

	if in.ShopId == 0 {
		l.Error("缺少shop_id参数")
		return &product.ProductAddResponse{}, errors.New("internal server error")
	}
	if in.Title == "" {
		l.Error("缺少title参数")
		return &product.ProductAddResponse{}, errors.New("product.title not set")
	}

	//var defaultImageErr error

	var defaultWidth int64

	imagesReq := make([]model.ImageData, 0)
	variantsReq := make([]model.VariantData, 0)

	//lockKey := fmt.Sprintf("adminapi_lock_product_title_%d_%s", in.ShopId, in.Title)
	//if !l.lock(lockKey, 5) {
	//	l.Error("添加商品失败，product title相同请求过多被限制")
	//	return &product.ProductAddResponse{}, errors.New(" product title repeated, too many request ")
	//}
	//defer l.unlock(lockKey)
	if in.Images != nil {
		if len(in.Images) != 0 {
			in.DefaultImageUrl = in.Images[0].Url
		}
		for _, image := range in.Images {
			ext := path.Ext(image.Url)
			if ext != ".png" && ext != ".jpg" && ext != ".bmp" && ext != ".jpeg" && ext != ".svg" && ext != ".gif" {
				l.Error("添加商品图片失败：图片格式有误")
				return &product.ProductAddResponse{}, errors.New(" image in wrong format ")
			}

			tempImage := model.ImageData{
				FileKey:    image.Url,
				ImageWidth: 0,
			}
			imagesReq = append(imagesReq, tempImage)
		}

	}

	if in.Variants != nil {
		if len(in.Variants) != 0 && in.DefaultImageUrl == "" {
			if in.Variants[0].ImageUrl != "" {
				in.DefaultImageUrl = in.Variants[0].ImageUrl
			}
		}
		for _, variant := range in.Variants {

			tempVariant := model.VariantData{
				Price:             variant.Price,
				ComparePrice:      variant.ComparePrice,
				Weight:            variant.Weight,
				WeightUnit:        variant.WeightUnit,
				RequiresShipping:  variant.RequiresShipping,
				InventoryQuantity: variant.InventoryQuantity,
				Spec:              variant.Spec,
				Sort:              variant.Sort,
				SkuCode:           variant.Sku,
				Title:             variant.Title,
				Options:           variant.Options,
				IsChecked:         variant.IsChecked,
				Image: model.ImageData{
					FileKey:    variant.ImageUrl,
					ImageWidth: 0,
				},
			}
			variantsReq = append(variantsReq, tempVariant)
		}
	}

	data := model.InsertProductData{
		Title:             strip.StripTags(in.Title),
		Price:             in.Price,
		ComparePrice:      in.ComparePrice,
		Weight:            in.Weight,
		WeightUnit:        in.WeightUnit,
		Sku:               in.Sku,
		SoldOutPolicy:     in.SoldOutPolicy,
		RequireShipping:   in.RequiresShipping,
		IsUseStock:        in.IsUseStock,
		BodyHtml:          in.BodyHtml,
		Status:            in.Status,
		ShopId:            in.ShopId,
		SeoTitle:          in.SeoTitle,
		SeoDesc:           in.SeoDesc,
		YoutubeVideoUrl:   in.YoutubeVideoUrl,
		YoutubeVideoPos:   strconv.Itoa(int(in.YoutubeVideoPos)),
		Attribute:         in.Attribute,
		InventoryQuantity: in.InventoryQuantity,
		Tags:              in.Tags,
		DefaultImage: model.ImageData{
			FileKey:    in.DefaultImageUrl,
			ImageWidth: defaultWidth,
		},
		Images:   &imagesReq,
		Variants: &variantsReq,
	}
	l.Error("imagesReq:", imagesReq)
	l.Error("variantsReq:", variantsReq)
	if len(imagesReq) == 0 {
		data.Images = nil
	} else {
		data.Images = &imagesReq
	}
	if len(variantsReq) == 0 {
		data.Variants = nil
	} else {
		data.Variants = &variantsReq
	}

	pro, respImageData, err := l.svcCtx.WriteModel.InsertProduct(data, l.svcCtx.RedisClientSaas)

	if err != nil {
		l.Error("创建产品失败：", err)
		errText := err.Error()
		if err.Error() == " product title repeated, too many request " {
			errText = " product title repeated, please retry later "
		}
		return &product.ProductAddResponse{}, errors.New(errText)
	}
	l.Error("pro:", pro)
	lastId, err := pro.LastInsertId()
	if err != nil {
		l.Error("创建产品失败：", err)
		return &product.ProductAddResponse{}, errors.New("internal server error")
	}
	l.Info("respImageData:", respImageData)
	for _, i2 := range respImageData.Variants {
		temp := model.ImageItem{Url: i2.Url}
		respImageData.Images = append(respImageData.Images, temp)
	}
	imagesMap := sync.Map{}
	for _, image := range respImageData.Images {
		if image.Url != "" {
			imagesMap.Store(image.Url, int64(0))
		}
	}
	imagesSlice := make([]model.ImageItem, 0)
	keyCheck := map[string]string{}
	for _, image := range respImageData.Images {
		if _, ok := keyCheck[image.Url]; ok {
			continue
		}
		temp := model.ImageItem{
			Url: image.Url,
		}
		imagesSlice = append(imagesSlice, temp)
		keyCheck[image.Url] = "1"
	}
	respImageData.Images = imagesSlice
	if respImageData != nil {
		if len(respImageData.Variants) != 0 || len(respImageData.Images) != 0 {
			go func() {
				imageAddLogic := NewProductImageAddLogic(l.ctx, l.svcCtx)
				//for s, i := range imagesMap {
				//	var urlImage string
				//	var width int64
				//	var errImage error
				//	if l.svcCtx.ProjectENV == "xshoppy" {
				//		urlImage, width, errImage = imageAddLogic.UploadImage(in.DefaultImageUrl)
				//	} else {
				//		urlImage, width, errImage = imageAddLogic.UploadImageEmy(in.DefaultImageUrl)
				//		urlImage = "uploader/" + path.Base(urlImage)
				//	}
				//
				//
				//
				//}
				l.Info("imagesmap:", &imagesMap)
				l.Info("respImagedata:", respImageData)
				l.Info("count:", len(respImageData.Images))
				if len(respImageData.Images) != 0 {
					l.Info("111111")
					failedProImg := FailedProImg{
						ProductId: lastId,
					}

					for _, image := range respImageData.Images {
						imageData := product.ProductImageAddRequest{
							ShopId:    in.ShopId,
							ProductId: lastId,
							Url:       image.Url,
						}
						respImages, errDefault := imageAddLogic.ProductImageAdd(&imageData)
						if errDefault != nil {
							l.Error("添加产品图片失败：", errDefault)
							failedProImg.Urls = append(failedProImg.Urls, image.Url)
							continue
						}
						l.Info("产品图链接：", image.Url, ":", respImages.Image.Id, ":", respImages.Image.Url)
						imagesMap.Store(image.Url, respImages.Image.Id)
					}
					l.Error("imagesmap:", &imagesMap)
					if len(failedProImg.Urls) != 0 {
						defaultByte, errDefaultByte := json.Marshal(failedProImg)
						if errDefaultByte != nil {
							l.Error("序列化failedProImage失败：", err)
						}
						_, errDefaultPush := l.svcCtx.RedisClient.Lpush("admin-api:failedProImgs", string(defaultByte))
						if errDefaultPush != nil {
							l.Error("添加failedProImage队列失败：", err)
						}
					}

				}

				if len(respImageData.Images) > 0 {
					in.DefaultImageUrl = respImageData.Images[0].Url
				}
				var imageId int64
				defaultImgId, ok := imagesMap.Load(in.DefaultImageUrl)
				if ok {
					if val, ok := defaultImgId.(int64); ok {
						imageId = val
					} else {
						imageId = int64(defaultImgId.(int))
					}
				}
				errDefault := l.svcCtx.WriteModel.UpdateDefaultImage(imageId, lastId)
				if errDefault != nil {
					tempFailedDefaultImg := FailedDefaultProImg{
						ProductId: lastId,
						Url:       in.DefaultImageUrl,
					}

					defaultByte, errDefaultByte := json.Marshal(tempFailedDefaultImg)
					if errDefaultByte != nil {
						l.Error("序列化failedDefaultImage失败：", errDefaultByte)
					}
					_, errDefaultPush := l.svcCtx.RedisClient.Lpush("admin-api:failedDefaultProImg", string(defaultByte))
					if errDefaultPush != nil {
						l.Error("加入failedDefaultImage队列失败：", err)
					}
				}

				if len(respImageData.Variants) != 0 {
					failedProImg := FailedProImg{
						ProductId: lastId,
					}
					for _, image := range respImageData.Variants {
						var imageId int64
						defaultImgId, ok := imagesMap.Load(image.Url)
						if ok {
							if val, ok := defaultImgId.(int64); ok {
								imageId = val
							} else {
								imageId = int64(defaultImgId.(int))
							}
						}
						errDefault := l.svcCtx.WriteVariantModel.UpdateImage(imageId, image.VariantId)
						if errDefault != nil {
							l.Error("添加产品图片失败：", errDefault)
							failedProImg.Urls = append(failedProImg.Urls, image.Url)
						}
					}
					if len(failedProImg.Urls) != 0 {
						defaultByte, errDefaultByte := json.Marshal(failedProImg)
						if errDefaultByte != nil {
							l.Error("序列化failedVariantImage失败：", errDefaultByte)
						}
						_, errDefaultPush := l.svcCtx.RedisClient.Lpush("admin-api:failedVariantImg", string(defaultByte))
						if errDefaultPush != nil {
							l.Error("添加failedVariantImage队列失败：", errDefaultPush)
						}
					}
				}
			}()
		}

	}

	resp, err := l.svcCtx.WriteModel.FindOne(in.ShopId, model.WithId(lastId))
	switch err {
	case nil:

	case sqlc.ErrNotFound:
		l.Error("product record not found")
		return &product.ProductAddResponse{}, errors.New("internal server error")
	default:
		l.Error(err)
		return nil, err
	}

	//var productDetail product.Product
	var defaultImage product.ProductImage
	var images []*product.ProductImage
	var variants []*product.ProductVariant
	var bodyHtml string

	err = mr.Finish(func() error {
		if strings.Contains(in.Fields, "body_html") {
			respDetail, err := l.svcCtx.ReadProductDetailModel.FindOne(lastId)
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
						ProductId:  lastId,
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
			variantResp, err := l.svcCtx.WriteVariantModel.FindList(in.ShopId, lastId)
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
			l.Info("imageSet:", imageSet)
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
							ProductId:  lastId,
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

	return &product.ProductAddResponse{Product: &product.Product{
		ProductTitle:    resp.Title,
		ProductId:       resp.Id,
		ShopUrl:         "",
		PreviewUrl:      "/products/" + resp.Handler,
		BodyHtml:        bodyHtml,
		Price:           resp.Price,
		CompareAtPrice:  resp.CompareAtPrice,
		Weight:          resp.Weight,
		WeightUnit:      resp.WeightUnit,
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
		Vendor:          in.Vendor,
		ProductType:     in.ProductType,
	}}, nil

}

type FailedDefaultProImg struct {
	ProductId int64  `json:"product_id"`
	Url       string `json:"url"`
}

type FailedProImg struct {
	ProductId int64    `json:"product_id"`
	Urls      []string `json:"urls"`
}

type FailedVariantImg struct {
	ProductId int64  `json:"product_id"`
	VariantId int64  `json:"variant_id"`
	Url       string `json:"url"`
}

func (l *ProductAddLogic) checkLock(key string) bool {
	resp, err := l.svcCtx.RedisClient.Get(key)
	if err != nil {
		l.Error("查询redis出错:", err)
		return false
	}
	if resp == "" {
		return false
	}
	lock := Lock{}
	err = json.Unmarshal([]byte(resp), &lock)
	if err != nil {
		l.Error("反序列化lock出错:", err)
		return false
	}
	if lock.Expires < time.Now().Unix() {
		return false
	}

	return true
}

func (l ProductAddLogic) lock(key string, ex int64) bool {
	if l.checkLock(key) {
		return false
	}
	ex = 5
	lock := Lock{}
	lock.Expires = time.Now().Unix() + 5
	lockStr, err := json.Marshal(lock)
	if err != nil {
		l.Error("序列化lock出错:", err)
		return false
	}
	err = l.svcCtx.RedisClient.Set(key, string(lockStr))
	if err != nil {
		l.Error("lock存入redis出错:", err)
		return false
	}

	return true
}

func (l ProductAddLogic) unlock(key string) bool {
	resp, err := l.svcCtx.RedisClient.Get(key)
	if err != nil {
		l.Error("redis查询lock出错:", err)
		return false
	}
	if resp != "" {
		_, err := l.svcCtx.RedisClient.Del(key)
		if err != nil {
			l.Error("redis删除lock出错:", err)
			return false
		}
	}
	return true
}

type Lock struct {
	Expires int64 `json:"expires"`
}

type UploadedImg struct {
	ImageId int64
}
