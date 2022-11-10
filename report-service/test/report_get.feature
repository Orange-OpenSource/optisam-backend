@report
Feature: Report Service Test - Get Report : Admin

  Background:
    # * def reportServiceUrl = "https://optisam-report-int.apps.fr01.paas.tech.orange"
    * url reportServiceUrl+'/api/v1'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

  @get
  Scenario: Get all report types
    Given path 'report/types'
    When method get
    Then status 200
  * match response.report_type[*].report_type_name contains ["Compliance"]
  * match response.report_type[*].report_type_name contains ["ProductEquipments"]

    
   @schema
  Scenario:  Schema validation for get Reports
    Given path 'reports'
    * def schema = {report_id: '#number', report_type: '#string', report_status: '#string', created_by: '#string', created_on: '##string', editor: '##string'}
    * params { page_num:1, page_size:50, sort_order:'desc', sort_by:'created_on',scope:'#(scope)'}
    When method get
    Then status 200
    * response.totalRecords == '#number? _ >= 0'
    * match response.reports == '#[] schema'

  @get
  Scenario: Get all the reports
    Given path 'reports'
    And params { page_num:1, page_size:50, sort_order:'desc', sort_by:'created_on',scope:'#(scope)'}
    When method get
    Then status 200
    * response.totalRecords == '#number? _ >= 0'
    * match response.reports[*].created_by contains ['admin@test.com']


  @pagination
  Scenario Outline: To verify Pagination on Reporting Page
    Given path  'reports'
    And params { page_num:1, page_size:'<page_size>', sort_by:'created_on', sort_order:'desc', scope:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match $.reports == '#[_ <= <page_size>]'
   Examples:
    | page_size |
    | 50 |
    | 100 |
    | 200 |

  @get
  Scenario: Sorting_Get all the reports sorted by report_id
    Given path 'reports'
    And params { page_num:1, page_size:50, sort_order:'desc', sort_by:'report_id', scope:'#(scope)'}
    When method get
    Then status 200
    * def actual = $response.reports[*].report_id
    * def sorted = sortNumber(actual,'desc')
    * match sorted == actual

  @get
  Scenario: Get reports by report_Id
    Given path 'reports'
    And params { page_num:1, page_size:50, sort_order:'asc', sort_by:'created_on',scope:'#(scope)'}
    When method get
    Then status 200
    # * response.totalRecords == '#number? _ >= 0'  
    * def report_id = karate.jsonPath(response.reports,"$.[?(@.report_status=='COMPLETED')].report_id")[0]  
    Given path 'report', report_id
    And params { page_num:1, page_size:10, sort_order:'asc', sort_by:'created_on',scope:'#(scope)'}
    * header Authorization = 'Bearer '+access_token
    When method get
    Then status 200
  #  * match response.report_data == '#string'

   @sort
  Scenario Outline: Sorting of Reports data 
    Given path 'reports'
    And params { page_num:1, page_size:100, sort_by:'<sortBy>', sort_order:'<sortOrder>', scope:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    * def actual = $response.reports[*].<sortBy>
    * def sorted = sort(actual,'<sortOrder>')
    * match sorted == actual
  Examples:
      | sortBy | sortOrder |  
      #  | report_status | desc |
       | created_on | asc |
       | created_on | desc|


   Scenario Outline: To verify Pagination on Reporting with Invalid inputs
    Given path 'reports'
    And params { page_num:'<page_num>', page_size:'<page_size>', sort_by:'name', sort_order:'desc', scope:'#(scope)'}
    When method get
    Then status 400
   Examples:
    | page_size | page_num |
    | 5 | 5 |
    | 10 | 0 |
    | "A" | 5 |        


   





    