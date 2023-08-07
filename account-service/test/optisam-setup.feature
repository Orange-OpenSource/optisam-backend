@setup 

Feature: Setup Pre-requisite for Optisam Test - API

  Background:
    * url accountServiceUrl+'/api/v1/account'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = "API"


# Create Test Scopes
  Scenario: Creating a scope
    Given path 'scopes'
    And request data.createScope
    When method post
    Then status 200
    * match response.success == true
    

#Delete Scope
  #Scenario: Delete created scope
  #  Given path 'scopes'
  #  * call read('common.feature') credentials
  #  * def access_token = response.access_token
  #  * header Authorization = 'Bearer '+access_token
  #  When method get
  #  Then  status 200
  #    * print 'scope_code: ' + data.createScope.scope_code
  #* def scope_code = karate.jsonPath(response.scopes,"$.[?(@.scope_code=='"+data.createScope.scope_code+"')].scope_code")[0]
  #
  #  Given path 'scope' , scope_code
  # * header Authorization = 'Bearer '+access_token
  #When method delete
  #Then status 200
  #And match response.success == true

## Create Test Groups at root level
  
  Scenario: Create Automation Test Group for Orange
    Given path 'admin/groups' 
    And request data.createGroup
    When method post
    Then status 200

   
    ## To check duplicate group not allowed
  Scenario: To validate duplicate group name is not allowed
  Given path 'admin/groups'
  And request data.createGroup
  When method post
  Then status 400
  * match response.message == data.Group_Response_message_For_Duplicate_Name.message

## Edit/update group name
  Scenario: To verify update group name
    Given path 'admin/groups/1/groups'
    When method get
    Then status 200
  * def group_id = karate.jsonPath(response.groups,"$.[?(@.name=='"+data.createGroup.name+"')].ID")[0] 
    * print 'group_id is: ' + group_id
    Given path 'admin/groups', group_id
    * header Authorization = 'Bearer '+access_token
    * set data.updateGroup.groupId = group_id
    And request data.updateGroup
    When method put
    Then status 200

## Delete Updated Group
  Scenario: To verify Delete group
    Given path 'admin/groups/1/groups'
    When method get
    Then status 200
  * def updated_group_id = karate.jsonPath(response.groups,"$.[?(@.name=='"+data.updateGroup.group.name+"')].ID")[0] 
    * print 'group_id is: ' + updated_group_id
    Given path 'admin/groups', updated_group_id
    * header Authorization = 'Bearer '+access_token
    When method delete
    Then status 200


    ## Create Group at Child level
  Scenario: To verify create group at child level
    Given path 'admin/groups/1/groups'
    When method get
    Then status 200
  * def get_group_id = karate.jsonPath(response.groups,"$.[?(@.name=='TestingGroup')].ID")[0]
    * print 'get_group_id: ' + get_group_id
   
    Given path 'admin/groups'
    * header Authorization = 'Bearer '+access_token
    * set data.createGroup_Child_Level.parent_id = get_group_id
    And request data.createGroup_Child_Level
    When method post
    Then status 200

  ## Duplicate group name can not be create at chield level
  Scenario: To verify duplicate group can't be create at chield level
    Given path 'admin/groups/1/groups'
    When method get
    Then status 200
  * def get_group_id = karate.jsonPath(response.groups,"$.[?(@.name=='TestingGroup')].ID")[0]
    * print 'get_group_id: ' + get_group_id
    Given path 'admin/groups'
    * header Authorization = 'Bearer '+access_token
    * set data.createGroup_Child_Level.parent_id = get_group_id
    And request data.createGroup_Child_Level
    When method post
    Then status 400
    * match response.message == data.Group_Response_message_For_Duplicate_Name.message


    # Delete group from Chield level
  Scenario: To verify Delete child group
    Given path 'admin/groups/1/groups'
    When method get
    Then status 200
  * def get_group_id = karate.jsonPath(response.groups,"$.[?(@.name=='TestingGroup')].ID")[0]
    * print 'get_group_id: ' + get_group_id
    Given path 'admin/groups'
    * header Authorization = 'Bearer '+access_token
    * set data.createGroup_Child_Level.parent_id = get_group_id
    And request data.createGroup_Child_Level
    When method post
    Then status 400
    * match response.message == data.Group_Response_message_For_Duplicate_Name.message
    Given path 'admin/groups', get_group_id , 'groups'
    * header Authorization = 'Bearer '+access_token
    When method get
    Then status 200
  * def child_group_ID = karate.jsonPath(response.groups,"$.[?(@.name=='"+data.createGroup_Child_Level.name+"')].ID")[0]
  * print 'Chield group iD: ' + child_group_ID
   Given path 'admin/groups', child_group_ID
   * header Authorization = 'Bearer '+access_token
   When method Delete
   Then status 200
   * match response.success == true

  ## Admin can add and delete user under group
  Scenario: To verify admin can add and delete user under group
    # Creating Admin user at root group
    Given path 'user' 
     And request data.createUser_For_group
     When method post
     Then status 200 
    And match response == data.createUser_For_group
   # Getting Id of group under user being created 
    Given path 'admin/groups/1/groups'
    * header Authorization = 'Bearer '+access_token
    When method get
    Then status 200
  * def get_group_id = karate.jsonPath(response.groups,"$.[?(@.name=='AAK')].ID")[0]
  * print 'get_group_id: ' + get_group_id
    # Passing the group id for creating user under group
    Given path 'admin/groups' , get_group_id , 'users/add'
    * header Authorization = 'Bearer '+access_token
    * set data.Add_User_For_Group.group_id = get_group_id
   # * set data.Add_User_For_Group.user_id = data.createUser_For_group.user_id
    And request data.Add_User_For_Group
    When method PUT
    Then status 200
    # Delete the added user
    Given path 'admin/groups' , get_group_id , 'users/delete'
    * header Authorization = 'Bearer '+access_token
    And request data.Add_User_For_Group
    When method PUT
    Then status 200

# Delete the user
  Scenario: Delete the created user 
    Given path data.createUser_For_group.user_id
    When method Delete
    Then status 200
    * match response.success == true
 

