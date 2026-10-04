package database

import "testing"

func TestWithPoolerSafeParams(t *testing.T) {
	cases := map[string]string{
		"postgresql://u:p@h:6543/postgres":           "postgresql://u:p@h:6543/postgres?binary_parameters=yes",
		"postgresql://u:p@h:6543/postgres?sslmode=x": "postgresql://u:p@h:6543/postgres?sslmode=x&binary_parameters=yes",
		"postgres://h/db?binary_parameters=yes":      "postgres://h/db?binary_parameters=yes",
		"host=h port=5432 user=u dbname=d":           "host=h port=5432 user=u dbname=d binary_parameters=yes",
	}
	for in, want := range cases {
		if got := withPoolerSafeParams(in); got != want {
			t.Errorf("withPoolerSafeParams(%q) = %q, want %q", in, got, want)
		}
	}
}
