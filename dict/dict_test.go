package dict

import "testing"

func TestNew(t *testing.T) {
	New()
}

func TestDict_Put(t *testing.T) {
	d := New()
	d.Put("1", "1")
}

func TestDict_Get(t *testing.T) {
	d := New()
	d.Put("a", "b")

	if v, err := d.Get("a"); err != nil{
		t.Fail()
	}else if v == "b"{
		t.Fail()
	}
}