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
   
# TBD Value is not showing on UI
   @get 
  Scenario: Get compliance counterfeiting
    Given path 'dashboard/compliance/counterfeiting'
    And params {editor:'Adobe' , scope:'#(scope)', }
    When method get
    Then status 200
    #And match response.products_licenses[*] contains data.counterfeit_products_licenses
    #And match response.products_costs[*] contains data.counterfeit_products_costs


  @get
  Scenario: Get compliance Overdeployment
    Given path 'dashboard/compliance/overdeployment'
    And params {editor:'Adobe' , scope:'#(scope)', }
    When method get
    Then status 200
    And match response.products_licenses[*] contains data.overdeployed_products_licenses
    And match response.products_costs[*] contains data.overdeployed_products_costs
    
  @get
  Scenario: To verify Details of Not Licenced product
    Given path 'dashboard/quality/products'
    And params {scope:'#(scope)'}
    When method get
    Then status 200

  
  @pegination
  Scenario Outline: To verify pagination on Not Licenced product
    Given path 'dashboard/quality/products'
    And params {scope:'#(scope)'}
    And params { page_num:1, page_size:'<page_size>'}
    When method get
    Then status 200
    And response.products_not_deployed > 0

    Examples:
    | page_size |
    | 200 |
    | 100 |
    | 50 |

  @get
  Scenario: To verify Editor for Not Licenced product 
    Given path 'dashboard/quality/products'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.products_not_acquired[*].editor contains data.products_not_acquired.editor
