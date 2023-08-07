@report
Feature: Report Service Test - Create Report : Admin

  Background:
    # * def reportServiceUrl = "https://optisam-report-int.apps.fr01.paas.tech.orange"
    * url reportServiceUrl+'/api/v1'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

  @create
  Scenario: Create the reports type compliance for oracle editor
    Given path 'report'
    And request data.compliance_report
    When method post
    Then status 200
    * match response.success == true

# Validate this case when issue getting fixed OPTISAM-3679
  Scenario: Create the reports type compliance with invalid editor
    Given path 'report'
    And request data.invalidcompliance_report
    When method post
    Then status 400
   # * match response.message == "Editor doesn't exist"

  @create
  Scenario: Create the reports type ProductEquipments
    Given path 'report'
    And request data.product_equipments_report
    When method post
    Then status 200
    * match response.success == true

    ## Validate this case when issue getting fixed OPTISAM-3679
   @create
  Scenario: Create the reports type ProductEquipments for invalid editor
    Given path 'report'
    And request data.invalidproduct_equipments_report
    When method post
    Then status 400
   # * match response.message == "Editor doesn't exist"
    
