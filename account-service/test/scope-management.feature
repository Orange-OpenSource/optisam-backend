
@account
Feature: Account Service Test

  Background:
  # * def accountServiceUrl = "https://ng-account-int.apps.fr01.paas.tech.orange"
    * url accountServiceUrl+'/api/v1/account'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    # * def err = {"error": "string","code": 0,"message": "string","details": [{"type_url": "string","value": "string"}]}


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
  * match response.scopes[*].scope_code contains ['DMO']


  # @create @ignore
  # Scenario: Create scopes
  #   Given path 'scopes'
  #   And request data.createScopes
  #   When method post
  #   Then status 200

  @create 
  Scenario: To verify scopeCode is unique
    Given path 'scopes'
    * set data.createScope.scope_code = 'DMO'
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