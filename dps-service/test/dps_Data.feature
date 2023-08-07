@dps
Feature: DPS Service Test - Data : admin user

  Background:
    * url dpsServiceUrl+'/api/v1/dps'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = "API"



#  Scenario: To verify user is able to Uploads global data  
#     Given path 'uploads/notify'
#     Given request { "scope":"TST", 	"type" :"globaldata", "uploadBy": "xyz", "files": ["TST_temp.xlsx"]}
#     And header Accept = 'application/json'
#     When method post
#     Then status 200 
#     And response.success == '#boolean? _ >= true'
 
Scenario: To get the details of Infrastructure inventory 
  Given path 'uploads/globaldata'
  And params {page_num:1,page_size:50,sort_order:'desc',sort_by:'uploaded_on',scope:'#(scope)'}
  When method get 
  Then status 200

Scenario Outline: To verify the Pagination on Infrastructure inventory Page
  Given path 'uploads/globaldata'
  And params { page_num:1,page_size:<pageSize>,sort_order:<sortOrder>,sort_by:'uploaded_on',scope:'#(scope)'}
When method get 
Then status 200
And response.totalRecords > 0
   

Examples:
     |pageSize|sortOrder|
     |50       |asc      |
     |100      |asc      |
     |200      |desc     |


    Scenario Outline: To verify the Pagination on Infrastructure inventory Page with invalid inputs 
      Given path 'uploads/globaldata'
      And params { page_num:1,page_size:<pageSize>,sort_order:<sortOrder>,sort_by:'uploaded_on',scope:'#(scope)'}
    When method get 
    Then status 400
    
    Examples:
         |pageSize|sortOrder|
         |5       |asc      |
         |1       |asc      |
         |'A'     |desc     |


    Scenario: To get Log Files 
      Given path 'uploads/data' 
      And params {page_num:1,page_size:50,sort_order:'desc',sort_by:'uploaded_on',scope:'#(scope)'}
      When method get 
      Then status 200
      And match response.totalRecords == '#number? _ >= 0'
      And match response.uploads[*].uploaded_on  contains ["2023-03-31T06:35:24.345745Z"]

    Scenario Outline: To verify the Pagination on Log Files
      Given path 'uploads/data' 
      And params { page_num:1,page_size:<pageSize>,sort_order:<sortOrder>,sort_by:'uploaded_on',scope:'#(scope)'}
      When method get 
      Then status 200 
      And response.totalRecords > 0
   

      Examples:
     |pageSize|sortOrder|
     |50       |asc      |
     |100      |asc      |
     |200      |desc     |


    Scenario Outline: To verify the Pagination on Log Files with invalid inputs 
      Given path 'uploads/data' 
      And params { page_num:1,page_size:<pageSize>,sort_order:<sortOrder>,sort_by:'uploaded_on',scope:'#(scope)'}
      When method get 
      Then status 400

      Examples:
      |pageSize|sortOrder|
      |5       |asc      |
      |1       |asc      |
      |'A'     |desc     |


    Scenario:To get the deletion log 
      Given path 'deletions'
      And params {scope:'#(scope)',sort_by:'created_on',sort_order:'asc',page_num:1,page_size:50}
      When method get 
      Then status 200
      And match response.totalRecords == '#number? _ >= 0'
      And match response.deletions[*].created_on contains ["2022-04-27T08:04:49.861867Z"]

    Scenario Outline: To verify the Pagination on Deletion Log 
      Given path 'deletions'
      And params {scope:'#(scope)',sort_by:'created_on',sort_order:<sortOrder>,page_num:1,page_size:<pageSize>}
      When method get 
      Then status 200

      Examples:
     |pageSize|sortOrder|
     |50       |asc      |
     |100      |asc      |
     |200      |desc     |

    Scenario Outline: To verify the Pagination on Deletion Log  with Invalid Input
      Given path 'deletions'
      And params {scope:'#(scope)',sort_by:'created_on',sort_order:<sortOrder>,page_num:1,page_size:<pageSize>}
      When method get 
      Then status 400

      Examples:
     |pageSize|sortOrder|
     |5       |asc      |
      |1       |asc      |
      |'A'     |desc     |





      



      








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

  

