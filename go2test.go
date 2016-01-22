package go2test

import (
	"os"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"runtime"
	"strconv"

	ghk "github.com/cucumber/gherkin-go"
	log "github.com/Sirupsen/logrus"
	"sort"
)

const G2T_STATUS_WAIT = 0
const G2T_STATUS_PASS = 1
const G2T_STATUS_FAIL = 2
const G2T_STATUS_SKIP = 3


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
func (v *Handle) ThrowException(format string, a ...interface{}) {
	panic(v.NewException(format, a ...))
}


// Create an *Exception
// @params:
//    message: Error message
// @returns:
//    (*Exception): New *Exception, you can panic it by yourself
func (v *Handle) NewException(format string, a ...interface{}) *Exception {
	e := new(Exception)
	e.Scenario = v.Scenario
	e.Feature = v.Feature
	e.Step = v.Step
	e.Message = fmt.Sprintf(format, a ...)

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
			log.Errorf(" ")
			log.Errorf("|    FAIL!!!")
			log.Errorf("|    %s", exception.Message)
			msgs := strings.Split(exception.Stack, "\n")
			for _, msg := range msgs {
				log.Errorf("|    %s ", msg)
			}
			log.Errorf(" ")
			v.Status = G2T_STATUS_FAIL
			panic(exception)
		}
	}()

	handle.Step = v
	log.Infof("[STEP] %s", v.Text)

	// rebuild the params with handle in the first
	params := make([]reflect.Value, len(v.Params)+1)
	params[0] = reflect.ValueOf(handle)
	for id, p := range v.Params {
		params[id+1]=p
	}

	// Step will ignore the action's return
	v.Action.Call(params)
	v.Status = G2T_STATUS_PASS
}

// Skip the step, not run it
func (v *Step) Skip() {
	v.Status = G2T_STATUS_SKIP
	log.Infof("[ SKIP ] %s", v.Text)
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

	log.Infof(" ")
	log.Infof("----------------------------------------")
	log.Infof("%s.%s", handle.Feature.Name, v.Name)
	log.Infof("----------------------------------------")

	for _, step := range v.Steps {
		step.Run(handle)
	}
	v.Status = G2T_STATUS_PASS
}

type Hook struct {
	key       string
	Priority  int
	Steps     []*ghk.Step
	Regex     *regexp.Regexp
}

func CreateHook(handle *Handle, gScenario *ghk.Scenario) (*Hook, *Exception) {
	hook := new(Hook)
	sName := strings.TrimSpace(gScenario.Name)

	checker, _ := regexp.Compile("@(.+)/(.+)")
	matched := checker.FindStringSubmatch(sName)
	if len(matched) != 0 {
		head := strings.TrimSpace(matched[1])
		body := strings.TrimSpace(matched[2])
		var key string
		var priority int
		var err error
		if strings.Contains(head, "(") && strings.Contains(head, ")") {
			head_checker, _ := regexp.Compile("(.+)\\((.+)\\)")
			head_matched := head_checker.FindStringSubmatch(head)
			if len(head_matched) != 3 {
				return nil, handle.NewException(fmt.Sprintf("Invalid Hook title [%s]", sName))
			}
			key = strings.ToLower(strings.TrimSpace(head_matched[1]))
			priority, err = strconv.Atoi(strings.TrimSpace(head_matched[2]))
			if err != nil {
				return nil, handle.NewException(fmt.Sprintf("Invalid Hook title [%s]: %s", sName, err.Error()))
			}
		} else {
			key = strings.ToLower(head)
			priority = 0
		}
		hook_checker, err := regexp.Compile(body)
		if err != nil {
			return nil, handle.NewException(fmt.Sprintf("Invalid Hook title [%s]: %s", sName, err.Error()))
		}
		hook.Regex = hook_checker
		hook.Steps = gScenario.Steps
		hook.key = key
		hook.Priority = priority
		return hook, nil
	} else {
		return nil, nil
	}
}

// ------------------------------------------------------
// Let []*Hook Support sort
// ------------------------------------------------------
type HookList []*Hook

func (v HookList) Len() int {
	return len(v)
}

