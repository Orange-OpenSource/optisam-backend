@dps
Feature: DPS Service Test - Data : admin user

  Background:
    * url dpsServiceUrl+'/api/v1/dps'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = "AUT"


## notify dps api is used internally by import - no test required

#  Scenario: To verify user is able to Uploads global data  
#     Given path 'uploads/notify'
#     Given request { "scope":"TST", 	"type" :"globaldata", "uploadBy": "admin@test.com", "files": ["TST_temp.xlsx"]}
#     And header Accept = 'application/json'
#     When method post
#     Then status 200 
#     And response.success == '#boolean? _ >= true'
 

  Scenario: Schema Validation for List uploads data 
    Given path 'uploads/data'
    * params { page_num:1, page_size:50, sort_by:'upload_id', sort_order:'desc', scope:'#(scope)'}
    When method get
    Then status 200 
    And response.totalRecords == '#number? _ >= 0'
    And match response.uploads == '#[] data.schema_data'


  @pagination
  Scenario Outline: To verify Pagination is working on for list data files for page_size <page_size>
    Given path 'uploads/data'
    And params { page_num:1, page_size:'<page_size>', sort_by:'status', sort_order:'desc', scope:'#(scope)'}
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
    Given path 'uploads/data'
    And params { page_num:'<page_num>', page_size:'<page_size>', sort_by:'name', sort_order:'desc', scope:'#(scope)'}
    When method get
    Then status 400
   Examples:
    | page_size | page_num |
    | 5 | 5 |
    | 10 | 0 |
    | "A" | 5 |   


 @sort
  Scenario Outline: To verify sorting of data on data management by <sortBy>
    Given path 'uploads/data'
    And params { page_num:1, page_size:50, sort_by:'<sortBy>', sort_order:'<sortOrder>', scope:'#(scope)'}
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
      | uploaded_by | desc|
      | uploaded_on | asc|
      | uploaded_on | desc|

  
    Scenario:  To verify the error for data when Scope Field is Missing
    Given path 'uploads/data'
    * params { page_num:1, page_size:50, sort_by:'upload_id', sort_order:'desc' }
    When method get
    Then status 400 
    And response.totalRecords == '#number? _ = 0'  

  ## TODO: Verify failed records (dependency to upload wrong file)
  # Scenario: Get Failed records of Uploaded data
  #   Given path 'failed/data'
  #   And params { page_num:1, page_size:10,  scope:'#(scope)'}
  #   And params { upload_id:200}
  #   When method get
  #   Then status 200 

