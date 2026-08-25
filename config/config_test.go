package config

import (
	"reflect"
	"testing"
)

func TestLoadParsesAllowedCORSOrigins(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{
			name:  "defaults to the local frontend",
			value: "",
			want:  []string{"http://localhost:5173"},
		},
		{
			name:  "trims whitespace and ignores empty or duplicate origins",
			value: " https://drive-academy.up.railway.app , , http://localhost:5173 , https://drive-academy.up.railway.app ",
			want: []string{
				"https://drive-academy.up.railway.app",
				"http://localhost:5173",
			},
		},
		{
			name:  "never enables wildcard origins",
			value: " *, https://drive-academy.up.railway.app ",
			want:  []string{"https://drive-academy.up.railway.app"},
		},
		{
			name:  "falls back safely when every origin is empty or wildcard",
			value: " *, , ",
			want:  []string{"http://localhost:5173"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("CORS_ALLOWED_ORIGINS", test.value)
			if got := Load().CORSAllowedOrigins; !reflect.DeepEqual(got, test.want) {
				t.Fatalf("CORS allowed origins = %v, want %v", got, test.want)
			}
		})
	}
}
