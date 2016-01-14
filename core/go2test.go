package core
import (
	"path/filepath"
	"reflect"
	"regexp"
	"errors"
)


const G2T_FEATURES = "@G2T_FEATURES"
const G2T_STEPS = "@G2T_STEPS"


// ----------------------------------------------------------------------------------
// The core struct of the framework
// Please use NewGo2Test() to create instance
// ----------------------------------------------------------------------------------
type Go2Test struct {
	options  map[string]interface{}   // properties for the struct
}


// -------------------------------------------------------------------------------------------------------------
// Create a instance of Go2Test
// @return
//    (*Go2Test):  Instance of Go2Test
// -------------------------------------------------------------------------------------------------------------
func NewGo2Test() (*Go2Test){
	v := new(Go2Test)
	v.options = make(map[string]interface{})
	v.SetOption(G2T_STEPS, make(map[*regexp.Regexp]reflect.Value))
	return v
}


// ----------------------------------------------------------------------------------
// Set the option value
// @param
//     key : (string) option's name
//     value : (interface{}) option's value
// @return
//     (interface{}): if the key is already existed, will return an old value.
//     (bool): is a replace action
// ----------------------------------------------------------------------------------
func (v *Go2Test) SetOption(key string, value interface{}) (interface{}, bool) {
	oldVal, ok := v.options[key]
	v.options[key] = value
	return oldVal, ok
}


// ----------------------------------------------------------------------------------
// Get the option value
// @param
//     key : (string) option's name
//
// @return
//     (interface{}): value
//     (bool): is the key existed
// ----------------------------------------------------------------------------------
func (v *Go2Test) GetOption(key string) (interface{}, bool) {
	val, ok := v.options[key]
	return val, ok
}


// ----------------------------------------------------------------------------------
// Set the *.feature files location.
// @param:
//    path: (string) the file path, such as
//          Examples:
//          =========
//          /root/features,
//          /root/features/sample.feature,
//          /root/features/sampel_*.feature
// @return
//    ([]string): file list which founded in path
//    (error): error with JSON {"msg":"", method: "Go2Test.SetFeaturesLocation"}
// ----------------------------------------------------------------------------------
func (v *Go2Test) SetFeaturesLocation(path string) ([]string, error) {
	matches, err := filepath.Glob(path)
	if err == nil {
		v.SetOption(G2T_FEATURES, matches)
	}
	return matches, err
}

// ----------------------------------------------------------------------------------
// Add step into framework step libs
// @param
//    reg: (string) describe step content with regular expressions
//    callback: (interface{}) the step function
// ----------------------------------------------------------------------------------
func (v *Go2Test) AddStep(reg string, callback interface{}) error {
	key, err := regexp.Compile(reg)
	if err != nil {
		return err
	}
	val := reflect.ValueOf(callback)
	steps, ok := v.GetOption(G2T_STEPS)
	if !ok {
		return errors.New("Please Call New2Go2Test() to create instance !!!")
	}
	steps.(map[*regexp.Regexp]reflect.Value)[key]=val
	return nil
}


// ----------------------------------------------------------------------------------
// Get step callback by step content
// @param
//    step: (string) step content
// @return
//    (*reflect.Value) callback with reflect type
// ----------------------------------------------------------------------------------
func (v *Go2Test) getStep(step string) (*reflect.Value, error) {
	steps, ok := v.GetOption(G2T_STEPS)
	if !ok {
		return nil, errors.New("Please Call New2Go2Test() to create instance !!!")
	}
	for reg, callback := range steps.(map[*regexp.Regexp]reflect.Value) {
		keywords := reg.FindStringSubmatch(step)
		if len(keywords) != 0 {
			return &callback, nil
		}
	}
	return nil, errors.New("Invalid step string [" + step + "]")
}


