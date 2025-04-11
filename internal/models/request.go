package models

type AddToCartRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	TeacherID string `json:"teacher_id" validate:"required"`
	StudentID string `json:"student_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,min=1"`
}

type UserRequest struct {
	TeacherID string `json:"teacher_id" validate:"required"`
	StudentID string `json:"student_id" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Types string `json:"types" validate:"required,oneof=cod bank_transfer"`
}

type UpdateCartItemRequest struct {
	Quantity  int    `json:"quantity" validate:"required,min=1"`
	Type      string `json:"types" validate:"required,oneof=increase decrease"`
	TeacherID string `json:"teacher_id" validate:"required"`
	StudentID string `json:"student_id" validate:"required"`
}