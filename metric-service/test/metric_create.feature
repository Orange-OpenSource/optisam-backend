@metric
Feature: Metric Service Test - Create new metrics : admin user

  Background:
  # * def metricServiceUrl = "https://optisam-metric-int.kermit-noprod-b.itn.intraorange"
    * url metricServiceUrl+'/api/v1'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'

 @create
  Scenario: To verify user can create instance metric and delete it
    Given path '/metric/inm'
    And request data.metric_inm
    When method post
    Then status 200
   * match response.Name == data.metric_inm.Name
   * header Authorization = 'Bearer '+access_token
    Given path 'metric' , data.metric_inm.Name
    * params {scope:'#(scope)'}
    When method delete
    Then status 200
    And response.success == true

  # @create
  # Scenario: Create IBM metrics
  #   Given path '/metric/ips'
  #   And request data.metric_ibm
  #   When method post
  #   Then status 200
  #  * match response.Name == data.metric_ibm.Name
  #  * header Authorization = 'Bearer '+access_token
  #   Given path 'metric' , data.metric_ibm.Name
  #   * params {scope:'#(scope)'}
  #   When method delete
  #   Then status 200
  #   And response.success == true
  
   Scenario: To verify user can create oracle processor metric and delete it
   # fetch equipment types
   * url equipmentServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'equipment/types'
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
   # metric oracle processor
   * url metricServiceUrl+'/api/v1'
   * header Authorization = 'Bearer '+access_token
   Given path 'metric/ops'
   * request {"Name":"apitest_ops","num_core_attr_id":'#(serv_core)',"core_factor_attr_id":'#(server_oracle)',"numCPU_attr_id":'#(serv_processor)',"base_eq_type_id":'#(server_eq_type.ID)',"start_eq_type_id":'#(server_eq_type.ID)',"aggerateLevel_eq_type_id":'#(cluster_eq_type.ID)',"end_eq_type_id":'#(vcenter_eq_type.ID)',"number_of_users":0,"scopes":['#(scope)']}
   When method post
   Then status 200
   And response.Name = "apitest_ops"
   * header Authorization = 'Bearer '+access_token
   Given path 'metric' , "apitest_ops"
   * params {scope:'#(scope)'}
   When method delete
   Then status 200
   And response.success == true