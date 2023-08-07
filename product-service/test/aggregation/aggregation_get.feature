@aggregation
Feature: Get Aggregation Test : admin user

  Background:
  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

  @SmokeTest
  @getagg
  Scenario: Schema validation for aggregation list
    Given path 'aggregations'
    And params {scope:'#(scope)'}
    And params {page_size:50, page_num:1, sort_by:'aggregation_name', sort_order:'asc'}
    When method get
    Then status 200
    #* match response.aggregations == '#[] data.schema_agg'
 
  @SmokeTest
  Scenario: Get Aggregation Editor
    Given path 'aggregations/editors'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
   And match response.editor[*] contains data.getAgg.product_editor

  
  Scenario: Get Aggregation Products
    Given path 'aggregations/products'
    And params {scope:'#(scope)',editor:'#(data.getAgg.editor)'}
    When method get
    Then status 200
    * response.totalRecords == '#number? _ >=0'
    
  @getAggID
  Scenario: To get the Aggegation ID
    Given path 'aggregations'
    And params {scope:'#(scope)'}
    And params {page_size:50, page_num:1, sort_by:'aggregation_name', sort_order:'asc'}
    When method get
    Then status 200
    * print 'Aggregation ID:' + response.aggregations[0].ID

 
