package handler

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	AddressID int                      `json:"address_id"`
	Items     []*CreateOrderItemRequest `json:"items"`
}

// CreateOrderItemRequest 创建订单项请求
type CreateOrderItemRequest struct {
	FlowerSKU string `json:"flower_sku"`
	Quantity  int    `json:"quantity"`
}
