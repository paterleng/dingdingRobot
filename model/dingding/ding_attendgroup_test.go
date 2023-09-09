package dingding

import (
	"reflect"
	"testing"
)

func TestTimeTransFrom(t *testing.T) {
	//TimeTransFrom()
}

func Test_timestampTran(t *testing.T) {
	type args struct {
		format string
		t      int64
	}
	tests := []struct {
		name  string
		args  args
		wantS string
	}{
		// TODO: Add test cases.
		{"case1", args{format: "2006:01:02 15:04:05", t: 1672976808000}, ""},
		{"case1", args{format: "15:04:05", t: 1672976808000}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotS := timestampTran(tt.args.format, tt.args.t); gotS != tt.wantS {
				t.Errorf("timestampTran() = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}

func TestDingAttendGroup_GetGroupDeptNumber(t *testing.T) {
	type fields struct {
		GroupId       int
		GroupName     string
		MemberCount   int
		WorkDayList   []string
		ClassesList   []string
		SelectedClass []struct {
			Setting struct {
				PermitLateMinutes int `json:"permit_late_minutes"` //允许迟到时长
			} `gorm:"-" json:"setting"`
			Sections []struct {
				Times []struct {
					CheckTime string `json:"check_time"` //打卡时间
					CheckType string `json:"check_type"` //打卡类型
				} `gorm:"-" json:"times"`
			} `gorm:"-" json:"sections"`
		}
		DingToken         DingToken
		IsRobotAttendance bool
		RobotAttendTaskID int
		IsSendFirstPerson int
	}
	tests := []struct {
		name          string
		fields        fields
		wantDeptUsers map[string][]DingUser
		wantErr       bool
	}{
		// TODO: Add test cases.
		{
			name:   "测试",
			fields: fields{GroupId: 952645016, DingToken: DingToken{Token: "c629a0352ef53cedaa7a8ee3f64a9540"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &DingAttendGroup{
				GroupId:           tt.fields.GroupId,
				GroupName:         tt.fields.GroupName,
				MemberCount:       tt.fields.MemberCount,
				WorkDayList:       tt.fields.WorkDayList,
				ClassesList:       tt.fields.ClassesList,
				SelectedClass:     tt.fields.SelectedClass,
				DingToken:         tt.fields.DingToken,
				IsRobotAttendance: tt.fields.IsRobotAttendance,
				RobotAttendTaskID: tt.fields.RobotAttendTaskID,
				IsSendFirstPerson: tt.fields.IsSendFirstPerson,
			}
			gotDeptUsers, err := a.GetGroupDeptNumber()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGroupDeptNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDeptUsers, tt.wantDeptUsers) {
				t.Errorf("GetGroupDeptNumber() gotDeptUsers = %v, want %v", gotDeptUsers, tt.wantDeptUsers)
			}
		})
	}
}
