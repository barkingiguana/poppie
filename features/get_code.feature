Feature: Get a TOTP code
  As a user or automated tool
  I want to retrieve a current TOTP code for a stored secret
  So that I can authenticate with the corresponding service

  Scenario: Get a code for a stored secret
    Given the vault is empty
    And I have stored a secret with label "github.com" and secret "JBSWY3DPEHPK3PXP"
    When I get a code for "github.com"
    Then I should receive a valid 6-digit code
    And I should see how many seconds the code is valid for

  Scenario: Get a code for a non-existent secret
    Given the vault is empty
    When I get a code for "nonexistent.com"
    Then the request should fail with a not found error
