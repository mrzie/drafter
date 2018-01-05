package db

import "testing"

func TestAccess(t *testing.T) {
	ac, err := Access()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ac)
}
