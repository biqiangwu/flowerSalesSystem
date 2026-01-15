package address

import (
	"testing"
)

func TestAddressValidation(t *testing.T) {
	tests := []struct {
		name    string
		address Address
		wantErr bool
		errMsg  string
	}{
		{
			name: "有效的完整地址",
			address: Address{
				UserID:  1,
				Label:   "家",
				Address: "北京市朝阳区xxx街道xxx号",
				Contact: "张三，13800138000",
			},
			wantErr: false,
		},
		{
			name: "有效的地址-无标签",
			address: Address{
				UserID:  1,
				Address: "北京市朝阳区xxx街道xxx号",
				Contact: "张三，13800138000",
			},
			wantErr: false,
		},
		{
			name: "无效-用户ID为0",
			address: Address{
				UserID:  0,
				Address: "北京市朝阳区xxx街道xxx号",
				Contact: "张三，13800138000",
			},
			wantErr: true,
			errMsg:  "用户ID不能为空",
		},
		{
			name: "无效-地址为空",
			address: Address{
				UserID:  1,
				Address: "",
				Contact: "张三，13800138000",
			},
			wantErr: true,
			errMsg:  "地址不能为空",
		},
		{
			name: "无效-联系方式为空",
			address: Address{
				UserID:  1,
				Address: "北京市朝阳区xxx街道xxx号",
				Contact: "",
			},
			wantErr: true,
			errMsg:  "联系方式不能为空",
		},
		{
			name: "无效-地址过短",
			address: Address{
				UserID:  1,
				Address: "abc",
				Contact: "张三，13800138000",
			},
			wantErr: true,
			errMsg:  "地址长度不能少于5个字符",
		},
		{
			name: "无效-联系方式过短",
			address: Address{
				UserID:  1,
				Address: "北京市朝阳区xxx街道xxx号",
				Contact: "123",
			},
			wantErr: true,
			errMsg:  "联系方式长度不能少于5个字符",
		},
		{
			name: "无效-地址过长",
			address: Address{
				UserID:  1,
				Address: string(make([]byte, 256)), // 256个字符
				Contact: "张三，13800138000",
			},
			wantErr: true,
			errMsg:  "地址长度不能超过255个字符",
		},
		{
			name: "无效-联系方式过长",
			address: Address{
				UserID:  1,
				Address: "北京市朝阳区xxx街道xxx号",
				Contact: string(make([]byte, 51)), // 51个字符
			},
			wantErr: true,
			errMsg:  "联系方式长度不能超过50个字符",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.address.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Address.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Address.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestNewAddress(t *testing.T) {
	address := NewAddress(1, "家", "北京市朝阳区xxx街道xxx号", "张三，13800138000")

	if address.UserID != 1 {
		t.Errorf("NewAddress() UserID = %v, want %v", address.UserID, 1)
	}
	if address.Label != "家" {
		t.Errorf("NewAddress() Label = %v, want %v", address.Label, "家")
	}
	if address.Address != "北京市朝阳区xxx街道xxx号" {
		t.Errorf("NewAddress() Address = %v, want %v", address.Address, "北京市朝阳区xxx街道xxx号")
	}
	if address.Contact != "张三，13800138000" {
		t.Errorf("NewAddress() Contact = %v, want %v", address.Contact, "张三，13800138000")
	}
	if address.CreatedAt.IsZero() {
		t.Errorf("NewAddress() CreatedAt should be set")
	}
	if address.UpdatedAt.IsZero() {
		t.Errorf("NewAddress() UpdatedAt should be set")
	}
}
