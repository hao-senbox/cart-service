package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"store/internal/models"
	"store/internal/service"
	"store/pkg/constants"

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

	adminCartGroup := r.Group("/api/v1/admin/cart").Use(Secured())
	{
		adminCartGroup.GET("", handlers.GetAllCartGroupedByTeacher)
		adminCartGroup.GET("/history", handlers.GetCartHistoryByTeacher)
	}

	cartGroup := r.Group("/api/v1/cart").Use(Secured())
	{
		cartGroup.GET("/items", handlers.GetCart)
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

	teacherID, exists := c.Get(constants.UserID)
	if !exists {
		SendError(c, http.StatusBadRequest, fmt.Errorf("user ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	if teacherID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("teacher ID cannot be empty"), models.ErrInvalidRequest)
		return
	}
	
	cart, err := h.cartService.GetCartByTeacher(c.Request.Context(), teacherID.(string))

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

	teacherID, exists := c.Get(constants.UserID)
	if !exists {
		SendError(c, http.StatusBadRequest, fmt.Errorf("user ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	if teacherID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("teacher ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	req.TeacherID = teacherID.(string)

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

	teacherID, exists := c.Get(constants.UserID)
	if !exists {
		SendError(c, http.StatusBadRequest, fmt.Errorf("user ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	if teacherID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("teacher ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	req.TeacherID = teacherID.(string)

	if req.StudentID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("student ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	if req.Type != "" && req.Type != "increase" && req.Type != "decrease" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("type must be 'increase', 'decrease', or empty"), models.ErrInvalidRequest)
		return
	}

	if req.Type == "" && *req.Quantity <= 0 {
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

	teacherID, exists := c.Get(constants.UserID)
	if !exists {
		SendError(c, http.StatusBadRequest, fmt.Errorf("user ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	if teacherID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("teacher ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	req.TeacherID = teacherID.(string)

	err := h.cartService.RemoveFromCart(c.Request.Context(), req.TeacherID, req.StudentID, productID)

	if err != nil {
		SendError(c, http.StatusInternalServerError, err, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Product deleted successfully", nil)																														
}

func (h *CartHandlers) ClearCart(c *gin.Context) {

	teacherID, exists := c.Get(constants.UserID)
	if !exists {
		SendError(c, http.StatusBadRequest, fmt.Errorf("user ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	if teacherID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("teacher ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	TeacherID := teacherID.(string)
	
	err := h.cartService.ClearCart(c.Request.Context(), TeacherID)

	if err != nil {
		SendError(c, http.StatusInternalServerError, err, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Cart deleted successfully", nil)

}

func (h *CartHandlers) CheckOutCart(c *gin.Context) {
	
	var req models.CheckOutCartRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidRequest)
		return 
	}

	teacherID, exists := c.Get(constants.UserID)
	if !exists {
		SendError(c, http.StatusBadRequest, fmt.Errorf("user ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	if teacherID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("teacher ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	token, ok := c.Get(constants.Token)
	if !ok {
		SendError(c, http.StatusForbidden, errors.New("unauthorized"), models.ErrInvalidRequest)
		return
	}

	ctx := context.WithValue(c, constants.TokenKey, token)
	
	req.TeacherID = teacherID.(string)

	err := h.cartService.CheckOutCart(ctx, &req)

	if err != nil {
		SendError(c, http.StatusInternalServerError, err, models.ErrInvalidOperation)
		return
	} 

	SendSuccess(c, http.StatusOK, "Checkout successfully", nil)
}

func (h *CartHandlers) GetCartHistoryByTeacher(c *gin.Context) {
	
	var req models.UserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidRequest)
		return 
	}


	cartHistory, err := h.cartService.GetCartHistoryByTeacher(c, req.TeacherID)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Cart history data of teacher retrieved successfully", cartHistory)

}