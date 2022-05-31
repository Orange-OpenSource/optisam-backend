@account @group
Feature: Account Service Test for Group Management : admin User

  Background:
  # * def accountServiceUrl = "https://optisam-account-int.apps.fr01.paas.tech.orange"
    * url accountServiceUrl+'/api/v1/account'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'
    # * def err = {"error": "string","code": 0,"message": "string","details": [{"type_url": "string","value": "string"}]}


  @get
  Scenario: Schema Validation for direct groups of a user
    Given path 'admin/direct_groups'
    When method get
    Then status 200
    * match response.groups == '#[_ > 0] data.schema_grp'

  @get
  Scenario: Schema Validation for all groups in scope
    Given path 'admin/groups'
    When method get
    Then status 200
    * match response.groups == '#[_ > 0] data.schema_grp'

  @get
  Scenario: To verify admin can get all the direct groups that he belongs
    Given path 'admin/direct_groups'
    When method get
    Then status 200
    * match response.groups[*].ID contains "1"
    * match response.groups[*].name contains ['ROOT']

  @get
  Scenario: To verify admin can get all the groups that he belongs
    Given path 'admin/groups'
    When method get
    Then status 200
    * match response.groups[*].name contains ['ROOT']
    * match response.groups contains data.group

  @get
  Scenario: To verify admin can get the group by group id
    Given path 'admin/groups/1/groups'
    When method get
    Then status 200
     * match response.groups contains data.group

  @get
  Scenario: To verify admin can get the users list of a group
    Given path 'admin/groups'
    When method get
    Then status 200
    * def group_id = karate.jsonPath(response.groups,"$.[?(@.name=='AUT')].ID")[0]  
    * header Authorization = 'Bearer '+access_token 
    Given path 'admin/groups',group_id,'users'
    When method get
    Then status 200
    * match response.users[*].user_id contains ['testadmin@test.com']
    * match response.users[*].user_id contains ['testuser@test.com']

  @create
  Scenario: To verify admin user can create groups and delete it
    Given path 'admin/groups' 
    * def createGroup = {"scopes": ['#(scope)'],"parent_id": "1"}
    * set createGroup.name = "apitest_grp_" + now()
    And request createGroup
    When method post
    Then status 200
    * match response contains createGroup
    * def group_id = response.ID
    Given path 'admin/groups/' + group_id
    * header Authorization = 'Bearer '+access_token
    When method delete
    Then status 200
    * response.success == true


  @create
  Scenario: Verify admin user can add and remove user from a group
  # create group
    Given path 'admin/groups' 
    * def createGroup = {"scopes": ['#(scope)'],"parent_id": "1"}
    * set createGroup.name = "apitest_grp_" + now()
    And request createGroup
    When method post
    Then status 200
    * match response contains createGroup
    * def group_id = response.ID
    # create user
    Given path 'user'
    * header Authorization = 'Bearer '+access_token
    * set data.createUserAccount.user_id = now() + "@test.com"
    * set data.createUserAccount.groups = null
    And request data.createUserAccount
    When method post
    Then status 200
    # add user to group
    Given path 'admin/groups/' + group_id , 'users/add'
    * header Authorization = 'Bearer '+access_token
    * set data.adduser.group_id = group_id 
    And request data.adduser
    When method put
    Then status 200
    # delete user from group
    Given path 'admin/groups/' + group_id , 'users/delete'
    * header Authorization = 'Bearer '+access_token
    And request data.adduser
    When method put
    Then status 200
    # delete user
    Given path data.createUserAccount.user_id
    * header Authorization = 'Bearer '+access_token
    When method delete
    Then status 200
    * response.success == true 
    # delete group
    Given path 'admin/groups/' + group_id 
    * header Authorization = 'Bearer '+access_token
    When method delete
    Then status 200
    * response.success == true 
    

  # @delete
  # Scenario:delete groups
  #   Given path 'admin/groups/' + 19
  #   When method delete
  #   Then status 200
  #   * response.success == true

  
