@gen @ignore
# Run only to setup the pre-requisite for GEN
Feature: Pre-Requisite Setup for GEN - Create Equipment type, Metric

  Background:
    * url dpsServiceUrl+'/api/v1'
    #* def credentials = {username:'admin@test.com', password: 'Welcome@123'}
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def scope = "GEN"


 Scenario: Upload Metadata files
    Given url importServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'import/metadata'
    * def file1_tmp = karate.readAsString('metadata/metadata_vcenter.csv')
    * def file2_tmp = karate.readAsString('metadata/metadata_cluster.csv')
    * def file3_tmp = karate.readAsString('metadata/metadata_server.csv')
    * def file4_tmp = karate.readAsString('metadata/metadata_softpartition.csv')
    * def file5_tmp = karate.readAsString('metadata/metadata_hardpartition.csv')
    * multipart file file = { value: '#(file1_tmp)', filename: 'metadata_vcenter.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file2_tmp)', filename: 'metadata_cluster.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file3_tmp)', filename: 'metadata_server.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file4_tmp)', filename: 'metadata_softpartition.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file5_tmp)', filename: 'metadata_hardpartition.csv', contentType: "text/csv" }
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
    # Create SoftPartition
    * def meta_sp_id = karate.jsonPath(metadata, "metadata[?(@.name=='metadata_softpartition.csv')].ID")
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments/types'
    * request { type:'softpartition', metadata_id:'#(meta_sp_id[0])', "parent_id":'#(server_id)', attributes:'#(attr.softpartition)', scopes:['#(scope)']}
    When method post
    Then status 200
    * def softpartition_id = response.ID
    # Create HardPartition
    * def meta_hp_id = karate.jsonPath(metadata, "metadata[?(@.name=='metadata_hardpartition.csv')].ID")
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments/types'
    * request { type:'hardpartition', metadata_id:'#(meta_hp_id[0])', "parent_id":'#(server_id)', attributes:'#(attr.hardpartition)', scopes:['#(scope)']}
    When method post
    Then status 200

 Scenario: Create Metrics
   # fetch equipment types
   * url equipmentServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'equipments/types'
   And params {scopes:'#(scope)'}
   When method get
   Then status 200
   * def server_eq_type = karate.jsonPath(response, "$.equipment_types[?(@.type=='server')]")[0]
   * def softpartition_eq_type = karate.jsonPath(response, "$.equipment_types[?(@.type=='softpartition')]")[0]
    * def hardpartition_eq_type = karate.jsonPath(response, "$.equipment_types[?(@.type=='hardpartition')]")[0]
   * def cluster_eq_type = karate.jsonPath(response, "$.equipment_types[?(@.type=='cluster')]")[0]
   * def vcenter_eq_type = karate.jsonPath(response, "$.equipment_types[?(@.type=='vcenter')]")[0]
   * def serv_core = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='cores')].ID")[0]
   * def serv_processor = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='server_processors_numbers')].ID")[0]
   * def server_oracle = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='core_factor')].ID")[0]
   # * def server_pvu = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='pvu')].ID")[0]
   # * def server_sag = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='sag')].ID")[0]
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   # metric instance
   Given path 'metric/inm'
   * request {"Name": "os_instance","Coefficient": "2","scopes": ['#(scope)']}
   When method post
    Then status 200
   * match response.Name == "os_instance"
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/inm'
   * request {"Name": "CEPH-1-NODE","Coefficient": "1","scopes": ['#(scope)']}
   When method post
   Then status 200
 
   #  metric attribute counter
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/acs'
   * request {"name": "attribute_counter_cpu","eq_type": "server","attribute_name": "cpu_model","value": "AMD Opteron(tm) Processor 6136","scopes":['#(scope)']}
   When method post
   Then status 200
   
