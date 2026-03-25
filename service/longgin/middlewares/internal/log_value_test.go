package internal

import "testing"

func Test_toSafeValue(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				v: "",
			},
			want: "",
		},
		{
			name: "1",
			args: args{
				v: "1",
			},
			want: "*",
		},
		{
			name: "2",
			args: args{
				v: "12",
			},
			want: "**",
		},
		{
			name: "3",
			args: args{
				v: "123",
			},
			want: "***",
		},
		{
			name: "4",
			args: args{
				v: "1234",
			},
			want: "1**4",
		},
		{
			name: "5",
			args: args{
				v: "12345",
			},
			want: "1**45",
		},
		{
			name: "6",
			args: args{
				v: "123456",
			},
			want: "1***56",
		},
		{
			name: "7",
			args: args{
				v: "1234567",
			},
			want: "12***67",
		},
		{
			name: "8",
			args: args{
				v: "12345678",
			},
			want: "12****78",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSafeValue(tt.args.v); got != tt.want {
				t.Errorf("ToSafeValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
