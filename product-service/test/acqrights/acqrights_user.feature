@acqrights
Feature: Product Service - Acquired Rights Test : Normal User

  Background:
  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product/'
    * def credentials = {username:#(UserAccount_Username), password:#(UserAccount_password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

  @SmokeTest
  @get
  Scenario: List acquired rights
    Given path 'acqrights'
    And params { page_num:1, page_size:10, sort_by:'SWID_TAG', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.acquired_rights[*].swid_tag contains ["Software_AG_WebMethods_Software_AG_10.11","Red_Hat_Openshift_Standard_Redhat_4.2"]


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
    * set data.createAcqrights.product_name = data.UpdateAcq.product_Name2
    * set data.createAcqrights.avg_unit_price = data.UpdateAcq.avg_unit_price
    And request data.createAcqrights
    When method put
    Then status 403
    

    @delete
  Scenario: Normal user cannot Delete Acquired Rights
    Given path 'acqright/hpud_2'
    And params {scope:'#(scope)'}
    When method delete
    Then status 403

  @saasmetriccases
  Scenario: To create acquired rights with concurrent saas metric
    Given path 'acqright'
    And request data.createAcqrights_concurrent
    When method post
    Then status 403

  Scenario: To edit the acquired rights concurrent saas metric
    Given path 'acqright', data.createAcqrights.sku
    And request data.editAcqrights_concurrent
    When method put
    Then status 403

  Scenario: To create acquired rights with nominative saas metric
    Given path 'acqright'
    And request data.createAcqrights_nominative
    When method post
    Then status 403

  Scenario: To edit the acquired rights nominative saas metric
    Given path 'acqright', data.createAcqrights.sku
    And request data.editAcqrights_nominative
    When method put
    Then status 403

  Scenario: To create acquired rights with concurrent saas metric aggregation
    Given path 'aggregations'
    And params {page_num:1, page_size:50, sort_by:aggregation_name, sort_order:asc, scope:'#(scope)'}
    When method get
    Then status 200
    * def agg_id = karate.jsonPath(response.aggregations,"$.[?(@.aggregation_name=='"+ data.names.aggregation_name_conc + "')].ID")[0]
    * header Authorization = 'Bearer '+access_token
    Given path 'aggregatedrights'
    * set data.create_aggrights_conc.aggregationID = agg_id
    And request data.create_aggrights_conc
    When method post
    Then status 403

  Scenario: To create acquired rights with nominative saas metric aggregation
    Given path 'aggregations'
    And params {page_num:1, page_size:50, sort_by:aggregation_name, sort_order:asc, scope:'#(scope)'}
    When method get
    Then status 200
    * def agg_id = karate.jsonPath(response.aggregations,"$.[?(@.aggregation_name=='"+ data.names.aggregation_name_nom + "')].ID")[0]
    * header Authorization = 'Bearer '+access_token
    Given path 'aggregatedrights'
    * set data.create_aggrights_nom.aggregationID = agg_id
    And request data.create_aggrights_nom
    When method post
    Then status 403

  Scenario: To test sharing licenses in Product
      Given path 'licenses'
      And request data.share_license_product
      When method put
      Then status 403

  Scenario: To test sharing licenses in Aggregations
      Given path 'aggrights/licenses'
      And request data.share_license_aggregations
      When method put
      Then status 403
  
  Scenario:To create acquired rights with maintainance for concurrent metric(Product)
    Given path 'acqright'
    And request data.maintainence_product_conc
    When method post
    Then status 403

  Scenario: To create acquired rights with maintainance for nominative metric(Product)
    Given path 'acqright'
    And request data.maintainence_product_nom
    When method post
    Then status 403

  Scenario: To create acquired rights with maintainance for concurrent metric(Aggregation)
    Given path 'aggregations'
    And params {page_num:1, page_size:50, sort_by:aggregation_name, sort_order:asc, scope:'#(scope)'}
    When method get
    Then status 200
    * def agg_id = karate.jsonPath(response.aggregations,"$.[?(@.aggregation_name=='"+data.names.aggregation_name_conc+"')].ID")[0]
    * header Authorization = 'Bearer '+access_token
    Given path 'aggregatedrights'
    * set data.maintainence_agg_conc.aggregationID = agg_id
    And request data.maintainence_agg_conc
    When method post
    Then status 403

  Scenario: To create acquired rights with maintainance for nominative metric(Aggregation)
    Given path 'aggregations'
    And params {page_num:1, page_size:50, sort_by:aggregation_name, sort_order:asc, scope:'#(scope)'}
    When method get
    Then status 200
    * def agg_id = karate.jsonPath(response.aggregations,"$.[?(@.aggregation_name=='"+data.names.aggregation_name_nom+"')].ID")[0]
    * header Authorization = 'Bearer '+access_token
    Given path 'aggregatedrights'
    * set data.maintainence_agg_nom.aggregationID = agg_id
    And request data.maintainence_agg_nom
    When method post
    Then status 403
