@product
Feature: Software Expenditure Test-Superadmin

  Background:
    * url productServiceUrl+'/api/v1'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'
  
  Scenario: To check the software exprenditure 
    Given path 'product/dashboard/compliance/soft_exp'
    And params { scope:'#(scope)'}
    When method get 
    Then status 200
    And match response.total_expenditure == data.Exp.total_expenditure