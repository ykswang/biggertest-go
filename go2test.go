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
	"runtime"
)


const SUB_FEA = ""
const SUB_SCE = "  "
const SUB_STE = "    "

const G2T_FEATURES = "@G2T_FEATURES"
const G2T_STEPS = "@G2T_STEPS"
const G2T_TAGS = "@G2T_TAGS"

const G2T_STATUS_WAIT = 0
const G2T_STATUS_PASS = 1
const G2T_STATUS_FAIL = 2
const G2T_STATUS_SKIP = 3

const STR_INVALID_INIT_GO2TEST = "Please Call New2Go2Test() to create instance !!!"
const STR_INVALID_STEP_TEXT = "Invalid step string [%s]"

// ----------------------------------------------------------------------------------
// @name: Exception
// Standard Error format for Go2Test Framework
// @values
//    - Feature: Error Feature
//    - Scenario: Error Scenario
//    - Step: Error Step
//    - Message: Error Message
//    - Stack: StaceTrace of call tree
// ----------------------------------------------------------------------------------
type Exception struct {
	Feature    *Feature
	Scenario   *Scenario
	Step       *Step
	Message    string
	Stack      string
}


// ----------------------------------------------------------------------------------
// @name: Handle
//   * Data bridge
//   * Common API
// @params:
//     - Buffer: Data buffer
// ----------------------------------------------------------------------------------
type Handle struct {
	Buffer       map[string]interface{}
	Feature      *Feature
	Scenario     *Scenario
	Step         *Step
}

// Clean the handle
func (v Handle) clean() {
	v.Buffer = make(map[string]interface{})
	v.Feature = nil
	v.Scenario = nil
	v.Step = nil
}

// Create an *Exception and panic it
// @params:
//    message: Error message
func (v *Handle) ThrowException(message string) {
	panic(v.NewException(message))
}


// Create an *Exception
// @params:
//    message: Error message
// @returns:
//    (*Exception): New *Exception, you can panic it by yourself
func (v *Handle) NewException(message string) *Exception {
	e := new(Exception)
	e.Scenario = v.Scenario
	e.Feature = v.Feature
	e.Step = v.Step
	e.Message = message

	bufStack := make([]byte, 1<<16)
	num := runtime.Stack(bufStack, true)
	e.Stack = string(bufStack[0:num])
	return e
}


// ----------------------------------------------------------------------------------
// @name: Step
// If step failed, framework will skip the remaining steps which belong the same scenario
// @returns:
//     Id: The order ID
//     Text: Statement of step, teh statement must cloud be matched by regex in step libs
//     Action: The callback
//     Params: Params pass to callback
//     Status: Result WAIT|PASS|FAIL|SKIP
// ----------------------------------------------------------------------------------
type Step struct {
	Id           int
	Text         string
	Action       reflect.Value
	Params       []reflect.Value
	Status       int
}

// Do the step, run step's action with params
// Step itself does not need to ensure that there is no error or panic, so it do not maintain the exception message
// @Params:
//    handle: *Handle, it's created by Go2Test
func (v *Step) Run(handle *Handle) {

	defer func(){
		if err:=recover(); err!=nil {
			name := reflect.TypeOf(err).Name()
			var exception *Exception
			if name != "*Exception" {
				exception = handle.NewException(fmt.Sprintf("%v+", err))
			} else {
				exception = err.(*Exception)
			}
			fmt.Printf("%s [FAIL] %s\n", SUB_STE, exception.Step.Text)
			fmt.Printf("%s        %s\n", SUB_STE, exception.Message)
			fmt.Printf("%s        %s\n", SUB_STE, exception.Stack)
			v.Status = G2T_STATUS_FAIL
			panic(exception)
		}
	}()

	handle.Step = v

	// rebuild the params with handle in the first
	params := make([]reflect.Value, len(v.Params)+1)
	params[0] = reflect.ValueOf(handle)
	for id, p := range v.Params {
		params[id+1]=p
	}

	// Step will ignore the action's return
	v.Action.Call(params)
	fmt.Printf("%s [ PASS ] %s\n", SUB_FEA, v.Text)
	v.Status = G2T_STATUS_PASS
}

// Skip the step, not run it
func (v *Step) Skip() {
	v.Status = G2T_STATUS_SKIP
	fmt.Printf("%s [ SKIP ] %s\n", SUB_FEA, v.Text)
}


// ----------------------------------------------------------------------------------
// @name: Scenario
// Groups of Steps
// @params:
//     Id: The order ID
//     Name: The name of Scenario
//     Description: The Description of Scenario
//     Steps: All Steps need to run(contains background)
//     Status: Result WAIT|PASS|FAIL
// ----------------------------------------------------------------------------------
type Scenario struct {
	Id              int
	Name            string
	Description     string
	Steps           []*Step
	Status          int
}

