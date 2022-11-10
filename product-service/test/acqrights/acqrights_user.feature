@acqrights
Feature: Product Service - Acquired Rights Test : Normal User

  Background:
  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product'
    * def credentials = {username:#(UserAccount_Username), password:#(UserAccount_password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

    
  @get
  Scenario: List acquired rights
    Given path 'acqrights'
    And params { page_num:1, page_size:10, sort_by:'SWID_TAG', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match $.acquired_rights == '#[_ <= 20]'


  @search
  Scenario: Searching_Filter Acquired Rights by Swid tag and Product Name
    Given path 'acqrights'
    And params { page_num:1, page_size:10, sort_by:'PRODUCT_NAME', sort_order:'desc', scopes:'#(scope)'}
    And params {search_params.swidTag.filteringkey: '#(data.getAcqrights.swid_tag)'}
    And params {search_params.productName.filteringkey: '#(data.getAcqrights.product_name)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match each response.acquired_rights[*].swid_tag == data.getAcqrights.swid_tag
    And match each response.acquired_rights[*].product_name == data.getAcqrights.product_name
   


  @sort
  Scenario: Sorting_sort Acquired Rights data by Swidtag in descending order
    Given path 'acqrights'
    And params { page_num:1, page_size:10, sort_by:'SWID_TAG', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    * def actual = $response.acquired_rights[*].swid_tag
    * def sorted = sort(actual,'desc')
    * match sorted contains actual


  @sort
  Scenario: Sorting_sort Acquired Rights data by Product Name in ascending order
    Given path 'acqrights'
    And params { page_num:1, page_size:10, sort_by:'PRODUCT_NAME', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    * def actual = $response.acquired_rights[*].product_name
    * def sorted = sort(actual,'asc')
    * match sorted == actual


  Scenario Outline: To verify Pagination on AcquiredRights Page with Invalid inputs
    Given path  'acqrights'
    And params { page_num:'<page_num>', page_size:'<page_size>', sort_by:'name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 400
   Examples:
    | page_size | page_num |
    | 5 | 5 |
    | 10 | 0 |
    | "A" | 5 |


# Acqrights 

  # @negative
    @create
    Scenario: Normal user can not Create Acquired Rights
    Given path 'acqright'
    And request data.createAcqrights
    When method post
    Then status 403
  

    @update
    Scenario:Normal user cannot Update Acquired Rights
    Given path 'acqright/hpud_2',
    * set data.createAcqrights.product_name = "APIProductUpdated"
    * set data.createAcqrights.avg_unit_price = "8"
    And request data.createAcqrights
    When method put
    Then status 403
    

    @delete
  Scenario: Normal user cannot Delete Acquired Rights
    Given path 'acqright/hpud_2'
    And params {scope:'#(scope)'}
    When method delete
    Then status 403