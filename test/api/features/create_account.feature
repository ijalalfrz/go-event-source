Feature: Create Account
  Scenario: create account - success
    Given I use default timestamp
    And I set a header key "x-transaction-id" with value "10001"
    And I send a POST with path "/accounts" with JSON:
    """
    {
        "account_id":20,
        "initial_balance":10
    }
    """
    Then the response code should be 201

  Scenario: create account - no x-transaction-id
    Given I use default timestamp
    And I send a POST with path "/accounts" with JSON:
    """
    {
        "account_id":20,
        "initial_balance":10
    }
    """
    Then the response code should be 400
    Then the response error message should contain "X-TRANSACTION-ID is required in header"


  Scenario: create account - already exists
    Given I use default timestamp
    And I set a header key "x-transaction-id" with value "123"
    And I send a POST with path "/accounts" with JSON:
    """
    {
        "account_id":1,
        "initial_balance":10
    }
    """
    Then the response code should be 409
    Then the response error message should contain "account already exists"

  Scenario: create account - invalid transaction id
    Given I use default timestamp
    And I set a header key "x-transaction-id" with value "tx-1"
    And I send a POST with path "/accounts" with JSON:
    """
    {
        "account_id":20,
        "initial_balance":10
    }
    """
    Then the response code should be 409
    Then the response error message should contain "transaction id already used by another operation"