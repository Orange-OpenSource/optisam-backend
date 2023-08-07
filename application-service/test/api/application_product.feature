Feature: Application Product  Test

Background:
  * url productServiceUrl+'/api/v1'
  #* def credentials = {username:'admin@test.com', password: 'Welcome@123'}
  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
  * callonce read('common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'

Scenario:To verfiy the  action after clicking on product count  
    Given path 'products'
    And params { page_num:1, page_size:50, sort_by:'name', sort_order:'asc', scopes:'#(scope)'}
And params { search_params.application_id.filter_type:'1', search_params.application_id.filteringkey:'#(data.product.application_id)'}
    When method get
    Then status 200
    And print response 
 And match response.products[*].name contains[ "Adobe Media Server"]

 

  Scenario: To get the Details of product
    Given path 'products'
    And params {page_num:1,page_size:50,sort_by:'name',sort_order:'asc',scopes:'API',search_params.application_id.filter_type:'1',search_params.application_id.filteringkey:'#(data.product.application_id)'}
    When method get 
    Then status 200 



  Scenario Outline: To verify Pagination on Product  page
    Given path 'products' 
    And params { page_num:1, page_size:'<page_size>', sort_by:'name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    #And match $.applications == '#[_ <= <page_size>]'
  Examples:
    | page_size |
    | 50 |
    | 100 |
    | 200 |
    # Pagination is not working on product 


  Scenario: To check the Details of product 
  And  path 'product', data.product.product_name1
  And params {scope:'#(scope)'}
  When method get 
  Then status 200
  # Not Giving response at backend 
  
  
  Scenario: To Check the maintenance detail of the product
    Given path 'product','acqrights'
    And params {page_num:1,page_size:50, sort_by:'PRODUCT_NAME',sort_order:'asc',scopes:'#(scope)',search_params.swidTag.filteringkey:'#(data.product.product_name1)', search_params.SKU.filteringkey:'null'}
  When method get 
  Then status 200 
  And match response.totalRecords == 0



