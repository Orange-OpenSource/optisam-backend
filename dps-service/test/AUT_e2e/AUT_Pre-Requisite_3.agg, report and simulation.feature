@ignore @aut-setup

Feature: Pre-Requisite Setup for AUT(Automation) - Aggregation, Report and Simulation

## Pre-requisite :  
# 1. equpiment type and metric is created
# 2. product and acquired rights Data is uploaded  

  Background:
    * url productServiceUrl+'/api/v1'
    #* def credentials = {username:'admin@test.com', password: 'Welcome@123'}
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def scope = "AUT"

 ## Aggregation
 Scenario: Create Aggregation - Oracle WebLogic, ops
    Given path 'aggregations'
    * def createAgg = {"ID":0, "name": "apitest_agg_oracleWL","editor": "Oracle", metric: "oracle.processor.standard", "products": ["oracle_wl_1","oracle_wl_2"],scope: '#(scope)'}
    And request createAgg
    When method post
    Then status 200
    And set createAgg.ID = response.ID
    * match response == createAgg

 Scenario: Create Aggregation - Oracle MySQL, nup
    Given path 'aggregations'
    * def createAgg = {"ID":0, "name": "apitest_agg_oracleSQL","editor": "Oracle", metric: "oracle.nup.standard", "products": ["oracle_mysql_2","oracle_mysql_3"],scope: '#(scope)'}
    And request createAgg
    When method post
    Then status 200
    And set createAgg.ID = response.ID
    * match response == createAgg


  ## Report
  Scenario: Create the Compliance type report
    * url reportServiceUrl+'/api/v1'
    Given path 'reports'
    And request {"scope": '#(scope)',"report_type_id": 1,"acqrights_report": {"editor": "Oracle","swidtag": ["oracle_wl_1"]}}
    When method post
    Then status 200

  @create
  Scenario: Create the ProductEquipments type report
    * url reportServiceUrl+'/api/v1'
    Given path 'reports'
    And request {"scope": '#(scope)',"report_type_id": 2,"product_equipments_report": {"editor": "Oracle","swidtag": ["oracle_wl_1"], "equipType": "server"}}
    When method post
    Then status 200

  ## Simulation Configuration

  @create
  Scenario: Create the Simulation Configuration
    * url importServiceUrl+'/api/v1'
    Given path 'config'
    * def file1_tmp = karate.readAsString('sim-server-config-cpu.csv')
    * multipart file server_cpu = { value: '#(file1_tmp)', filename: 'sim-server-config-cpu.csv', contentType: "text/csv" }
    * multipart field scopes = [scope]
    * multipart field config_name = "apitest_sim_server_cpu"
    * multipart field equipment_type = "server"
    When method post
    Then status 200
