@metric
Feature: To validate CRUD operation on metrics : admin user

  Background:
    * url metricServiceUrl+'/api/v1'
    #* def credentials = {username:'admin@test.com', password: 'Welcome@123'}
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

  @create
  Scenario: To verify user can create instance metric.
    Given path '/metric/inm'
    And request data.crud_metric_inm
    When method post
    Then status 200
   * match response.Name == data.crud_metric_inm.Name
    And response.success == true
    # validate schema
   * def schema = read('schema_data.json')
    And match response == '#(schema.instance_metric)'

  @create
  Scenario: To verify user can update instance metric
    # get instance metric
      Given path 'metric/config'
    * def metric_name = data.crud_metric_inm.Name
    * params {metric_info.type:'instance.number.standard' , metric_info.name:'#(metric_name)' , scopes:'#(scope)'}
      When method get
      Then status 200
    * json json_res = karate.jsonPath(response, "$.metric_config")
    * match json_res.Name == data.crud_metric_inm.Name
    # update istance metric
    * header Authorization = 'Bearer '+access_token
      Given path '/metric/inm'
    * def inn_payload = {"ID": "", "Name": '#(json_res.Name)', "num_of_deployments": "2","scopes": ['#(scope)'] }
      And request inn_payload
      When method put
      Then status 200
      And response.success == true
      #validate metric is updated or not
    * header Authorization = 'Bearer '+access_token
      Given path 'metric/config'
    * def metric_name = data.crud_metric_inm.Name
    * params {metric_info.type:'instance.number.standard' , metric_info.name:'#(metric_name)' , scopes:'#(scope)'}
      When method get
      Then status 200
    * json metric_res = karate.jsonPath(response, "$.metric_config")
    * match metric_res.Coefficient == 2

  Scenario: To verify user can delete instance metric
      Given path 'metric/config'
    * def metric_name = data.crud_metric_inm.Name
    * params {metric_info.type:'instance.number.standard' , metric_info.name:'#(metric_name)' , scopes:'#(scope)'}
      When method get
      Then status 200
    * json json_res = karate.jsonPath(response, "$.metric_config")
    * match json_res.Name == data.crud_metric_inm.Name
    * header Authorization = 'Bearer '+access_token
      Given path 'metric' , json_res.Name
    * params {scope:'#(scope)'}
      When method delete
      Then status 200
      And response.success == true

  Scenario: To verify user can create oracle processor metric.
      * def eq_data = call read('get_equipments_id.feature')
      # creating metric oracle processor
      * url metricServiceUrl+'/api/v1'
      * header Authorization = 'Bearer '+access_token
      Given path 'metric/ops'
      * request {"Name":"apitest_ops","num_core_attr_id":'#(eq_data.serv_core.ID)',"core_factor_attr_id":'#(eq_data.server_oracle.ID)',"numCPU_attr_id":'#(eq_data.serv_processor.ID)',"base_eq_type_id":'#(eq_data.server_eq_type.ID)',"start_eq_type_id":'#(eq_data.server_eq_type.ID)',"aggerateLevel_eq_type_id":'#(eq_data.cluster_eq_type.ID)',"end_eq_type_id":'#(eq_data.vcenter_eq_type.ID)',"number_of_users":1,"scopes":['#(scope)']}
      When method post
      Then status 200
      And response.Name == "apitest_ops"
      # validate schema
      * def schema = read('schema_data.json')
      And match response == '#(schema.oracle_processor)'

  @create
  Scenario: To verify user can update oracle processor metric.
      # get oracle metric
      Given path 'metric/config'
    * params {metric_info.type:'oracle.processor.standard' , metric_info.name:'apitest_ops' , scopes:'#(scope)'}
      When method get
      Then status 200
    * json json_res = karate.jsonPath(response, "$.metric_config")
    * match json_res.Name == 'apitest_ops'
    * def eq_data = call read('get_equipments_id.feature')
    # update oracle metric
    * url metricServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    * def oracle_payload = {"Name":"apitest_ops","num_core_attr_id":'#(eq_data.server_pvu.ID)',"core_factor_attr_id":'#(eq_data.server_oracle.ID)',"numCPU_attr_id":'#(eq_data.server_sag.ID)',"base_eq_type_id":'#(eq_data.server_eq_type.ID)',"start_eq_type_id":'#(eq_data.server_eq_type.ID)',"aggerateLevel_eq_type_id":'#(eq_data.cluster_eq_type.ID)',"end_eq_type_id":'#(eq_data.vcenter_eq_type.ID)',"number_of_users":1,"scopes":['#(scope)']}
      And request oracle_payload
      Given path 'metric/ops'
      When method put
      Then status 200
      And response.success == true
      #validate metric is updated or not
      Given path 'metric/config'
    * header Authorization = 'Bearer '+access_token
    * params {metric_info.type:'oracle.processor.standard' , metric_info.name:'apitest_ops' , scopes:'#(scope)'}
      When method get
      Then status 200
    * json json_res = karate.jsonPath(response, "$.metric_config")
    * print json_res
    * match json_res.NumCoreAttr == 'ibm_pvu'
    * match json_res.NumCPUAttr == 'sag_uvu'

  @delete
  Scenario: To verify user can delete oracle processor metric.
    # get oracle metric
    Given path 'metric/config'
  * params {metric_info.type:'oracle.processor.standard' , metric_info.name:'apitest_ops' , scopes:'#(scope)'}
    When method get
    Then status 200
  * json json_res = karate.jsonPath(response, "$.metric_config")
  * match json_res.Name == 'apitest_ops'
  * header Authorization = 'Bearer '+access_token
    Given path 'metric' , json_res.Name
  * params {scope:'#(scope)'}
    When method delete
    Then status 200
    And response.success == true
 
  @create
  Scenario: To verfiy user can create IBM metric
     * def eq_data = call read('get_equipments_id.feature')
     # creating metric sag processor
     * url metricServiceUrl+'/api/v1'
     * header Authorization = 'Bearer '+access_token
     Given path '/metric/ips'
     * request {"Name":"apitest_ibm_pvu_std","num_core_attr_id":'#(eq_data.serv_core.ID)',"core_factor_attr_id":'#(eq_data.server_oracle.ID)',"numCPU_attr_id":'#(eq_data.serv_processor.ID)',"base_eq_type_id":'#(eq_data.server_eq_type.ID)',"scopes":['#(scope)']}
     When method post
     Then status 200
     And match response.Name == 'apitest_ibm_pvu_std'
     # validate schema
     * def schema = read('schema_data.json')
     And match response == '#(schema.ibm_metric)'


  @update
  Scenario Outline: To verfiy user can update IBM metric 
    # get IBM metric
    Given path 'metric/config'
    * params {metric_info.type:'ibm.pvu.standard' , metric_info.name:'apitest_ibm_pvu_std' , scopes:'#(scope)'}
    When method get
    Then status 200
    * json json_res = karate.jsonPath(response, "$.metric_config")
    * match json_res.Name == 'apitest_ibm_pvu_std'
    * def eq_data = callonce read('get_equipments_id.feature')
    #update & validate ibm metrics
    * url metricServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path '/metric/ips'
    * def inn_payload = {"ID": "", "Name": '#(json_res.Name)', "num_core_attr_id": <numCore>, "numCPU_attr_id":<numCpu>, "core_factor_attr_id":<corefactor> ,"base_eq_type_id":'#(eq_data.server_eq_type.ID)',"scopes": ['#(scope)'] }
    * print inn_payload
    And request inn_payload
    When method put
    Then status 200
    And response.success == true
    #validate metric is updated or not
    Given path 'metric/config'
    * header Authorization = 'Bearer '+access_token
    * params {metric_info.type:'ibm.pvu.standard' , metric_info.name:'apitest_ibm_pvu_std' , scopes:'#(scope)'}
    When method get
    Then status 200
    * json json_res = karate.jsonPath(response, "$.metric_config")
    * def attr_expec = ['cores_per_processor', 'ibm_pvu', 'oracle_core_factor', 'sag_uvu', 'server_processors_numbers']
    * match attr_expec contains json_res.NumCoreAttr
    * match attr_expec contains json_res.NumCPUAttr
    * match attr_expec contains json_res.CoreFactorAttr
  Examples:
    | numCore | numCpu | corefactor |
    | '#(eq_data.serv_core.ID)' | '#(eq_data.serv_processor.ID)' | '#(eq_data.server_oracle.ID)' |
    | '#(eq_data.server_pvu.ID)' | '#(eq_data.serv_core.ID)' | '#(eq_data.serv_processor.ID)' |
    | '#(eq_data.server_oracle.ID)' | '#(eq_data.server_sag.ID)' | '#(eq_data.server_pvu.ID)' |
    | '#(eq_data.serv_processor.ID)' | '#(eq_data.server_pvu.ID)' | '#(eq_data.server_sag.ID)'|
    | '#(eq_data.serv_core.ID)' | '#(eq_data.server_pvu.ID)' | '#(eq_data.serv_processor.ID)' |
    | '#(eq_data.server_pvu.ID)' | '#(eq_data.server_sag.ID)' | '#(eq_data.serv_processor.ID)' |
    | '#(eq_data.server_oracle.ID)' | '#(eq_data.server_pvu.ID)' | '#(eq_data.server_sag.ID)' |
    | '#(eq_data.serv_processor.ID)' | '#(eq_data.serv_core.ID)' | '#(eq_data.server_sag.ID)' |
    | '#(eq_data.serv_core.ID)' | '#(eq_data.server_sag.ID)' | '#(eq_data.server_pvu.ID)' |
    | '#(eq_data.server_sag.ID)' | '#(eq_data.serv_core.ID)' | '#(eq_data.serv_processor.ID)' |

 
  @delete
  Scenario: To verify user can delete IBM metric.
    # get IBM metric
    Given path 'metric/config'
  * params {metric_info.type:'ibm.pvu.standard' , metric_info.name:'apitest_ibm_pvu_std' , scopes:'#(scope)'}
    When method get
    Then status 200
  * json json_res = karate.jsonPath(response, "$.metric_config")
  * match json_res.Name == 'apitest_ibm_pvu_std'
  * header Authorization = 'Bearer '+access_token
    Given path 'metric' , json_res.Name
  * params {scope:'#(scope)'}
    When method delete
    Then status 200
    And response.success == true


  Scenario: To verify user can create oracle nup metric.
    * def eq_data = call read('get_equipments_id.feature')
    # creating metric oracle nup
    * url metricServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'metric/oracle_nup'
    * request {"Name":"apitest_oracle_nup","num_core_attr_id":'#(eq_data.serv_core.ID)',"core_factor_attr_id":'#(eq_data.server_oracle.ID)',"numCPU_attr_id":'#(eq_data.serv_processor.ID)',"base_eq_type_id":'#(eq_data.server_eq_type.ID)',"start_eq_type_id":'#(eq_data.server_eq_type.ID)',"aggerateLevel_eq_type_id":'#(eq_data.cluster_eq_type.ID)',"end_eq_type_id":'#(eq_data.vcenter_eq_type.ID)',"number_of_users":1,"scopes":['#(scope)']}
    When method post
    Then status 200
    And match response.Name == "apitest_oracle_nup"
    # validate schema
    * def schema = read('schema_data.json')
    And match response == '#(schema.oracle_nup)'

  @update
  Scenario: To verify user can update oracle nup metric.
    # get oracle nup metric
    Given path 'metric/config'
    * params {metric_info.type:'oracle.nup.standard' , metric_info.name:'apitest_oracle_nup' , scopes:'#(scope)'}
    When method get
    Then status 200
    * json json_res = karate.jsonPath(response, "$.metric_config")
    * match json_res.Name == 'apitest_oracle_nup'
    * def eq_data = call read('get_equipments_id.feature')
    # update the oracle nup metric
    * url metricServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    * def oracle_nup_payload = {"Name":"apitest_oracle_nup","num_core_attr_id":'#(eq_data.server_pvu.ID)',"core_factor_attr_id":'#(eq_data.server_oracle.ID)',"numCPU_attr_id":'#(eq_data.server_sag.ID)',"base_eq_type_id":'#(eq_data.server_eq_type.ID)',"start_eq_type_id":'#(eq_data.server_eq_type.ID)',"aggerateLevel_eq_type_id":'#(eq_data.cluster_eq_type.ID)',"end_eq_type_id":'#(eq_data.vcenter_eq_type.ID)',"number_of_users":6,"scopes":['#(scope)']}
    And request oracle_nup_payload
    Given path 'metric/oracle_nup'
    When method put
    Then status 200
    And response.success == true
    #validate metric is updated or not
    Given path 'metric/config'
    * header Authorization = 'Bearer '+access_token
    * params {metric_info.type:'oracle.nup.standard' , metric_info.name:'apitest_oracle_nup' , scopes:'#(scope)'}
    When method get
    Then status 200
    * json json_res = karate.jsonPath(response, "$.metric_config")
    * print json_res
    * match json_res.NumCoreAttr == 'ibm_pvu'
    * match json_res.NumCPUAttr == 'sag_uvu'
    * match json_res.NumberOfUsers == 6

  @delete
  Scenario: To verify user can delete oracle nup metric.
    # get oracle nup metric
    Given path 'metric/config'
  * params {metric_info.type:'oracle.nup.standard' , metric_info.name:'apitest_oracle_nup' , scopes:'#(scope)'}
    When method get
    Then status 200
  * json json_res = karate.jsonPath(response, "$.metric_config")
  * match json_res.Name == 'apitest_oracle_nup'
  * header Authorization = 'Bearer '+access_token
    Given path 'metric' , json_res.Name
  * params {scope:'#(scope)'}
    When method delete
    Then status 200
    And response.success == true
  
  Scenario: To verify user can create sag processor metric.
    * def eq_data = call read('get_equipments_id.feature')
    # creating metric sag processor
    * url metricServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'metric/sps'
    * request {"Name":"apitest_sag_processor","num_core_attr_id":'#(eq_data.serv_core.ID)',"core_factor_attr_id":'#(eq_data.server_oracle.ID)',"numCPU_attr_id":'#(eq_data.serv_processor.ID)',"base_eq_type_id":'#(eq_data.server_eq_type.ID)',"scopes":['#(scope)']}
    When method post
    Then status 200
    And match response.Name == "apitest_sag_processor"
    # validate schema
    * def schema = read('schema_data.json')
    And match response == '#(schema.sag_metric)'
  
  @update
  Scenario Outline: To verfiy user can update sag processor metric 
    # get sag processor metric
    Given path 'metric/config'
    * params {metric_info.type:'sag.processor.standard' , metric_info.name:'apitest_sag_processor' , scopes:'#(scope)'}
    When method get
    Then status 200
    * json json_res = karate.jsonPath(response, "$.metric_config")
    * match json_res.Name == 'apitest_sag_processor'
    * def eq_data = callonce read('get_equipments_id.feature')
    #update sag processor metric
    * url metricServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path '/metric/sps'
    * def inn_payload = {"ID": "", "Name": '#(json_res.Name)', "num_core_attr_id": <numCore>, "numCPU_attr_id":<numCpu>, "core_factor_attr_id":<corefactor> ,"base_eq_type_id":'#(eq_data.server_eq_type.ID)',"scopes": ['#(scope)'] }
    And request inn_payload
    When method put
    Then status 200
    And response.success == true
    #validate metric is updated or not
    Given path 'metric/config'
    * header Authorization = 'Bearer '+access_token
    * params {metric_info.type:'sag.processor.standard' , metric_info.name:'apitest_sag_processor' , scopes:'#(scope)'}
      When method get
      Then status 200
    * json json_res = karate.jsonPath(response, "$.metric_config")
    * def attr_expec = ['cores_per_processor', 'ibm_pvu', 'oracle_core_factor', 'sag_uvu', 'server_processors_numbers']
    * match attr_expec contains json_res.NumCoreAttr
    * match attr_expec contains json_res.NumCPUAttr
    * match attr_expec contains json_res.CoreFactorAttr
  Examples:
    | numCore | numCpu | corefactor |
    | '#(eq_data.serv_core.ID)' | '#(eq_data.serv_processor.ID)' | '#(eq_data.server_oracle.ID)' |
    | '#(eq_data.server_pvu.ID)' | '#(eq_data.serv_core.ID)' | '#(eq_data.serv_processor.ID)' |
    | '#(eq_data.server_oracle.ID)' | '#(eq_data.server_sag.ID)' | '#(eq_data.server_pvu.ID)' |
    | '#(eq_data.serv_processor.ID)' | '#(eq_data.server_pvu.ID)' | '#(eq_data.server_sag.ID)'|
    | '#(eq_data.serv_core.ID)' | '#(eq_data.server_pvu.ID)' | '#(eq_data.serv_processor.ID)' |
    | '#(eq_data.server_pvu.ID)' | '#(eq_data.server_sag.ID)' | '#(eq_data.serv_processor.ID)' |
    | '#(eq_data.server_oracle.ID)' | '#(eq_data.server_pvu.ID)' | '#(eq_data.server_sag.ID)' |
    | '#(eq_data.serv_processor.ID)' | '#(eq_data.serv_core.ID)' | '#(eq_data.server_sag.ID)' |
    | '#(eq_data.serv_core.ID)' | '#(eq_data.server_sag.ID)' | '#(eq_data.serv_core.ID)' |
    | '#(eq_data.server_sag.ID)' | '#(eq_data.serv_core.ID)' | '#(eq_data.serv_processor.ID)' |

  @delete
  Scenario: To verify user can delete sag processor metric.
    # get sag processor metric
    Given path 'metric/config'
  * params {metric_info.type:'sag.processor.standard' , metric_info.name:'apitest_sag_processor' , scopes:'#(scope)'}
    When method get
    Then status 200
  * json json_res = karate.jsonPath(response, "$.metric_config")
  * match json_res.Name == 'apitest_sag_processor'
  * header Authorization = 'Bearer '+access_token
    Given path 'metric' , json_res.Name
  * params {scope:'#(scope)'}
    When method delete
    Then status 200
    And response.success == true

  Scenario: To verify user can create static standard metric.  
    Given path 'metric/ss'
    * request data.crud_metric_static_standard
    When method post
    Then status 200
    And response.success == true
    And match response.Name == data.crud_metric_static_standard.Name
    # validate schema
    * def schema = read('schema_data.json')
    And match response == '#(schema.static_std_metric)'

  Scenario: To verify user can update static standard metric.
    Given path 'metric/config'
    And params {metric_info.type:'static.standard' , metric_info.name:'#(data.crud_metric_static_standard.Name)' , scopes:'#(scope)'}
    When method get
    Then status 200
    * json json_res = karate.jsonPath(response, "$.metric_config")
    # update istance metric
    * header Authorization = 'Bearer '+access_token
    Given path '/metric/ss'
    * def inn_payload = {"ID": "", "Name": '#(json_res.Name)', "reference_value": 8,"scopes": ["#(scope)"] }
    And request inn_payload
    When method put
    Then status 200
    And response.success == true
    
    Scenario: To verify user can delete static standard metric.
      Given path 'metric/config'
    * params {metric_info.type:'static.standard' , metric_info.name:'#(data.crud_metric_static_standard.Name)' , scopes:'#(scope)'}
      When method get
      Then status 200
    * json json_res = karate.jsonPath(response, "$.metric_config")
    * header Authorization = 'Bearer '+access_token
      Given path 'metric' , json_res.Name
    * params {scope:'#(scope)'}
      When method delete
      Then status 200
      And response.success == true

    Scenario: To verify user can create user sum standard metric
      Given path 'metric/uss'
      * request data.crud_metric_user_sum_standard
      When method post
      Then status 200
      And response.success == true
      And match response.Name == data.crud_metric_user_sum_standard.Name
      # validate schema
      * def schema = read('schema_data.json')
      And match response == '#(schema.user_sum_metric)'
    
    Scenario: To verify user can delete user sum standard metric
      Given path 'metric/config'
      * params {metric_info.type:'user.sum.standard' , metric_info.name:'#(data.crud_metric_user_sum_standard.Name)', scopes:'#(scope)'}
      When method get
      Then status 200
      * json json_res = karate.jsonPath(response, "$.metric_config")
      * header Authorization = 'Bearer '+access_token
      Given path 'metric' , json_res.Name
      * params {scope:'#(scope)'}
      When method delete
      Then status 200
      And response.success == true

    Scenario: To verify user can create attribute sum standard
      Given path 'metric/attr_sum'
      * request data.crud_metric_attr_sum_std
      When method post
      Then status 200
      And response.success == true
      And match response.name == data.crud_metric_attr_sum_std.name
      # validate schema
      * def schema = read('schema_data.json')
      And match response == '#(schema.attribute_sum_std_metric)'
      
      @update
    Scenario Outline: To verify user can update attribute sum standard
    # get metric config of attribute sum standard
      Given path 'metric/config'
      * params {metric_info.type:'attribute.sum.standard' , metric_info.name:'#(data.crud_metric_attr_sum_std.name)' , scopes:'#(scope)'}
      When method get
      Then status 200
      * json json_res = karate.jsonPath(response, "$.metric_config")
      * url equipmentServiceUrl+'/api/v1'
      * header Authorization = 'Bearer '+access_token
      * def eq_data = callonce read('get_equipments_id.feature')
      #update attribute sum standard metric
      * url metricServiceUrl+'/api/v1'
      * header Authorization = 'Bearer '+access_token
      Given path '/metric/attr_sum'
      * def inn_payload = {"ID": "", "name": '#(json_res.Name)', "eq_type": 'server' , "attribute_name":<attr_name> , "reference_value":7 , "scopes":['#(scope)']}
      * print inn_payload
      And request inn_payload
      When method put
      Then status 200
      And response.success == true
      #  validate metric is updated or not
      Given path 'metric/config'
      * header Authorization = 'Bearer '+access_token
      * params {metric_info.type:'attribute.sum.standard' , metric_info.name:'#(data.crud_metric_attr_sum_std.name)' , scopes:'#(scope)'}
      When method get
      Then status 200
      Examples:
        | attr_name | 
        | '#(eq_data.serv_core.name)' |
        | '#(eq_data.serv_processor.name)' |
        | '#(eq_data.server_oracle.name)' |
        | '#(eq_data.server_pvu.name)' |
        | '#(eq_data.server_sag.name)' |
    
    Scenario: To verify user can delete attribute sum standard metric
      Given path 'metric/config'
      * params {metric_info.type:'attribute.sum.standard' , metric_info.name:'#(data.crud_metric_attr_sum_std.name)' , scopes:'#(scope)'}
      When method get
      Then status 200
      * json json_res = karate.jsonPath(response, "$.metric_config")
      * header Authorization = 'Bearer '+access_token
      Given path 'metric' , json_res.Name
      * params {scope:'#(scope)'}
      When method delete
      Then status 200
      And response.success == true

    Scenario: To verify user can create attribute counter standard
      Given path 'metric/acs'
      * request data.crud_metric_attr_counter_std
      When method post
      Then status 200
      And response.success == true
      And match response.name == data.crud_metric_attr_counter_std.name
      # validate schema
      * def schema = read('schema_data.json')
      And match response == '#(schema.attribute_counter_metric)'

    Scenario Outline: To verify user can update attribute counter standard
      # get metric config of attribute counter standard
      Given path 'metric/config'
      * params {metric_info.type:'attribute.counter.standard' , metric_info.name:'#(data.crud_metric_attr_counter_std.name)' , scopes:'#(scope)'}
      When method get
      Then status 200
      * json json_res = karate.jsonPath(response, "$.metric_config")
      * url equipmentServiceUrl+'/api/v1'
      * header Authorization = 'Bearer '+access_token
      * def eq_data = callonce read('get_equipments_id.feature')
       #update attribute counter standard metric
      * url metricServiceUrl+'/api/v1'
      * header Authorization = 'Bearer '+access_token
      Given path '/metric/acs'
      * def inn_payload = {"ID": "", "name": '#(json_res.Name)', "eq_type":<rel_eq>, "attribute_name":<attr_val>, "value":<pass_val> , "scopes":['#(scope)']}
      And request inn_payload
      When method put
      Then status 200
      And response.success == true
      #  validate metric is updated or not
      Given path 'metric/config'
      * header Authorization = 'Bearer '+access_token
      * params {metric_info.type:'attribute.counter.standard' , metric_info.name:'#(data.crud_metric_attr_counter_std.name)', scopes:'#(scope)'}
      When method get
      Then status 200
    Examples:
      | rel_eq | attr_val | pass_val |
      | 'server' | '#(eq_data.serv_core.name)' | "1" |
      | 'server' | '#(eq_data.serv_processor.name)' | "4" |   
      | 'server' | '#(eq_data.server_sag.name)' | "3" |
      | 'server' | '#(eq_data.cpu_manufacture.name)' | "no" |   
      | 'server' | '#(eq_data.parent_id.name)' | "yes" |
      | 'cluster' | '#(eq_data.cluster_name.name)' | "no" |   
      | 'cluster' | '#(eq_data.cluster_parent_id.name)' | "yes" |
      | 'vcenter' | '#(eq_data.vcenter_name.name)' | "no" |
      | 'vcenter' | '#(eq_data.vcenter_version.name)' | "yes" |
      | 'softpartition' | '#(eq_data.parent_id.name)' | "no" |
      | 'softpartition' | '#(eq_data.softpartition_name.name)' | "yes" |
      | 'softpartition' | '#(eq_data.softpartition_id.name)' | "no" |  

  Scenario: To verify user can delete attribute counter standard metric
    Given path 'metric/config'
    * params {metric_info.type:'attribute.counter.standard' , metric_info.name:'#(data.crud_metric_attr_counter_std.name)', scopes:'#(scope)'}
      When method get
      Then status 200
    * json json_res = karate.jsonPath(response, "$.metric_config")
    * header Authorization = 'Bearer '+access_token
      Given path 'metric' , json_res.Name
    * params {scope:'#(scope)'}
      When method delete
      Then status 200
      And response.success == true
  
  