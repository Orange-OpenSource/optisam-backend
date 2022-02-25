@metric
Feature: Metric Service Test - Get metrics and configurations : Normal user

  Background:
  # * def metricServiceUrl = "https://optisam-metric-int.kermit-noprod-b.itn.intraorange"
    * url metricServiceUrl+'/api/v1'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'

  @schema
  Scenario: Validate Schema for get metrics list
    Given path 'metrics'
    * params {scopes:'#(scope)'}
    * def schema = {type: '#string', name: '#string', description: '##string'}
    When method get
    Then status 200
    #* response.totalRecords == '#number? _ >= 0'
    * match response.metrices == '#[] schema'

  @get
  Scenario: Get all metric types
    Given path 'metric/types'
    * params {scopes:'#(scope)'}
    When method get
    Then status 200
  * match response.types[*].name contains ['oracle.processor.standard']
  * match response.types[*].name contains ['oracle.nup.standard']
  * match response.types[*].name contains ['sag.processor.standard']
  * match response.types[*].name contains ['attribute.counter.standard']
  * match response.types[*].name contains ['ibm.pvu.standard']
  * match response.types[*].name contains ['instance.number.standard']


    @get
  Scenario: Get Metric configuration
    Given path 'metric/config'
    * params {metric_info.type:'oracle.processor.standard' , metric_info.name:'oracle.processor.standard' , scopes:'#(scope)'}
    When method get
    Then status 200
  * response.metric_config.Name == 'oracle_processor'


  @get
  Scenario Outline: Get metric configuration for metric <type>
    Given path 'metric/config'
    * params {metric_info.type:<type>, metric_info.name:<name>,scopes:'#(scope)'}
    When method get
    Then status 200
    * match response.metric_config == '#notnull'
    #  TODO : update api repsonse 
    # * match response.metric_config.Name == <name>
  Examples:
    | type | name |
    | 'oracle.processor.standard' | 'oracle.processor.standard' |
    | 'oracle.nup.standard' | 'oracle.nup.standard' |
    # | 'sag.processor.standard' | 'sag' |
    # | 'ibm.pvu.standard' | 'ibm_pvu' |
    # | 'instance.number.standard' | 'os_instance' |
    # | 'attribute.counter.standard' | 'attribute_counter_core' |

