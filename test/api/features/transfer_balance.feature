Feature: Transfer Balance
  Scenario: transfer balance - success
    Given I use default timestamp
    And I set a header key "x-transaction-id" with value "tx-3"
    And I send a POST with path "/transactions" with JSON:
    """
    {
        "source_account_id": 1,
        "destination_account_id": 2,
        "amount": 100.00
    }
    """
    Then the response code should be 204

  Scenario: transfer balance - source account not found
    Given I use default timestamp
    And I set a header key "x-transaction-id" with value "tx-4"
    And I send a POST with path "/transactions" with JSON:
    """
    {
        "source_account_id": 999,
        "destination_account_id": 2,
        "amount": 100.00
    }
    """
    Then the response code should be 404
    Then the response error message should contain "source account not found"

  Scenario: transfer balance - destination account not found
    Given I use default timestamp
    And I set a header key "x-transaction-id" with value "tx-5"
    And I send a POST with path "/transactions" with JSON:
    """
    {
        "source_account_id": 1,
        "destination_account_id": 999,
        "amount": 100.00
    }
    """
    Then the response code should be 404
    Then the response error message should contain "destination account not found"

  Scenario: transfer balance - insufficient balance
    Given I use default timestamp
    And I set a header key "x-transaction-id" with value "tx-6"
    And I send a POST with path "/transactions" with JSON:
    """
    {
        "source_account_id": 1,
        "destination_account_id": 2,
        "amount": 10000.00
    }
    """
    Then the response code should be 400
    Then the response error message should contain "insufficient balance"

  Scenario: transfer balance - same account
    Given I use default timestamp
    And I set a header key "x-transaction-id" with value "tx-7"
    And I send a POST with path "/transactions" with JSON:
    """
    {
        "source_account_id": 1,
        "destination_account_id": 1,
        "amount": 100.00
    }
    """
    Then the response code should be 409
    Then the response error message should contain "source and destination account cannot be the same"

  Scenario: transfer balance - invalid transaction id
    Given I use default timestamp
    And I set a header key "x-transaction-id" with value "tx-1"
    And I send a POST with path "/transactions" with JSON:
    """
    {
        "source_account_id": 1,
        "destination_account_id": 2,
        "amount": 100.00
    }
    """
    Then the response code should be 409
    Then the response error message should contain "transaction id already used by another operation"
