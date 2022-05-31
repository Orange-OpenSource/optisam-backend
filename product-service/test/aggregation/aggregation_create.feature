@aggregation
Feature: Create Aggregation : admin 

  Background:
  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'

  @create
  Scenario: To verify admin can create new Aggregation and delete it - ibm
    Given path 'aggregations'
    * set data.createAgg.scope = scope
    And request data.createAgg
    When method post
    Then status 200
    And set data.createAgg.ID = response.ID
    #  * match response == data.createAgg
    # # * set data.createAgg.product_names = '#[]'
    # Given path 'aggregations'
    # * header Authorization = 'Bearer '+access_token
    # * params {scopes:'#(scope)'}
    # When method get
    # Then status 200
    #  * match response.aggregations contains data.createAgg
    Given path 'aggregations',data.createAgg.ID
    * header Authorization = 'Bearer '+access_token
    And  params {scope:'#(scope)'}
    When method delete
    Then status 200
    * match response.success == true


  @create
  Scenario: To verify aggregation name is unique
    Given path 'aggregations'
    * set data.createAgg.name = data.getAgg.name 
    * set data.createAgg.editor = data.getAgg.editor 
    * set data.createAgg.metric = data.getAgg.metric
    * set data.createAgg.products = data.getAgg.products
    * set data.createAgg.scope = scope
    And request data.getAgg
    When method post
    Then status 400


  @create
  Scenario: User can not create new Aggregation with same swidtag
    Given path 'aggregations'
    * set data.createAgg.name = "apitest_sameswidtag"
    * set data.createAgg.editor = data.getAgg.editor 
    * set data.createAgg.metric = data.getAgg.metric
    * set data.createAgg.products = data.getAgg.products
    * set data.createAgg.scope = scope
    And request data.createAgg
    When method post
    Then status 400

  #   @create
  # Scenario: User can not create Aggregation of product with different editor
  #   Given path 'aggregations'
  #   * set data.createAgg.products = data.getAgg.products
  #   * set data.createAgg.scope = scope
  #   And request data.createAgg
  #   When method post
  #   Then status 400  
  #   And match response.message contains "ProductNotAvailable"

## TODO : Add more swidtag for same metric(ibm) to update the aggregation
  # @update 
  # Scenario: Update Aggregation
  #   Given path 'aggregations'
  #   When method get
  #   Then status 200
  #   Given path 'aggregations/'+ 47
  #   And set data.createAgg.name = "Tst_aggregation111"
  #   And set data.createAgg.ID= responseID
  #   And request data.createAgg
  #   When method put
  #   Then status 200
  #   And match response.ID == responseID
