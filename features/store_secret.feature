Feature: Store a TOTP secret
  As a user
  I want to store a TOTP secret in poppie
  So that I can generate codes for it later

  Scenario: Store a valid secret
    Given the vault is empty
    When I store a secret with label "github.com" and secret "JBSWY3DPEHPK3PXP"
    Then the store should succeed
    And I should receive a 6-digit verification code
    And the secret "github.com" should be in the vault

  Scenario: Store a secret with custom settings
    Given the vault is empty
    When I store a secret with label "aws.amazon.com" and secret "JBSWY3DPEHPK3PXP" using SHA256 with 8 digits
    Then the store should succeed
    And I should receive an 8-digit verification code

  Scenario: Attempt to store a duplicate label
    Given the vault is empty
    And I have stored a secret with label "github.com" and secret "JBSWY3DPEHPK3PXP"
    When I store a secret with label "github.com" and secret "ORSXG5DJNZTQ"
    Then the store should fail with an error

  Scenario: Attempt to store an invalid secret
    Given the vault is empty
    When I store a secret with label "bad" and secret "!!!not-valid!!!"
    Then the store should fail with an error
