@simulation
Feature: Simulation Service Test : Admin

  Background:
  # * def simulationServiceUrl = "https://optisam-simulation-int.apps.fr01.paas.tech.orange"
    * url simulationServiceUrl+'/api/v1'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')

#  TODO: add simulation for all metrics that exist
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

