@aggregation
Feature: Aggregation Test for Acqrights : Normal user

  Background:
  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product'
    * def credentials = {username:#(UserAccount_Username), password:#(UserAccount_password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'


  @schema
  Scenario: Schema validation for Aggregation list in Acqrights
    Given path 'aggregated_acqrights'
    And params { page_num:1, page_size:50, scope:'#(scope)', sort_order:'asc', sort_by:'SKU'}
    * def schema = data.schema_acq_agg
    When method get
    Then status 200
    * response.totalRecords == '#number? _ > 0'
    * match response.aggregations == '#[_ > 0] data.schema_acqrights_agg'

   
   @search
  Scenario Outline: To verify Searching is working on Acqrights Aggregation by single column by <searchBy>
    Given path 'aggregated_acqrights'
    * params {scope:'#(scope)'}
    * params { page_num:1, page_size:50, sort_by:'SKU', sort_order:'asc'}
    * params {search_params.<searchBy>.filteringkey: '<searchValue>'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.aggregations[*].<searchBy1> contains '<searchValue1>'
    Examples:
    | searchBy | searchValue | searchBy1 | searchValue1 |
    | name | Openshift | aggregation_name |redhat_openshift |
    | editor | Redhat | product_editor | Redhat |


   
    @search 
  Scenario Outline: To verify Searching is working on Acqrights Aggregation by Multiple columns
    Given path 'aggregated_acqrights'
    * params {scope:'#(scope)'}
    * params { page_num:1, page_size:50, sort_by:'SKU', sort_order:'asc'}
    And params {search_params.<searchBy1>.filteringkey: '<searchValue1>'}
    And params {search_params.<searchBy2>.filteringkey: '<searchValue2>'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.aggregations[*].<searchmatchResp1> contains '<searchmatchRespVal>'
    And match response.aggregations[*].<searchmatchResp2> contains '<searchmatchRespVal2>'
  Examples:
    | searchBy1 | searchValue1 | searchBy2 | searchValue2 | searchmatchResp1 | searchmatchRespVal | searchmatchResp2 | searchmatchRespVal2 |
    | name | Openshift| editor | Redhat | aggregation_name | redhat_openshift | product_editor | Redhat |


  Scenario Outline: To verify Sorting is working on Acqrights Aggregation by <sortBy>
    Given path 'aggregated_acqrights'
    * params {scope:'#(scope)'}
    And params { page_num:1, page_size:50, sort_by:'<sortBy>', sort_order:'<sortOrder>', scopes:'#(scope)'}
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
      | METRIC | desc |
      | METRIC | asc |


 @pagination
  Scenario Outline: To verify Pagination is working on Acqrights Aggregation Page for <page_size>
    Given path 'aggregated_acqrights'
    * params {scope:'#(scope)'}
    And params { page_num:1, page_size:<page_size>, sort_by:'<sortBy>', sort_order:'<sortOrder>'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match $.aggregations == '#[_ <= <page_size>]'
   Examples:
   | page_size | sortBy | sortOrder |  
   | 200 | SKU | asc |
   | 100 | SKU | asc |
   | 50 | SKU | asc |

