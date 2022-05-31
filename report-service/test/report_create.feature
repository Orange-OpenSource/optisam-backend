@report
Feature: Report Service Test - Create Report : Admin

  Background:
    # * def reportServiceUrl = "https://optisam-report-int.apps.fr01.paas.tech.orange"
    * url reportServiceUrl+'/api/v1'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'

  @create
  Scenario: Create the reports type compliance
    Given path 'report'
    And request data.acqrights_report
    When method post
    Then status 200
    * match response.success == true

  @create
  Scenario: Create the reports type ProductEquipments
    Given path 'report'
    And request data.product_equipments_report
    When method post
    Then status 200
    * match response.success == true
    
