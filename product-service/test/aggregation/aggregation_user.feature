@aggregation
Feature: Aggregation CRUD Test : Normal User 

  Background:
  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product'
    * def credentials = {username:'testuser@test.com', password: 'password'}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'

  @create
  Scenario: To verify normal user can not create new Aggregation - ibm
    Given path 'aggregations'
    * set data.createAgg.name = "apitest_agg_user"
    * set data.createAgg.scope = scope
    And request data.createAgg
    When method post
    Then status 403


  @getagg
  Scenario: Schema validation for aggregation list
    Given path 'aggregations'
    And params {scopes:'#(scope)'}
    When method get
    Then status 200
    * match response.aggregations == '#[] data.schema_agg'
 

  Scenario: Get Aggregation Editor
    Given path 'aggregations/editors'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
   And match response.editor[*] contains data.getAgg.editor

  Scenario: Get Aggregation Metric
    Given path 'aggregations/metrics'
    And params {scope:'#(scope)'}
    When method get
    Then status 200
    And match response.metric[*] contains data.getAgg.metric

  Scenario: Get Aggregation Products
    Given path 'aggregations/products'
    And params {scope:'#(scope)',editor:'#(data.getAgg.editor)',metric:'#(data.getAgg.metric)'}
    When method get
    Then status 200

 
