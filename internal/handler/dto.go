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

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

// CreateFlowerRequest 创建鲜花请求
type CreateFlowerRequest struct {
	SKU           string  `json:"sku"`
	Name          string  `json:"name"`
	Origin        string  `json:"origin"`
	ShelfLife     string  `json:"shelf_life"`
	Preservation  string  `json:"preservation"`
	PurchasePrice float64 `json:"purchase_price"`
	SalePrice     float64 `json:"sale_price"`
	Stock         int     `json:"stock"`
}

// UpdateFlowerRequest 更新鲜花请求
type UpdateFlowerRequest struct {
	Name          *string  `json:"name,omitempty"`
	Origin        *string  `json:"origin,omitempty"`
	ShelfLife     *string  `json:"shelf_life,omitempty"`
	Preservation  *string  `json:"preservation,omitempty"`
	PurchasePrice *float64 `json:"purchase_price,omitempty"`
	SalePrice     *float64 `json:"sale_price,omitempty"`
}

// CreateAddressRequest 创建地址请求
type CreateAddressRequest struct {
	Label   *string `json:"label,omitempty"`
	Address string  `json:"address"`
	Contact string  `json:"contact"`
}

// UpdateAddressRequest 更新地址请求
type UpdateAddressRequest struct {
	Label   *string `json:"label,omitempty"`
	Address *string `json:"address,omitempty"`
	Contact *string `json:"contact,omitempty"`
}

// AddStockRequest 进货入库请求
type AddStockRequest struct {
	Quantity int `json:"quantity"`
}
