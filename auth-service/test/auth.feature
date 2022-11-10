@auth
Feature: Account Service Test

  Background:
    * url authServiceUrl+'/api/v1'

  Scenario: Superadmin user is able to login
    Given path 'token'
    #* def credentials = {username:'admin@test.com', password: 'Welcome@123'}
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * form field grant_type = 'password'
    * form fields credentials
    When method post
    Then status 200
    And match response.token_type == 'Bearer'


  Scenario: Admin user is able to login
    Given path 'token'
    * def credentials = {username:'testadmin@test.com', password: 'password'}
    * form field grant_type = 'password'
    * form fields credentials
    When method post
    Then status 200
    And match response.token_type == 'Bearer'

  Scenario: Normal user is able to login
    Given path 'token'
    * def credentials = {username:'testuser@test.com', password: 'password'}
    * form field grant_type = 'password'
    * form fields credentials
    When method post
    Then status 200
    And match response.token_type == 'Bearer'


  Scenario: User is not able to login with invalid credentials
    Given path 'token'
    * def credentials = {username:'invalid@invalid.com', password: 'invalid'}
    * form field grant_type = 'password'
    * form fields credentials
    When method post
    Then status 401
    * response.error == 'unauthorised'
    * response.error_description == 'Invalid username or password'

  # Scenario: User account gets locked after 3 wrong password attempts
  #   Given path 'token'
  #   * def credentials = {username:'invalid@invalid.com', password: 'invalid'}
  #   * form field grant_type = 'password'
  #   * form fields credentials
  #   When method post
  #   Then status 400
  #   Given path 'token'
  #   * def credentials = {username:'invalid@invalid.com', password: 'invalid'}
  #   * form field grant_type = 'password'
  #   * form fields credentials
  #   When method post
  #   Then status 400
  #       Given path 'token'
  #   * def credentials = {username:'invalid@invalid.com', password: 'invalid'}
  #   * form field grant_type = 'password'
  #   * form fields credentials
  #   When method post
  #   Then status 400