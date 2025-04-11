package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartItem struct {
	ProductID   primitive.ObjectID `bson:"product_id" json:"product_id"`
	ProductName string             `bson:"product_name" json:"product_name"`
	Price       float64            `bson:"price" json:"price"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	ImageURL    string             `bson:"image_url" json:"image_url"`
}

type Cart struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TeacherID  string             `bson:"teacher_id" json:"teacher_id"`
	StudentID  string             `bson:"student_id" json:"student_id"`
	Items      []CartItem         `bson:"items" json:"items"`
	TotalPrice float64            `bson:"total_price" json:"total_price"`
	CreateAt   time.Time          `bson:"create_at" json:"create_at"`
	UpdateAt   time.Time          `bson:"update_at" json:"update_at"`
}
