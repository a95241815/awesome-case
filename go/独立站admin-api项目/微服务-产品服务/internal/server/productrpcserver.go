// Code generated by goctl. DO NOT EDIT!
// Source: rpc-product.proto

package server

import (
	"context"

	"gitlab.jhongnet.com/mall/rpc-product-server/internal/logic"
	"gitlab.jhongnet.com/mall/rpc-product-server/internal/svc"
	"gitlab.jhongnet.com/mall/rpc-product-server/product"
)

type ProductRPCServer struct {
	svcCtx *svc.ServiceContext
}

func NewProductRPCServer(svcCtx *svc.ServiceContext) *ProductRPCServer {
	return &ProductRPCServer{
		svcCtx: svcCtx,
	}
}

func (s *ProductRPCServer) Ping(ctx context.Context, in *product.PingRequest) (*product.PingResponse, error) {
	l := logic.NewPingLogic(ctx, s.svcCtx)
	return l.Ping(in)
}

func (s *ProductRPCServer) ProductList(ctx context.Context, in *product.ProductListRequest) (*product.ProductListResponse, error) {
	l := logic.NewProductListLogic(ctx, s.svcCtx)
	return l.ProductList(in)
}

func (s *ProductRPCServer) ProductDetail(ctx context.Context, in *product.ProductDetailRequest) (*product.ProductDetailResponse, error) {
	l := logic.NewProductDetailLogic(ctx, s.svcCtx)
	return l.ProductDetail(in)
}

func (s *ProductRPCServer) ProductAdd(ctx context.Context, in *product.ProductAddRequest) (*product.ProductAddResponse, error) {
	l := logic.NewProductAddLogic(ctx, s.svcCtx)
	return l.ProductAdd(in)
}

func (s *ProductRPCServer) ProductUpdate(ctx context.Context, in *product.ProductUpdateRequest) (*product.ProductUpdateResponse, error) {
	l := logic.NewProductUpdateLogic(ctx, s.svcCtx)
	return l.ProductUpdate(in)
}

func (s *ProductRPCServer) ProductCount(ctx context.Context, in *product.ProductCountRequest) (*product.ProductCountResponse, error) {
	l := logic.NewProductCountLogic(ctx, s.svcCtx)
	return l.ProductCount(in)
}

func (s *ProductRPCServer) ProductDelete(ctx context.Context, in *product.ProductDeleteRequest) (*product.ProductDeleteResponse, error) {
	l := logic.NewProductDeleteLogic(ctx, s.svcCtx)
	return l.ProductDelete(in)
}

func (s *ProductRPCServer) ProductImageList(ctx context.Context, in *product.ProductImageListRequest) (*product.ProductImageListResponse, error) {
	l := logic.NewProductImageListLogic(ctx, s.svcCtx)
	return l.ProductImageList(in)
}

func (s *ProductRPCServer) ProductImageDetail(ctx context.Context, in *product.ProductImageDetailRequest) (*product.ProductImageDetailResponse, error) {
	l := logic.NewProductImageDetailLogic(ctx, s.svcCtx)
	return l.ProductImageDetail(in)
}

func (s *ProductRPCServer) ProductImageAdd(ctx context.Context, in *product.ProductImageAddRequest) (*product.ProductImageAddResponse, error) {
	l := logic.NewProductImageAddLogic(ctx, s.svcCtx)
	return l.ProductImageAdd(in)
}

func (s *ProductRPCServer) ProductImageCount(ctx context.Context, in *product.ProductImageCountRequest) (*product.ProductImageCountResponse, error) {
	l := logic.NewProductImageCountLogic(ctx, s.svcCtx)
	return l.ProductImageCount(in)
}

func (s *ProductRPCServer) ProductVariantList(ctx context.Context, in *product.ProductVariantListRequest) (*product.ProductVariantListResponse, error) {
	l := logic.NewProductVariantListLogic(ctx, s.svcCtx)
	return l.ProductVariantList(in)
}

func (s *ProductRPCServer) ProductVariantDetail(ctx context.Context, in *product.ProductVariantDetailRequest) (*product.ProductVariantDetailResponse, error) {
	l := logic.NewProductVariantDetailLogic(ctx, s.svcCtx)
	return l.ProductVariantDetail(in)
}

