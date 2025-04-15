package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/api"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"net/http"
	"store/internal/models"
	"store/internal/repository"
	"store/pkg/consul"
	"time"
)

type CartService interface {
	AddToCart(ctx context.Context, req *models.AddToCartRequest) (*models.CartItem, error)
	GetCartByTeacher(ctx context.Context, teacherID string) ([]bson.M, error)
	GetAllCartGroupedByTeacher(ctx context.Context) ([]bson.M, error)
	UpdateQuantityItem(ctx context.Context, productID string, req *models.UpdateCartItemRequest) error
	RemoveFromCart(ctx context.Context, teacherID string, studentID string, productID string) error
	ClearCart(ctx context.Context, teacherID string) error
	CheckOutCart(ctx context.Context, req *models.CheckOutCartRequest) error
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
	productID, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID format: %v", err)
	}

	// productRes := s.productAPI.GetProductByID(req.ProductID)

	// if productRes == nil {
	// 	return nil, fmt.Errorf("product not found")
	// }

	// product := productRes["data"].(map[string]interface{})

	// name := product["product_name"].(string)
	// price := product["original_price"].(float64)
	// imageURL := product["cover_image"].(string)
	// topic := product["topic_name"].(string)

	sampleProducts := []struct {
		Name     string
		Price    float64
		ImageURL string
	}{
		{"High-Performance Gaming Laptop with RTX Graphics", 1499.99, "https://example.com/images/laptop.jpg"},
		{"Mechanical RGB Backlit Keyboard for Gaming and Office", 79.99, "https://example.com/images/keyboard.jpg"},
		{"Ergonomic Wireless Mouse with Adjustable DPI Settings", 39.99, "https://example.com/images/mouse.jpg"},
		{"27-Inch 4K Ultra HD Monitor with HDR Support", 299.99, "https://example.com/images/monitor.jpg"},
		{"All-in-One Wireless Color Printer with Scanner", 189.99, "https://example.com/images/printer.jpg"},
		{"Latest Generation Smartphone with 5G and Triple Camera", 999.99, "https://example.com/images/smartphone.jpg"},
		{"10.1-Inch Android Tablet with Stylus Support", 349.99, "https://example.com/images/tablet.jpg"},
		{"Fitness Smartwatch with Heart Rate and GPS Tracker", 149.99, "https://example.com/images/smartwatch.jpg"},
		{"Noise-Cancelling Over-Ear Headphones with Deep Bass", 119.99, "https://example.com/images/headphones.jpg"},
		{"Portable Bluetooth Speaker with Waterproof Design", 69.99, "https://example.com/images/speaker.jpg"},
		{"1080p Full HD Webcam with Built-in Microphone", 49.99, "https://example.com/images/webcam.jpg"},
		{"1TB USB 3.0 External Hard Drive for Backup and Storage", 109.99, "https://example.com/images/hdd.jpg"},
		{"64GB USB Flash Drive with High-Speed File Transfer", 24.99, "https://example.com/images/usb.jpg"},
		{"Ergonomic Gaming Chair with Adjustable Armrests", 259.99, "https://example.com/images/gaming-chair.jpg"},
		{"NVIDIA RTX 4070 Graphics Card with 12GB GDDR6 Memory", 599.99, "https://example.com/images/gpu.jpg"},
		{"ATX Motherboard for Intel Processors with WiFi Support", 179.99, "https://example.com/images/motherboard.jpg"},
		{"16GB DDR4 RAM Kit (2x8GB) for Desktop Computers", 74.99, "https://example.com/images/ram.jpg"},
		{"750W Modular Power Supply with 80+ Gold Certification", 129.99, "https://example.com/images/psu.jpg"},
		{"Adjustable LED Desk Lamp with USB Charging Port", 39.99, "https://example.com/images/desk-lamp.jpg"},
		{"Dual-Band Wireless Router with Parental Controls", 109.99, "https://example.com/images/router.jpg"},
	}

	now := time.Now()
	// Tạo seed cho random
	rand.Seed(now.UnixNano())

	// Lấy ngẫu nhiên 1 sản phẩm
	randomIndex := rand.Intn(len(sampleProducts))
	selected := sampleProducts[randomIndex]

	cartItem := &models.CartItem{
		ProductID:   productID,
		Quantity:    req.Quantity,
		ProductName: selected.Name,
		Price:       selected.Price,
		ImageURL:    selected.ImageURL,
	}

	// cartItem := &models.CartItem{
	// 	ProductID:   productID,
	// 	Quantity:    req.Quantity,
	// 	ProductName: name,
	// 	TopicName: topic,
	// 	Price:       float64(price),
	// 	ImageURL:    imageURL,
	// }

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

	id, err := primitive.ObjectIDFromHex(productID)

	productRes := s.productAPI.GetProductByID(productID)

	if productRes == nil {
		return fmt.Errorf("product not found")
	}

	product := productRes["data"].(map[string]interface{})

	name := product["product_name"].(string)
	price := product["original_price"].(float64)
	imageURL := product["cover_image"].(string)
	topic := product["topic"].(map[string]interface{})
	topicName := topic["topic_name"].(string)

	cartItem := &models.CartItem{
		ProductID:   id,
		Quantity:    1,
		ProductName: name,
		TopicName:   topicName,
		Price:       float64(price),
		ImageURL:    imageURL,
	}

	if err != nil {
		return fmt.Errorf("invalid product ID format: %w", err)
	}

	return s.repoCart.UpdateCartItemQuantity(ctx, req.TeacherID, req.StudentID, id, req.Quantity, req.Type, *cartItem)
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

	response, err := s.orderAPI.CreateOrderByUserID(req.TeacherID, req.Types, req.Email, req.Street, req.City, req.Country, req.Phone, req.State)
	if err != nil {
		return fmt.Errorf("failed to create order: %v", err)
	}
	
	// Có thể log để debug
	fmt.Printf("Response from API: %+v\n", response)

	// Kiểm tra response có chứa thông tin lỗi không
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

func (c *callAPI) CreateOrderByUserID(userID, types, email, street, city, country, phone string, state *string) (interface{}, error) {
	// Tạo dữ liệu body cho request POST
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
		"Content-Type": "application/json",
	}

	// Gọi API sử dụng phương thức POST
	endpoint := "/api/orders/items"
	res, err := c.client.CallAPI(c.clientServer, endpoint, http.MethodPost, jsonData, headers)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}

	// In ra response để debug
	fmt.Printf("Raw API response: %s\n", res)

	// Xử lý kết quả trả về
	var responseData interface{}
	err = json.Unmarshal([]byte(res), &responseData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return responseData, nil
}