#    ###
# #   # metric attribute counter
#    * url metricServiceUrl+'/api/v1'
#    * header Authorization = 'Bearer '+access_token
#    Given path 'metric/acs'
#    * request {"name": "attribute_counter_cpu","eq_type": "server","attribute_name": "server_cpu","value": "Intel","scopes":['#(scope)']}
#    When method post
#    Then status 200
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/acs'
   * request {"name": "attribute_counter_core","eq_type": "server","attribute_name": "cores","value": "8","scopes":['#(scope)']}
   When method post
   Then status 200
  #  # metric ibm pvu
#    * url metricServiceUrl+'/api/v1'
#    * header Authorization = 'Bearer '+access_token
#    Given path 'metric/ips'
#    * request {"Name":"ibm_pvu","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_pvu)',"numCPU_attr_id":null,"base_eq_type_id":'#(server_eq_type.ID)',"number_of_users":0,"scopes":['#(scope)']}
#    When method post
#    Then status 200
#    * url metricServiceUrl+'/api/v1'
#    * header Authorization = 'Bearer '+access_token
#    Given path 'metric/ips'
#    * request {"Name":"ibm_pvu_75","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_pvu)',"numCPU_attr_id":null,"base_eq_type_id":'#(server_eq_type.ID)',"number_of_users":0,"scopes":['#(scope)']}
#    When method post
#    Then status 200
#   #  # metric sag
#    * url metricServiceUrl+'/api/v1'
#    * header Authorization = 'Bearer '+access_token
#    Given path 'metric/sps'
#    * request {"Name":"sag","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_sag)',"numCPU_attr_id":null,"base_eq_type_id":'#(server_eq_type.ID)',"number_of_users":0,"scopes":['#(scope)']}
#    When method post
#    Then status 200
   # metric oracle processor
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/ops'
   * request {"Name":"oracleprocesser","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_oracle)',"numCPU_attr_id":'#(serv_processor)',"base_eq_type_id":'#(server_eq_type.ID)',"start_eq_type_id":'#(server_eq_type.ID)',"aggerateLevel_eq_type_id":'#(cluster_eq_type.ID)',"end_eq_type_id":'#(vcenter_eq_type.ID)',"number_of_users":0,"scopes":['#(scope)']}
   When method post
   Then status 200
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/ops'
   * request {"Name":"ops_partition","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_oracle)',"numCPU_attr_id":'#(serv_processor)',"base_eq_type_id":'#(server_eq_type.ID)',"start_eq_type_id":'#(softpartition_eq_type.ID)',"aggerateLevel_eq_type_id":'#(cluster_eq_type.ID)',"end_eq_type_id":'#(vcenter_eq_type.ID)',"number_of_users":0,"scopes":['#(scope)']}
   When method post
   Then status 200
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/ops'
   * request {"Name":"ops_serv","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_oracle)',"numCPU_attr_id":'#(serv_processor)',"base_eq_type_id":'#(server_eq_type.ID)',"start_eq_type_id":'#(server_eq_type.ID)',"aggerateLevel_eq_type_id":'#(cluster_eq_type.ID)',"end_eq_type_id":'#(vcenter_eq_type.ID)',"number_of_users":0,"scopes":['#(scope)']}
   When method post
   Then status 200
   # metric oracle nup
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/oracle_nup'
   * request {"Name":"oracle.nup.standard1","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_oracle)',"numCPU_attr_id":'#(serv_processor)',"base_eq_type_id":'#(server_eq_type.ID)',"start_eq_type_id":'#(server_eq_type.ID)',"aggerateLevel_eq_type_id":'#(cluster_eq_type.ID)',"end_eq_type_id":'#(vcenter_eq_type.ID)',"number_of_users":25,"scopes":['#(scope)']}
   When method post
   Then status 200
   * url metricServiceUrl+'/api/v1' 
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/oracle_nup'
   * request {"Name":"nup_partition","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_oracle)',"numCPU_attr_id":'#(serv_processor)',"base_eq_type_id":'#(server_eq_type.ID)',"start_eq_type_id":'#(softpartition_eq_type.ID)',"aggerateLevel_eq_type_id":'#(cluster_eq_type.ID)',"end_eq_type_id":'#(vcenter_eq_type.ID)',"number_of_users":30,"scopes":['#(scope)']}
   When method post
   Then status 200
