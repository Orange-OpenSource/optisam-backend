@account

Feature: Account Service Test

  Background:
  # * def accountServiceUrl = "https://optisam-account-int.apps.fr01.paas.tech.orange"
    * url accountServiceUrl+'/api/v1/account'
    #* def credentials = {username:'admin@test.com', password: 'Welcome@123'}
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
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
    

  # @get
  # Scenario: Verify Get user by userID
  #   Given path 'admin@test.com'
  #   When method get
  #   Then status 200
    
    
  @create
  Scenario: Create User account with Admin role and delete it
    Given path 'admin/groups'
    When method get
    Then status 200
    * def group_id = karate.jsonPath(response.groups,"$.[?(@.name=='API')].ID")[0]  
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
    * def group_id = karate.jsonPath(response.groups,"$.[?(@.name=='API')].ID")[0]  
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

  @update
  Scenario: Verify admin can Update user role to admin for the account 
    Given path 'admin/groups'
    When method get
    Then status 200
    * def group_id = karate.jsonPath(response.groups,"$.[?(@.name=='API')].ID")[0]  
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
     * remove data.createUserAccount.groups
     * set data.createUserAccount.role = "ADMIN"
     And request data.createUserAccount
     When method put
     Then status 200
     Given path  data.createUserAccount.user_id
     * header Authorization = 'Bearer '+access_token
     When method delete
     Then status 200
     And match response.success == true


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


    @update
  Scenario: Verify admin can Update admin to user role for the account 
    Given path 'admin/groups'
    When method get
    Then status 200
    * def group_id = karate.jsonPath(response.groups,"$.[?(@.name=='API')].ID")[0]  
    Given path 'user' 
    * header Authorization = 'Bearer '+access_token
    * set data.createAdminAccount.groups[0] = group_id
    * set data.createAdminAccount.user_id = now() + "@test.com"
    And request data.createAdminAccount
    When method post
    Then status 200
    And match response == data.createAdminAccount
    Given path  data.createAdminAccount.user_id
     * header Authorization = 'Bearer '+access_token
     * remove data.createAdminAccount.groups
     * set data.createAdminAccountt.role = "USER"
     And request data.createAdminAccount
     When method put
     Then status 200
     Given path  data.createAdminAccount.user_id
     * header Authorization = 'Bearer '+access_token
     When method delete
     Then status 200
     And match response.success == true