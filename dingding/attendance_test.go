package dingding

import (
	"ding/model/dingding"
	"reflect"
	"testing"
)

func TestGetGroupDeptNumber(t *testing.T) {
	type args struct {
		token   string
		groupId int
	}
	tests := []struct {
		name          string
		args          args
		wantDeptUsers map[string][]dingding.DingUser
	}{
		// TODO: Add test cases.
		{args: args{
			token:   "58f17416ab6437789a23209cbef4ca19",
			groupId: 952645016,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDeptUsers := GetGroupDeptNumber(tt.args.token, tt.args.groupId); !reflect.DeepEqual(gotDeptUsers, tt.wantDeptUsers) {
				t.Errorf("GetGroupDeptNumber() = %v, want %v", gotDeptUsers, tt.wantDeptUsers)
			}
		})
	}
}
