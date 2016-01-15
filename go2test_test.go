package go2test

import (
	"testing"
	"fmt"
)


func Test_001(t *testing.T) {
	go2test := NewGo2Test()
	go2test.SetFeaturesLocation("./examples/simple.*")
	go2test.AddStep("^(.+)'s name:(.+)$", func(handle *Handle, person string, name string) error {
		fmt.Printf("Found Person: %s , Found Name: %s\n", person, name)
		return nil
	})
	fmt.Print(go2test.Run())
}
