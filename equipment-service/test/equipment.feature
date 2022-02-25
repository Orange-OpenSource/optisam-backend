@equipment @ignore
Feature: Equipment Service Test

  Background:
  # * def equipmentServiceUrl = "https://optisam-equipment-int.kermit-noprod-b.itn.intraorange"
    * url equipmentServiceUrl+'/api/v1'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'


  @get
  Scenario: Get Equipment Server
     * url equipmentServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'equipments/types'
   And params {scopes:'#(scope)'}
   When method get
   Then status 200
   * def server_eq_type = karate.jsonPath(response, "$.equipment_types[?(@.type=='server')]")[0]
   * def partition_eq_type = karate.jsonPath(response, "$.equipment_types[?(@.type=='partition')]")[0]
   * def cluster_eq_type = karate.jsonPath(response, "$.equipment_types[?(@.type=='cluster')]")[0]
   * def vcenter_eq_type = karate.jsonPath(response, "$.equipment_types[?(@.type=='vcenter')]")[0]
   * def serv_core = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='server_coresNumber')].ID")[0]
   * def serv_processor = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='server_processorsNumber')].ID")[0]
   * def server_oracle = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='corefactor_oracle')].ID")[0]
   * def server_pvu = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='pvu')].ID")[0]
   * def server_sag = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='sag')].ID")[0]
    Given path 'equipments', server_eq_type.ID , 'equipments'
    * params { page_num:1, page_size:10, sort_by:'server_code', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And match response.equipments != 'W10='

  @get  
  Scenario: Get Details of an Equipment - Server
    Given path 'equipments', data.server.server_id,'equipments', data.server.server_code
    * params {scopes:'#(scope)'}
    When method get
    Then status 200
    And response.server_code == data.server.server_code


  @get
  Scenario: Get Parent of an equipment - server
    Given path 'equipments', data.server.server_id, data.server.server_code_id, 'parents'
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


  @get
  Scenario: Get Equipments of a product
    Given path 'products',data.server.swid_tag, 'equipments',data.server.server_id
    * params { page_num:1, page_size:10, sort_by:'server_code', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And match response.equipments != 'W10='

  @get
  Scenario: Get Equipments of an Aggregation
    Given path 'products/aggregations',data.server.agg_name, 'equipments',data.server.server_id
    * params { page_num:1, page_size:10, sort_by:'server_code', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    And match response.equipments != 'W10='


 @sort
  Scenario Outline: Sorting_sort Equipment data for Datacenter
    Given path 'equipments' , data.sorting.datacenter_id , 'equipments'
    And params { page_num:1, page_size:10, sort_by:'<sortBy>', sort_order:'<sortOrder>', scopes:'#(scope)', search_params:''}
    When method get
    Then status 200
    And match response.equipments != 'W10='
   
  Examples:
      | sortBy | sortOrder |  
      | datacenter_name | desc |
      | datacenter_name | asc |


 @sort
  Scenario Outline: Sorting_sort Equipment data for Vcenter
    Given path 'equipments' , data.sorting.vcenter_id , 'equipments'
    And params { page_num:1, page_size:10, sort_by:'<sortBy>', sort_order:'<sortOrder>', scopes:'#(scope)', search_params:''}
    When method get
    Then status 200
    And match response.equipments != 'W10='
   
  Examples:
      | sortBy | sortOrder |  
      | vcenter_name | desc | 
      | vcenter_name | asc |   

@sort
  Scenario Outline: Sorting_sort Equipment data for Cluster
    Given path 'equipments' , data.sorting.cluster_id , 'equipments'
    And params { page_num:1, page_size:10, sort_by:'<sortBy>', sort_order:'<sortOrder>', scopes:'#(scope)', search_params:''}
    When method get
    Then status 200
    And match response.equipments != 'W10='
   
  Examples:
      | sortBy | sortOrder |  
      | cluster_name | desc | 
      | cluster_name | asc |   

  @sort
  Scenario Outline: Sorting_sort Equipment data for Server
    Given path 'equipments' , data.sorting.server_id , 'equipments'
    And params { page_num:1, page_size:10, sort_by:'<sortBy>', sort_order:'<sortOrder>', scopes:'#(scope)', search_params:''}
    When method get
    Then status 200
    And match response.equipments != 'W10='
   
  Examples:
      | sortBy | sortOrder |  
      | server_code | desc |
      | server_cpu | desc | 
      | server_hostname | asc|

 @pagination
  Scenario Outline: To verify Pagination on Server Page
    Given path 'equipments' , data.sorting.server_id ,'equipments'
    And params {page_num:1, page_size:'<page_size>', sort_by:'server_code', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.equipments != 'W10='

  Examples:
    | page_size | 
    | 20 |
    | 30 |
    | 50 |



    Scenario Outline: To verify Pagination on Equipment Page(server) with Invalid inputs
    Given path 'equipments' , data.sorting.server_id , 'equipments'
    And params { page_num:'<page_num>', page_size:'<page_size>', sort_by:'name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 400
   Examples:
    | page_size | page_num |
    | 5 | 5 |
    | 10 | 0 |
    | "A" | 5 | 

    @pagination
  Scenario Outline: To verify Pagination on Dtacenter Page
    Given path 'equipments' , data.sorting.datacenter_id ,'equipments'
    And params {page_num:1, page_size:'<page_size>', sort_by:'datacenter_name', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.equipments != 'W10='

  Examples:
    | page_size | 
    | 20 |
    | 30 |
    | 50 |


      @pagination
  Scenario Outline: To verify Pagination on Cluster Page
    Given path 'equipments' , data.sorting.cluster_id ,'equipments'
    And params {page_num:1, page_size:'<page_size>', sort_by:'cluster_name', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.equipments != 'W10='

  Examples:
    | page_size | 
    | 20 |
    | 30 |
    | 50 |



  Scenario Outline: To verify Pagination on Equipment Page(cluster) with Invalid inputs
    Given path 'equipments' , data.sorting.cluster_id , 'equipments'
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
    Given path 'equipments' , data.sorting.vcenter_id , 'equipments'
    And params {page_num:1, page_size:'<page_size>', sort_by:'vcenter_name', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match response.equipments != 'W10='

  Examples:
    | page_size | 
    | 20 |
    | 30 |
    | 50 |
        

  Scenario Outline: To verify Pagination on Equipment Page(vcenter) with Invalid inputs
    Given path 'equipments' , data.sorting.vcenter_id , 'equipments'
    And params { page_num:'<page_num>', page_size:'<page_size>', sort_by:'name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 400
   Examples:
    | page_size | page_num |
    | 5 | 5 |
    | 10 | 0 |
    | "A" | 5 |       
             
                
      