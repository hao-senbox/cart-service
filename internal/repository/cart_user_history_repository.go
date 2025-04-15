package repository

import (
	"context"
	"store/internal/models"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CartHistoryRepository struct {
	collectionHistory *mongo.Collection
	collectionCart    *mongo.Collection
}

func NewCartHistoryRepository(collectionHistory *mongo.Collection, collectionCart *mongo.Collection) *CartHistoryRepository {
	return &CartHistoryRepository{
		collectionHistory: collectionHistory,
		collectionCart:    collectionCart,
	}
}

func (r *CartHistoryRepository) AddCartHistory(ctx context.Context, teacherID string, studentID string, productID primitive.ObjectID, eventType string, quantity int) error {

	history := models.CartHistory{
		TeacherID: teacherID,
		StudentID: studentID,
		ProductID:  productID,
		EventType:  eventType,
		OcccuredOn: time.Now(),
		Quantity:   quantity,
	}

	_, err := r.collectionHistory.InsertOne(ctx, history)

	if err != nil {
		return err
	}

	return nil
}

func (r *CartHistoryRepository) AddAllCartHistory(ctx context.Context, teacherID string) error {
	var cart models.Cart

	cursor, err := r.collectionCart.Find(ctx, bson.M{"teacher_id": teacherID})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		if err := cursor.Decode(&cart); err != nil {
			return err
		}

		var historyRecords []interface{}
	
		for _, item := range cart.Items {
			historyRecord := models.CartHistory{
				TeacherID:  teacherID,
				StudentID:  cart.StudentID,
				ProductID:  item.ProductID,
				EventType:  "order",
				Quantity:   item.Quantity,
				OcccuredOn: time.Now(),
			}
			historyRecords = append(historyRecords, historyRecord)
		}
	
		if len(historyRecords) > 0 {
			_, err := r.collectionHistory.InsertMany(ctx, historyRecords)
			if err != nil {
				return err
			}
		}

	}

	if err := cursor.Err(); err != nil {
		return err
	}

	return nil
}