// Run Scenario
// Scenario has defer, handle the panic
// @params:
//    handle: *Handle, it's created by Go2Test
func (v *Scenario) Run(handle *Handle) {

	defer func() {
		if err := recover(); err != nil {
			v.Status = G2T_STATUS_FAIL
			exception := err.(*Exception)
			for _, step := range v.Steps[exception.Step.Id+1:] {
				step.Skip()
			}
		}
	}()

	handle.Scenario = v

	fmt.Printf("%s Scenario: %s\n", SUB_SCE, v.Name)

	for _, step := range v.Steps {
		step.Run(handle)
	}
	v.Status = G2T_STATUS_PASS
}


// ----------------------------------------------------------------------------------
// @name: Feature
// Scenario groups
// @params:
//     Name: The feature's name
//     Description: The feature's description
//     Scenarios: All scenarios need to run(contains background)
//     Status: Result WAIT|PASS|FAIL
// ----------------------------------------------------------------------------------
type Feature struct {
	Name         string
	Scenarios    []*Scenario
	Description  string
	Status       int
}

// Do the Feature
// @params:
//    handle: *Handle, it's created by Go2Test
func (v *Feature) Run(handle *Handle) {
	fmt.Printf("Feature %s\n", v.Name)
	v.Status = G2T_STATUS_PASS
	handle.Feature = v
	for _, scenario := range v.Scenarios {
		scenario.Run(handle)
		if scenario.Status == G2T_STATUS_FAIL {
			v.Status = G2T_STATUS_FAIL
		}
	}
}


// ----------------------------------------------------------------------------------
// @name: Go2Test
// The main framework
// Please use NewGo2Test() to create new *Go2Test
// ----------------------------------------------------------------------------------
type Go2Test struct {
	options  map[string]interface{}   // properties for the struct
	handle   *Handle
}

// Create new *Go2Test and init it
// @returns:
//    (*Go2Test): new *Go2Test
func NewGo2Test() (*Go2Test){
	v := new(Go2Test)
	v.options = make(map[string]interface{})
	v.setOption(G2T_STEPS, make(map[*regexp.Regexp]reflect.Value))
	v.handle = new(Handle)
	v.handle.Buffer = make(map[string]interface{})
	return v
}

// Set property
// @params:
//     key : property's name
//     value : property's value
// @returns:
//     (interface{}): if the key is already saved before, will return the old value
//     (bool): is have returned an old value
func (v *Go2Test) setOption(key string, value interface{}) (interface{}, bool) {
	oldVal, ok := v.options[key]
	v.options[key] = value
	return oldVal, ok
}


// Get property
// @params:
//     key : property's name
// @returns:
//     (interface{}): property's value
//     (bool): if the key is valid
func (v *Go2Test) getOption(key string) (interface{}, bool) {
	val, ok := v.options[key]
	return val, ok
}


// Set the location of feature files
// @params:
//    path: the location
//          ex:
//          =========
//          /root/features,
//          /root/features/sample.feature,
//          /root/features/sampel_*.feature
// @returns:
//    ([]string): the feature list which founded in the path
//    (error): error message
func (v *Go2Test) SetFeaturesLocation(path string) ([]string, error) {
	matches, err := filepath.Glob(path)
	if err == nil {
		v.setOption(G2T_FEATURES, matches)
	}
	return matches, err
}

// Config the running tags
// @params
//     tags: the tag array, help user to filter the features and scenarios
func (v *Go2Test) SetTags(tags []string) {
	v.setOption(G2T_TAGS, tags)
}


// Add steps
// @params:
//    reg: the regex
//    callback: the action
// @returns:
//    (error): Errors
func (v *Go2Test) AddStep(reg string, callback interface{}) error {
	key, err := regexp.Compile(reg)
	if err != nil {
		return err
	}
	val := reflect.ValueOf(callback)
	steps, ok := v.getOption(G2T_STEPS)
	if !ok {
		return errors.New(STR_INVALID_INIT_GO2TEST)
	}
	steps.(map[*regexp.Regexp]reflect.Value)[key]=val
	return nil
}


// Find matched action
// @params:
//    step: step's text
// @returns:
//    (*reflect.Value) the step action
func (v *Go2Test) findStepAction(step string) ([]string, *reflect.Value, error) {
	steps, ok := v.getOption(G2T_STEPS)
	if !ok {
		return nil, nil, errors.New(STR_INVALID_INIT_GO2TEST)
	}
	for reg, callback := range steps.(map[*regexp.Regexp]reflect.Value) {
		keywords := reg.FindStringSubmatch(step)
		if len(keywords) != 0 {
			return keywords, &callback, nil
		}
	}
	return nil, nil, errors.New(fmt.Sprintf(STR_INVALID_STEP_TEXT, step))
}

