package core

import (
	"testing"
)

// Test to set option with string value
func Test_StringOption(t *testing.T) {
	go2test, _ := NewGo2Test()
	_, replaced := go2test.SetOption("string", "string_value")
	if replaced {
		t.Fatalf("replaced signal shoud be false while the key is new")
	}

	val, replaced := go2test.SetOption("string", "string_value 2")
	if !replaced {
		t.Fatalf("replaced signal shoud be true while the key is existed!")
	}

	if val != "string_value" {
		t.Fatalf("wrong old value!")
	}

	val, existed := go2test.GetOption("string")
	if !existed {
		t.Fatalf("Can't found a value which is saved before!")
	}

	if val != "string_value 2" {
		t.Fatalf("Get wrong value!")
	}
}
