Feature: To get equipments id: admin user

Scenario: To get all the details of equipment types
    # fetch equipment types
    * url equipmentServiceUrl+'/api/v1'

    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token

    * header Authorization = 'Bearer '+access_token
    Given path 'equipment/types'
    #And params {scopes:'#(scope)'}
   And params {scopes:DEM}
    When method get
    Then status 200
    * def server_eq_type = karate.jsonPath(response, "$.equipment_types[?(@.type=='server')]")[0]
    * def partition_eq_type = karate.jsonPath(response, "$.equipment_types[?(@.type=='softpartition')]")[0]
    * def cluster_eq_type = karate.jsonPath(response, "$.equipment_types[?(@.type=='cluster')]")[0]
    * def vcenter_eq_type = karate.jsonPath(response, "$.equipment_types[?(@.type=='vcenter')]")[0]
    * def serv_core = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='cores_per_processor')]")[0]
    * def serv_core1 = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='cores_per_processor')]")[0]
    #* def serv_core = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='cores per processor')]")[0]
    * def serv_processor = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='server_processors_numbers')]")[0]
    * def server_oracle = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='oracle_core_factor')]")[0]
    * def server_oracle1 = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='oracle_core_factor')]")[0]   
    #* def server_oracle = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='oracle core factor')]")[0]   
    * def server_pvu = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='ibm_pvu')]")[0]
    * def server_sag = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='sag_uvu')]")[0]
    * def cpu_manufacture = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='cpu_manufacturer')]")[0]
    * def parent_id = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='parent_id')]")[0]
    * def datacenter_name = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='datacenter_name')]")[0]
    * def hyperthreading = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='hyperthreading')]")[0]
    * def server_id = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='server_id')]")[0]
    * def server_type = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='server_type')]")[0]
    * def cpu_model = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='cpu_model')]")[0]
    * def server_os = karate.jsonPath(server_eq_type, "$.attributes[?(@.name=='server_os')]")[0]
    * def cluster_name = karate.jsonPath(cluster_eq_type, "$.attributes[?(@.name=='cluster_name')]")[0]
    * def cluster_parent_id = karate.jsonPath(cluster_eq_type, "$.attributes[?(@.name=='parent_id')]")[0]
    * def vcenter_name = karate.jsonPath(vcenter_eq_type, "$.attributes[?(@.name=='vcenter_name')]")[0]
    * def vcenter_version = karate.jsonPath(vcenter_eq_type, "$.attributes[?(@.name=='vcenter_version')]")[0]
    * def softpartition_id = karate.jsonPath(partition_eq_type, "$.attributes[?(@.name=='softpartition_id')]")[0]
    * def softpartition_name = karate.jsonPath(partition_eq_type, "$.attributes[?(@.name=='softpartition_name')]")[0]
    * def sp_parent_id = karate.jsonPath(partition_eq_type, "$.attributes[?(@.name=='parent_id')]")[0]
    * def softpartition_vcpu = karate.jsonPath(partition_eq_type, "$.attributes[?(@.name=='vcpu')]")[0]
    * def environment = karate.jsonPath(partition_eq_type, "$.attributes[?(@.name=='environment')]")[0]

