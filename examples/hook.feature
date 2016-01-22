Feature: Test Hook

  Scenario: @before(1) / ^S1$
    Given Name Hook Before1

  Scenario: @before(2) / ^S1$
    Given Name Hook Before2

  Scenario: @after(1) / ^S1$
    Given Name Hook After1

  Scenario: @after(2) / ^S1$
    Given Name Hook After2

  Scenario: S1
    Given Name S1