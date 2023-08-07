@acqrights

Feature: Acquired Rights Service Test

  Background:
  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

  @SmokeTest
  @schema
  Scenario: Schema validation for get Acquired Rights
    Given path 'acqrights'
    * params { page_num:1, page_size:50, sort_by:'SKU', sort_order:'asc', scopes:'#(scope)'}
    * def schema = data.schema_acq
    When method get
    Then status 200
    * response.totalRecords == '#number? _ > 0'
    * match response.acquired_rights[*].product_name contains ["Adobe Media Server"]
   

  @get
  Scenario: Pagination_get 20 records of acquired rights
    Given path 'acqrights'
    And params { page_num:1, page_size:20, sort_by:'SKU', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match $.acquired_rights == '#[_ <= 20]'

  @get
  Scenario: Pagination_get 30 records of acquired rights
    Given path 'acqrights'
    And params { page_num:1, page_size:30, sort_by:'SKU', sort_order:'desc', scopes:'#(scope)'}
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
    And params { page_num:1, page_size:50, sort_by:'SKU', sort_order:'asc', scopes:'#(scope)'}
    And params {search_params.swidTag.filteringkey: '#(data.getAcqrights.swid_tag)'}
    And params {search_params.productName.filteringkey: '#(data.getAcqrights.product_name)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match   response.acquired_rights[*].swid_tag contains data.getAcqrights.swid_tag
    And match each response.acquired_rights[*].product_name == data.getAcqrights.product_name
   # And match  response.acquired_rights contains data.getAcqrights

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
    And match response.acquired_rights[*].editor contains data.getAcqrights.editor
    #And match  response.acquired_rights contains data.getAcqrights


  @search
  Scenario: Search Acquired Rights by metric name
    Given path 'acqrights'
    And params { page_num:1, page_size:50, sort_by:'METRIC', sort_order:'asc', scopes:'#(scope)'}
    And params {search_params.metric.filteringkey: '#(data.getAcqrights.metric)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match each response.acquired_rights[*].metric contains data.getAcqrights.metric
   # And match  response.acquired_rights contains data.getAcqrights

  @sort
  Scenario: Sorting_sort Acquired Rights data by Swidtag in descending order
    Given path 'acqrights'
    And params { page_num:1, page_size:50, sort_by:'SWID_TAG', sort_order:'desc', scopes:'#(scope)'}
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

  
## UPSERT APIs for Acqrights

# Working
  @create
  Scenario: Create Acquired Rights without maintenance details
    Given path 'acqright'
    And request data.createAcqrights
    When method post
    Then status 200
    And match response.success == true
    # after running this program to rerun kindly provide a diffrent sku name or it will give error
# Working
    Scenario: To verify Acquired Rights is not created with same sku
    Given path 'acqright'
     And request data.createAcqrights
     When method post
     Then status 400

    @SmokeTest
   Scenario: To verify Acquired Rights is not created without sku
     Given path 'acqright'
    * remove data.createAcqrights.sku
      And request data.createAcqrights
      When method post
      Then status 400
     

  @update
  Scenario: Update Acquired Rights
    Given path 'acqright',data.createAcqrights.sku
    * set data.createAcqrights.product_name = data.UpdateAcq.product_Name2
    * set data.createAcqrights.avg_unit_price = data.UpdateAcq.avg_unit_price
    And request data.createAcqrights
    When method put
    Then status 200
    And match response.success == true

  @delete
  Scenario: Delete Acquired Rights
    Given path 'acqright',data.createAcqrights.sku
    And params {scope:'#(scope)'}
    When method delete
    Then status 200
    And match response.success == true

   
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
    * set data.createAcqrightswithmaintenance.end_of_maintenance = data.UpdateAcq.end_of_maintenance
    * set data.createAcqrightswithmaintenance.start_of_maintenance = data.UpdateAcq.start_of_maintenance
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
#----------------------Acquired Right Without version creation--------------------#
  Scenario: Create Acquired Rights without selecting version
    Given path 'acqright'
    And request data.createAcqrightsWithoutVersion
    When method post
    Then status 200
     # after running this program to rerun kindly provide a diffrent sku name or it will give error
# Working

Scenario: Update Acquired Rights versions 
  Given path 'acqright',data.createAcqrightsWithoutVersion.sku
  * set data.createAcqrightsWithoutVersion.version = data.UpdateAcq.version
  And request data.createAcqrightsWithoutVersion
  When method put
  Then status 200 
  And match response.success == true 


  Scenario: delete  Acquired Rights without selecting version
    Given path 'acqright'
    Given path data.createAcqrightsWithoutVersion.sku
    And params {scope:'#(scope)'}
    When method delete
    Then status 200 

  


  



    








    
