@dashboard
Feature: Dashboard  Test

  Background:
  # * def productServiceUrl = "https://optisam-product-int.kermit-noprod-b.itn.intraorange"
    * url productServiceUrl+'/api/v1/product'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'


 @get
  Scenario: Get Total license cost 
    Given path 'dashboard/overview'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.total_license_cost == data.overview.total_license_cost


  @schema
   Scenario: Schema validation for acquiredRights on dashboard
   Given path 'dashboard/overview'
    And params {scope:'#(scope)'}
    * def schema = data.schema_overview
    When method get
    Then status 200

   @get
  Scenario: Get Metric Products
    Given path 'dashboard/metrics/products'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.metrics_products[*] contains data.metrics_products
   

   @get @ignore
  Scenario: Get compliance counterfeiting
    Given path 'dashboard/compliance/counterfeiting'
    And params {editor:'IBM' , scope:'#(scope)', }
    When method get
    Then status 200
    And match response.products_licenses[*] contains data.counterfeit_products_licenses
    And match response.products_costs[*] contains data.counterfeit_products_costs


  @get
  Scenario: Get compliance Overdeployment
    Given path 'dashboard/compliance/overdeployment'
    And params {editor:'Micro Focus' , scope:'#(scope)', }
    When method get
    Then status 200
    And match response.products_licenses[*] contains data.overdeployed_products_licenses
    And match response.products_costs[*] contains data.overdeployed_products_costs
    
  
   