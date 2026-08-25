package models

import "testing"

func TestCoursePackageTotalSessions(t *testing.T) {
	tests := []struct {
		hours int16
		want  int
	}{
		{hours: 6, want: 3},
		{hours: 8, want: 4},
		{hours: 10, want: 5},
		{hours: 12, want: 6},
	}

	for _, test := range tests {
		pkg := CoursePackage{TotalHours: test.hours}
		if got := pkg.TotalSessions(); got != test.want {
			t.Errorf("TotalSessions(%d) = %d, want %d", test.hours, got, test.want)
		}
	}
}

func TestEnrollmentRequiredSessions(t *testing.T) {
	enrollment := StudentEnrollment{TotalHours: 12}
	if got := enrollment.RequiredSessions(); got != 6 {
		t.Fatalf("RequiredSessions() = %d, want 6", got)
	}
}
