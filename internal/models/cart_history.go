package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartHistory struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TeacherID  string             `bson:"teacher_id" json:"teacher_id"`
	StudentID  string             `bson:"student_id" json:"student_id"`
	ProductID  primitive.ObjectID `bson:"product_id" json:"product_id"`
	EventType  string             `bson:"event_type" json:"event_type"`
	Quantity   int                `bson:"quantity" json:"quantity"`
	OcccuredOn time.Time          `bson:"occcured_on" json:"occcured_on"`
}
