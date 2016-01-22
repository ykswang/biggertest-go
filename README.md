#### Go2Test

Go2Test is a Cucumber framework, but not only Cucumber


#### How to run it

```go
go2test := NewGo2Test()
go2test.AddAction("^Name is (.*)$", func(handle *Handle, name string){
		handle.Buffer["name"] = name
})
go2test.AddAction("^print name$", func(handle *Handle){
		fmt.Printf("%+v", handle.Buffer["name"].(string))
})

exp := go2test.Run("./examples/*.feature", make([]string, 0))
    if exp != nil {
		fmt.Printf("%s", exp.Message)
		msgs := strings.Split(exp.Stack, "\n")
		for _, msg := range msgs {
			fmt.Printf("%s", msg)
		}
}
```


#### Hooks Before & After

Hook is a special Scenario with name `@tag_name(Priority Level)/regex`

If `print name` will print the given name before:
```gherkin
Feature: Sample

  Scenario: @before(1) / ^(.+)$
    Given Name is before_every
    Then print name

  Scenario: @after(1) / ^(.+)$
    Given Name is after_every
    Then print name

  Scenario: @before(2) / ^Sample1$
    Given Name is before_Sample1
    Then print name

  Scenario: @after(2) / ^Sample1$
    Given Name is after_Sample1
    Then print name

  Scenario: Sample1
    Given Name is Sample1
    Then print name

  Scenario: Sample2
    Given Name is Sample2
    Then print name
```

Run Results:

```
before_every
before_Sample1
Sample1
after_Sample1
after_every
before_every
Sample2
after_every
```


