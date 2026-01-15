package flower

import (
	"testing"
)

func TestDecimalFromFloat64(t *testing.T) {
	tests := []struct {
		name    string
		input   float64
		want    int64
		wantErr bool
	}{
		{"正整数", 100.00, 10000, false},
		{"正小数一位", 100.5, 10050, false},
		{"正小数两位", 100.55, 10055, false},
		{"零值", 0, 0, false},
		{"负数", -50.25, -5025, false},
		{"多位小数-四舍五入", 100.555, 10056, false},
		{"多位小数-向下舍", 100.554, 10055, false},
		{"多位小数-向上入", 100.556, 10056, false},
		{"非常大的数字", 999999.99, 99999999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecimalFromFloat64(tt.input)
			if got.Value != tt.want {
				t.Errorf("DecimalFromFloat64(%v) = %v, want %v", tt.input, got.Value, tt.want)
			}
		})
	}
}

func TestDecimalToFloat64(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  float64
	}{
		{"正整数", 10000, 100.00},
		{"正小数一位", 10050, 100.50},
		{"正小数两位", 10055, 100.55},
		{"零值", 0, 0},
		{"负数", -5025, -50.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Decimal{Value: tt.input}
			got := d.ToFloat64()
			if got != tt.want {
				t.Errorf("Decimal.ToFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimalAdd(t *testing.T) {
	tests := []struct {
		name  string
		a     int64
		b     int64
		want  int64
	}{
		{"两个正数", 10000, 5000, 15000},
		{"正数加零", 10000, 0, 10000},
		{"零加正数", 0, 10000, 10000},
		{"正数加负数", 10000, -3000, 7000},
		{"两个负数", -5000, -3000, -8000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Decimal{Value: tt.a}
			b := Decimal{Value: tt.b}
			got := a.Add(b)
			if got.Value != tt.want {
				t.Errorf("Decimal.Add() = %v, want %v", got.Value, tt.want)
			}
		})
	}
}

func TestDecimalSub(t *testing.T) {
	tests := []struct {
		name  string
		a     int64
		b     int64
		want  int64
	}{
		{"正数减正数", 10000, 3000, 7000},
		{"正数减零", 10000, 0, 10000},
		{"零减正数", 0, 10000, -10000},
		{"正数减负数", 10000, -3000, 13000},
		{"负数减正数", -5000, 3000, -8000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Decimal{Value: tt.a}
			b := Decimal{Value: tt.b}
			got := a.Sub(b)
			if got.Value != tt.want {
				t.Errorf("Decimal.Sub() = %v, want %v", got.Value, tt.want)
			}
		})
	}
}

func TestDecimalMul(t *testing.T) {
	tests := []struct {
		name  string
		a     int64
		b     int64
		want  int64
	}{
		{"正数乘正数", 10000, 2, 20000},
		{"乘零", 10000, 0, 0},
		{"零乘正数", 0, 100, 0},
		{"正数乘负数", 10000, -2, -20000},
		{"负数乘负数", -10000, -2, 20000},
		{"计算小计-10件", 10000, 10, 100000}, // 100.00 * 10 = 1000.00
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Decimal{Value: tt.a}
			got := a.Mul(tt.b)
			if got.Value != tt.want {
				t.Errorf("Decimal.Mul() = %v, want %v", got.Value, tt.want)
			}
		})
	}
}

func TestDecimalCmp(t *testing.T) {
	tests := []struct {
		name string
		a    int64
		b    int64
		want int
	}{
		{"a大于b", 10000, 5000, 1},
		{"a等于b", 10000, 10000, 0},
		{"a小于b", 5000, 10000, -1},
		{"负数比较", -5000, -3000, -1},
		{"零与正数", 0, 10000, -1},
		{"零与负数", 0, -10000, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Decimal{Value: tt.a}
			b := Decimal{Value: tt.b}
			got := a.Cmp(b)
			if got != tt.want {
				t.Errorf("Decimal.Cmp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimalEqual(t *testing.T) {
	tests := []struct {
		name  string
		a     int64
		b     int64
		want  bool
	}{
		{"相等", 10000, 10000, true},
		{"不相等", 10000, 5000, false},
		{"与零相等", 0, 0, true},
		{"负数相等", -5000, -5000, true},
		{"负数不相等", -5000, -3000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Decimal{Value: tt.a}
			b := Decimal{Value: tt.b}
			got := a.Equal(b)
			if got != tt.want {
				t.Errorf("Decimal.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimalGreaterThan(t *testing.T) {
	tests := []struct {
		name  string
		a     int64
		b     int64
		want  bool
	}{
		{"a大于b", 10000, 5000, true},
		{"a等于b", 10000, 10000, false},
		{"a小于b", 5000, 10000, false},
		{"负数比较", -3000, -5000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Decimal{Value: tt.a}
			b := Decimal{Value: tt.b}
			got := a.GreaterThan(b)
			if got != tt.want {
				t.Errorf("Decimal.GreaterThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimalLessThan(t *testing.T) {
	tests := []struct {
		name  string
		a     int64
		b     int64
		want  bool
	}{
		{"a小于b", 5000, 10000, true},
		{"a等于b", 10000, 10000, false},
		{"a大于b", 10000, 5000, false},
		{"负数比较", -5000, -3000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Decimal{Value: tt.a}
			b := Decimal{Value: tt.b}
			got := a.LessThan(b)
			if got != tt.want {
				t.Errorf("Decimal.LessThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimalString(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{"正整数", 10000, "100.00"},
		{"正小数", 10055, "100.55"},
		{"零值", 0, "0.00"},
		{"负数", -5025, "-50.25"},
		{"一位小数", 10050, "100.50"},
		{"不足分位补零", 5, "0.05"},
		{"不足角位补零", 100, "1.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Decimal{Value: tt.input}
			got := d.String()
			if got != tt.want {
				t.Errorf("Decimal.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

