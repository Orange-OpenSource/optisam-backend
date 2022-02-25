@equipment
Feature: Equipment Service Test

  Background:
  # * def equipmentServiceUrl = "https://optisam-equipment-int.kermit-noprod-b.itn.intraorange"
    * url equipmentServiceUrl+'/api/v1/equipment'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'


## Metadata 

  @schema
  Scenario: Schema validation for get Metadata
    Given path 'metadata'
    * params { type:'ALL', scopes:'#(scope)'}
    When method get
    Then status 200
    * match response.metadata == '#[] data.metadata_schema'
    
  # @metadata
  # Scenario: Get Metadata by ID
  #   Given path 'equipments/metadata',data.getMetadata.ID
  #   * params { type:'ALL', scopes:'#(scope)'}
  #   When method get
  #   Then status 200
  #  * match response == data.getMetadata


## Equipments Types

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
