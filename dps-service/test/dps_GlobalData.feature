@dps
Feature: DPS Service Test - Global Data : admin user

  Background:
    * url dpsServiceUrl+'/api/v1/dps'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = "AUT"



  Scenario: Schema Validation for List uploads global data 
    Given path 'uploads/globaldata'
    * params { page_num:1, page_size:50, sort_by:'upload_id', sort_order:'desc' , scope : '#(scope)'}
    When method get
    Then status 200 
    And response.totalRecords == '#number? _ >= 0'
    And match response.uploads == '#[] data.schema_data'


  Scenario:  To verify the error for globaldata when Scope Field is Missing
    Given path 'uploads/globaldata'
    * params { page_num:1, page_size:50, sort_by:'upload_id', sort_order:'desc' }
    When method get
    Then status 400 
    And response.totalRecords == '#number? _ = 0'  
 

