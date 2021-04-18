package tinyurl

import (
	"math"
	"math/big"
	"reflect"
	"testing"
)

func TestBase58_Encode(t *testing.T) {
	type fields struct {
		Value big.Int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"", fields{*big.NewInt(-1)}, ""},
		{"", fields{*big.NewInt(0)}, "1"},
		{"", fields{*big.NewInt(1)}, "2"},
		{"", fields{*big.NewInt(10)}, "B"},
		{"", fields{*big.NewInt(57)}, "z"},
		{"", fields{*big.NewInt(58)}, "21"},
		{"", fields{*big.NewInt(math.MaxInt16)}, "Ajx"},
		{"", fields{*big.NewInt(math.MaxInt32)}, "4GmR58"},
		{"", fields{*big.NewInt(math.MaxInt64)}, "NQm6nKp8qFC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Base58{
				Value: &tt.fields.Value,
			}
			if got := b.Encode(); got != tt.want {
				t.Errorf("Base58.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBase58_Decode(t *testing.T) {
	type fields struct {
		Value big.Int
	}
	type args struct {
		text string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *big.Int
	}{
		{"", fields{*big.NewInt(-1)}, args{""}, big.NewInt(-1)},
		{"", fields{*big.NewInt(0)}, args{"1"}, big.NewInt(0)},
		{"", fields{*big.NewInt(1)}, args{"2"}, big.NewInt(1)},
		{"", fields{*big.NewInt(10)}, args{"B"}, big.NewInt(10)},
		{"", fields{*big.NewInt(57)}, args{"z"}, big.NewInt(57)},
		{"", fields{*big.NewInt(58)}, args{"21"}, big.NewInt(58)},
		{"", fields{*big.NewInt(math.MaxInt16)}, args{"Ajx"}, big.NewInt(math.MaxInt16)},
		{"", fields{*big.NewInt(math.MaxInt32)}, args{"4GmR58"}, big.NewInt(math.MaxInt32)},
		{"", fields{*big.NewInt(math.MaxInt64)}, args{"NQm6nKp8qFC"}, big.NewInt(math.MaxInt64)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Base58{
				Value: &tt.fields.Value,
			}
			if got := b.Decode(tt.args.text); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Base58.Decode(%v) = %v, want %v", tt.args.text, got, tt.want)
			}
		})
	}
}
