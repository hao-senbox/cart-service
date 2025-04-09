package repository

import (
	"context"
	"fmt"
	"store/internal/models"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CartRepository interface {
	GetCartByTeacherStudent(ctx context.Context, teacherID string, studentID string) (*models.Cart, error)
	GetAllCartGroupedByTeacher(ctx context.Context) ([]bson.M, error)
	GetCartByTeacher(ctx context.Context, teacherID string) ([]bson.M, error)
	UpdateCart(ctx context.Context, cart *models.Cart) error
	AddItemToCart(ctx context.Context, teacherID string, studentID string, item models.CartItem) error
	UpdateCartItemQuantity(ctx context.Context, teacherID string, studentID string, productID primitive.ObjectID, quantity int, types string) error
	RemoveFromCart(ctx context.Context, teacherID string, studentID string, productID primitive.ObjectID) error
	ClearCart(ctx context.Context, teacherID string) error 
}

type cartRepository struct {
	collection        *mongo.Collection
	collectionHistory *mongo.Collection
}

func NewCartRepository(collection *mongo.Collection, collectionHistory *mongo.Collection) CartRepository {
	return &cartRepository{
		collection:        collection,
		collectionHistory: collectionHistory,
	}
}

func (r *cartRepository) GetAllCartGroupedByTeacher(ctx context.Context) ([]bson.M, error) {
	// Pipeline gom nhóm theo teacher_id
	pipeline := mongo.Pipeline{
		{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$teacher_id"},            // Nhóm theo teacher_id
				{Key: "carts", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}, // Đẩy toàn bộ cart vào mảng `carts`
			}},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *cartRepository) GetCartByTeacherStudent(ctx context.Context, teacherID string, studentID string) (*models.Cart, error) {

	var cart models.Cart

	filter := bson.M{"teacher_id": teacherID, "student_id": studentID}

	err := r.collection.FindOne(ctx, filter).Decode(&cart)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			newCart := &models.Cart{
				ID:         primitive.NewObjectID(),
				TeacherID:     teacherID,
				StudentID:    studentID,
				Items:      []models.CartItem{},
				TotalPrice: 0,
				CreateAt:   time.Now(),
				UpdateAt:   time.Now(),
			}
			_, insertErr := r.collection.InsertOne(ctx, newCart)
			if insertErr != nil {
				return nil, insertErr
			}
			return newCart, nil
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &cart, nil

}

func (r *cartRepository) GetCartByTeacher(ctx context.Context, teacherID string) ([]bson.M, error) {
	
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"teacher_id": teacherID}}},
		{{
			Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$student_id"},
				{Key: "items", Value: bson.M{"$first": "$items"}},
				{Key: "total_price", Value: bson.M{"$sum": "$total_price"}},
			},
		}},
	}
	

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *cartRepository) UpdateCart(ctx context.Context, cart *models.Cart) error {

	cart.UpdateAt = time.Now()

	filter := bson.M{"_id": cart.ID}

	update := bson.M{
		"$set": bson.M{
			"items":       cart.Items,
			"total_price": cart.TotalPrice,
			"update_at":   cart.UpdateAt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return err
	}

	return nil
}

func (r *cartRepository) AddItemToCart(ctx context.Context, teacherID string, studentID string, item models.CartItem) error {

	cart, err := r.GetCartByTeacherStudent(ctx, teacherID, studentID)

	if err != nil {
		return err
	}

	found := false

	for i, existingItem := range cart.Items {
		if existingItem.ProductID == item.ProductID {
			cart.Items[i].Quantity += item.Quantity
			found = true
			break
		}
	}

	if !found {
		cart.Items = append(cart.Items, item)
	}

	return r.UpdateCartTotalPrice(ctx, cart)
}

func (r *cartRepository) UpdateCartItemQuantity(ctx context.Context, teacherID string, studentID string, productID primitive.ObjectID, quantity int, types string) error {
	if types == "increase" {
		return r.IncreaseCartItemQuantity(ctx, teacherID, studentID, productID)
	} else {
		return r.DecreaseCartItemQuantity(ctx, teacherID, studentID, productID)
	}
}

func (r *cartRepository) IncreaseCartItemQuantity(ctx context.Context, teacherID string, studentID string, productID primitive.ObjectID) error {
	cart, err := r.GetCartByTeacherStudent(ctx, teacherID, studentID)
	if err != nil {
		return err
	}
	found := false
	for i, item := range cart.Items {
		if item.ProductID.Hex() == productID.Hex() {
			found = true
			cart.Items[i].Quantity += 1

			history := models.CartHistory {
				TeacherID: teacherID,
				StudentID: studentID,
				ProductID: productID,
				EventType: "add",
				Quantity: 1,
				OcccuredOn: time.Now(),
			}

			_, err := r.collectionHistory.InsertOne(ctx, history)
			if err != nil {
				return err
			}
		}
	}

	if !found {
		return fmt.Errorf("product not found")
	}
	return r.UpdateCartTotalPrice(ctx, cart)
}

func (r *cartRepository) DecreaseCartItemQuantity(ctx context.Context, teacherID string, studentID string, productID primitive.ObjectID) error {
	cart, err := r.GetCartByTeacherStudent(ctx, teacherID, studentID)
	if err != nil {
		return err
	}
	found := false
	for i, item := range cart.Items {
		if item.ProductID.Hex() == productID.Hex() {
			found = true
			if item.Quantity > 1 {
				cart.Items[i].Quantity -= 1

				history := models.CartHistory {
					TeacherID: teacherID,
					StudentID: studentID,
					ProductID: productID,
					EventType: "remove",
					Quantity: 1,
					OcccuredOn: time.Now(),
				}
	
				_, err := r.collectionHistory.InsertOne(ctx, history)
				if err != nil {
					return err
				}
			} else {
				filter := bson.M{"teacher_id": teacherID, "student_id": studentID}
				update := bson.M{
					"$pull": bson.M{
						"items": bson.M{"product_id": productID},
					},
				}

				_, err := r.collection.UpdateOne(ctx, filter, update)
				if err != nil {
					return err
				}

				cart, err = r.GetCartByTeacherStudent(ctx, teacherID, studentID)
				if err != nil {
					return err
				}

				// Ghi lịch sử xóa sản phẩm
				history := models.CartHistory{
					TeacherID:  teacherID,
					StudentID:  studentID,
					ProductID:  productID,
					EventType:  "remove",
					Quantity:   1,
					OcccuredOn: time.Now(),
				}

				_, err = r.collectionHistory.InsertOne(ctx, history)
				if err != nil {
					return err
				}
			}
			break
		}
	}

	
	if !found {
		return fmt.Errorf("product not found")
	}

	return r.UpdateCartTotalPrice(ctx, cart)
}

func (r *cartRepository) RemoveFromCart(ctx context.Context, teacherID string, studentID string, productID primitive.ObjectID) error {

	cart, err := r.GetCartByTeacherStudent(ctx, teacherID, studentID)
	if err != nil {
		return err
	}

	found := false

	for _, item := range cart.Items {
		if item.ProductID == productID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("product not found in cart")
	}

	filter := bson.M{
		"teacher_id": teacherID,
		"student_id": studentID,
	}

	update := bson.M{
		"$pull": bson.M{
			"items": bson.M{
				"product_id": productID,
			},
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return err
	}

	cart, err = r.GetCartByTeacherStudent(ctx, teacherID, studentID)

	if err != nil {
		return err
	}

	return r.UpdateCartTotalPrice(ctx, cart)

}

func (r *cartRepository) ClearCart(ctx context.Context, teacherID string) error {

	var cart models.Cart

	filter := bson.M{"teacher_id": teacherID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	
	for cursor.Next(ctx) {
		if err := cursor.Decode(&cart); err != nil {
			return err
		}
		cart.Items = []models.CartItem{}
		cart.TotalPrice = 0
		cart.UpdateAt = time.Now()

		if err := r.UpdateCart(ctx, &cart); err != nil {
			return err
		}
	}

	return nil
}

func (r *cartRepository) UpdateCartTotalPrice(ctx context.Context, cart *models.Cart) error {
	totalPrice := 0.0
	for _, item := range cart.Items {
		totalPrice += item.Price * float64(item.Quantity)
	}
	cart.TotalPrice = totalPrice
	return r.UpdateCart(ctx, cart)
}
