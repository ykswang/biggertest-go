package go2test

import (
	"testing"
	log "github.com/Sirupsen/logrus"
//	"fmt"
	"os"
	"strings"
)
//
//
//func Test_001(t *testing.T) {
//	go2test := NewGo2Test()
//	go2test.SetFeaturesLocation("./examples/simple.*")
//	go2test.AddStep(
//		"^(.+)'s name:(.+)$",
//		func(handle *Handle, person string, name string) error {
//			fmt.Printf("Found Person: %s , Found Name: %s\n", person, name)
//			return nil
//		})
//	if err:=go2test.Run(); err!=nil{
//		t.Fail()
//	}
//}
//
//
//func Test_002(t *testing.T) {
//	go2test := NewGo2Test()
//	go2test.SetFeaturesLocation("./examples/list.*")
//	go2test.AddStep(
//		"^Params$",
//		func(handle *Handle, values []string) error {
//			for _, val := range values {
//				fmt.Println(val)
//			}
//			return nil
//		})
//	if err:=go2test.Run(); err!=nil{
//		t.Fail()
//	}
//}
//
//
//func Test_003(t *testing.T) {
//	go2test := NewGo2Test()
//	go2test.SetFeaturesLocation("./examples/table.*")
//	go2test.AddStep(
//		"^Params$",
//		func(handle *Handle, values []map[string]string) error {
//			for num, val := range values {
//				for k, v := range val {
//					fmt.Printf("[Line %d] Key: %s, Value: %s\n", num, k, v)
//				}
//			}
//			return nil
//		})
//	if err:=go2test.Run(); err!=nil{
//		t.Fail()
//	}
//}
//
//
//func Test_004(t *testing.T) {
//	go2test := NewGo2Test()
//	go2test.SetFeaturesLocation("./examples/outline.*")
//	go2test.AddStep(
//		"^LineNum: (.*)$",
//		func(handle *Handle, nu interface{}) error {
//			fmt.Printf("Line: %d\n", nu.(int))
//			return nil
//		})
//	go2test.AddStep(
//		"^Name: (.*)$",
//		func(handle *Handle, name string) error {
//			fmt.Printf("Line: %d\n", name)
//			return nil
//		})
//	if err:=go2test.Run(); err!=nil{
//		t.Fail()
//	}
//}
//
//
//func Test_005(t *testing.T) {
//	go2test := NewGo2Test()
//	go2test.AddStep(
//		"^Name(.*)$",
//		func(handle *Handle, name string) error {
//			fmt.Printf("Name: %s\n", name)
//			return nil
//		})
//	go2test.AddStep(
//		"^LineNum: (.*)$",
//		func(handle *Handle, nu interface{}) error {
//			fmt.Printf("Line: %d\n", nu.(int))
//			return nil
//		})
//	go2test.AddStep(
//		"^Name: (.*)$",
//		func(handle *Handle, name string) error {
//			fmt.Printf("Line: %d\n", name)
//			return nil
//		})
//	go2test.SetTags([]string{"@Tag1"})
//	go2test.SetFeaturesLocation("./examples/tags.*")
//	if err:=go2test.Run(); err!=nil{
//		t.Fail()
//	}
//}
//
//
//func Test_006(t *testing.T) {
//	go2test := NewGo2Test()
//	go2test.AddStep(
//		"^Name(.*)$",
//		func(handle *Handle, name string) error {
//			fmt.Printf("Name: %s\n", name)
//			return nil
//		})
//	go2test.AddStep(
//		"^LineNum: (.*)$",
//		func(handle *Handle, nu interface{}) error {
//			fmt.Printf("Line: %d\n", nu.(int))
//			return nil
//		})
//	go2test.AddStep(
//		"^Name: (.*)$",
//		func(handle *Handle, name string) error {
//			fmt.Printf("Line: %d\n", name)
//			return nil
//		})
//	go2test.SetTags([]string{"@Tag3"})
//	go2test.SetFeaturesLocation("./examples/tags.*")
//	if err:=go2test.Run(); err!=nil{
//		t.Fail()
//	}
//}
//
//
//func Test_007(t *testing.T) {
//	go2test := NewGo2Test()
//	go2test.AddStep(
//		"^Name(.*)$",
//		func(handle *Handle, name string) error {
//			fmt.Printf("Name: %s\n", name)
//			return nil
//		})
//	go2test.SetFeaturesLocation("./examples/background.*")
//	if err:=go2test.Run(); err!=nil{
//		t.Fail()
//	}
//}
//
//func Test_008(t *testing.T) {
//	go2test := NewGo2Test()
//	go2test.AddStep(
//		"^Name(.*)$",
//		func(handle *Handle, name string) error {
//			fmt.Printf("Name: %s\n", name)
//			return nil
//		})
//	go2test.SetFeaturesLocation("./examples/hook.*")
//	if err:=go2test.Run(); err!=nil{
//		t.Fail()
//	}
//}

func Test_009(t *testing.T) {
	log.SetFormatter(&log.TextFormatter{FullTimestamp:true, TimestampFormat:"01-02 15:04:05"})
	log.SetOutput(os.Stdout)
	go2test := NewGo2Test()
	go2test.AddAction("^Name(.*)$", func(handle *Handle, name string){
		log.Infof("Name: %s", name)
	})
	go2test.AddAction("^Failed$", func(handle *Handle){
		panic("err")
	})
	exp := go2test.Run("./examples/hook.feature", make([]string, 0))
	if exp != nil {
		log.Errorf("%s", exp.Message)
		msgs := strings.Split(exp.Stack, "\n")
		for _, msg := range msgs {
			log.Errorf("%s", msg)
		}
		t.Fail()
	}
}