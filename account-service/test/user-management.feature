@account
Feature: Account Service Test

  Background:
  # * def accountServiceUrl = "https://optisam-account-int.apps.fr01.paas.tech.orange"
    * url accountServiceUrl+'/api/v1/account'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    # * def err = {"error": "string","code": 0,"message": "string","details": [{"type_url": "string","value": "string"}]}


  @schema
  Scenario: Schema validation for get all users request
    Given path 'users'
    And params { user_filter.all_users:false}
    When method get
    Then status 200
    * match response.users == '#[] data.schema_users'

  @get
  Scenario: Verify Get all the users present
    Given path 'users'
    When method get
    Then status 200
     And match response.users[*] contains data.testadmin
    
  @schema
  Scenario: Schema validation for get User by UserID
    Given path 'admin@test.com'
    * def schema = {role:'#string' , user_id: '#string', first_name: '#string', last_name: '##string', locale:'#string', "profile_pic":'##',"first_login":'##boolean'}
    When method get
    Then status 200
  * match response == schema

  @get
  Scenario: Verify Get user by userID
    Given path 'admin@test.com'
    When method get
    Then status 200
    * match response.user_id == 'admin@test.com'
    * match response.role == 'SUPER_ADMIN'
    
  @create
  Scenario: Create User account with Admin role and delete it
    Given path 'admin/groups'
    When method get
    Then status 200
    * def group_id = karate.jsonPath(response.groups,"$.[?(@.name=='AUT')].ID")[0]  
    Given path 'user' 
    * header Authorization = 'Bearer '+access_token
    * set data.createAdminAccount.groups[0] = group_id
    * set data.createAdminAccount.user_id = now() + "@test.com"
    And request data.createAdminAccount
    When method post
    Then status 200
    And match response == data.createAdminAccount
    * path data.createAdminAccount.user_id
    * header Authorization = 'Bearer '+access_token
    * method delete
    * status 200
    * match response.success == true

  @create
  Scenario: Create User account with User role and delete it
    Given path 'admin/groups'
    When method get
    Then status 200
    * def group_id = karate.jsonPath(response.groups,"$.[?(@.name=='AUT')].ID")[0]  
    Given path 'user' 
    * header Authorization = 'Bearer '+access_token
    * set data.createUserAccount.groups[0] = group_id
    * set data.createUserAccount.user_id = now() + "@test.com"
    And request data.createUserAccount
    When method post
    Then status 200
    And match response == data.createUserAccount
    Given path  data.createUserAccount.user_id
    * header Authorization = 'Bearer '+access_token
    When method delete
    Then status 200
    And match response.success == true

  # @update
  # Scenario: Verify user can Update the account
  #   Given path 'accounts' 
  #   And request data.createUserAccount
  #   When method post
  #   Then status 200
  #   Given path 'accounts' ,data.createUserAccount.user_id
  #   * header Authorization = 'Bearer '+access_token
  #   * set data.createUserAccount.role = 'ADMIN'
  #   * remove data.createUserAccount.groups
  #   And request data.createUserAccount
  #   When method put
  #   Then status 200
  #   # * match response == data.createUserAccount
  #   * path 'accounts' ,data.createAdminAccount.user_id
  #   * header Authorization = 'Bearer '+access_token
  #   * method delete
  #   * status 200
  #   * match response.success == true


# TODO: get account uses token info and not path parameter
  # @delete @ignore
  # Scenario:Delete account
  #   Given path 'accounts' ,data.createUserAccount.user_id
  #   When method delete
  #   Then status 200
  #   # * path 'accounts',data.createUserAccount.user_id
  #   # * header Authorization = 'Bearer '+access_token
  #   # * method get
  #   # * status 500

  @delete
  Scenario:Delete account that does not exist
    Given path 'invaid@invalid.com'
    When method delete
    Then status 404
    * response.message = "user does not exist"
    
  @create
  Scenario: Verify UserID is unique field
    Given path 'user'
    And request data.testuser
    When method post
    Then status 400
    # * response.error = data

  @create
  Scenario: Verify User can not be created with SUPER_ADMIN role
    Given path 'user'
    * set data.createAdminAccount.role = 'SUPER_ADMIN'
    And request data.createAdminAccount
    When method post
    Then status 400
    * response.error = "only admin and user roles are allowed"


   @create
  Scenario: Verify user role is updated by admin
    Given path 'admin/groups'
    When method get
    Then status 200
    * def group_id = karate.jsonPath(response.groups,"$.[?(@.name=='AUT')].ID")[0]  
    Given path 'user' 
    * header Authorization = 'Bearer '+access_token
    * set data.createUserAccount.groups[0] = group_id
    * set data.createUserAccount.user_id = now() + "@test.com"
    And request data.createUserAccount
    When method post
    Then status 200
    And match response == data.createUserAccount
    Given  path  data.createUserAccount.user_id
    * header Authorization = 'Bearer '+access_token
    * set data.createUserAccount.groups[0] = group_id
     * set data.createUserAccount.role = 'ADMIN'
    And request data.createUserAccount
    When method put
    Then status 200
    Given  path  data.createUserAccount.user_id
    * header Authorization = 'Bearer '+access_token
    When method delete
    Then status 200
    And match response.success == true  
