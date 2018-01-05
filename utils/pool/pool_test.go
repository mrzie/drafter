package pool

import (
	"testing"
)

func TestCreatePool(t *testing.T) {
	var db_url = "localhost"
	p, err := CreatePool(db_url, 10, 1000)
	if err != nil {
		t.Fatal(err)
		t.Fail()
		return
	}
	t.Log(p.count)
	t.Log(p)
}
