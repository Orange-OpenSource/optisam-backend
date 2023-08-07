@NomenativeUser
Feature: Concurrent_user test for Admin user

  Background:
  * url productServiceUrl+'/api/v1/product'
  * def credentials = {username:#(UserAccount_Username), password:#(UserAccount_password)}
  * callonce read('../common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'

  Scenario: Get all Concurrent User
    Given path 'concurrent'
    And params { page_num: 1,page_size: 50, sort_by: purchase_date, sort_order: asc, scopes: '#(scope)', is_aggregation: false,search_params.product_editor.filteringkey: 'hajd' , search_params.product_editor.filter_type: true,search_params.product_name.filteringkey: 'dddd', search_params.product_name.filter_type: true,search_params.product_version.filteringkey: 'dff',search_params.product_version.filter_type: true }
    When method get
    Then status 200

  Scenario: Create Concurrent User
    Given path 'concurrent'
    * set data.Create_Concurrent_User.scope = scope
    And request data.Create_Concurrent_User
    When method post
    Then status 403
