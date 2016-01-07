package core

type Cucu struct {
	Path  string  // The path of *.feature File
}

// --
// Create a new instance
// @param path : The path of *.feature File
// @return : The instance or error with string format
// --
func NewCucu(path string) (*Cucu, error){
	inst := new(Cucu)
	inst.Path = path
	return inst, nil
}
