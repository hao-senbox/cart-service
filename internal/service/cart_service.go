package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"store/internal/models"
	"store/internal/repository"
	"store/pkg/constants"
	"store/pkg/consul"

	"github.com/hashicorp/consul/api"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartService interface {
	AddToCart(ctx context.Context, req *models.AddToCartRequest) (*models.CartItem, error)
	GetCartByTeacher(ctx context.Context, teacherID string) ([]bson.M, error)
	GetAllCartGroupedByTeacher(ctx context.Context) ([]bson.M, error)
	UpdateQuantityItem(ctx context.Context, productID string, req *models.UpdateCartItemRequest) error
	RemoveFromCart(ctx context.Context, teacherID string, studentID string, productID string) error
	ClearCart(ctx context.Context, teacherID string) error
	CheckOutCart(ctx context.Context, req *models.CheckOutCartRequest) error
	GetCartHistoryByTeacher(ctx context.Context, teacherID string) ([]bson.M, error)
}

type cartService struct {
	repoCart    repository.CartRepository
	repoHistory repository.CartHistoryRepository
	productAPI  *callAPI
	orderAPI    *callAPI
}

type callAPI struct {
	client       consul.ServiceDiscovery
	clientServer *api.CatalogService
}

var (
	productService = "product-service"
	orderService   = "order-service"
)

func NewCartService(repo repository.CartRepository, repoHistory repository.CartHistoryRepository, client *api.Client) CartService {

	productAPI := NewServiceAPI(client, productService)
	orderAPI := NewServiceAPI(client, orderService)
	return &cartService{
		repoCart:    repo,
		repoHistory: repoHistory,
		productAPI:  productAPI,
		orderAPI:    orderAPI,
	}
}

func (s *cartService) GetAllCartGroupedByTeacher(ctx context.Context) ([]bson.M, error) {
	return s.repoCart.GetAllCartGroupedByTeacher(ctx)
}

func (s *cartService) GetCartByTeacher(ctx context.Context, teacherID string) ([]bson.M, error) {
	return s.repoCart.GetCartByTeacher(ctx, teacherID)
}

func (s *cartService) AddToCart(ctx context.Context, req *models.AddToCartRequest) (*models.CartItem, error) {
	var topic string
	var imageURL string
	var category string
	productID, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID format: %v", err)
	}

	productRes := s.productAPI.GetProductByID(req.ProductID)

	if productRes == nil {
		return nil, fmt.Errorf("product not found")
	}
	product := productRes["data"].(map[string]interface{})
	name := product["product_name"].(string)
	priceStore := product["original_price_store"].(float64)
	priceService := product["original_price_service"].(float64)
	if product["cover_image"] == nil {
		imageURL = ""
	} else {
		imageURL = product["cover_image"].(string)
	}
	if rawTopic, ok := product["topic"]; ok && rawTopic != nil {
		topicMap, ok := rawTopic.(map[string]interface{})
		if ok {
			topic = topicMap["topic_name"].(string)
		}
	} else {
		topic = ""
	}

	if rawCategory, ok := product["category"]; ok && rawCategory != nil {
		categoryMap, ok := rawCategory.(map[string]interface{})
		if ok {
			category = categoryMap["category_name"].(string)
		}
	} else {
		category = ""
	}

	cartItem := &models.CartItem{
		ProductID:    productID,
		Quantity:     req.Quantity,
		ProductName:  name,
		TopicName:    topic,
		CategoryName: category,
		PriceStore:   priceStore,
		PriceService: priceService,
		ImageURL:     imageURL,
	}

	if err = s.repoCart.AddItemToCart(ctx, req.TeacherID, req.StudentID, *cartItem); err != nil {
		return nil, err
	}

	if err = s.repoHistory.AddCartHistory(ctx, req.TeacherID, req.StudentID, productID, "add", req.Quantity); err != nil {
		return nil, fmt.Errorf("unable to add cart history: %w", err)
	}

	cart, err := s.repoCart.GetCartByTeacherStudent(ctx, req.TeacherID, req.StudentID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving updated cart: %w", err)
	}

	for _, item := range cart.Items {
		if item.ProductID == productID {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("item not found in cart")
}

func (s *cartService) UpdateQuantityItem(ctx context.Context, productID string, req *models.UpdateCartItemRequest) error {
	
	var topic string
	var imageURL string
	var category string
	var quantity int
	id, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return fmt.Errorf("invalid product ID format: %v", err)
	}
	

	productRes := s.productAPI.GetProductByID(productID)

	if productRes == nil {
		return fmt.Errorf("product not found")
	}

	product := productRes["data"].(map[string]interface{})
	name := product["product_name"].(string)
	priceStore := product["original_price_store"].(float64)
	priceService := product["original_price_service"].(float64)
	if product["cover_image"] == nil {
		imageURL = ""
	} else {
		imageURL = product["cover_image"].(string)
	}
	if rawTopic, ok := product["topic"]; ok && rawTopic != nil {
		topicMap, ok := rawTopic.(map[string]interface{})
		if ok {
			topic = topicMap["topic_name"].(string)
		}
	} else {
		topic = ""
	}

	if rawCategory, ok := product["category"]; ok && rawCategory != nil {
		categoryMap, ok := rawCategory.(map[string]interface{})
		if ok {
			category = categoryMap["category_name"].(string)
		}
	} else {
		category = ""
	}

	if req.Quantity == nil {
		quantity = 1
	} else {
		quantity = *req.Quantity
		fmt.Printf("Quantity: %d\n", quantity)
	}

	cartItem := &models.CartItem{
		ProductID:    id,
		Quantity:     quantity,
		ProductName:  name,
		TopicName:    topic,
		CategoryName: category,
		PriceStore:   priceStore,
		PriceService: priceService,
		ImageURL:     imageURL,
	}

	return s.repoCart.UpdateCartItemQuantity(ctx, req.TeacherID, req.StudentID, id, quantity, req.Type, *cartItem)
}

