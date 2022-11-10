@aut @ignore @aut-setup
Feature: Pre-Requisite Setup for AUT(Automation) - Create Equipment type, Metric

  Background:
    * url dpsServiceUrl+'/api/v1'
    #* def credentials = {username:'admin@test.com', password: 'Welcome@123'}
   * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def scope = "AUT"

 Scenario: Upload Metadata files
    Given url importServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'import/metadata'
    * def file1_tmp = karate.readAsString('metadata/metadata_datacenter.csv')
    * def file2_tmp = karate.readAsString('metadata/metadata_vcenter.csv')
    * def file3_tmp = karate.readAsString('metadata/metadata_cluster.csv')
    * def file4_tmp = karate.readAsString('metadata/metadata_server.csv')
    * def file5_tmp = karate.readAsString('metadata/metadata_partition.csv')
    * multipart file file = { value: '#(file1_tmp)', filename: 'metadata_datacenter.csv', contentType: "text/csv" }
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
    # Create Datacenter
    * def meta_datacenter_id = karate.jsonPath(metadata, "metadata[?(@.name=='metadata_datacenter.csv')].ID")
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments/types'
    * request { type:'datacenter', metadata_id:'#(meta_datacenter_id[0])', attributes:'#(attr.datacenter)', scopes:['#(scope)']}
    When method post
    Then status 200
    # dgraph uid for datacenter
    * def datacenter_id = response.ID 
    # Create Vcenter
    * def meta_vcenter_id = karate.jsonPath(metadata, "metadata[?(@.name=='metadata_vcenter.csv')].ID")
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments/types'
    * request { type:'vcenter', metadata_id:'#(meta_vcenter_id[0])', "parent_id":'#(datacenter_id)', attributes:'#(attr.vcenter)', scopes:['#(scope)']}
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


 Scenario: Create Metrics
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
   * def serv_core = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='server_coresNumber')].ID")[0]
   * def serv_processor = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='server_processorsNumber')].ID")[0]
   * def server_oracle = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='corefactor_oracle')].ID")[0]
   * def server_pvu = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='pvu')].ID")[0]
   * def server_sag = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='sag')].ID")[0]
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   # metric instance
   Given path 'metric/inm'
   * request {"Name": "os_instance","Coefficient": "2","scopes": ['#(scope)']}
   When method post
   Then status 200
   # * match response.name == "os_instance"
   # metric attribute counter
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/acs'
   * request {"name": "attribute_counter_cpu","eq_type": "server","attribute_name": "server_cpu","value": "Intel","scopes":['#(scope)']}
   When method post
   Then status 200
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/acs'
   * request {"name": "attribute_counter_core","eq_type": "server","attribute_name": "server_coresNumber","value": "2","scopes":['#(scope)']}
   When method post
   Then status 200
   # metric ibm pvu
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/ips'
   * request {"Name":"ibm_pvu","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_pvu)',"numCPU_attr_id":null,"base_eq_type_id":'#(server_eq_type.ID)',"number_of_users":0,"scopes":['#(scope)']}
   When method post
   Then status 200
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/ips'
   * request {"Name":"ibm_pvu_75","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_pvu)',"numCPU_attr_id":null,"base_eq_type_id":'#(server_eq_type.ID)',"number_of_users":0,"scopes":['#(scope)']}
   When method post
   Then status 200
   # metric sag
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/sps'
   * request {"Name":"sag","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_sag)',"numCPU_attr_id":null,"base_eq_type_id":'#(server_eq_type.ID)',"number_of_users":0,"scopes":['#(scope)']}
   When method post
   Then status 200
   # metric oracle processor
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/ops'
   * request {"Name":"oracle.processor.standard","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_oracle)',"numCPU_attr_id":'#(serv_processor)',"base_eq_type_id":'#(server_eq_type.ID)',"start_eq_type_id":'#(server_eq_type.ID)',"aggerateLevel_eq_type_id":'#(cluster_eq_type.ID)',"end_eq_type_id":'#(vcenter_eq_type.ID)',"number_of_users":0,"scopes":['#(scope)']}
   When method post
   Then status 200
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/ops'
   * request {"Name":"ops_partition","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_oracle)',"numCPU_attr_id":'#(serv_processor)',"base_eq_type_id":'#(server_eq_type.ID)',"start_eq_type_id":'#(partition_eq_type.ID)',"aggerateLevel_eq_type_id":'#(cluster_eq_type.ID)',"end_eq_type_id":'#(vcenter_eq_type.ID)',"number_of_users":0,"scopes":['#(scope)']}
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
   * request {"Name":"oracle.nup.standard","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_oracle)',"numCPU_attr_id":'#(serv_processor)',"base_eq_type_id":'#(server_eq_type.ID)',"start_eq_type_id":'#(server_eq_type.ID)',"aggerateLevel_eq_type_id":'#(cluster_eq_type.ID)',"end_eq_type_id":'#(vcenter_eq_type.ID)',"number_of_users":25,"scopes":['#(scope)']}
   When method post
   Then status 200
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/oracle_nup'
   * request {"Name":"nup_partition","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_oracle)',"numCPU_attr_id":'#(serv_processor)',"base_eq_type_id":'#(server_eq_type.ID)',"start_eq_type_id":'#(partition_eq_type.ID)',"aggerateLevel_eq_type_id":'#(cluster_eq_type.ID)',"end_eq_type_id":'#(vcenter_eq_type.ID)',"number_of_users":30,"scopes":['#(scope)']}
   When method post
   Then status 200