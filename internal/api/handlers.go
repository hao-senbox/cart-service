package api

import (
	"fmt"
	"net/http"
	"store/internal/models"
	"store/internal/service"
	"github.com/gin-gonic/gin"
)

type CartHandlers struct {
	cartService service.CartService
}

func NewCartHandlers(cartService service.CartService) *CartHandlers {
	return &CartHandlers{
		cartService: cartService,
	}
}

func RegisterHandlers(r *gin.Engine, cartService service.CartService) {

	handlers := NewCartHandlers(cartService)

	adminCartGroup := r.Group("/api/v1/admin/cart")
	{
		adminCartGroup.GET("", handlers.GetAllCartGroupedByTeacher)
	}

	cartGroup := r.Group("/api/v1/cart")
	{
		cartGroup.GET("/items/:teacher_id", handlers.GetCart)
		cartGroup.POST("/items", handlers.AddToCart)
		cartGroup.PUT("/items/:product_id", handlers.UpdateQuantity)
		cartGroup.DELETE("/items/:product_id", handlers.RemoveFromCart)
		cartGroup.DELETE("/items", handlers.ClearCart)
		cartGroup.POST("/items/checkout", handlers.CheckOutCart)
	}
}


func (h *CartHandlers) GetAllCartGroupedByTeacher(c *gin.Context) {

	cart, err := h.cartService.GetAllCartGroupedByTeacher(c.Request.Context())

	if err != nil {
		SendError(c, http.StatusInternalServerError, err, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Cart data retrieved successfully", cart)
}

func (h *CartHandlers) GetCart(c *gin.Context) {

	teacherID := c.Param("teacher_id")

	if teacherID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("teacher ID cannot be empty"), models.ErrInvalidRequest)
		return
	}
	
	cart, err := h.cartService.GetCartByTeacher(c.Request.Context(), teacherID)

	if err != nil {
		SendError(c, http.StatusInternalServerError, err, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Cart data of user retrieved successfully", cart)
}

func (h *CartHandlers) AddToCart(c *gin.Context) {

	var req models.AddToCartRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidRequest)
		return 
	}

	if req.TeacherID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("teacher ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	if req.StudentID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("student ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	if req.ProductID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("product ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	if req.Quantity <= 0 {
		SendError(c, http.StatusBadRequest, fmt.Errorf("quantity must be greater than 0"), models.ErrInvalidRequest)
		return
	}

	cartItem, err := h.cartService.AddToCart(c.Request.Context(), &req)

	if err != nil {
		SendError(c, http.StatusInternalServerError, err, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Product has been added to cart", cartItem)
}

func (h *CartHandlers) UpdateQuantity(c *gin.Context) {

	productID := c.Param("product_id")

	var req models.UpdateCartItemRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidRequest)
		return
	}

	if req.TeacherID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("teacher ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	if req.StudentID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("student ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	if req.Type != "" && req.Type != "increase" && req.Type != "decrease" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("type must be 'increase', 'decrease', or empty"), models.ErrInvalidRequest)
		return
	}

	if req.Type == "" && req.Quantity <= 0 {
		SendError(c, http.StatusBadRequest, fmt.Errorf("quantity must be greater than 0 when type is not specified"), models.ErrInvalidRequest)
		return
	}

	err := h.cartService.UpdateQuantityItem(c.Request.Context(), productID, &req)

	if err != nil {
		SendError(c, http.StatusInternalServerError, err, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Update successfully", nil)

}

func (h *CartHandlers) RemoveFromCart(c *gin.Context) {

	productID := c.Param("product_id")

	var req models.UserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidRequest)
		return
	}

	err := h.cartService.RemoveFromCart(c.Request.Context(), req.TeacherID, req.StudentID, productID)

	if err != nil {
		SendError(c, http.StatusInternalServerError, err, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Product deleted successfully", nil)																														
}

func (h *CartHandlers) ClearCart(c *gin.Context) {

	var req models.UserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidRequest)
		return
	}
	
	err := h.cartService.ClearCart(c.Request.Context(), req.TeacherID)

	if err != nil {
		SendError(c, http.StatusInternalServerError, err, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Cart deleted successfully", nil)

}

func (h *CartHandlers) CheckOutCart(c *gin.Context) {
	
	var req models.UserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidRequest)
		return 
	}

	err := h.cartService.CheckOutCart(c.Request.Context(), req.TeacherID, req.Types, req.Email)

	if err != nil {
		SendError(c, http.StatusInternalServerError, err, models.ErrInvalidOperation)
		return
	} 

	SendSuccess(c, http.StatusOK, "Checkout successfully", nil)
}