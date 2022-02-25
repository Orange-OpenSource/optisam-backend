@equipment @ignore

Feature: Equipment Service Test - user

  Background:
  # * def equipmentServiceUrl = "https://optisam-equipment-int.kermit-noprod-b.itn.intraorange"
    * url equipmentServiceUrl+'/api/v1/equipment'
    * def credentials = {username:'testuser@test.com', password: 'password'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'


  # @schema
  # Scenario: Schema validation for get Equipments
  #   Given path 'equipments/0x18e1'
  #   * params { page_num:1, page_size:10, sort_by:'server_hostname', sort_order:'desc', scopes:'#(scope)'}
  #   When method get
  #   Then status 200


  @get
  Scenario: Get Equipment Server
    Given path data.equipmentID.server_id , 'equipments'
    * params { page_num:1, page_size:10, sort_by:'server_code', sort_order:'ASC', scopes:'#(scope)'}
    When method get
    Then status 200
    And match response.equipments != 'W10='

  @get
  Scenario: Get Details of an Equipment - Server
    Given path  data.equipmentID.server_id,'equipments', data.server.server_code
    * params {scopes:'#(scope)'}
    When method get
    Then status 200
    And response.server_code == data.server.server_code


  @get
  Scenario: Get Parent of an equipment - server
    Given path  data.equipmentID.server_id, data.server.server_code_id, 'parents'
    * params {scopes:'#(scope)'}
    When method get 
    Then status 200
    And response.totalRecords==1

  # @get
  # Scenario: Get Children of an equipment - server
  #   Given path 'equipments', data.server.server_id, data.server.server_code_id, 'childs/0xd956'
  #   * params { page_num:1, page_size:10, sort_by:'server_code', sort_order:'desc', scopes:'#(scope)'}
  #   When method get
  #   Then status 200
  #   And response.totalRecords > 0


  # @get
  # Scenario: Get Equipments of a product
  #   Given path 'products',data.server.swid_tag, 'equipments',data.server.server_id
  #   * params { page_num:1, page_size:50, sort_by:'server_code', sort_order:'desc', scopes:'#(scope)'}
  #   When method get
  #   Then status 200
  #   And match response.equipments != 'W10='

  # @get
  # Scenario: Get Equipments of an Aggregation
  #   Given path 'products/aggregations',data.server.agg_name, 'equipments',data.server.server_id
  #   * params { page_num:1, page_size:50, sort_by:'server_code', sort_order:'desc', scopes:'#(scope)'}
  #   When method get
  #   Then status 200
  #   And match response.equipments != 'W10='


## Equipment Metadata

  # @schema
  # Scenario: Schema validation for get Metadata
  #   Given path 'equipments/metadata'
  #   * params { type:'ALL', scopes:'#(scope)'}
  #   When method get
  #   Then status 200
  #   * match response.metadata == '#[] data.metadata_schema'
    
  # @metadata
  # Scenario: Get Metadata by ID
  #   Given path 'equipments/metadata',data.getMetadata.ID
  #   * params { type:'ALL', scopes:'#(scope)'}
  #   When method get
  #   Then status 200
  #   * match response == data.getMetadata


## Equipment Type
  
  Scenario: Get Equipment Type List
    Given path 'types'
    * params {scopes:'#(scope)'}
    When method get
    Then status 200
    * match response.equipment_types contains data.equiptype_server


        @pagination
  Scenario Outline: To verify Pagination on Cluster Page
    Given path   data.equipmentID.cluster_id ,  'equipments'
    And params {page_num:1, page_size:'<page_size>', sort_by:'cluster_name', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.equipments != 'W10='

  Examples:
    | page_size | 
    | 50 |
    | 100 |
    | 200 |


  Scenario Outline: To verify Pagination on Equipment Page(cluster) with Invalid inputs
    Given path  data.equipmentID.cluster_id ,  'equipments'
    And params { page_num:'<page_num>', page_size:'<page_size>', sort_by:'name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 400
   Examples:
    | page_size | page_num |
    | 5 | 5 |
    | 10 | 0 |
    | "A" | 5 | 
    

      @pagination
  Scenario Outline: To verify Pagination on Vcenter Page
    Given path  data.equipmentID.vcenter_id , 'equipments'
    And params {page_num:1, page_size:'<page_size>', sort_by:'vcenter_name', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.equipments != 'W10='

  Examples:
    | page_size | 
    | 50 |
    | 100 |
    | 200 |

  Scenario Outline: To verify Pagination on Equipment Page(vcenter) with Invalid inputs
    Given path  data.equipmentID.vcenter_id ,  'equipments'
    And params { page_num:'<page_num>', page_size:'<page_size>', sort_by:'name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 400
   Examples:
    | page_size | page_num |
    | 5 | 5 |
    | 10 | 0 |
    | "A" | 5 |   
