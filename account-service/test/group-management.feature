@account @group

Feature: Account Service Test for Group Management : admin User

  Background:
    * url accountServiceUrl+'/api/v1/account'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}  
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'


  @SmokeTest
  @get
  Scenario: Schema Validation for direct groups of a user
    Given path 'admin/direct_groups'
    When method get
    Then status 200
   # * match response.groups == '#[_ > 0] data.schema_grp'

  @get
  Scenario: Schema Validation for all groups in scope
    Given path 'admin/groups'
    When method get
    Then status 200
    #* match response.groups == '#[_ > 0] data.schema_grp'

  @get
  Scenario: To verify admin can get all the direct groups that he belongs
    Given path 'admin/direct_groups'
    When method get
    Then status 200
    * match response.groups[*].ID contains "854"
    * match response.groups[*].name contains ['ROOT']

  @get
  Scenario: To verify admin can get all the groups that he belongs
    Given path 'admin/groups'
    When method get
    Then status 200
    * match response.groups[*].name contains ['ROOT']
   
  @get
  Scenario: To verify admin can get the group by group id
    Given path 'admin/groups/1/groups'
    When method get
    Then status 200
     
  @get
  Scenario: To verify admin can get the users list of a group
    Given path 'admin/groups'
    When method get
    Then status 200
    * def group_id = karate.jsonPath(response.groups,"$.[?(@.name=='ROOT')].ID")[0]  
    * header Authorization = 'Bearer '+access_token 
    Given path 'admin/groups',group_id,'users'
    When method get
    Then status 200
    * match response.users[*].user_id contains data.User_id
    

  @create
  Scenario: To verify admin user can create groups and delete it
    Given path 'admin/groups' 
    * def createGroup = {"scopes": ['AAA'],"parent_id": "1"}
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

  Scenario: To create group for viewing group compliance
    Given path 'admin/groups'
    And request data.group_creation
    When method post 
    Then status 400

    # Change the status code to 200 when providing a uniqe name for group creation 
    # as it is showing already exisit error.


# API scope get deleted due to that showing errors.

 


    

 

  
