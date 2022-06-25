package logic

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/model"

	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"

	"strconv"
	"strings"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"gitlab.jhongnet.com/mall/rpc-product-server/internal/svc"
	"gitlab.jhongnet.com/mall/rpc-product-server/product"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/tal-tech/go-zero/core/logx"
)

type ProductImageAddLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProductImageAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductImageAddLogic {
	return &ProductImageAddLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ProductImageAddLogic) ProductImageAdd(in *product.ProductImageAddRequest) (*product.ProductImageAddResponse, error) {
	if in.Url == "" {
		l.Error("添加图片失败,缺少url参数")
		return &product.ProductImageAddResponse{}, errors.New("need url param")
	}
	ext := path.Ext(in.Url)
	if ext != ".png" && ext != ".jpg" && ext != ".bmp" && ext != ".jpeg" && ext != ".svg" && ext != ".gif" {
		l.Error("image.url is invalid")
		return &product.ProductImageAddResponse{}, errors.New("image.url is invalid")
	}
	var url string
	var width int64 = 0
	var err error
	if l.svcCtx.ProjectENV == "xshoppy" {
		url, width, err = l.UploadImage(in.Url)
	} else {
		url, width, err = l.UploadImageEmy(in.Url)
		url = "uploader/" + path.Base(url)
	}

	if err != nil {
		l.Error("上传商品图片出错：", err)
		return &product.ProductImageAddResponse{}, err
	}
	data := model.SailUpload{}
	data.ShopId = in.ShopId
	data.ImageWidth = width
	data.FileKey = url
	h := md5.New()
	h.Write([]byte(url))
	md5Str := hex.EncodeToString(h.Sum(nil))
	data.FileMd5 = md5Str
	resp, err := l.svcCtx.WriteImageModel.InsertProductImage(data, in.ProductId)
	if err != nil {
		l.Error("添加图片失败：", err)
		return &product.ProductImageAddResponse{}, errors.New("internal server error")
	}
	lastId, err := resp.LastInsertId()
	if err != nil {
		l.Error("添加图片失败：", err)
		return &product.ProductImageAddResponse{}, errors.New("internal server error")
	}
	respAdd, err := l.svcCtx.WriteImageModel.FindOne(lastId)
	if err != nil {
		l.Error("添加图片失败：", err)
		return &product.ProductImageAddResponse{}, errors.New("internal server error")
	}
	variantIds := make([]int64, 0)
	respVariants, err := l.svcCtx.ReadVariantModel.FindListByImageId(in.ShopId, in.ProductId, lastId)
	switch err {
	case nil:
		if respVariants != nil {
			for _, variant := range *respVariants {
				variantIds = append(variantIds, variant.Id)
				l.Info(variant.Id)
			}
		}
	case sqlc.ErrNotFound:
		l.Error("该图片关联的子商品列表为空")
		//return &product.ProductImageAddResponse{}, sqlc.ErrNotFound
	default:
		l.Error("查询图片关联的子商品列表出错：", err)
		//return &product.ProductImageAddResponse{}, err
	}

	return &product.ProductImageAddResponse{Image: &product.ProductImage{
		Id:         respAdd.Id,
		ProductId:  in.ProductId,
		Sort:       0,
		Width:      respAdd.ImageWidth,
		Url:        respAdd.FileKey,
		VariantIds: variantIds,
		FileKey:    respAdd.FileKey,
		CreatedAt:  respAdd.CreatedAt.Local().Format(time.RFC3339),
		UpdatedAt:  respAdd.UpdatedAt.Local().Format(time.RFC3339),
	}}, nil
}

