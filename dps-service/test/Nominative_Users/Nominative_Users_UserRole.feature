@NomenativeUser
Feature: Nominative_user test for Admin user

  Background:
  * url productServiceUrl+'/api/v1/product'
  #* def credentials = {username:#(UserAccount_Username), password:#(UserAccount_password)}
  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
  * callonce read('../common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'

@SmokeTest
    Scenario: Get all Nominative_user
        Given path 'nominative/users'
        And params {page_num:1, page_size:50, sort_by:'activation_date', sort_order:'asc', scopes:'#(scope)'}
        When method get
        Then status 200

    Scenario: Validating Nominative Page API 
        Given path 'nominative/users' 
      # And params {scope:'#(scope)'}
        And params {page_num:1, page_size:50, sort_by:product_name, sort_order:asc, scopes:'#(scope)', is_product:true}
        When method get
        Then status 200

    Scenario: Check the Nominative user list in Individual tab
      Given path 'nominative/users' 
      And params {page_num:1, page_size:50, sort_by:aggregation_name, sort_order:asc, scopes:'#(scope)', is_product:true}
      When method get
      Then status 200

    Scenario: Check the Nominative user list in Aggregation tab
      Given path 'nominative/users'
      And params {page_num:1, page_size:50, sort_by:aggregation_name, sort_order:asc, scopes:'#(scope)', is_aggregation:false}
      When method get
      Then status 200


    Scenario: Add Nominative Product user(Individual)
      Given path 'nominative/users'
      And request data.add_nominative_user
      When method post 
      Then status 200
      
    Scenario: Edit the user in Individual list
      Given path 'nominative/users' 
      And request data.Edit_Nominative_user
      When method post
      Then status 200
  
    Scenario: Export the Products list in Individual
      Given path 'nominative/users/export'
      And params {sort_by:aggregation_name,sort_order:asc,scopes:'#(scope)',is_aggregation:false}
      When method get
      Then status 200
    
    Scenario Outline: Search products in Individual list
      Given path 'nominative/users' 
      And params {page_num:1, page_size:50, sort_by:product_name, sort_order:asc, scopes:'#(scope)', is_aggregation:false}
      And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
      When method get
      Then status 200
      Examples:
      | searchBy | searchValue |
      | product_editor | Axicon Verifer ID |
      | profile_user | QA |
      | product_name |  Axicon Verifer ID | 
     
    
    Scenario: To verfiy deletion of User in Individual List
      Given path 'nominative/users'
      And params {page_num: 1,page_size: 50,sort_by: 'product_name',sort_order: 'asc',scopes:'#(scope)',is_product: 'true'}
      When method get
      Then status 200
      * def del_id = karate.jsonPath(response.nominative_user,'$.[?(@.first_name== "Test_page")].id')[0]  
      * header Authorization = 'Bearer '+access_token 
      Given path 'nominative/users',del_id
        And params {scope:'#(scope)'}
        When method Delete
        Then status 200
  
      
    Scenario: To verify admin can create new Aggregation -Redhat
      Given path 'nominative/users'
      And request data.createAgg
      When method post
      Then status 200
      * header Authorization = 'Bearer '+access_token
      Given path 'nominative/users'
      And params {page_num: 1,page_size: 50,sort_by: 'product_name',sort_order: 'asc',scopes:'#(scope)',is_product: 'true'}
      When method get
      Then status 200
      
        
    Scenario: Validating Concurrent Page API (Aggregation)
      Given path 'concurrent' 
      And params {page_num:1, page_size:50, sort_by:product_name, sort_order:asc, scopes:'#(scope)', is_aggregation:true}
      When method get
      Then status 200
  
    Scenario: Add Nominative Product user(Aggregation)
      Given path 'nominative/users'
      And request data.create_nominative_user
      When method post
      Then status 200
  
    
    Scenario: Edit the user in Aggregation list
      Given path 'nominative/users' 
      And params {page_num:1, page_size:50, sort_by:product_name, sort_order:asc, scopes:'#(scope)', is_aggregation:true}
      When method get
      Then status 200
     And def aggId = karate.jsonPath(response.nominative_user, '$.[?(@.aggregation_name=="redhat_CS")].id')[0]
      * header Authorization = 'Bearer '+access_token 
      Given path 'nominative/users'
      And request data.edit_aggregation_user
      When method post
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
      Given path 'nominative/users' 
      And params {page_num:1, page_size:50, sort_by:product_name, sort_order:asc, scopes:'#(scope)', is_aggregation:true}
      When method get
      Then status 200
      And def aggId = karate.jsonPath(response.nominative_user, '$.[?(@.aggregation_name=="Oracle_RAC")].id')[0]
      * header Authorization = 'Bearer '+access_token 
      Given path 'nominative/users',aggId
      And params {scope:'#(scope)'}
      When method Delete
      Then status 200