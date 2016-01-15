package go2test
import (
	"path/filepath"
	"reflect"
	"regexp"
	"errors"
	"os"

	ghk "github.com/cucumber/gherkin-go"
	"strings"
	"fmt"
)


const G2T_FEATURES = "@G2T_FEATURES"
const G2T_STEPS = "@G2T_STEPS"

type Handle struct {
	Buffer  map[string]interface{}
}


type Step struct {
	Text         string
	Action       reflect.Value
	Params       []reflect.Value
}

func (v *Step) Run() (error) {
	results := v.Action.Call(v.Params)
	if results[0].Interface() == nil {
		return nil
	} else {
		return results[0].Interface().(error)
	}
}


type Scenario struct {
	Name            string
	Description     string
	Tags            map[string]interface{}
	Steps           []*Step
	ExName          string
	ExDescription   string
	ExID            int

	run_idx         int
}


func (v *Scenario) Run() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Print("----\n")
			fmt.Print(err)
			fmt.Print("\n----\n")
			fmt.Printf("[ FAIL ] %s\n", v.Steps[v.run_idx].Text)
			for _, step := range v.Steps[v.run_idx+1:] {
				fmt.Printf("[ SKIP ] %s\n", step.Text)
			}
		}

	}()

	v.run_idx = 0
	if v.ExID != -1 {
		fmt.Printf("Scenario %s -- %s -- %d\n", v.Name, v.ExName, v.ExID)
	} else {
		fmt.Printf("Scenario %s\n", v.Name)
	}

	for _, step := range v.Steps {
		err := step.Run()
		if err != nil {
			panic(err)
		} else {
			fmt.Printf("[ PASS ] %s", step.Text)
			v.run_idx += 1
		}
	}
}


type Feature struct {
	Name         string
	Scenarios    []*Scenario
	Description  string
	Tags         map[string]interface{}
}

func (v *Feature) Run() {

	fmt.Printf("Feature %s\n", v.Name)
	for _, scenario := range v.Scenarios {
		scenario.Run()
	}
}


