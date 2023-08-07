@aggregation
Feature: Create Aggregation : admin 

  Background:
  # * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * url productServiceUrl+'/api/v1/product'
   
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'


  @createDelete
  Scenario: To verify admin can create new Aggregation and delete it -Adobe
    Given path 'aggregations'
    * set data.createAgg.scope = scope
    And request data.createAgg
    When method post
    Then status 200
  * header Authorization = 'Bearer '+access_token
    Given path 'aggregations'
    And params {scope:'#(scope)'}
    And params {page_size:50, page_num:1, sort_by:'aggregation_name', sort_order:'asc'}
    When method get
    Then status 200
  * def agg_id = karate.jsonPath(response.aggregations,"$.[?(@.aggregation_name=='"+data.getProdAgg.aggregation_name+"')].ID")[0]
    * print 'aggregation_id:' + agg_id
    And set data.createAgg.ID = agg_id
    Given path 'aggregations' , data.createAgg.ID
    * header Authorization = 'Bearer '+access_token
    And  params {scope:'#(scope)'}
    When method delete
    Then status 200
    * match response.success == true

  @create
  Scenario: To verify admin can create new Aggregation
    Given path 'aggregations'
    * set data.createAgg.scope = scope
    And request data.createAgg
    When method post
    Then status 200

  
  @verfiy
  Scenario: To verify aggregation name is unique
  Given path 'aggregations'
  * set data.createAgg.scope = scope
  And request data.createAgg
  When method post
  Then status 400
  * match response.message == data.Response_Message.message

@verify
Scenario: To verify aggregation created for one product than should not allowed to create again
  Given path 'aggregations'
  * set data.createAgg.aggregation_name = data.Aggregation_name.agg_name
  * set data.createAgg.product_editor = data.getAgg.product_editor 
  * set data.createAgg.product_names = data.getAgg.product_names
  * set data.createAgg.scope = scope
  And request data.createAgg
  When method post
  Then status 400
  * match response.message == data.Response_Message.prd_message


  @UpdateAgg
  Scenario: To verify update aggregation record using PUT call
    Given path 'aggregations'
    And params {scope:'#(scope)'}
    And params {page_size:50, page_num:1, sort_by:'aggregation_name', sort_order:'asc'}
    When method get
    Then status 200
  * def agg_id = karate.jsonPath(response.aggregations,"$.[?(@.aggregation_name=='"+data.getProdAgg.aggregation_name+"')].ID")[0]
    * print 'aggregation_id:' + agg_id
    And set data.createAgg.ID = agg_id
    Given path 'aggregations' , data.createAgg.ID
    * header Authorization = 'Bearer '+access_token
    * set data.updateAgg.scope = scope
    And request data.updateAgg
    When method PUT
    Then status 200
    * match response.success == true

  @UpdateAgg
  Scenario: To verify Update Aggregation record using PATCH call
    Given path 'aggregations'
    And params {scope:'#(scope)'}
    And params {page_size:50, page_num:1, sort_by:'aggregation_name', sort_order:'asc'}
    When method get
    Then status 200
 * def agg_id = karate.jsonPath(response.aggregations,"$.[?(@.aggregation_name=='"+data.updateAgg.aggregation_name+"')].ID")[0]
    And set data.createAgg.ID = agg_id
    Given path 'aggregations' , data.createAgg.ID
    * header Authorization = 'Bearer '+access_token
    * set data.updateAggPATCH.scope = scope
    And request data.updateAggPATCH
    When method PATCH
    Then status 200
    * match response.success == true


  @Delete
  Scenario: To verify get the aggregation and delete the created record
    Given path 'aggregations'
    And params {scope:'#(scope)'}
    And params {page_size:50, page_num:1, sort_by:'aggregation_name', sort_order:'asc'}
     When method get
     Then status 200
    * def agg_id = karate.jsonPath(response.aggregations,"$.[?(@.aggregation_name=='"+data.updateAggPATCH.aggregation_name+"')].ID")[0]
    And set data.createAgg.ID = agg_id
    Given path 'aggregations' , data.createAgg.ID
    * header Authorization = 'Bearer '+access_token
    And  params {scope:'#(scope)'}
    When method delete
    Then status 200
    * match response.success == true



