package utils

import "testing"

func TestSHA224String(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"sha",
			args{"asd"},
			"cda1d665441ef8120c3d3e82610e74ab0d3b043763784676654d8ef1",
		},
		{"sha1",
			args{"1243t44213"},
			"1463ae1b2452056618847cdeb7be269b4600705c7f94d8cdbde2b797",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SHA224String(tt.args.password); got != tt.want {
				t.Errorf("SHA224String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitPathAndFile(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			"base",
			args{"./foo/boo.yaml"},
			"./foo",
			"boo.yaml",
			false,
		},
		{
			"base1",
			args{"../../../foo/boo.yaml"},
			"../../../foo",
			"boo.yaml",
			false,
		},
		{
			"base2",
			args{"/root/foo/boo.yaml"},
			"/root/foo",
			"boo.yaml",
			false,
		},
		{
			"emptyPath",
			args{"boo.yaml"},
			"",
			"boo.yaml",
			false,
		},
		{
			"emptyFile",
			args{"/look/sas/"},
			"/look/sas/",
			"",
			true,
		},
		{
			"emptyS",
			args{""},
			"",
			"",
			true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := SplitPathAndFile(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitPathAndFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SplitPathAndFile() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("SplitPathAndFile() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
