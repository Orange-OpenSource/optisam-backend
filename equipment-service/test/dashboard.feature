@dashboard
Feature: DashboardTest

  Background:
  # * def equipmentServiceUrl = "https://optisam-equipment-int.kermit-noprod-b.itn.intraorange"
    * url equipmentServiceUrl+'/api/v1/equipment'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'

 @get
  Scenario: Get Equipment Types on Dashboard
    Given path 'dashboard/types/equipments'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.types_equipments contains data.overview

  @schema
   Scenario: Schema validation for Equipments on dashboard
   Given path 'dashboard/types/equipments'
    And params {scope:'#(scope)'}
    * def schema = data.schema_overview
    When method get
    Then status 200
    * match response.types_equipments == '#[] data.schema_overview'

  @get @ignore
  Scenario: Get Equipment Types where there is no data in scope
    Given path 'dashboard/types/equipments'
    And params {scope:'CLR'}
    When method get
    Then status 200