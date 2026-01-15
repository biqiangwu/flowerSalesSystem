package address

import (
	"context"
	"fmt"
)

// AddressService 定义地址业务逻辑接口
type AddressService interface {
	CreateAddress(ctx context.Context, userID int, req *CreateAddressRequest) error
	GetAddress(ctx context.Context, userID, id int) (*AddressResponse, error)
	ListAddresses(ctx context.Context, userID int) ([]*AddressResponse, error)
	UpdateAddress(ctx context.Context, userID, id int, req *UpdateAddressRequest) error
	DeleteAddress(ctx context.Context, userID, id int) error
}

// CreateAddressRequest 创建地址请求
type CreateAddressRequest struct {
	Label   *string
	Address string
	Contact string
}

// UpdateAddressRequest 更新地址请求
type UpdateAddressRequest struct {
	Label   *string
	Address *string
	Contact *string
}

// AddressResponse 地址响应
type AddressResponse struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Label     string `json:"label"`
	Address   string `json:"address"`
	Contact   string `json:"contact"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// addressService 实现 AddressService 接口
type addressService struct {
	repo AddressRepository
}

// NewAddressService 创建 AddressService 实例
func NewAddressService(repo AddressRepository) AddressService {
	return &addressService{repo: repo}
}

// CreateAddress 创建地址
func (s *addressService) CreateAddress(ctx context.Context, userID int, req *CreateAddressRequest) error {
	// 创建 Address 实体
	address := &Address{
		UserID:  userID,
		Address: req.Address,
		Contact: req.Contact,
	}
	if req.Label != nil {
		address.Label = *req.Label
	}

	// 验证数据
	if err := address.Validate(); err != nil {
		return err
	}

	// 保存到数据库
	return s.repo.Create(ctx, address)
}

// GetAddress 获取地址详情（验证用户权限）
func (s *addressService) GetAddress(ctx context.Context, userID, id int) (*AddressResponse, error) {
	address, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 验证用户只能访问自己的地址
	if address.UserID != userID {
		return nil, fmt.Errorf("无权访问该地址")
	}

	return s.toResponse(address), nil
}

// ListAddresses 获取用户地址列表
func (s *addressService) ListAddresses(ctx context.Context, userID int) ([]*AddressResponse, error) {
	addresses, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]*AddressResponse, len(addresses))
	for i, a := range addresses {
		responses[i] = s.toResponse(a)
	}

	return responses, nil
}

// UpdateAddress 更新地址信息（验证用户权限）
func (s *addressService) UpdateAddress(ctx context.Context, userID, id int, req *UpdateAddressRequest) error {
	// 获取现有地址
	address, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 验证用户只能操作自己的地址
	if address.UserID != userID {
		return fmt.Errorf("无权操作该地址")
	}

	// 更新字段
	if req.Label != nil {
		address.Label = *req.Label
	}
	if req.Address != nil {
		address.Address = *req.Address
	}
	if req.Contact != nil {
		address.Contact = *req.Contact
	}

	// 验证更新后的数据
	if err := address.Validate(); err != nil {
		return err
	}

	// 保存到数据库
	return s.repo.Update(ctx, address)
}

// DeleteAddress 删除地址（验证用户权限）
func (s *addressService) DeleteAddress(ctx context.Context, userID, id int) error {
	// 获取现有地址
	address, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 验证用户只能操作自己的地址
	if address.UserID != userID {
		return fmt.Errorf("无权操作该地址")
	}

	// 删除地址
	return s.repo.Delete(ctx, id)
}

// toResponse 将 Address 实体转换为响应 DTO
func (s *addressService) toResponse(a *Address) *AddressResponse {
	label := a.Label
	if label == "" {
		label = "未命名"
	}

	return &AddressResponse{
		ID:        a.ID,
		UserID:    a.UserID,
		Label:     label,
		Address:   a.Address,
		Contact:   a.Contact,
		CreatedAt: a.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: a.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
