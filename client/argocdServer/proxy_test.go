package argocdServer

import "testing"

func Test_modifyLocation1(t *testing.T) {
	type args struct {
		location string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "location rewriting",
			args: args{location: "https://xxx.yyy.abc:3249843/api/dex/auth?client_id=argo-cd&redirect_uri=https%3A%2F%2Fxxx.yyy.abc%3A3249843%2Fauth%2Fcallback&response_type=code&scope=openid+profile+email+groups&state=TjIiKPucNS"},
			want: "https://xxx.yyy.abc:3249843/api/dex/auth?client_id=argo-cd&redirect_uri=https%3A%2F%2Fxxx.yyy.abc%3A3249843%2Forchestrator%2Fauth%2Fcallback&response_type=code&scope=openid+profile+email+groups&state=TjIiKPucNS",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := modifyLocation(tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("modifyLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("modifyLocation() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rewriteRequestUrl(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with orchestrator in url",
			args: args{path: "/orchestrator/main"},
			want: "/main",
		},
		{
			name: "without orchestrator in url",
			args: args{path: "/path1/main"},
			want: "/path1/main",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rewriteRequestUrl(tt.args.path); got != tt.want {
				t.Errorf("rewriteRequestUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}