package flower

import (
	"fmt"
	"math"
)

// Decimal 表示一个精确的十进制数，内部存储为"分"（整数）
// 用于处理金额等需要精确计算的场景
type Decimal struct {
	Value int64 // 存储为分（整数）
}

// DecimalFromFloat64 从 float64 创建 Decimal，自动四舍五入到分
func DecimalFromFloat64(f float64) Decimal {
	return Decimal{Value: roundToCent(f)}
}

// roundToCent 将 float64 四舍五入到分（2位小数）
func roundToCent(f float64) int64 {
	return int64(math.Round(f*100) / 100 * 100)
}

// ToFloat64 将 Decimal 转换为 float64
func (d Decimal) ToFloat64() float64 {
	return float64(d.Value) / 100
}

// Add 返回 d + other
func (d Decimal) Add(other Decimal) Decimal {
	return Decimal{Value: d.Value + other.Value}
}

// Sub 返回 d - other
func (d Decimal) Sub(other Decimal) Decimal {
	return Decimal{Value: d.Value - other.Value}
}

// Mul 返回 d * factor（整数乘法）
func (d Decimal) Mul(factor int64) Decimal {
	return Decimal{Value: d.Value * factor}
}

// Cmp 比较两个 Decimal
// 返回值：-1 表示 d < other, 0 表示 d == other, 1 表示 d > other
func (d Decimal) Cmp(other Decimal) int {
	if d.Value < other.Value {
		return -1
	}
	if d.Value > other.Value {
		return 1
	}
	return 0
}

// Equal 返回 d == other
func (d Decimal) Equal(other Decimal) bool {
	return d.Value == other.Value
}

// GreaterThan 返回 d > other
func (d Decimal) GreaterThan(other Decimal) bool {
	return d.Value > other.Value
}

// LessThan 返回 d < other
func (d Decimal) LessThan(other Decimal) bool {
	return d.Value < other.Value
}

// String 返回 Decimal 的字符串表示，格式化为两位小数
func (d Decimal) String() string {
	sign := ""
	if d.Value < 0 {
		sign = "-"
		d.Value = -d.Value
	}
	yuan := d.Value / 100
	cents := d.Value % 100
	return fmt.Sprintf("%s%d.%02d", sign, yuan, cents)
}
