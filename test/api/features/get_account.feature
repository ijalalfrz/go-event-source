Feature: Get Account by account id
  Scenario: get account - success
    Given I send a GET with path "/accounts/1"
    Then the response code should be 200
  Scenario: get account - not found
    Given I send a GET with path "/accounts/10"
    Then the response code should be 404