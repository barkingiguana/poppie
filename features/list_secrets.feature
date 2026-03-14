Feature: List stored secrets
  As a user
  I want to see which TOTP secrets are stored
  So that I know what's available

  Scenario: List secrets in an empty vault
    Given the vault is empty
    When I list all secrets
    Then I should see 0 secrets

  Scenario: List secrets after storing some
    Given the vault is empty
    And I have stored a secret with label "github.com" and secret "JBSWY3DPEHPK3PXP"
    And I have stored a secret with label "aws.amazon.com" and secret "ORSXG5DJNZTQ"
    When I list all secrets
    Then I should see 2 secrets
    And the list should include "github.com"
    And the list should include "aws.amazon.com"
