@application
Feature: Application Service Test

  Background:
  # * def applicationServiceUrl = "https://optisam-application-int.kermit-noprod-b.itn.intraorange"
    * url applicationServiceUrl+'/api/v1'
    * def credentials = {username:'testuser@test.com', password: 'password'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'

  @schema
  Scenario: Schema validation for get Applications
    Given path 'applications'
    * params { page_num:1, page_size:10, sort_by:'name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    * response.totalRecords == '#number? _ >= 0'
    * match response.applications == '#[] data.schema_app'

   @pagination
  Scenario Outline: To verify Pagination on Application page
    Given path 'applications' 
    And params { page_num:1, page_size:'<page_size>', sort_by:'name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match $.applications == '#[_ <= <page_size>]'

  Examples:
    | page_size |
    | 50 |
    | 100 |
    | 200 |

    Scenario Outline: To verify Pagination on Application Page with Invalid inputs
    Given path   'applications'
    And params { page_num:'<page_num>', page_size:'<page_size>', sort_by:'name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 400
   Examples:
    | page_size | page_num |
    | 5 | 5 |
    | 10 | 0 |
    | "A" | 5 |  


  Scenario Outline: Search Applications by single column
    Given path 'applications' 
    And params { page_num:1, page_size:50, sort_by:'name', sort_order:'asc', scopes:'#(scope)'}
    And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.applications[*].<searchBy> contains '<searchValue>'
  Examples:
    | searchBy | searchValue |
    | name | General Application 1 |
    | owner | Orange Money |
    | obsolescence_risk | Medium | 


 @search
  Scenario Outline: Search Applications by Multiple columns
    Given path 'applications' 
    And params { page_num:1, page_size:50, sort_by:'name', sort_order:'asc', scopes:'#(scope)'}
    And params {search_params.<searchBy1>.filteringkey: '<searchValue1>'}
    And params {search_params.<searchBy2>.filteringkey: '<searchValue2>'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.applications[*].<searchBy1> contains '<searchValue1>'
    And match response.applications[*].<searchBy2> contains '<searchValue2>'
  Examples:
    | searchBy1 | searchValue1 | searchBy2 | searchValue2 |
    | name | Random Application 1| owner | Random |
    | domain | Payment | owner | Orange Money |


  @sort
  Scenario Outline: Sorting_sort Applications data 
    Given path 'applications'
    And params { page_num:1, page_size:10, sort_by:'<sortBy>', sort_order:'<sortOrder>', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    * def actual = $response.applications[*].<sortBy>
    * def sorted = sort(actual,'<sortOrder>')
    * match sorted == actual
  Examples:
      | sortBy | sortOrder |  
     # | num_of_products | asc |
      | name | desc |
      | domain | asc |
      | obsolescence_risk | desc|
    


  @get
  Scenario: get Application
    Given path 'applications'
    And params { page_num:1, page_size:100, sort_by:'name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    # And match response.applications[*].application_id contains data.getApp.application_id
    

## Instances 

  @schema
  Scenario: Schema validation for get Instances
    Given path 'application/instances'
    * params { page_num:1, page_size:10, sort_by:'instance_id', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    * response.totalRecords == '#number? _ >= 0'
    * match response.instances == '#[] data.schema_instance'

  @search
  Scenario: Searching_Filter Instances by Application Id
    Given path 'application/instances'
    And params { page_num:1, page_size:10, sort_by:'instance_environment', sort_order:'desc', scopes:'#(scope)'}
    And params {search_params.application_id.filteringkey: '#(data.getInstance.application_id)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    # And match each response.instances[*].name == data.getInstance.name
    * remove data.getInstance.application_id
    And match response.instances contains data.getInstance

  # @search
  # Scenario: Searching_Filter Instances by product Id
  #   Given path 'instances'
  #   And params { page_num:1, page_size:10, sort_by:'instance_environment', sort_order:'desc', scopes:'#(scope)'}
  #   And params {search_params.product_id.filteringkey: '#(data.getInstance.products[0])'}
  #   When method get
  #   Then status 200
  #   And response.totalRecords > 0
  #   * remove data.getInstance.application_id
  #   And match  response.instances contains data.getInstance



## Creation API

# @create @ignore
#   Scenario: Create Application
#     Given path 'applications'
#     And request data.createApp
#     When method post
#     Then status 200
#     And match response.success == true

  # Scenario: To verify Application is not created for incorrect body
  #   Given path 'applications'
  #   And request { "application_id": 'wrong_4' ,"wrong_param":"value" , "scope": "France"}
  #   When method post
  #   Then status 400
  #   And match response.success == false


  # Scenario: To verify scope is mandetory
  #   Given path 'applications'
  #   * remove data.createApp.scope
  #   * request data.createApp
  #   When method post
  #   Then status 400
  #   And match response.success == false

  # @update @ignore
  # Scenario: Update Application
  #   Given path 'applications'
  #   And request ({ "application_id": data.createApp.application_id, "name": 'dummyUpdated', "version": "0.1.4", "owner": "OrangeUpdated", "scope": "France"})
  #   When method post
  #   Then status 200
  #   And match response.success == true

  # @update
  # Scenario: To verify scope is not Updated
  #   Given path 'applications'
  #   And request ({ "application_id": data.createApp.application_id, "name": 'dummyUpdated', "version": "0.1.4", "owner": "OrangeUpdated", "scope": "Dummy"})
  #   When method post
  #   Then status 200
  #   And match response.success == true


  # @delete
  # Scenario: Delete Application
  #   Given path 'applications',application_id
  #   When method delete
  #   Then status 200