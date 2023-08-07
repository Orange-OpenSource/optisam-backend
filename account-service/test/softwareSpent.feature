Feature: Softwarwe Spent _Test 

Background:
  * url accountServiceUrl+'/api/v1'
  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
  * callonce read('common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'

  Scenario Outline: To Modifiy  the software spent     Positive  
    Given path '/account/scopes/expenses'
    * def expense =  { "scope_code":<scope>,"expenses":<Expense>}
    And request expense
    When method post 
    Then status 200
    Examples:
    | Expense| scope|
    | 52345.35 | AZU|
    | 34343434343434343434 |TST|
    | 378701320.453|AAK|
    
    
  Scenario Outline: To Modifiy  the software spent     Negative 
    Given path '/account/scopes/expenses'
    * def expense =  { "scope_code":<scope>,"expenses":<Expense>}
    And request expense
    When method post 
    Then status 400
    Examples:
    | Expense| scope|
    | avcdsdf | AZU|
    | -343434sv4545sdsds | ACQ|
    | 333334343434br545545 |TST|
    | -1234 | OLN|


  Scenario: To Get  the software spent 
    Given path '/account/scopes/expenses',scope
    When method get
    Then status 200
    


