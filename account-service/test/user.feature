

@account
Feature: Account Service Test

  Background:
  # * def accountServiceUrl = "https://optisam-account-int.apps.fr01.paas.tech.orange"
    * url accountServiceUrl+'/api/v1'
    * def credentials = {username:#(UserAccount_Username), password:#(UserAccount_password)}
     * callonce read('common.feature') credentials
     * def access_token = response.access_token
     * header Authorization = 'Bearer '+access_token
     * def data = read('data.json')
     * def scope = 'API'

    Scenario: Create User account with user role
        Given path 'account/user' 
        And request data.createAdminAccount
        When method post
        Then status 403
        And match response.message == data.User_Account_Create_usermessage.message
       
    Scenario: To verify user can't create Scope from user account
        Given path 'account/scopes'
        And request data.createScope
        When method post
        Then status 403
        * match response.message == data.User_Account_Create_usermessage.message_scope

    @create
    Scenario: To verify user can't create groups
      Given path 'account/admin/groups' 
      * def createGroup = {"scopes": ['#(scope)'],"parent_id": "1"}
      * set createGroup.name = "apitest_grp_" + now()
      And request createGroup
      When method post
      Then status 403     
      * match response.message == data.User_Account_Create_usermessage.message_group


    @SmokeTest
    Scenario: To get all users
       Given path 'account/users'
       And params {user_filter.all_users:'true'}
       When method get 
       Then status 200 
#------------------------User Acess For Software spend-----------------------------------------#
      Scenario: To Get  the software spent  from User Account
        Given path '/account/scopes/expenses', scope
        When method get
        Then status 200
        
      Scenario Outline: To verify that User can not modify the software Spent 
        Given path '/account/scopes/expenses'
        * def expense =  { "scope_code":<scope>,"expenses":<Expense>}
        And request expense
        When method post 
        Then status 403
        * match response.message == data.User_Account_Create_usermessage.message_expense
        Examples:
        | Expense| scope|
        | 52345.35 | API|
        | 34343434343434343434 |AZU|
        | 378701320.453|AAK|

      Scenario Outline: To verify that User can not modify the software Spent 
        Given path '/account/scopes/expenses'
        * def expense =  { "scope_code":<scope>,"expenses":<Expense>}
        And request expense
        When method post 
        Then status 400
        * match response.message == data.User_Account_Create_usermessage.message_expense
        Examples:
        | -345345345345 | ACQ|
       


        
        

