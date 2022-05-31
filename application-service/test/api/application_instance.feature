@application
Feature: Application Service Test for Instances

  Background:
  # * def applicationServiceUrl = "https://optisam-application-int.apps.fr01.paas.tech.orange"
    * url applicationServiceUrl+'/api/v1/application'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'    

  @schema
  Scenario: Schema validation for get Instances
    Given path 'instances'
    * params { page_num:1, page_size:10, sort_by:'instance_id', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    * response.totalRecords == '#number? _ >= 0'
    * match response.instances == '#[] data.schema_instance'

  @search
  Scenario:   
    Given path 'instances'
    And params { page_num:1, page_size:10, sort_by:'instance_environment', sort_order:'desc', scopes:'#(scope)'}
    And params {search_params.application_id.filteringkey: '#(data.getInstance.application_id)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    # And match each response.instances[*].name == data.getInstance.name
    * remove data.getInstance.application_id
    And match response.instances contains data.getInstance

  # @search
  # Scenario: Searching_Filter Instances by product Id
  #   Given path 'instances'
  #   And params { page_num:1, page_size:10, sort_by:'instance_environment', sort_order:'desc', scopes:'#(scope)'}
  #   And params {search_params.product_id.filteringkey: '#(data.getInstance.products[0])'}
  #   When method get
  #   Then status 200
  #   And response.totalRecords > 0
  #   * remove data.getInstance.application_id
  #   And match  response.instances contains data.getInstance