func (v HookList) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v HookList) Less(i, j int) bool {
	return v[i].Priority < v[j].Priority
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
	handle      *Handle
	actions     map[*regexp.Regexp]reflect.Value
}

// Create new *Go2Test and init it
// @returns:
//    (*Go2Test): new *Go2Test
func NewGo2Test() (*Go2Test) {
	v := new(Go2Test)
	v.actions = make(map[*regexp.Regexp]reflect.Value)
	v.handle = new(Handle)
	v.handle.Buffer = make(map[string]interface{})
	return v
}


// Add regex && action
// @params:
//    reg: the regex to match step text
//    action: func need to run if matched
// @returns:
//    (error): Errors
func (v *Go2Test) AddAction(reg string, action interface{}) *Exception {
	key, err := regexp.Compile(reg)
	if err != nil {
		return v.handle.NewException(err.Error())
	}
	v.actions[key] = reflect.ValueOf(action)
	return nil
}


// Find matched action
// @params:
//    step: step's text
// @returns:
//    ([]string) matched words
//    (*reflect.Value) action
//    (*Exception) error
func (v *Go2Test) findAction(step string) ([]string, *reflect.Value, *Exception) {
	buf := make([]reflect.Value, 0)
	matched := make([]string, 0)
	for reg, action := range v.actions {
		keywords := reg.FindStringSubmatch(step)
		if len(keywords) != 0 {
			matched = keywords
			buf = append(buf, action)
		}
	}

	switch len(buf) {
	case 0:
		return []string{}, nil, v.handle.NewException(fmt.Sprintf("Matched 0 function [%s]", step))
	case 1:
		return matched, &buf[0], nil
	default:
		return nil, nil, v.handle.NewException(fmt.Sprintf("Matched >1 functions [%s]", step))
	}
}


