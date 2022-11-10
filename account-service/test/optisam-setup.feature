@setup @ignore

Feature: Setup Pre-requisite for Optisam Test - API

  Background:
    #* def credentials = {username:'admin@test.com', password: 'Welcome@123'}
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * def scope = "API"


# Create Test Scopes
  Scenario: Create API scope
    * url accountServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'scopes'
    And request {"scopeCode": "API","scopeName": "API"}
    When method post
    Then eval if (responseStatus  == 409) karate.abort()
    * status 200


## Create Test Groups
  @ignore
  Scenario: Create Automation Test Group for Orange
    * url accountServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'admin/groups' 
    And request {"name": "API","scopes": ["API"],"parentId": "1"}
    When method post
    Then eval if (responseStatus == 409) karate.abort()
    Then status 200

## Create Users
  @ignore
  Scenario: Create Normal User
    * url accountServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'admin/groups'
    When method get
    Then status 200
    * def group_id = karate.jsonPath(response.groups,"$.[?(@.name=='API')].ID")[0]
    * url accountServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'accounts' 
    And request {"userId": "testuser@test.com","firstName": "Test","lastName": "User","locale": "en","groups": [#(group_id)],"role": "USER"}
    When method post
    Then eval if (responseStatus  == 409) karate.abort()
    Then status 200

  @ignore
  Scenario: Create Admin User
    * url accountServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'admin/groups'
    When method get
    Then status 200
    * def group_id = karate.jsonPath(response.groups,"$.[?(@.name=='AUT')].ID")[0]
    * url accountServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'accounts' 
    And request {"userId": "testadmin@test.com","firstName": "Test","lastName": "Admin","locale": "en","groups": [#(group_id)],"role": "ADMIN"}
    When method post
    Then eval if (responseStatus  == 409) karate.abort()
    Then status 200