package proxy

import (
	"testing"

	"github.com/ikethecoder/prom-multi-tenant-proxy/internal/pkg"
)

func Test_isAuthorized(t *testing.T) {
	authConfig := pkg.Authn{
		[]pkg.User{
			pkg.User{
				"User-a",
				"pass-a",
				"tenant-a",
			},
			pkg.User{
				"User-b",
				"pass-b",
				"tenant-b",
			},
		},
	}
	type args struct {
		user       string
		pass       string
		authConfig *pkg.Authn
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 string
	}{
		{
			"Valid User",
			args{
				"User-a",
				"pass-a",
				&authConfig,
			},
			true,
			"tenant-a",
		}, {
			"Invalid User",
			args{
				"invalid",
				"pass-a",
				&authConfig,
			},
			false,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := isAuthorized(tt.args.user, tt.args.pass, tt.args.authConfig)
			if got != tt.want {
				t.Errorf("isAuthorized() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("isAuthorized() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
