@account
Feature: Account Service Test

  Background:
  # * def accountServiceUrl = "https://optisam-account-int.apps.fr01.paas.tech.orange"
    * url accountServiceUrl+'/api/v1'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
     * callonce read('common.feature') credentials
     * def access_token = response.access_token
     * header Authorization = 'Bearer '+access_token
     * def data = read('data.json')
    # * def err = {"error": "string","code": 0,"message": "string","details": [{"type_url": "string","value": "string"}]}


  # @Update
  # Scenario:Change password
  #    Given path 'account/changepassword'
  #   And request data.change_password
  #   When method put
  #   Then status 200


   Scenario: Create User account with Admin role and delete it
     Given path 'account/user' 
     And request data.createAdminAccount
     When method post
     Then status 200 
    And match response == data.createAdminAccount
    
     * path 'account' ,data.createAdminAccount.user_id
     * header Authorization = 'Bearer '+access_token
     * method delete
    * status 200
    * match response.success == true

  Scenario: Create User account with User role and delete it
    Given path 'account/user' 
    And request data.createUserAccount
    When method post
    Then status 200 
   And match response == data.createUserAccount
   * print response
    * path 'account' ,data.createUserAccount.user_id
    * header Authorization = 'Bearer '+access_token
    * method delete
   * status 200
   * match response.success == true
