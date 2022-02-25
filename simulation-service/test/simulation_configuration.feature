@simulation @ignore
Feature: Simulation Service Test for Configuration : Admin

  Background:
    # * def simulationServiceUrl = "https://optisam-simulation-int.kermit-noprod-b.itn.intraorange"
    * url simulationServiceUrl+'/api/v1/simulation'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')


  @schema
  Scenario: Schema Validation for simulation Configuration and Simulation Metadata Configuration
    Given path 'config'
    * params { equipment_type:"server"}
    When method get
    Then status 200
    * match response.configurations == '#[] data.configurations'
    * def config_id = response.configurations[0].config_id
    * def simulation_id = response.configurations[0].config_attributes[0].attribute_id
    Given path 'config',config_id,simulation_id
    * header Authorization = 'Bearer '+access_token
    When method get
    Then status 200
    # * def schema = {"config_id": '#number',"config_name": '#string',"equipment_type": '#string',"created_by": '#string',"created_on": '#string',"config_attributes": '#[]'}
    # TODO : decode base64 data and validate
    * match response.data == '#string'


  @create
  Scenario: Create the Simulation Configuration for Server and delete it
    * url importServiceUrl+'/api/v1'
    Given path 'config'
    * def file1_tmp = karate.readAsString('sim-server-config-manuf.csv')
    * multipart file server_manufacturer = { value: '#(file1_tmp)', filename: 'sim-server-config-manuf.csv', contentType: "text/csv" }
    * multipart field scopes = [scope]
    * multipart field config_name = "apitest_sim_server_manuf"
    * multipart field equipment_type = "server"
    When method post
    Then status 200
    * url simulationServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'config'
    * params { equipment_type:"server"}
    When method get
    Then status 200
    # * match response.configurations[*].config_id contains data.simulationconfig.config_id
    * def config_id = karate.jsonPath(response.configurations,"$.[?(@.config_name=='apitest_sim_server_manuf')].config_id")[0]  
    * url simulationServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'config',config_id
    When method delete
    Then status 200
    * url simulationServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'config'
    * params { equipment_type:"server"}
    When method get
    Then status 200
    * match response.configurations[*].config_id != config_id
