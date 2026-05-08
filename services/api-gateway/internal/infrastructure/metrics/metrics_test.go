package metrics

import "testing"

func TestNormalizedHTTPPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "static route",
			path: "/api/v1/posts:search",
			want: "/api/v1/posts:search",
		},
		{
			name: "numeric id",
			path: "/api/v1/posts/123",
			want: "/api/v1/posts/{id}",
		},
		{
			name: "nested numeric id",
			path: "/api/v1/trainers/1001/tiers",
			want: "/api/v1/trainers/{id}/tiers",
		},
		{
			name: "hex id",
			path: "/api/v1/files/7c33a8a8882f53ad0f199e6e40f677c8",
			want: "/api/v1/files/{id}",
		},
		{
			name: "plain slug",
			path: "/api/v1/sport-types",
			want: "/api/v1/sport-types",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizedHTTPPath(test.path); got != test.want {
				t.Fatalf("normalizedHTTPPath(%q) = %q, want %q", test.path, got, test.want)
			}
		})
	}
}