// ----------------------------------------------------------------------------------
// The core struct of the framework
// Please use NewGo2Test() to create instance
// ----------------------------------------------------------------------------------
type Go2Test struct {
	options  map[string]interface{}   // properties for the struct
	handle   *Handle
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
	v.handle = new(Handle)
	v.handle.Buffer = make(map[string]interface{})
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
func (v *Go2Test) findStepAction(step string) ([]string, *reflect.Value, error) {
	steps, ok := v.GetOption(G2T_STEPS)
	if !ok {
		return nil, nil, errors.New("Please Call New2Go2Test() to create instance !!!")
	}
	for reg, callback := range steps.(map[*regexp.Regexp]reflect.Value) {
		keywords := reg.FindStringSubmatch(step)
		if len(keywords) != 0 {
			return keywords, &callback, nil
		}
	}
	return nil, nil, errors.New("Invalid step string [" + step + "]")
}

// ----------------------------------------------------------------------------------
// Create Feature form *.feature file
// @param
//    path: (string) feature's path
// @return
//    (*Feature) The Feature{} instance
//    (error) if anything failed
// ----------------------------------------------------------------------------------
func (v *Go2Test) createFeature(path string) (*Feature, error) {
	feature := new(Feature)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	gFeature, err := ghk.ParseFeature(f)
	f.Close()
	if err != nil {
		return nil, err
	}

	// Description
	feature.Description = gFeature.Description
	feature.Name = gFeature.Name

	// Tags
	feature.Tags = make(map[string]interface{})
	for _, tag := range gFeature.Tags {
		feature.Tags[tag.Name] = make(map[string]interface{})
	}

	// Scenario
	feature.Scenarios = make([]*Scenario, 0)
	for _, s := range gFeature.ScenarioDefinitions {
		gScenario, ok := s.(*ghk.Scenario)
		var scenario *Scenario
		if ok {
			scenario, err = v.createScenario(gScenario)
		} else {
			s, err = v.createScenarioArray(s.(*ghk.ScenarioOutline))
		}
		if err!= nil {
			return nil, err
		}
		feature.Scenarios = append(feature.Scenarios, scenario)
	}
	return feature,nil
}


// ----------------------------------------------------------------------------------
// Create *Scenario form *ghk.Scenario
// @param
//    gScenario: (*ghk.Scenario) The instance of *ghk.Scenario
// @return
//    (*Scenario) The Scenario{} instance
//    (error) if anything failed
// ----------------------------------------------------------------------------------
func (v *Go2Test) createScenario(gScenario *ghk.Scenario) (*Scenario, error) {
	scenario := new(Scenario)

	// Description
	scenario.Name = gScenario.Name
	scenario.Description = gScenario.Description

	// No Examples
	scenario.ExID = -1

	// Tags
	scenario.Tags = make(map[string]interface{})
	for _, tag := range gScenario.Tags {
		scenario.Tags[tag.Name] = make(map[string]interface{})
	}

	// Step
	scenario.Steps = make([]*Step, 0)
	for _, gStep := range gScenario.Steps {
		step, err := v.createStep(gStep, map[string]string{})
		if err != nil {
			return nil, err
		}
		scenario.Steps = append(scenario.Steps, step)
	}

	return scenario, nil
}


// ----------------------------------------------------------------------------------
// Create *Scenario form *ghk.ScenarioOutline
// @param
//    gScenario: (*ghk.ScenarioOutline) The instance of *ghk.ScenarioOutline
// @return
//    (*Scenario) The Scenario{} instance
//    (error) if anything failed
// ----------------------------------------------------------------------------------
func (v *Go2Test) createScenarioArray( gScenario *ghk.ScenarioOutline ) ([]*Scenario, error) {

	scenarios := make([]*Scenario, 0)
	for _, gExample := range gScenario.Examples {
		for id, body := range gExample.TableBody {
			scenario := new(Scenario)
			scenario.Name = gScenario.Name
			scenario.Description = gScenario.Description
			scenario.ExName = gExample.Name
			scenario.Description = gExample.Description
			scenario.ExID = id
			data := map[string]string{}
			for i, cell := range body.Cells {
				data[gExample.TableHeader.Cells[i].Value] = cell.Value
			}
			scenario.Steps = make([]*Step, 0)
			for _, gStep := range gScenario.Steps {
				step, err := v.createStep(gStep, data)
				if err != nil {
					return nil, err
				}
				scenario.Steps = append(scenario.Steps, step)
			}
			scenarios = append(scenarios, scenario)
		}
	}
	return scenarios, nil
}


// ----------------------------------------------------------------------------------
// Create Step form *ghk.Step
// @param
//    gStep: (*ghk.Step) Instance of *ghk.Step
//    example: (map[string]string) Line of Example
// @return
//    (*Step) The Step{} instance
//    (error) if anything failed
// ----------------------------------------------------------------------------------
func (v *Go2Test) createStep(gStep *ghk.Step, example map[string]string, ) (*Step, error) {
	step := new(Step)
	step.Text = gStep.Text
	step.Params = make([]reflect.Value, 0)
	step.Params = append(step.Params, reflect.ValueOf(v.handle))

	// update step text with example data
	for key, val := range example {
		step.Text = strings.Replace(step.Text, "<" + key + ">", val, -1)
	}

	// If with a special param
	data, ok := gStep.Argument.(*ghk.DataTable)
	if ok {
		if len(data.Rows[0].Cells) > 1 {
			// It's a map
			param := make([]map[string]string, 0)
			for _, row := range data.Rows[1:] {
				r := make(map[string]string)
				for idx, cell := range row.Cells {
					r[data.Rows[0].Cells[idx].Value] = cell.Value
				}
				param = append(param, r)
			}
			step.Params = append(step.Params, reflect.ValueOf(param))
		} else {
			// It's a slice
			param := make([]string, 0)
			for _, row := range data.Rows {
				param = append(param, row.Cells[0].Value)
			}
			step.Params = append(step.Params, reflect.ValueOf(param))
		}
	}

	// Find Kyewords, Action
	keywords, action, err := v.findStepAction(step.Text)
	if err != nil {
		return nil, err
	}
	step.Action = *action

	// If with regex params
	if len(keywords) > 1 {
		for _, keyword := range keywords[1:] {
			step.Params = append(step.Params, reflect.ValueOf(keyword))
		}
	}

	return step, nil
}


func (v *Go2Test) Run() (error) {

	paths, ok := v.GetOption(G2T_FEATURES)
	if !ok {
		return errors.New("go2test:{\"message\":\"Please call SetFeaturesLocation before Run !\"}")
	}

	features := make([]*Feature, 0)
	for _, path := range paths.([]string) {
		feature, err := v.createFeature(path)
		if err != nil {
			return err
		}
		features = append(features, feature)
	}

	for _, feature := range features {
		feature.Run()
	}

	return nil
}