func (s *cartService) RemoveFromCart(ctx context.Context, teacherID string, studentID string, productID string) error {

	id, err := primitive.ObjectIDFromHex(productID)

	if err != nil {
		return fmt.Errorf("invalid product ID")
	}

	cart, err := s.repoCart.GetCartByTeacherStudent(ctx, teacherID, studentID)

	if err != nil {
		return fmt.Errorf("failed to get cart: %w", err)
	}

	var quantity int

	for _, item := range cart.Items {
		if item.ProductID == id {
			quantity = item.Quantity
			break
		}
	}

	if quantity == 0 {
		return fmt.Errorf("product not found in cart")
	}

	if err = s.repoHistory.AddCartHistory(ctx, teacherID, studentID, id, "remove", quantity); err != nil {
		return fmt.Errorf("unable to add cart history: %w", err)
	}

	return s.repoCart.RemoveFromCart(ctx, teacherID, studentID, id)
}

func (s *cartService) ClearCart(ctx context.Context, teacherID string) error {

	err := s.repoHistory.AddAllCartHistory(ctx, teacherID)
	if err != nil {
		return fmt.Errorf("unable to add all cart history: %w", err)
	}

	return s.repoCart.ClearCart(ctx, teacherID)
}

func (s *cartService) CheckOutCart(ctx context.Context, req *models.CheckOutCartRequest) error {

	if req.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if req.Types == "" {
		return fmt.Errorf("type cannot be empty")
	}

	if req.Street == "" {
		return fmt.Errorf("street cannot be empty")
	}

	if req.City == "" {
		return fmt.Errorf("city cannot be empty")
	}

	if req.Country == "" {
		return fmt.Errorf("country cannot be empty")
	}

	if req.Phone == "" {
		return fmt.Errorf("phone cannot be empty")
	}

	response, err := s.orderAPI.CreateOrderByUserID(ctx, req.TeacherID, req.Types, req.Email, req.Street, req.City, req.Country, req.Phone, req.State)
	if err != nil {
		return fmt.Errorf("failed to create order: %v", err)
	}

	if respMap, ok := response.(map[string]interface{}); ok {
		if statusCode, exists := respMap["status_code"].(float64); exists && statusCode >= 400 {
			errorMsg := respMap["error"]
			errorCode := respMap["error_code"]
			return fmt.Errorf("API Error: %v, Code: %v, Status: %v", errorMsg, errorCode, statusCode)
		}
	}

	if err := s.ClearCart(ctx, req.TeacherID); err != nil {
		return fmt.Errorf("order created, but failed to clear cart: %v", err)
	}

	return nil
}

func NewServiceAPI(client *api.Client, serviceName string) *callAPI {
	sd, err := consul.NewServiceDiscovery(client, serviceName)
	if err != nil {
		fmt.Printf("Error creating service discovery: %v\n", err)
		return nil
	}

	service, err := sd.DiscoverService()
	if err != nil {
		fmt.Printf("Error discovering service: %v\n", err)
		return nil
	}

	return &callAPI{
		client:       sd,
		clientServer: service,
	}
}

func (c *callAPI) GetProductByID(productID string) map[string]interface{} {
	endpoint := fmt.Sprintf("/api/v1/products/%s", productID)
	res, err := c.client.CallAPI(c.clientServer, endpoint, http.MethodGet, nil, nil)
	if err != nil {
		fmt.Printf("Error calling API: %v\n", err)
		return nil
	}

	var productData interface{}
	json.Unmarshal([]byte(res), &productData)

	if productData == nil {
		fmt.Println("Product data is nil")
		return nil
	}

	myMap := productData.(map[string]interface{})

	return myMap
}

func (c *callAPI) CreateOrderByUserID(ctx context.Context, userID, types, email, street, city, country, phone string, state *string) (interface{}, error) {

	requestBody := map[string]string{
		"teacher_id": userID,
		"email":      email,
		"types":      types,
		"street":     street,
		"city":       city,
		"country":    country,
		"phone":      phone,
	}

	if state != nil {
		requestBody["state"] = *state
	}

	// Chuyển đổi dữ liệu thành JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Thiết lập headers
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + ctx.Value(constants.TokenKey).(string),
	}

	// Gọi API sử dụng phương thức POST
	endpoint := "/api/v1/orders/items"
	res, err := c.client.CallAPI(c.clientServer, endpoint, http.MethodPost, jsonData, headers)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}
	
	// Xử lý kết quả trả về
	var responseData interface{}
	err = json.Unmarshal([]byte(res), &responseData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return responseData, nil
}

func (s *cartService) GetCartHistoryByTeacher(ctx context.Context, teacherID string) ([]bson.M, error) {
	return s.repoCart.GetCartHistoryByTeacher(ctx, teacherID)
}
