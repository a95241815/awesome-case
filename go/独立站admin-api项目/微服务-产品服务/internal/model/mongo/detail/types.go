package detail

import "github.com/globalsign/mgo/bson"

//go:generate goctl model mongo -t ProductDetail
type ProductDetail struct {
	ID        bson.ObjectId `bson:"_id" json:"id"`
	ProductID string        `bson:"product_id" json:"product_id"`
	BodyHtml  string        `bson:"body_html" json:"body_html"`
}
