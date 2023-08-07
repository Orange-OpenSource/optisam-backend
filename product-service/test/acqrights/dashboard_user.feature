@dashboard
Feature: Dashboard  Test

  Background:
  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'


 @get
  Scenario: Get Total license cost 
    Given path 'dashboard/overview'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
   And match response.total_license_cost == data.overview.total_license_cost

  @SmokeTest
  @get
  Scenario: Get Metric Products
    Given path 'dashboard/metrics/products'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    #And match response.metrics_products[*] contains data.metrics_products

  #@get
  #Scenario: Get compliance counterfeiting
   # Given path 'dashboard/compliance/counterfeiting'
    #And params {editor:'Adobe' , scope:'#(scope)' }
    #When method get
    #Then status 200
    #And match response.products_licenses contains data.counterfeit_products_licenses
    #And match response.products_costs contains data.counterfeit_products_costs

  @get
  Scenario: Get compliance Overdeployment
    Given path 'dashboard/compliance/overdeployment'
    And params {editor:'Adobe' , scope:'#(scope)' }
    When method get
    Then status 200
    #And match response.products_licenses contains data.overdeployed_products_licenses
    #And match response.products_costs contains data.overdeployed_products_costs

  
  Scenario: Get Total shared license cost 
    Given path 'sharedamount'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
   And match response.total_shared_amount == data.sharedamount.total_shared_amount

  
  Scenario: Get Total received license cost 
    Given path 'sharedamount'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
   And match response.total_recieved_amount == data.sharedamount.total_recieved_amount

