Feature: Test Scenario Tags

  @Tag1
  Scenario: Tag1
    Given Name Tag1

  @Tag2
  Scenario: Tag2
    Given Name Tag2

  @Tag3
  Scenario Outline: Tag3
    Given LineNum: <NO>
    Given Name: <NAME>

    Examples: Name List1
      |NO|NAME|
      |1 |Tom |
      |2 |Eric|