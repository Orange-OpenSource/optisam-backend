@product
Feature: Product Aggregation Test - admin user

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
  Scenario: Schema validation for get Product aggregation list
    Given path 'aggregations'
    And params { page_num:1, page_size:10, sort_by:'aggregation_name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    * response.totalRecords == '#number? _ >= 0'
    # * match response.aggregations == '#[_ > 0] data.schema_prod_agg'


 @get
  Scenario: To verify user can get product aggregation Details
    Given path 'aggregations'
    And params {scopes:'#(scope)'}
    When method get
    Then status 200
    * def agg_id = karate.jsonPath(response.aggregations,"$.[?(@.name=='"+data.getProdAgg.name+"')].ID")[0]  
    * header Authorization = 'Bearer '+access_token
    Given path 'aggregations/productview',agg_id,'details'
    * params { scope:'#(scope)'}
    When method get
    Then status 200
    And match response.ID contains data.getProdAgg.ID
    And match response.products contains data.getProdAgg.swidtags

  @getproductagg
  Scenario: To verify user can get product view from product aggregation
    Given path 'aggregations'
    And params {scopes:'#(scope)'}
    When method get
    Then status 200
    * def agg_id = karate.jsonPath(response.aggregations,"$.[?(@.name=='"+data.getProdAgg.name+"')].ID")[0]  
    * header Authorization = 'Bearer '+access_token
    Given path 'aggregations',agg_id,'products'
    And params { scopes:'#(scope)'}
    When method get
    Then status 200
    And match response.products[*].swidTag == data.getProdAgg.swidtags

   @search
  Scenario Outline: To verify Searching is working on product Aggregation by single column by <searchBy>
    Given path 'aggregations'
    * params { page_num:1, page_size:10, sort_by:'aggregation_name', sort_order:'asc', scopes:'#(scope)'}
    * params {search_params.<searchBy>.filteringkey: '<searchValue>'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.aggregations[*].<searchBy> contains '<searchValue>'
    Examples:
    | searchBy | searchValue |
    | name | apitest_agg_oracleWL |
    | editor | Oracle |


    @search 
  Scenario Outline: To verify Searching is working on product Aggregation by Multiple columns
    Given path 'aggregations'
    And params { page_num:1, page_size:10, sort_by:'aggregation_name', sort_order:'asc', scopes:'#(scope)'}
    And params {search_params.<searchBy1>.filteringkey: '<searchValue1>'}
    And params {search_params.<searchBy2>.filteringkey: '<searchValue2>'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.aggregations[*].<searchBy1> contains '<searchValue1>'
    And match response.aggregations[*].<searchBy2> contains '<searchValue2>'
  Examples:
    | searchBy1 | searchValue1 | searchBy2 | searchValue2 |
    | name | apitest_agg_oracleWL| editor | Oracle |

# TODO: add more sorting
  Scenario Outline: To verify Sorting is working on product Aggregation by <sortBy>
    Given path 'aggregations'
    And params { page_num:1, page_size:10, sort_by:'<sortBy>', sort_order:'<sortOrder>', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    * def actual = $response.aggregations[*].<sortBy>
    * def sorted = sort(actual,'<sortOrder>')
    * match sorted == actual
  Examples:
      | sortBy | sortOrder |  
      # | editor | asc |
      | aggregation_name | desc |
   
 @pagination
  Scenario Outline: To verify Pagination is working on Product Aggregation Page for <page_size>
    Given path  'aggregations'
    And params { page_num:1, page_size:'<page_size>', sort_by:'aggregation_name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match $.aggregations == '#[_ <= <page_size>]'
   Examples:
    | page_size |
    | 200 |
     | 100 |
     | 50 |