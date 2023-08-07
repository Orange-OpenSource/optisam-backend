
@account
Feature: Account Service Test

  Background:
    * url accountServiceUrl+'/api/v1/account'
    
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}  
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'
    # * def err = {"error": "string","code": 0,"message": "string","details": [{"type_url": "string","value": "string"}]}

  @SmokeTest
  @get
  Scenario: To validate Scope Schema
    Given path 'scopes'
    When method get
    Then status 200
    * match response.scopes == '#[] data.scopeSchema'

  @get
  Scenario: To verify Get scopes is working
    Given path 'scopes'
    When method get
    Then status 200
  * match response.scopes[*].scope_code contains ['CLR']


  # @create @ignore
 #  Scenario: Create scopes
  #   Given path 'scopes'
  #   And request data.createScopes
  #   When method post
  #   Then status 200

  @create 
  Scenario: To verify scopeCode is unique
    Given path 'scopes'
    * set data.createScope.scope_code = 'CLR'
    And request data.createScope
    When method post
    Then status 409
    * match response.message == "Scope already exists"

  @create 
  Scenario: To verify 3 chars ScopeId is manadatory for scope creation.
    Given path 'scopes'
    * set data.createScope.scope_code = 'DMOS'
    And request data.createScope
    When method post
    Then status 400
    * match response.message == "invalid CreateScopeRequest.ScopeCode: value does not match regex pattern \"\\\\b[A-Z]{3}\\\\b\""

   @create 
  Scenario: To verify spaces are not allowed in ScopeId
    Given path 'scopes'
    * set data.createScope.scope_code = 'XY  '
    And request data.createScope
    When method post
    Then status 400
    * match response.message == "invalid CreateScopeRequest.ScopeCode: value does not match regex pattern \"\\\\b[A-Z]{3}\\\\b\""  

  Scenario: To verify that the edited expense is getting updated
    Given path 'scopes/expenses'
    And request data.createScope
    When method post
    Then status 200 

  Scenario: To verify the updated expense is getting updated on the dashboard
    Given path 'scopes/expenses'
    And request data.createScope
    When method post
    Then status 200
    * header Authorization = 'Bearer '+access_token 
    Given path 'scopes/expenses/', scope
    When method get
    Then status 200
    * match response.expenses == data.createScope.expenses

     # Delete the Created Scope 
  Scenario: Delete the Created Scope
    Given path 'scope',data.createScope.scope_code
    When method Delete
    Then status 200 