func (s *ProductRPCServer) ProductVariantAdd(ctx context.Context, in *product.ProductVariantAddRequest) (*product.ProductVariantAddResponse, error) {
	l := logic.NewProductVariantAddLogic(ctx, s.svcCtx)
	return l.ProductVariantAdd(in)
}

func (s *ProductRPCServer) ProductVariantUpdate(ctx context.Context, in *product.ProductVariantUpdateRequest) (*product.ProductVariantUpdateResponse, error) {
	l := logic.NewProductVariantUpdateLogic(ctx, s.svcCtx)
	return l.ProductVariantUpdate(in)
}

func (s *ProductRPCServer) ProductVariantCount(ctx context.Context, in *product.ProductVariantCountRequest) (*product.ProductVariantCountResponse, error) {
	l := logic.NewProductVariantCountLogic(ctx, s.svcCtx)
	return l.ProductVariantCount(in)
}

func (s *ProductRPCServer) ProductVariantDelete(ctx context.Context, in *product.ProductVariantDeleteRequest) (*product.ProductVariantDeleteResponse, error) {
	l := logic.NewProductVariantDeleteLogic(ctx, s.svcCtx)
	return l.ProductVariantDelete(in)
}

func (s *ProductRPCServer) ProductCommentScoreList(ctx context.Context, in *product.ProductCommentScoreRequest) (*product.ProductCommentScoreResponse, error) {
	l := logic.NewProductCommentScoreListLogic(ctx, s.svcCtx)
	return l.ProductCommentScoreList(in)
}

func (s *ProductRPCServer) CategoryList(ctx context.Context, in *product.CategoryListRequest) (*product.CategoryListResponse, error) {
	l := logic.NewCategoryListLogic(ctx, s.svcCtx)
	return l.CategoryList(in)
}

func (s *ProductRPCServer) CategoryDetail(ctx context.Context, in *product.CategoryDetailRequest) (*product.CategoryDetailResponse, error) {
	l := logic.NewCategoryDetailLogic(ctx, s.svcCtx)
	return l.CategoryDetail(in)
}

func (s *ProductRPCServer) CategoryAdd(ctx context.Context, in *product.CategoryAddRequest) (*product.CategoryAddResponse, error) {
	l := logic.NewCategoryAddLogic(ctx, s.svcCtx)
	return l.CategoryAdd(in)
}

func (s *ProductRPCServer) CategoryCount(ctx context.Context, in *product.CategoryCountRequest) (*product.CategoryCountResponse, error) {
	l := logic.NewCategoryCountLogic(ctx, s.svcCtx)
	return l.CategoryCount(in)
}

func (s *ProductRPCServer) CategoryProductList(ctx context.Context, in *product.CategoryProductListRequest) (*product.CategoryProductListResponse, error) {
	l := logic.NewCategoryProductListLogic(ctx, s.svcCtx)
	return l.CategoryProductList(in)
}

func (s *ProductRPCServer) CategoryProductDetail(ctx context.Context, in *product.CategoryProductDetailRequest) (*product.CategoryProductDetailResponse, error) {
	l := logic.NewCategoryProductDetailLogic(ctx, s.svcCtx)
	return l.CategoryProductDetail(in)
}

func (s *ProductRPCServer) CategoryProductAdd(ctx context.Context, in *product.CategoryProductAddRequest) (*product.CategoryProductAddResponse, error) {
	l := logic.NewCategoryProductAddLogic(ctx, s.svcCtx)
	return l.CategoryProductAdd(in)
}

func (s *ProductRPCServer) CategoryProductCount(ctx context.Context, in *product.CategoryProductCountRequest) (*product.CategoryProductCountResponse, error) {
	l := logic.NewCategoryProductCountLogic(ctx, s.svcCtx)
	return l.CategoryProductCount(in)
}

func (s *ProductRPCServer) CategoryProductDelete(ctx context.Context, in *product.CategoryProductDeleteRequest) (*product.CategoryProductDeleteResponse, error) {
	l := logic.NewCategoryProductDeleteLogic(ctx, s.svcCtx)
	return l.CategoryProductDelete(in)
}
