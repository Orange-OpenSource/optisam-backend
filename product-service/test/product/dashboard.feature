@dashboard
Feature: Dashboard Test

  Background:

  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product'
    #* def credentials = {username:'admin@test.com', password: 'Welcome@123'}
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'


 
  @schema
  Scenario: Schema validation for Products on dashboard
   Given path 'dashboard/overview'
    And params {scope:'#(scope)'}
    * def schema = data.schema_overview
    When method get
    Then status 200 
    * response.totalRecords == '#number? _ > 0'

    
     @get
  Scenario: Get Total no. of editors 
    Given path 'dashboard/overview'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.num_editors == data.overview.num_editors


  @get
  Scenario: Get Total no.of products
    Given path 'dashboard/overview'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.num_products == data.overview.num_products
   

  @schema
  Scenario: Schema validation for Products on dashboard
   Given path 'dashboard/overview'
    And params {scope:'#(scope)'}
    * def schema = data.schema_overview
    When method get
    Then status 200
    And match response == schema

  @get
  Scenario: Get Editor's Product 
    Given path 'dashboard/editors/products'
    And params {scope:'#(scope)'}
    When method get
    Then status 200


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
    And match response.products_not_acquired[*] contains data.dashboard_products.products_not_acquired

    
    Scenario: Get counterfeiting_percentage on Dashboard overview
      Given path 'dashboard/alert/compliance'
      And params {scope:'#(scope)'}
      When method get
      Then status 200
      And match response.counterfeiting_percentage == data.overview.counterfeiting_percentage_val

    Scenario: Get overdeployment_percentage on Dashboard overview 
      Given path 'dashboard/alert/compliance'
      And params {scope:'#(scope)'}
      When method get
      Then status 200
      And match response.overdeployment_percentage == data.overview.overdeployment_percentage_val

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

  @get 
  Scenario: Get Total no. of products when there is no data in scope
    Given path 'dashboard/overview'
    And params {scope:'CLR'}
    When method get
    Then status 200

  @get 
  Scenario: Get Total no. of editors WHEN there is no data in scope
    Given path 'dashboard/overview'
    And params {scope:'CLR'}
    When method get
    Then status 200