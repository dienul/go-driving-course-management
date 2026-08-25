package seeds

import (
	"strings"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	valid := Config{
		AdminName:     "Administrator",
		AdminEmail:    "admin@example.com",
		AdminPassword: "strong-password",
	}
	if err := validateConfig(valid); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}

	tests := []Config{
		{AdminEmail: valid.AdminEmail, AdminPassword: valid.AdminPassword},
		{AdminName: valid.AdminName, AdminPassword: valid.AdminPassword},
		{AdminName: valid.AdminName, AdminEmail: valid.AdminEmail},
		{
			AdminName:     valid.AdminName,
			AdminEmail:    valid.AdminEmail,
			AdminPassword: strings.Repeat("a", 73),
		},
	}
	for _, test := range tests {
		if err := validateConfig(test); err == nil {
			t.Errorf("invalid config accepted: %+v", test)
		}
	}
}
