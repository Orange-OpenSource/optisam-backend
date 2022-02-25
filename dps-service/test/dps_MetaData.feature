@dps
Feature: DPS Service Test - Metadata : admin user

  Background:
    * url dpsServiceUrl+'/api/v1/dps'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = "AUT"


  Scenario: Schema validation for List uploads Metadata 
    Given path 'uploads/metadata'
    * params { page_num:1, page_size:50, sort_by:'upload_id', sort_order:'desc' ,scope:'#(scope)'}
    When method get
    Then status 200 
    And response.totalRecords == '#number? _ >= 0'
    And match response.uploads == '#[] data.schema_data'


  @pagination
  Scenario Outline: To verify Pagination is working for list metadata for page_size <page_size>
    Given path 'uploads/metadata'
    And params { page_num:1, page_size:'<page_size>', sort_by:'upload_id', sort_order:'desc', scope:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match $.uploads == '#[_ <= <page_size>]'

  Examples:
    | page_size |
    | 50 |
    | 100 |
    | 200 |

   Scenario Outline: To verify Pagination with invalid inputs for list data files
    Given path 'uploads/metadata'
    And params { page_num:'<page_num>', page_size:'<page_size>', sort_by:'name', sort_order:'desc', scope:'#(scope)'}
    When method get
    Then status 400
   Examples:
    | page_size | page_num |
    | 5 | 5 |
    | 10 | 0 |
    | "A" | 5 |   


  # TODO: verify file name and status
  @sort @ignore
  Scenario Outline: To verify sorting of Meta Data on data management by <sortBy>
    Given path 'uploads/metadata'
    And params { page_num:1, page_size:50, sort_by:'<sortBy>', sort_order:'<sortOrder>' , scope:'#(scope)'}
    When method get
    Then status 200
    And  response.totalRecords > 0
    * def actual = $response.uploads[*].<sortBy>
    * def sorted = sort(actual,'<sortOrder>')
    * match sorted == actual
   
  Examples:
      | sortBy | sortOrder |  
      | status | desc | 
      | status | asc | 
      | uploaded_by | asc|
      | uploaded_on | asc|


  Scenario:  To verify the error for metadata when Scope Field is Missing
    Given path 'uploads/metadata'
    * params { page_num:1, page_size:50, sort_by:'upload_id', sort_order:'desc' }
    When method get
    Then status 400 
    And response.totalRecords == '#number? _ = 0'  