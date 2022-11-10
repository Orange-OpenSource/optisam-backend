@product
Feature: Product Service Test : Normal user

  Background:
    * url productServiceUrl+'/api/v1'
    * def credentials = {username:#(UserAccount_Username), password:#(UserAccount_password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'


  @schema
  Scenario: Schema Validation for get product list
    Given path 'products'
    * params { page_num:1, page_size:50, sort_by:'name', sort_order:'asc', scopes:'#(scope)'}
    * def schema = data.schema_prod
    When method get
    Then status 200
    * response.totalRecords == '#number? _ >= 0'
    * match response.products == '#[_ > 0] schema'
    * match response.products == '#[_ <= 50] schema'

     @get
  Scenario: To verify user can get list of all products for the scope
    Given path 'products'
    And params { page_num:1, page_size:50, sort_by:'name', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
     And match response.products contains data.getProduct
    * def result = karate.jsonPath(response, "$.products[?(@.swidTag=='"+data.getProduct.swidTag+"')]")[0]
    * match result == data.getProduct


     @search
  Scenario Outline: To verify Searching is working on list of Products by <searchBy>
    Given path 'products' 
    And params { page_num:1, page_size:50, sort_by:'name', sort_order:'asc', scopes:'#(scope)'}
    And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.products[*].<searchBy> contains '<searchValue>'
  Examples:
    | searchBy | searchValue |
    | name | Adobe Media Server |
    | swidTag | Adobe_Media_Server_Adobe_5.0.16 |
    | editor | Adobe |  

     @search
  Scenario Outline: To verify Searching is working on list of Products by <searchBy1> and <searchBy2>
    Given path 'products' 
    And params { page_num:1, page_size:50, sort_by:'name', sort_order:'asc', scopes:'#(scope)'}
    And params {search_params.<searchBy1>.filteringkey: '<searchValue1>'}
    And params {search_params.<searchBy2>.filteringkey: '<searchValue2>'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.products[*].<searchBy1> contains '<searchValue1>'
    And match response.products[*].<searchBy2> contains '<searchValue2>'
  Examples:
    | searchBy1 | searchValue1 | searchBy2 | searchValue2 |
    | name | Adobe Media Server | swidTag | Adobe_Media_Server_Adobe_5.0.16 |
    | name | IBM DB2 | editor | IBM |


  @pagination
  Scenario Outline: To verify Pagination is working on Products list
    Given path  'products'
    And params { page_num:1, page_size:'<page_size>', sort_by:'name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match $.products == '#[_ <= <page_size>]'
   Examples:
    | page_size |
    | 200 |
    | 100 |
    | 50 |

  Scenario Outline: To verify Pagination on Product Page with Invalid inputs
    Given path  'products'
    And params { page_num:'<page_num>', page_size:'<page_size>', sort_by:'name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 400
   Examples:
    | page_size | page_num |
    | 5 | 5 |
    | 10 | 0 |
    | "A" | 5 |
    
 @sort
  Scenario Outline: To verify Sorting is working on list of Products by <sortBy>
    Given path 'products'
    And params { page_num:1, page_size:10, sort_by:'<sortBy>', sort_order:'<sortOrder>', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    * def actual = $response.products[*].<sortBy>
    * def sorted = sort(actual,'<sortOrder>')
    * match sorted == actual
  Examples:
      | sortBy | sortOrder |  
      | editor | asc |
      | editor | desc |
      | name | asc |
      #| name | desc |
      | swidtag | desc |
    
  @get
  Scenario: To verify user can get all Editors
    Given path 'product/editors'
    * params {scopes:'#(scope)'}
    When method get
    Then status 200
    # And match response.editor == 'REDHAT'
    And match response.editors contains data.getProduct.editor



    @getdetail
  Scenario: To verify user can get details of a product
    Given path 'product/'+data.getProduct.swidTag
    * params { scope:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0

  # Working For Admin
  #@get
  #Scenario: To verify user can get Products of a given Editor
   # Given path 'product/editors/products'
   # * params { scopes:'#(scope)', editor:'IBM'}
   # When method get
   # Then status 200
    # And match response.products[*].name contains data.getProduct.name
   # And match response.products[*].swidTag contains data.getProduct.swidTag