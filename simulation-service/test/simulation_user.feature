@simulation
Feature: Simulation Service Test for Normal User

  Background:
  # * def simulationServiceUrl = "https://optisam-simulation-int.apps.fr01.paas.tech.orange"
    * url simulationServiceUrl+'/api/v1'
    #* def credentials = {username:'testuser@test.com', password: 'password'}
    * def credentials = {username:#(UserAccount_Username), password:#(UserAccount_password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')

  @metricsimulation
  Scenario: verify Metric Simulation
    Given path 'simulation/metric'
    And request data.metricsimulation.request
    When method post
    Then status 200
  * match response.metric_sim_result[*] contains data.metric_sim_result


  @hardwaresimulation
  Scenario: verify Hardware Simulation
    Given path 'simulation/hardware'
    And request data.hardwaresimulation
    When method post
    Then status 200


  # Simulation Confiugration
  @schema
  Scenario: Schema Validation for simulation Configuration and Simulation Metadata Configuration
    Given path 'simulation/config'
    * params { equipment_type:"server"}
    When method get
    Then status 200
    * match response.configurations == '#[] data.configurations'
    * def config_id = response.configurations[0].config_id
    * def simulation_id = response.configurations[0].config_attributes[0].attribute_id
    Given path 'simulation/config',config_id,simulation_id
    * header Authorization = 'Bearer '+access_token
    When method get
    Then status 200
    # * def schema = {"config_id": '#number',"config_name": '#string',"equipment_type": '#string',"created_by": '#string',"created_on": '#string',"config_attributes": '#[]'}
    # TODO : decode base64 data and validate
    * match response.data == '#string'


  @create
  Scenario: Normal user can not create simulation configuration
    * url importServiceUrl+'/api/v1'
    Given path 'simulation/config'
    * def file1_tmp = karate.readAsString('sim-server-config-manuf.csv')
    * multipart file server_manufacturer = { value: '#(file1_tmp)', filename: 'sim-server-config-manuf.csv', contentType: "text/csv" }
    * multipart field scopes = [scope]
    * multipart field config_name = "apitest_sim_server_manuf_user"
    * multipart field equipment_type = "server"
    When method post
    Then status 403

  #  TODO : create new simulation with admin and delete with normal user
  # @delete 
  # Scenario: Normal user can not delete simulation configuration
  #   Given path 'config/2'
  #   When method delete
  #   Then status 502
  #   Given path 'config'
  #   * header Authorization = 'Bearer '+access_token
  #   * params { equipment_type:"server"}
  #   When method get
  #   Then status 200
  #   * match response.configurations[*].config_id contains 2
