@acqrights

Feature: Acquired Rights Service Test

  Background:
  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'

  @schema
  Scenario: Schema validation for get Acquired Rights
    Given path 'acqrights'
    * params { page_num:1, page_size:10, sort_by:'SWID_TAG', sort_order:'asc', scopes:'#(scope)'}
    * def schema = data.schema_acq
    When method get
    Then status 200
    * response.totalRecords == '#number? _ > 0'
    * match response.acquired_rights == '#[_ > 0] schema'
    * match response.acquired_rights == '#[_ <= 10] schema'

  @get
  Scenario: Pagination_get 20 records of acquired rights
    Given path 'acqrights'
    And params { page_num:1, page_size:20, sort_by:'SWID_TAG', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match $.acquired_rights == '#[_ <= 20]'

  @get
  Scenario: Pagination_get 30 records of acquired rights
    Given path 'acqrights'
    And params { page_num:1, page_size:30, sort_by:'SWID_TAG', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match $.acquired_rights == '#[_ <= 30]'


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

      @search
  Scenario: Search AcquiredRights by multiple  column
    Given path 'acqrights'
    And params { page_num:1, page_size:10, sort_by:'PRODUCT_NAME', sort_order:'desc', scopes:'#(scope)'}
    And params {search_params.swidTag.filteringkey: '#(data.getAcqrights.swid_tag)'}
    And params {search_params.productName.filteringkey: '#(data.getAcqrights.product_name)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match each response.acquired_rights[*].swid_tag == data.getAcqrights.swid_tag
    And match each response.acquired_rights[*].product_name == data.getAcqrights.product_name
    And match  response.acquired_rights contains data.getAcqrights


  

  @search
  Scenario: Searching_filter Acquired Rights by SKU and editor
    Given path 'acqrights'
    And params { page_num:1, page_size:10, sort_by:'SWID_TAG', sort_order:'desc', scopes:'#(scope)'}
    And params {search_params.SKU.filteringkey: '#(data.getAcqrights.SKU)'}
    And params {search_params.editor.filteringkey: '#(data.getAcqrights.editor)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match each response.acquired_rights[*].SKU == data.getAcqrights.SKU
    And match each response.acquired_rights[*].editor == data.getAcqrights.editor
    And match  response.acquired_rights contains data.getAcqrights


  @search
  Scenario: Search Acquired Rights by metric name
    Given path 'acqrights'
    And params { page_num:1, page_size:10, sort_by:'METRIC', sort_order:'asc', scopes:'#(scope)'}
    And params {search_params.metric.filteringkey: '#(data.getAcqrights.metric)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match each response.acquired_rights[*].metric contains data.getAcqrights.metric
    And match  response.acquired_rights contains data.getAcqrights

  @sort
  Scenario: Sorting_sort Acquired Rights data by Swidtag in descending order
    Given path 'acqrights'
    And params { page_num:1, page_size:10, sort_by:'SWID_TAG', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    * def actual = $response.acquired_rights[*].swid_tag
    * def sorted = sort(actual,'desc')
    * match sorted == actual


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


## UPSERT APIs for Acqrights

  @create
  Scenario: Create Acquired Rights without maintenance details
    Given path 'acqright'
    And request data.createAcqrights
    When method post
    Then status 200
    And match response.success == true

    Scenario: To verify Acquired Rights is not created with same sku
    Given path 'acqright'
     And request data.createAcqrights
     When method post
     Then status 400


   Scenario: To verify Acquired Rights is not created without sku
     Given path 'acqright'
    * remove data.createAcqrights.sku
      And request data.createAcqrights
      When method post
      Then status 400
     

  @update
  Scenario: Update Acquired Rights
    Given path 'acqright',data.createAcqrights.sku
    * set data.createAcqrights.product_name = "APIProductUpdated"
    * set data.createAcqrights.avg_unit_price = "8"
    And request data.createAcqrights
    When method put
    Then status 200
    And match response.success == true

  # @update
  # Scenario: To verify scope is not Updated
  #   Given path 'acqrights'
  #   And request ({ "application_id": data.createApp.application_id, "name": 'dummyUpdated', "version": "0.1.4", "owner": "OrangeUpdated", "scope": "Dummy"})
  #   When method post
  #   Then status 200
  #   And match response.success == true

  @delete
  Scenario: Delete Acquired Rights
    Given path 'acqright',data.createAcqrights.sku
    And params {scope:'#(scope)'}
    When method delete
    Then status 200

    @create
  Scenario: Create Acquired Rights with maintenance details
    Given path 'acqright'
    And request data.createAcqrightswithmaintenance
    When method post
    Then status 200
    And match response.success == true


    @update
  Scenario: Miantenance End date cannot be less than start date
    Given path 'acqright',data.createAcqrightswithmaintenance.sku
    * set data.createAcqrightswithmaintenance.end_of_maintenance = "2021-07-02T18:30:00.000Z"
    * set data.createAcqrightswithmaintenance.start_of_maintenance = "2029-07-31T18:30:00.000Z"
    And request data.createAcqrightswithmaintenance
    When method put
    Then status 400
    

    @delete
  Scenario: Delete Acquired Rights with maintenance details
    Given path 'acqright',data.createAcqrightswithmaintenance.sku
    And params {scope:'#(scope)'}
    When method delete
    Then status 200

  Scenario: To verify scope is mandetory to create Acquired rights
    Given path 'acqright'
    * remove data.createAcqrights.scope
    And request data.createAcqrights
    When method post
    Then status 400