// read *.feature to create new *Feature
// @params:
//    path: the path of *.feature
// @returns
//    (*Feature) new *Feature
//    (error) Error
func (v *Go2Test) createFeature(path string, tags []string) (*Feature, error) {

	if tags == nil {
		tags = []string{}
	}

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

	if gFeature.Tags != nil && len(gFeature.Tags) > 0 && len(tags) > 0 {
		bOK := false
		for _, gTag := range gFeature.Tags {
			for _, tag := range tags {
				if tag == gTag.Name {
					bOK = true
					break
				}
			}
		}
		if !bOK {
			return nil, nil
		}
	}


	// Description
	feature.Description = gFeature.Description
	feature.Name = gFeature.Name

	// Background
	gBgSteps := []*ghk.Step{}
	if gFeature.Background != nil {
		gBgSteps = gFeature.Background.Steps
	}

	// Scenario
	feature.Scenarios = make([]*Scenario, 0)
	for _, s := range gFeature.ScenarioDefinitions {
		gScenario, ok := s.(*ghk.Scenario)

		if ok {
			scenario, err := v.createScenario(gScenario, gBgSteps, tags)
			if err!= nil {
				return nil, err
			}
			if scenario != nil {
				scenario.Id = len(feature.Scenarios)
				feature.Scenarios = append(feature.Scenarios, scenario)
			}
		} else {
			scenarios, err := v.createScenarioArray(s.(*ghk.ScenarioOutline), gBgSteps, tags)
			if err!= nil {
				return nil, err
			}
			for _, scenario := range scenarios {
				scenario.Id = len(feature.Scenarios)
				feature.Scenarios = append(feature.Scenarios, scenario)
			}
		}
	}
	return feature,nil
}


// Create new *Scenario
// @params:
//    gScenario: ghk.Scenario
// @returns:
//    (*Scenario) new *Scenario
//    (error) errors
func (v *Go2Test) createScenario(gScenario *ghk.Scenario, bgSteps []*ghk.Step, tags[] string) (*Scenario, error) {
	scenario := new(Scenario)

	if len(tags) > 0 {
		bOK := false
		if gScenario.Tags != nil && len(gScenario.Tags) > 0 {
			for _, gTag := range gScenario.Tags {
				for _, tag := range tags {
					if tag == gTag.Name {
						bOK = true
						break
					}
				}
			}
		}
		if !bOK {
			return nil, nil
		}
	}

	// Description
	scenario.Name = gScenario.Name
	scenario.Description = gScenario.Description

	// Step
	scenario.Steps = make([]*Step, 0)

	for _, gStep := range bgSteps {
		step, err := v.createStep(gStep, map[string]string{})
		if err != nil {
			return nil, err
		}
		step.Id = len(scenario.Steps)
		scenario.Steps = append(scenario.Steps, step)
	}

	for _, gStep := range gScenario.Steps {
		step, err := v.createStep(gStep, map[string]string{})
		if err != nil {
			return nil, err
		}
		step.Id = len(scenario.Steps)
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
func (v *Go2Test) createScenarioArray( gScenario *ghk.ScenarioOutline, bgSteps []*ghk.Step, tags[] string) ([]*Scenario, error) {
	scenarios := make([]*Scenario, 0)

	// Check Tags
	// Check Tags
	// For scenario, if gTags is empty and tags is not empty, is not allowed
	if len(tags) > 0 {
		bOK := false
		if gScenario.Tags != nil && len(gScenario.Tags) > 0 {
			for _, gTag := range gScenario.Tags {
				for _, tag := range tags {
					if tag == gTag.Name {
						bOK = true
						break
					}
				}
			}
		}
		if !bOK {
			return nil, nil
		}
	}


	for _, gExample := range gScenario.Examples {
		for id, body := range gExample.TableBody {
			scenario := new(Scenario)
			scenario.Name = gScenario.Name + " | " + gExample.Name + " | " + string(id)
			scenario.Description = gScenario.Description
			data := map[string]string{}
			for i, cell := range body.Cells {
				data[gExample.TableHeader.Cells[i].Value] = cell.Value
			}
			scenario.Steps = make([]*Step, 0)
			for _, gStep := range bgSteps {
				step, err := v.createStep(gStep, map[string]string{})
				if err != nil {
					return nil, err
				}
				scenario.Steps = append(scenario.Steps, step)
			}
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

	paths, ok := v.getOption(G2T_FEATURES)
	if !ok {
		return errors.New("Please call SetFeaturesLocation before Run")
	}

	tags, ok := v.getOption(G2T_TAGS)
	if !ok {
		tags = []string{}
	}

	features := make([]*Feature, 0)
	for _, path := range paths.([]string) {
		feature, err := v.createFeature(path, tags.([]string))
		if err != nil {
			return err
		}
		if feature != nil {
			features = append(features, feature)
		}
	}

	for _, feature := range features {
		feature.Run(v.handle)
	}

	return nil
}



