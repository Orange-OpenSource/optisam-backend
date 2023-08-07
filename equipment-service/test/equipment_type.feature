@equipment
Feature: Equipment Service Test

  Background:
  # * def equipmentServiceUrl = "https://optisam-equipment-int.apps.fr01.paas.tech.orange"
    * url equipmentServiceUrl+'/api/v1/equipment'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'


## Equipments Types
@SmokeTest
  @schema
  Scenario: Schema validation for get Equipment Types
    Given path 'types'
    * params {scopes:'#(scope)'}
    When method get
    Then status 200
    * match response.equipment_types == '#[] data.equiptype_schema'
    
  
  Scenario: Get Equipment Type List
    Given path 'types'
    * params {scopes:'#(scope)'}
    When method get
    Then status 200
    * match response.equipment_types contains data.equiptype_server
