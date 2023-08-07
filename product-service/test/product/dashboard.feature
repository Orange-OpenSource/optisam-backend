@dashboard
Feature: Dashboard Test

  Background:

  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product'
    
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

    
    Scenario: Get counterfeiting_percentage on Dashboard overview
      Given path 'dashboard/alert/compliance'
      And params {scope:'#(scope)'}
      When method get
      Then status 200
      And match response.counterfeiting_percentage == data.overview.counterfeiting_percentage

    Scenario: Get overdeployment_percentage on Dashboard overview 
      Given path 'dashboard/alert/compliance'
      And params {scope:'#(scope)'}
      When method get
      Then status 200
      And match response.overdeployment_percentage == data.overview.overdeployment_percentage

    Scenario: Get total_counterfeiting_amount on Dashboard overview
      Given path 'dashboard/overview'
      And params {scope:'#(scope)'}
      When method get
      Then status 200
      And match response.total_counterfeiting_amount == data.overview.total_counterfeiting_amount

    Scenario: Get total_license_cost on Dashboard overview
      Given path 'dashboard/overview'
      And params {scope:'#(scope)'}
      When method get
      Then status 200
      And match response.total_license_cost == data.overview.total_license_cost

    Scenario: Get total_maintenance_cost on Dashboard overview
      Given path 'dashboard/overview'
      And params {scope:'#(scope)'}
      When method get
      Then status 200
      And match response.total_maintenance_cost == data.overview.total_maintenance_cost

    Scenario: Get total_underusage_amount on Dashboard overview
      Given path 'dashboard/overview'
      And params {scope:'#(scope)'}
      When method get
      Then status 200
      And match response.total_underusage_amount == data.overview.total_underusage_amount

  @get
  Scenario: Get Product banner
    Given path 'banner'
    And params {scope:'#(scope)', time_zone :'CEST'}
    When method get
    Then status 200


  @get
  Scenario: Get compliance alert details 
    Given path 'dashboard/alert/compliance'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
And match response == data.compliance
   

  @Get
  Scenario: Get Metric product details
    Given path 'dashboard/metrics/products'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    
  @SmokeTest
  @schema
  Scenario: Schema validation for Metric product on dashboard
    Given path 'dashboard/metrics/products'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.metrics_products[*] contains data.Metric_product
   
  @get 
  Scenario: Get Editor's Product when there is no data in scope
    Given path 'dashboard/editors/products'
    And params {scope:'CLR'}
    When method get
    Then status 200 
    # not working 

  @get 
  Scenario: Get Total no. of products when there is no data in scope
    Given path 'dashboard/overview'
    And params {scope:'CLR'}
    When method get
    Then status 200

  @get 
  Scenario: Get Total no. of editors When there is no data in scope
    Given path 'dashboard/overview'
    And params {scope:'CLR'}
    When method get
    Then status 200
    # not working 