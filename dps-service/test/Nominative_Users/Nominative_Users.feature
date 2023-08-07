@NomenativeUser
Feature: Nominative_user test for Admin user

  Background:
  * url productServiceUrl+'/api/v1/product'
  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
  * callonce read('../common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'


@SmokeTest
  Scenario: Schema validation for Nominative_user
    Given path 'nominative/users'
    And params {page_num:1, page_size:50, sort_by:'activation_date', sort_order:'asc', scopes:'#(scope)'}
    * def schema = data.Schema_nominative_user
    When method get
    Then status 200
   # * match response.nominative_user == '#[_ > 0] schema'

  @SmokeTest
    Scenario: Get all Nominative_user
        Given path 'nominative/users'
        And params {page_num:1, page_size:50, sort_by:'activation_date', sort_order:'asc', scopes:'#(scope)'}
        When method get
        Then status 200
        And response.totalRecords > 0

      Scenario: To verify  Nominative_user without filter params page_num
        Given path 'nominative/users'
        And params { page_size:50, sort_by:'activation_date', sort_order:'asc'}
        And params {scopes:'#(scope)'}
        When method get
        Then status 400
       
      Scenario: To verify  Nominative_user without filter params page_size
        Given path 'nominative/users'
        And params { page_num:1, sort_by:'activation_date', sort_order:'asc'}
        And params {scopes:'#(scope)'}
        When method get
        Then status 400

      Scenario: To verify Nominative_user without filter sorted_by
        Given path 'nominative/users'
        And params { page_num:1, page_size:50, sort_order:'asc'}
        And params {scopes:'#(scope)'}
        When method get
        Then status 400

      Scenario: To verify Nominative_user without scope
        Given path 'nominative/users'
        And params {page_num:1, page_size:50, sort_by:'activation_date', sort_order:'asc'}
        When method get
        Then status 400

      Scenario: To verify Nominative_user with invalid page num
        Given path 'nominative/users'
        And params {page_num:abc, page_size:50, sort_by:'activation_date', sort_order:'asc'}
        And params {scopes:'#(scope)'}
        When method get
        Then status 400

        