@Shared_License
Feature: Shared_License From User Role

  Background:
  * url productServiceUrl+'/api/v1/product'
  * def credentials = {username:#(UserAccount_Username), password:#(UserAccount_password)}
  * callonce read('../common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'

  Scenario: Get all the shared License deatils
    Given path 'acqrights/licenses'
    And params {}