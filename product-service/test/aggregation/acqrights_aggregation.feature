@aggregation
Feature: Aggregation Test for Acqrights : admin user

  Background:
  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'


  @schema
  Scenario: Schema validation for Aggregation list in Acqrights
    Given path 'acqrights/aggregations'
    And params { page_num:1, page_size:10, sort_by:'NAME', sort_order:'desc', scopes:'#(scope)'}
    * def schema = data.schema_acq_agg
    When method get
    Then status 200
    * response.totalRecords == '#number? _ > 0'
    * match response.aggregations == '#[_ > 0] data.schema_acqrights_agg'

  @agg
  Scenario: User can get Aggregation records from Acqrights
    Given path 'aggregations'
    And params {scopes:'#(scope)'}
    When method get
    Then status 200
    * def agg_id = karate.jsonPath(response.aggregations,"$.[?(@.name=='"+data.getProdAgg.name+"')].ID")[0]  
    * header Authorization = 'Bearer '+access_token
    Given path 'acqrights/aggregations',agg_id,'records'
    And params { page_num:1, page_size:10, sort_by:'NAME', sort_order:'desc',scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.acquired_rights contains data.getAcqrightsAgg

   @search
  Scenario Outline: To verify Searching is working on Acqrights Aggregation by single column by <searchBy>
    Given path 'acqrights/aggregations'
    * params { page_num:1, page_size:10, sort_by:'NAME', sort_order:'desc', scopes:'#(scope)'}
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
  Scenario Outline: To verify Searching is working on Acqrights Aggregation by Multiple columns
    Given path 'acqrights/aggregations'
    And params { page_num:1, page_size:10, sort_by:'NAME', sort_order:'desc', scopes:'#(scope)'}
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
   
  Scenario Outline: To verify Sorting is working on Acqrights Aggregation by <sortBy>
    Given path 'acqrights/aggregations'
    And params { page_num:1, page_size:10, sort_by:'<sortBy>', sort_order:'<sortOrder>', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    * def actual = $response.aggregations[*].<sortBy>
    * def sorted = sort(actual,'<sortOrder>')
    * match sorted == actual
  Examples:
      | sortBy | sortOrder |  
      | EDITOR | asc |
      | EDITOR | desc |
      | NAME | desc |
      | METRIC | desc |
      | METRIC | asc |

 @pagination
  Scenario Outline: To verify Pagination is working on Acqrights Aggregation Page for <page_size>
    Given path  'acqrights/aggregations'
    And params { page_num:1, page_size:<page_size>, sort_by:'NAME', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match $.aggregations == '#[_ <= <page_size>]'
   Examples:
    | page_size |
    | 200 |
    | 100 |
    | 50 |

