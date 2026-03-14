Feature: Delete a stored secret
  As a user
  I want to remove a TOTP secret from poppie
  So that I can clean up secrets I no longer need

  Scenario: Delete an existing secret
    Given the vault is empty
    And I have stored a secret with label "github.com" and secret "JBSWY3DPEHPK3PXP"
    When I delete the secret "github.com"
    Then the deletion should succeed
    And the secret "github.com" should not be in the vault

  Scenario: Delete a non-existent secret
    Given the vault is empty
    When I delete the secret "nonexistent.com"
    Then the deletion should report nothing was deleted