func (l *ProductImageAddLogic) UploadImage(url string) (src string, width int64, err error) {

	imgSplit := strings.Split(url, "/")
	length := len(imgSplit)
	if length < 4 {
		l.Error("url参数不可用")
		return "", 0, errors.New("url param invalid")
	}
	respHead, err := http.Head(url)
	if err != nil {
		l.Error("上传图片失败：", err)
		return "", 0, err
	}
	defer respHead.Body.Close()
	if respHead.ContentLength > 8388608 {
		l.Error("图片超过8M")
		return "", 0, errors.New(" image size beyond limit ")
	}

	name := imgSplit[len(imgSplit)-1:][0]
	now := time.Now()
	rand.Seed(now.UnixNano())
	timeStr := strconv.Itoa(int(now.UnixNano()))
	numStr := strconv.Itoa(rand.Intn(999))
	s := timeStr + numStr
	h := sha1.New()
	h.Write([]byte(s))
	bs := hex.EncodeToString(h.Sum(nil))

	name = bs + path.Ext(imgSplit[len(imgSplit)-2 : len(imgSplit)-1][0])

	resp1, err := http.Get(url)

	if err != nil {
		l.Error("上传图片失败：", err)
		return "", 0, err
	}
	defer resp1.Body.Close()
	body, err := ioutil.ReadAll(resp1.Body)
	if err != nil {
		l.Error("上传图片失败：", err)
		return "", 0, err
	}
	if resp1.StatusCode != 200 {
		l.Error("上传图片失败：", resp1.Status)
		return "", 0, errors.New("invalid image url")
	}
	out, err := os.Create(name + path.Ext(url))
	if err != nil {
		l.Error("上传图片失败：", err)
		return "", 0, err
	}
	defer os.Remove(name + path.Ext(url))
	_, err = io.Copy(out, bytes.NewReader(body))
	if err != nil {
		l.Error("上传图片失败：", err)
		return "", 0, errors.New("internal server error")
	}

	endpoint := l.svcCtx.AliOss.Endpoint
	accessKeyId := l.svcCtx.AliOss.AccessKeyId
	accessKeySecret := l.svcCtx.AliOss.AccessKeySecret
	bucketName := l.svcCtx.AliOss.BucketName
	objectName := "uploader/" + name + path.Ext(url)
	localFileName := name + path.Ext(url)
	// 创建OSSClient实例。
	client, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		l.Error(err)
		return "", 0, err
	}
	// 获取存储空间。
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		l.Error(err)
		return "", 0, err
	}
	// 上传文件。
	err = bucket.PutObjectFromFile(objectName, localFileName)
	if err != nil {
		l.Error(err)
		return "", 0, err
	}
	return "uploader/" + name + path.Ext(url), width, nil
}

func (l ProductImageAddLogic) UploadImageEmy(url string) (src string, width int64, err error) {
	imgSplit := strings.Split(url, "/")
	length := len(imgSplit)
	if length < 4 {
		l.Error("url参数不可用")
		return "", 0, errors.New("url param invalid")
	}
	respHead, err := http.Head(url)
	if err != nil {
		l.Error("上传图片失败：", err)
		return "", 0, err
	}
	defer respHead.Body.Close()
	if respHead.ContentLength > 8388608 {
		l.Error("图片超过8M")
		return "", 0, errors.New(" image size beyond limit ")
	}
	name := imgSplit[len(imgSplit)-1:][0]
	now := time.Now()
	rand.Seed(now.UnixNano())
	timeStr := strconv.Itoa(int(now.UnixNano()))
	numStr := strconv.Itoa(rand.Intn(999))
	s := timeStr + numStr
	h := sha1.New() // md5加密类似md5.New()
	h.Write([]byte(s))
	bs := hex.EncodeToString(h.Sum(nil))

	name = bs + path.Ext(imgSplit[len(imgSplit)-2 : len(imgSplit)-1][0])
	l.Info("imageset:", imgSplit)

	resp1, err := http.Get(url)
	if err != nil {
		l.Error("上传图片失败：", err)
		return "", 0, err
	}
	defer resp1.Body.Close()

	body, err := ioutil.ReadAll(resp1.Body)
	if err != nil {
		l.Error("上传图片失败：", err)
		return "", 0, err
	}
	out, err := os.Create(name + path.Ext(url))
	if err != nil {
		l.Error("上传图片失败：", err)
		return "", 0, err
	}
	defer os.Remove(name + path.Ext(url))
	_, err = io.Copy(out, bytes.NewReader(body))
	if err != nil {
		l.Error("上传图片失败：", err)
		return "", 0, errors.New("internal server error")
	}

	accessKey := l.svcCtx.StaticStorage.AccessKey
	accessSecret := l.svcCtx.StaticStorage.AccessSecret
	bucket := l.svcCtx.StaticStorage.Bucket
	region := l.svcCtx.StaticStorage.Region
	ext := path.Ext(url)
	contentType := ""

	switch ext {
	case ".png":
		contentType = "image/png"
	case ".jpg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".bmp":
		contentType = "image/bmp"
	case ".jpeg":
		contentType = "image/jpeg"
	case ".svg":
		contentType = "image/svg"
	default:
		return "", 0, errors.New("unsupported image format")
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, accessSecret, ""),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(false), //virtual-host style方式，不要修改
	})
	uploader := s3manager.NewUploader(sess)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String("uploader/" + name + path.Ext(url)),
		Body:        bytes.NewReader([]byte(body)),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		l.Error("上传图片失败：", err)
		return "", 0, errors.New("upload image failed")
	}

	return result.Location, width, nil

}
