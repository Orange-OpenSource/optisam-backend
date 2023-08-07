@NomenativeUser
Feature: Concurrent_user test for Normal user

  Background:
  * url productServiceUrl+'/api/v1/product'
  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
  * callonce read('../common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'

  Scenario: Validating Concurrent Page API (Individual)
    Given path 'concurrent' 
   # And params {scope:'#(scope)'}
    And params {page_num:1, page_size:50, sort_by:product_name, sort_order:asc, scopes:'#(scope)', is_aggregation:false}
    When method get
    Then status 200
    And response.concurrent_user.product_name == '#name? _ = string'
    And response.concurrent_user.product_editor == '#name? _ = string'

  Scenario: Add concurrent Product user(Individual)
    Given path 'concurrent'
    And request data.Create_Concurrent_User
    When method post 
    Then status 200

  Scenario: Check the concurrent user list in Individual tab
    Given path 'concurrent' 
    And params {page_num:1, page_size:50, sort_by:product_name, sort_order:asc, scopes:'#(scope)', is_aggregation:false}
    When method get
    Then status 200

  Scenario: Edit the user in Individual list
    Given path 'concurrent' 
    And params {page_num:1, page_size:50, sort_by:product_name, sort_order:asc, scopes:'#(scope)', is_aggregation:false}
    When method get
    Then status 200
    And def myId = karate.jsonPath(response, '$.concurrent_user[?(@.product_name=="Adobe Media Server")].id')[0]
    * header Authorization = 'Bearer '+access_token 
    Given path 'concurrent/' + myId 
    And request data.Edit_Concurrent_user
    When method put 
    Then status 200

  Scenario: Export the Products list in Individual
    Given path 'concurrent/users/export'
    And params {sort_by:aggregation_name,sort_order:asc,scopes:'#(scope)',is_aggregation:false}
    When method get
    Then status 200
  
  Scenario Outline: Search products in Individual list
    Given path 'concurrent' 
    And params {page_num:1, page_size:50, sort_by:product_name, sort_order:asc, scopes:'#(scope)', is_aggregation:false}
    And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
    When method get
    Then status 200
    Examples:
    | searchBy | searchValue |
    | product_editor | Adobe |
    | profile_user | QA |
    | product_name | Adobe Media Server | 
    | product_version | 1.0.1 |
  
  Scenario: Delete the user in Individual list
    Given path 'concurrent' 
    And params {page_num:1, page_size:50, sort_by:product_name, sort_order:asc, scopes:'#(scope)', is_aggregation:false}
    When method get
    Then status 200
    And def myId = karate.jsonPath(response, '$.concurrent_user[?(@.product_name=="Adobe Media Server")].id')[0]
    * header Authorization = 'Bearer '+access_token 
    Given path 'concurrent/' + myId
    And params {scope:'#(scope)'}
    When method Delete
    Then status 200

  Scenario: Validating Concurrent Page API (Aggregation)
    Given path 'concurrent' 
    And params {page_num:1, page_size:50, sort_by:product_name, sort_order:asc, scopes:'#(scope)', is_aggregation:true}
    When method get
    Then status 200

  Scenario: Add concurrent Product user(Aggregation)
    Given path 'concurrent'
    And request data.add_concurrent_user
    When method post 
    Then status 200 

  
  Scenario: Edit the user in Aggregation list
    Given path 'concurrent' 
    And params {page_num:1, page_size:50, sort_by:product_name, sort_order:asc, scopes:'#(scope)', is_aggregation:true}
    When method get
    Then status 200
    And def aggId = karate.jsonPath(response, '$.concurrent_user[?(@.aggregation_name=="Oracle_RAC")].id')[0]
    * header Authorization = 'Bearer '+access_token 
    Given path 'concurrent/' + aggId
    And request data.edit_aggregation_user
    When method put
    Then status 200

  Scenario: Export the Products list in Aggregation
    Given path 'concurrent/users/export'
    And params {sort_by:aggregation_name,sort_order:asc,scopes: '#(scope)',is_aggregation:true}
    When method get
    Then status 200

  Scenario Outline: Search products in Aggregation list
    Given path 'concurrent' 
    And params {page_num:1, page_size:50, sort_by:product_name, sort_order:asc, scopes:'#(scope)', is_aggregation:true}
    And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
    When method get
    Then status 200
    Examples:
    | searchBy | searchValue |
    | number_of_users | 1211 |
    | aggregation_name | simtest1 |
    | team | QA | 
    | product_version | 1.0.1 |
  
  Scenario: Delete the user in Aggregation list
    Given path 'concurrent' 
    And params {page_num:1, page_size:50, sort_by:product_name, sort_order:asc, scopes:'#(scope)', is_aggregation:true}
    When method get
    Then status 200
    And def aggId = karate.jsonPath(response, '$.concurrent_user[?(@.aggregation_name=="Oracle_RAC")].id')[0]
    * header Authorization = 'Bearer '+access_token 
    Given path 'concurrent/' + aggId
    And params {scope:'#(scope)'}
    When method Delete
    Then status 200 
  

