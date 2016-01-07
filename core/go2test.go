package core

// Use NewGo2Test() to create one

type Go2Test struct {
	options  map[string]interface{}
}

// -----------------------------------------
// Create a instance of Go2Test
// @return
//    New instance.
//    Error with string message.
// -----------------------------------------
func NewGo2Test() (*Go2Test, error){
	v := new(Go2Test)
	v.options = make(map[string]interface{})
	return v, nil
}


// -----------------------------------------
// Set the option value
// @param
//     key : option's name
//     value : option's value
// @return
//     interface{}: if the key is already existed, will return an old value.
//     bool: is a replace action
// -----------------------------------------
func (v *Go2Test) SetOption(key string, value interface{}) (interface{}, bool) {
	oldVal, ok := v.options[key]
	v.options[key] = value
	return oldVal, ok
}

// -----------------------------------------
// Set the option value
// @param
//     key : option's name
// @return
//     interface{}: value
//     bool: is the key existed
// -----------------------------------------
func (v *Go2Test) GetOption(key string) (interface{}, bool) {
	val, ok := v.options[key]
	return val, ok
}


