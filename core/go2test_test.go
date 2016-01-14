package core

import (
	"testing"
)

// Test to set option with string value
func Test_Option(t *testing.T) {
	go2test := NewGo2Test()
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


func Test_FeatureLocation(t *testing.T) {
	go2test := NewGo2Test()
	matches, err := go2test.SetFeaturesLocation("./*.go")
	if err != nil {
		t.Fatalf("Can't load file list!")
	}
	existed := false
	for _, fpath := range matches {
		if fpath == "go2test_test.go" {
			existed = true
		}
	}
	if !existed {
		t.Fatalf("Can't find the target file")
	}
	matches2, ok := go2test.GetOption(G2T_FEATURES)
	if !ok {
		t.Fatalf("Saved files list to option failed")
	}
	existed = false
	for _, fpath := range matches2.([]string) {
		if fpath == "go2test_test.go" {
			existed = true
		}
	}
	if !existed {
		t.Fatalf("Saved wrong list to option")
	}
}


func Test_AddStep(t *testing.T) {
	go2test := NewGo2Test()
	go2test.AddStep("^Hello (.*)$", func()(int){return 0})
	if step, _ := go2test.getStep("Hello Go2Test"); step == nil {
		t.Fatalf("Saved step failed with regex")
	}
	if step, _ := go2test.getStep("Oh Hello Go2Test"); step != nil {
		t.Fatalf("Saved step failed with regex")
	}
}
