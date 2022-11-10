@ofr @ignore

Feature: Pre-Requisite Setup for OFR - Create Equipment type, Metric

  Background:
    * url dpsServiceUrl+'/api/v1'
    #* def credentials = {username:'admin@test.com', password: 'Welcome@123'}
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def scope = "OFR"

 Scenario: Upload Metadata files
    Given url importServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'import/metadata'
    * def file2_tmp = karate.readAsString('metadata/metadata_vcenter.csv')
    * def file3_tmp = karate.readAsString('metadata/metadata_cluster.csv')
    * def file4_tmp = karate.readAsString('metadata/metadata_server.csv')
    * def file5_tmp = karate.readAsString('metadata/metadata_partition.csv')
    * multipart file file = { value: '#(file2_tmp)', filename: 'metadata_vcenter.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file3_tmp)', filename: 'metadata_cluster.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file4_tmp)', filename: 'metadata_server.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file5_tmp)', filename: 'metadata_partition.csv', contentType: "text/csv" }
    * multipart field scope = scope
    When method post
    Then status 200 

 Scenario: Create Equipment Types
    * url equipmentServiceUrl+'/api/v1'
    Given path 'equipments/metadata'
    * params { type:'ALL', scopes:'#(scope)'}
    When method get
    Then status 200
    * def metadata = response
    * def attr = read('type_attributes.json')
    # Create Vcenter
    * def meta_vcenter_id = karate.jsonPath(metadata, "metadata[?(@.name=='metadata_vcenter.csv')].ID")
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments/types'
    * request { type:'vcenter', metadata_id:'#(meta_vcenter_id[0])', attributes:'#(attr.vcenter)', scopes:['#(scope)']}
    When method post
    Then status 200
    * def vcenter_id = response.ID
    # Create Cluster
    * def meta_cluster_id = karate.jsonPath(metadata, "metadata[?(@.name=='metadata_cluster.csv')].ID")
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments/types'
    * request { type:'cluster', metadata_id:'#(meta_cluster_id[0])', "parent_id":'#(vcenter_id)', attributes:'#(attr.cluster)', scopes:['#(scope)']}
    When method post
    Then status 200
    * def cluster_id = response.ID
    # Create Server
    * def meta_server_id = karate.jsonPath(metadata, "metadata[?(@.name=='metadata_server.csv')].ID")
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments/types'
    * request { type:'server', metadata_id:'#(meta_server_id[0])', "parent_id":'#(cluster_id)', attributes:'#(attr.server)', scopes:['#(scope)']}
    When method post
    Then status 200
    * def server_id = response.ID
    # Create partition
    * def meta_sp_id = karate.jsonPath(metadata, "metadata[?(@.name=='metadata_partition.csv')].ID")
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments/types'
    * request { type:'partition', metadata_id:'#(meta_sp_id[0])', "parent_id":'#(server_id)', attributes:'#(attr.partition)', scopes:['#(scope)']}
    When method post
    Then status 200


 Scenario: Create Metric - processor
   # fetch equipment types
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
   * def serv_core = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='core_per_processor')].ID")[0]
   * def serv_processor = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='used_processor_sockets')].ID")[0]
   * def server_oracle = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='core_factor')].ID")[0]
   # metric oracle processor
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/ops'
   * request {"Name":"processor","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_oracle)',"numCPU_attr_id":'#(serv_processor)',"base_eq_type_id":'#(server_eq_type.ID)',"start_eq_type_id":'#(partition_eq_type.ID)',"aggerateLevel_eq_type_id":'#(cluster_eq_type.ID)',"end_eq_type_id":'#(vcenter_eq_type.ID)',"number_of_users":0,"scopes":['#(scope)']}
   When method post
   Then status 200
