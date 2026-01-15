package address

import (
	"fmt"
	"time"
)

// Address 表示收货地址实体
type Address struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Label     string    `json:"label"`      // 可选，如：家、公司
	Address   string    `json:"address"`    // 详细地址
	Contact   string    `json:"contact"`    // 联系方式（电话/微信）
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewAddress 创建一个新的地址实体
func NewAddress(userID int, label, address, contact string) *Address {
	now := time.Now()
	return &Address{
		UserID:    userID,
		Label:     label,
		Address:   address,
		Contact:   contact,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate 验证地址数据是否有效
func (a *Address) Validate() error {
	if a.UserID <= 0 {
		return fmt.Errorf("用户ID不能为空")
	}
	if a.Address == "" {
		return fmt.Errorf("地址不能为空")
	}
	if a.Contact == "" {
		return fmt.Errorf("联系方式不能为空")
	}
	if len(a.Address) < 5 {
		return fmt.Errorf("地址长度不能少于5个字符")
	}
	if len(a.Contact) < 5 {
		return fmt.Errorf("联系方式长度不能少于5个字符")
	}
	if len(a.Address) > 255 {
		return fmt.Errorf("地址长度不能超过255个字符")
	}
	if len(a.Contact) > 50 {
		return fmt.Errorf("联系方式长度不能超过50个字符")
	}
	return nil
}
