@dashboard
Feature: Dashboard Test

  Background:
  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product'
    * def credentials = {username:#(UserAccount_Username), password:#(UserAccount_password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'


     @get
  Scenario: Get Total no. of editors 
    Given path 'dashboard/overview'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.num_editors == data.overview.num_editors

   
       @get
  Scenario: Get Total no. of products
    Given path 'dashboard/overview'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.num_products == data.overview.num_products
  
  @SmokeTest
    @schema
  Scenario: Schema validation for Products on dashboard
   Given path 'dashboard/overview'
    And params {scope:'#(scope)'}
    * def schema = data.schema_overview
    When method get
    Then status 200 
    * response.totalRecords == '#number? _ > 0'

   
    @get
  Scenario: Get Editor's Product 
    Given path 'dashboard/editors/products'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    * response.totalRecords == '#number? _ >0'

  #@get @ignore
  #Scenario: Get Total no. of editors WHEN there is no data in scope
   # Given path 'dashboard/overview'
   # And params {scope:'CLR'}
   # When method get
   # Then status 200

  @get
  Scenario: Get Non-Acquired Products count
    Given path 'dashboard/product/quality'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.not_acquired_products == data.dashboard_products.not_acquired_products

   
    @get
  Scenario: Get Non-deployed Products count
    Given path 'dashboard/product/quality'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.not_deployed_products == data.dashboard_products.not_deployed_products

    
    Scenario: Get Non-deployed Products percentage
    Given path 'dashboard/product/quality'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.not_deployed_products_percentage == data.dashboard_products.not_deployed_products_percentage
    
    Scenario: Get Non-Acquired Products percentage
    Given path 'dashboard/product/quality'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.not_acquired_products_percentage == data.dashboard_products.not_acquired_products_percentage

   
    Scenario: Get Non-Deployed Products list
    Given path 'dashboard/quality/products'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.products_not_deployed[*] contains data.dashboard_products.products_not_deployed


    Scenario: Get Non-Acquired Products list
    Given path 'dashboard/quality/products'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    #And match response.products_not_acquired[*] contains data.dashboard_products.products_not_acquired

#-------------TO get Shared and Received License amount on the Dashboard------------------#
    Scenario:To get the shared and Received License Amount on Dashboard 
      Given path 'sharedamount'
      And params {scope:"AAK"}
      When method get 
      Then status 200 
      


      


    
    
    
    
