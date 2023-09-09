package localTime

import (
	"testing"
	"time"
)

func TestMySelfTime_StringToStamp(t1 *testing.T) {
	type fields struct {
		TimeStamp int64
		Format    string
		Time      time.Time
		Duration  string
	}
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "测试",
			args: args{
				s: "2023-01-31 20:29:00",
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &MySelfTime{
				TimeStamp: tt.fields.TimeStamp,
				Format:    tt.fields.Format,
				Time:      tt.fields.Time,
				Duration:  tt.fields.Duration,
			}
			got, err := t.StringToStamp(tt.args.s)
			if (err != nil) != tt.wantErr {
				t1.Errorf("StringToStamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t1.Errorf("StringToStamp() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMySelfTime_StampToString(t1 *testing.T) {
	type fields struct {
		TimeStamp int64
		Format    string
		Time      time.Time
		Duration  string
	}
	type args struct {
		s int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "测试",
			args: args{
				s: 1675470600000,
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &MySelfTime{
				TimeStamp: tt.fields.TimeStamp,
				Format:    tt.fields.Format,
				Time:      tt.fields.Time,
				Duration:  tt.fields.Duration,
			}
			if got := t.StampToString(tt.args.s); got != tt.want {
				t1.Errorf("StampToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
