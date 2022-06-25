package model

import (
	"errors"
	"github.com/tal-tech/go-zero/core/logx"
)

type Handler interface {
	SetId(int64)
	SetHandler(string)
}

type HandlerFactory interface {
	Create() Handler
}

type HandlerItem struct {
	Id      int64
	Handler string
}

type HandlerOption func(option *HandlerItem)

func (h *HandlerItem) SetId(id int64) {
	h.Id = id
}

func (h *HandlerItem) SetHandler(handler string) {
	h.Handler = handler
}

func WithId(id int64) HandlerOption {
	return func(h *HandlerItem) {
		if id > 0 {
			h.SetId(id)
		}
	}
}

func WithHandler(handler string) HandlerOption {
	return func(h *HandlerItem) {
		if handler != "" {
			h.SetHandler(handler)
		}
	}
}

func SetOptions(opts ...HandlerOption) *HandlerItem {
	defaultProduct := HandlerItem{}
	for _, opt := range opts {
		opt(&defaultProduct)
	}

	return &defaultProduct
}

func SetQueryStr(filter []HandlerOption) (args []interface{}, extraWhere string, err error) {
	if len(filter) == 0 {
		logx.Error("need query argument id or handler")
		return nil, "", errors.New("need query argument id or handler")
	}
	filters := SetOptions(filter...)

	if filters.Id != 0 {
		args = append(args, filters.Id)
		extraWhere += " and `id` = ? "
	}
	if filters.Handler != "" {
		args = append(args, filters.Handler)
		extraWhere += " and `handler` = ? "
	}

	return args, extraWhere, nil
}
