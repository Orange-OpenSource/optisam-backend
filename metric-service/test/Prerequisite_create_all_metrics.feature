@metric
Feature: To validate CRUD operation on metrics : admin user

  Background:
    * url metricServiceUrl+'/api/v1'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

  Scenario Outline: To verify user can create all metrics
    Given path <path>
    * def eq_data = callonce read('get_equipments_id.feature')
    * def ibm_payload = {"Name":"ibm_pvu","num_core_attr_id":'#(eq_data.serv_core.ID)',"core_factor_attr_id":'#(eq_data.server_oracle.ID)',"numCPU_attr_id":'#(eq_data.serv_processor.ID)',"base_eq_type_id":'#(eq_data.server_eq_type.ID)', "scopes":['#(scope)']}
    * def sag_payload = {"Name":"sag.processor","num_core_attr_id":'#(eq_data.serv_core.ID)',"core_factor_attr_id":'#(eq_data.server_oracle.ID)',"numCPU_attr_id":'#(eq_data.serv_processor.ID)',"base_eq_type_id":'#(eq_data.server_eq_type.ID)',"scopes":['#(scope)']}
    * def ops_payload = {"Name":"oracle.processor","num_core_attr_id":'#(eq_data.serv_core.ID)',"core_factor_attr_id":'#(eq_data.server_oracle.ID)',"numCPU_attr_id":'#(eq_data.serv_processor.ID)',"base_eq_type_id":'#(eq_data.server_eq_type.ID)',"start_eq_type_id":'#(eq_data.server_eq_type.ID)',"aggerateLevel_eq_type_id":'#(eq_data.cluster_eq_type.ID)',"end_eq_type_id":'#(eq_data.vcenter_eq_type.ID)',"scopes":['#(scope)']}
    * def nup_payload = {"Name":"oracle.nup","num_core_attr_id":'#(eq_data.serv_core.ID)',"core_factor_attr_id":'#(eq_data.server_oracle.ID)',"numCPU_attr_id":'#(eq_data.serv_processor.ID)',"base_eq_type_id":'#(eq_data.server_eq_type.ID)',"start_eq_type_id":'#(eq_data.server_eq_type.ID)',"aggerateLevel_eq_type_id":'#(eq_data.cluster_eq_type.ID)',"end_eq_type_id":'#(eq_data.vcenter_eq_type.ID)',"number_of_users":1,"scopes":['#(scope)']}
    * request <payload>
    When method post
    Then status 200
    And response.success == true
    * def result = karate.jsonPath(response, "$.['name','Name']")
    And match ((result.name==<m_name>) || (result.Name==<m_name>)) == true
  Examples:
    | path | payload | m_name |
    | 'metric/ips' | ibm_payload | 'ibm_pvu' |
    | 'metric/sps' | sag_payload | 'sag.processor' |
    | 'metric/ops' | ops_payload | 'oracle.processor' |
    | 'metric/oracle_nup' | nup_payload | 'oracle.nup'|
    | 'metric/inm' | data.metric_instance_number_standard_1 | data.metric_instance_number_standard_1.Name |
    | 'metric/inm' | data.metric_instance_number_standard_62 | data.metric_instance_number_standard_62.Name |
    | 'metric/inm' | data.metric_instance_number_standard_8 | data.metric_instance_number_standard_8.Name |
    | 'metric/acs' | data.metric_attribute_counter_standard | data.metric_attribute_counter_standard.name |
    | 'metric/attr_sum' | data.metric_attribute_sum_standard | data.metric_attribute_sum_standard.name |
    | 'metric/uss' | data.metric_user_sum_standard | data.metric_user_sum_standard.Name |
    | 'metric/ss' | data.metric_static_standard | data.metric_static_standard.Name |