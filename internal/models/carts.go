package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartItem struct {
	ProductID    primitive.ObjectID `bson:"product_id" json:"product_id"`
	ProductName  string             `bson:"product_name" json:"product_name"`
	TopicName    string             `bson:"topic_name" json:"topic_name"`
	CategoryName string             `bson:"category_name" json:"category_name"`
	PriceStore   float64            `bson:"price_store" json:"price_store"`
	PriceService float64            `bson:"price_service" json:"price_service"`
	Quantity     int                `bson:"quantity" json:"quantity"`
	ImageURL     string             `bson:"image_url" json:"image_url"`
}

type Cart struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TeacherID         string             `bson:"teacher_id" json:"teacher_id"`
	StudentID         string             `bson:"student_id" json:"student_id"`
	Items             []CartItem         `bson:"items" json:"items"`
	TotalPriceStore   float64            `bson:"total_price_store" json:"total_price_store"`
	TotalPriceService float64            `bson:"total_price_service" json:"total_price_service"`
	CreateAt          time.Time          `bson:"create_at" json:"create_at"`
	UpdateAt          time.Time          `bson:"update_at" json:"update_at"`
}
