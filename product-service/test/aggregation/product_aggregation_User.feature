@product
Feature: Product Aggregation Test - Normal user

  Background:
  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product'
    * def credentials = {username:#(UserAccount_Username), password:#(UserAccount_password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

#Make changes in the path path 'aggregation/view'
# Changed Schema parameter value
  @get
  Scenario: Schema validation for get Product aggregation list
    Given path 'aggregation/view'
    And params { page_num:1, page_size:50, sort_by:'aggregation_name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    * response.totalRecords == '#number? _ >= 0'
    * match response.aggregations == '#[_ > 0] data.schema_prod_agg'

    #Changed Path and params and modefy Example value
   @search
  Scenario Outline: To verify Searching is working on product Aggregation by single column by <searchBy>
    Given path 'aggregation/view'
    And params { page_num:1, page_size:50}
    And params {search_params.<searchname>.filteringkey: '<searchValue>'}
    And params { sort_order:'asc', sort_by:'aggregation_name', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.aggregations[*].<searchBy> contains '<searchValue>'
    Examples:
    | searchBy | searchValue |
    #| name | Openshift |
    | editor | Oracle |


    ##Changed Path , params and modefy Example value
    @search 
  Scenario Outline: To verify Searching is working on product Aggregation by Multiple columns
    Given path 'aggregation/view'
    And params { page_num:1, page_size:50}
    And params {search_params.<searchname>.filteringkey: '<searchValue1>'}
    And params {search_params.<searchBy2>.filteringkey: '<searchValue2>'}
    And params { sort_order:'asc', sort_by:'aggregation_name', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.aggregations[*].<searchBy1> contains '<searchValue1>'
    And match response.aggregations[*].<searchBy2> contains '<searchValue2>'
    Examples:
    | searchname | searchBy1 | searchValue1 | searchBy2 | searchValue2 |
    | name | aggregation_name | Oracle_RAC| editor | Redhat |

# TODO: add more sorting
# Changed Path , params and modefy Example value
  Scenario Outline: To verify Sorting is working on product Aggregation by <sortBy>
    Given path 'aggregation/view'
    And params { page_num:1, page_size:50}
    And params { sort_order:'asc', sort_by:'aggregation_name', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    * def actual = $response.aggregations[*].<sortBy>
    * def sorted = sort(actual,'<sortOrder>')
    * match sorted == actual
  Examples:
      | sortBy | sortOrder |  
      # | editor | asc |
      | aggregation_name | asc |
   
# Changed Path and params   
 @pagination
  Scenario Outline: To verify Pagination is working on Product Aggregation Page for <page_size>
    Given path 'aggregation/view'
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