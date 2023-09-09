package utils

import (
	"reflect"
	"testing"
)

func TestDiffArray(t *testing.T) {
	type args struct {
		a []int
		b []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "测试",
			args: args{a: []int{1, 2, 3}, b: []int{2, 3, 4}},
		},
		{
			name: "测试2",
			args: args{a: []int{2, 3, 4}, b: []int{1, 2, 3}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DiffArray(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DiffArray() = %v, want %v", got, tt.want)
			}
		})
	}
}