// read *.feature to create new *Feature
// @params:
//    path: the path of *.feature
// @returns
//    (*Feature) new *Feature
//    (error) Error
func (v *Go2Test) createFeature(path string, tags []string) (*Feature, *Exception) {

	if tags == nil {
		tags = make([]string, 0)
	}

	feature := new(Feature)

	f, err := os.Open(path)
	if err != nil {
		return nil, v.handle.NewException(err.Error())
	}
	gFeature, err := ghk.ParseFeature(f)
	f.Close()
	if err != nil {
		return nil, v.handle.NewException(err.Error())
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

	// Find Hooks
	// Hook is a special which make its name started with @before or @after
	// Example:
	// Scenario: @before/^(/*)$  match all scenarios
	// Scenario: @before/^(.*)stg(.*)$  match all scenarios contains "stg" in its name
	hooklib_be := make(HookList, 0)
	hooklib_af := make(HookList, 0)
	normal_sce := make([]interface{}, 0)

	for _, s := range gFeature.ScenarioDefinitions {
		gScenario, ok := s.(*ghk.Scenario)
		if !ok {
			// Hook Scenario must not be ScenarioOutline
			normal_sce = append(normal_sce, s)
			continue
		}

		hook,exp := CreateHook(v.handle, gScenario)
		if exp != nil {
			return nil, exp
		}

		if hook == nil {
			normal_sce = append(normal_sce, s)
			continue
		}

		switch hook.key {
		case "before":
			hooklib_be = append(hooklib_be, hook)
		case "after":
			hooklib_af = append(hooklib_af, hook)
		default:
			return nil, v.handle.NewException("Find unsupported hook tag: [%s]", hook.key)
		}
	}

	// sort hooklib
	sort.Sort(hooklib_be)
	sort.Sort(hooklib_af)

	// Scenario
	feature.Scenarios = make([]*Scenario, 0)
	for _, s := range normal_sce {
		gScenario, ok := s.(*ghk.Scenario)

		var hook_b []*ghk.Step
		var hook_a []*ghk.Step
//		hook_a := make(map[int][]*ghk.Step)

		if ok {
			// Search matched before hook
			hook_b = GetHookSteps(hooklib_be, strings.TrimSpace(gScenario.Name))
			hook_a = GetHookSteps(hooklib_af, strings.TrimSpace(gScenario.Name))
			scenario, err := v.createScenario(gScenario, gBgSteps, hook_b, hook_a, tags)
			if err!= nil {
				return nil, err
			}
			if scenario != nil {
				scenario.Id = len(feature.Scenarios)
				feature.Scenarios = append(feature.Scenarios, scenario)
			}
		} else {
			hook_b = GetHookSteps(hooklib_be, strings.TrimSpace(s.(*ghk.ScenarioOutline).Name))
			hook_a = GetHookSteps(hooklib_af, strings.TrimSpace(s.(*ghk.ScenarioOutline).Name))
			scenarios, err := v.createScenarioArray(s.(*ghk.ScenarioOutline), gBgSteps, hook_b, hook_a, tags)
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
//    (*Exception) *Exception
func (v *Go2Test) createScenario(gScenario *ghk.Scenario, bgSteps []*ghk.Step,
				hook_b []*ghk.Step, hook_a []*ghk.Step, tags[] string ) (*Scenario, *Exception) {

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

	for i:=0; i<len(bgSteps); i++ {
		step, err := v.createStep(bgSteps[i], map[string]string{})
		if err != nil {
			return nil, err
		}
		step.Id = len(scenario.Steps)
		scenario.Steps = append(scenario.Steps, step)
	}

	for i:=0; i<len(hook_b); i++  {
		step, err := v.createStep(hook_b[i], map[string]string{})
		if err != nil {
			return nil, err
		}
		step.Id = len(scenario.Steps)
		scenario.Steps = append(scenario.Steps, step)
	}

	for i:=0; i<len(gScenario.Steps); i++ {
		step, err := v.createStep(gScenario.Steps[i], map[string]string{})
		if err != nil {
			return nil, err
		}
		step.Id = len(scenario.Steps)
		scenario.Steps = append(scenario.Steps, step)
	}

	for i:=len(hook_a)-1; i>=0; i-- {
		step, err := v.createStep(hook_a[i], map[string]string{})
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
func (v *Go2Test) createScenarioArray( gScenario *ghk.ScenarioOutline,
                  bgSteps []*ghk.Step, hook_b []*ghk.Step, hook_a []*ghk.Step, tags[] string) ([]*Scenario, *Exception) {
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

			for _, gStep := range hook_b {
				step, err := v.createStep(gStep, map[string]string{})
				if err != nil {
					return nil, err
				}
				step.Id = len(scenario.Steps)
				scenario.Steps = append(scenario.Steps, step)
			}

			for _, gStep := range gScenario.Steps {
				step, err := v.createStep(gStep, data)
				if err != nil {
					return nil, err
				}
				scenario.Steps = append(scenario.Steps, step)
			}

			for _, gStep := range hook_a {
				step, err := v.createStep(gStep, map[string]string{})
				if err != nil {
					return nil, err
				}
				step.Id = len(scenario.Steps)
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
func (v *Go2Test) createStep(gStep *ghk.Step, example map[string]string, ) (*Step, *Exception) {
	step := new(Step)
	step.Text = strings.TrimSpace(gStep.Text)
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

	// Find Keywords, Action
	keywords, action, err := v.findAction(step.Text)
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


// Start to run Go2Test framework
// @params:
//     path: test files location ( where *.feature is )
//     tags: filter by @tag
func (v *Go2Test) Run(path string, tags []string) *Exception {

	v.handle.clean()

	log.Infof("Search *.feature by [%s]", path)
	files, err := filepath.Glob(path)
	if err != nil {
		return v.handle.NewException(err.Error())
	}

	features := make([]*Feature, 0)
	for _, p := range files {
		log.Infof("- %s", p)
		feature, err := v.createFeature(p, tags)
		if err != nil {
			log.Errorf("Reading %s", p)
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


func GetHookSteps(lib HookList, s_name string) ([]*ghk.Step) {
	ret := make([]*ghk.Step, 0)
	lib_size := len(lib)
	for i:=0; i<lib_size; i++ {
		hook := lib[i]
		if len(hook.Regex.FindStringSubmatch(s_name)) != 0 {
			step_size := len(hook.Steps)
			for j:=0; j<step_size; j++ {
				ret = append(ret, hook.Steps[j])
			}
		}
	}
	return ret
}


