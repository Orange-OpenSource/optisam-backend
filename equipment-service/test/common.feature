@ignore
Feature: Common utilities and authentication

  Background: pre-requisite
    # * def authServiceUrl = "https://optisam-auth-int.kermit-noprod-b.itn.intraorange"
    * url authServiceUrl+'/api/v1'
    # Common configurations
    * karate.configure('logPrettyRequest', true)
    * karate.configure('logPrettyResponse', true)
    * karate.configure('ssl', true)
    ## Utilities Functions
    * def now = function(){ return java.lang.System.currentTimeMillis() }
    * def uuid = function(){ return java.util.UUID.randomUUID() + '' } 
    * def replace = function(str, old_val, new_val){ return str.replace(old_val,new_val) } 
    * def pause = function(pause){ java.lang.Thread.sleep(pause) }
    * def sort = 
      """
      function(actual, order) {
        var ArrayList = Java.type('java.util.ArrayList')
        var Collections = Java.type('java.util.Collections')
        var list = new ArrayList();
        for (var i = 0; i < actual.length; i++) {
          list.add(actual[i]);
        }
        if (order=='asc') {
          Collections.sort(list, java.lang.String.CASE_INSENSITIVE_ORDER)
        } else if (order=='desc') {
          Collections.sort(list, Collections.reverseOrder(java.lang.String.CASE_INSENSITIVE_ORDER))
        }
        return list;
      }
      """
    * def sortNumber = 
      """
      function(actual, order) {
        var ArrayList = Java.type('java.util.ArrayList')
        var Collections = Java.type('java.util.Collections')
        var list = new ArrayList();
        for (var i = 0; i < actual.length; i++) {
          list.add(actual[i]);
        }
        if (order=='asc') {
          Collections.sort(list)
        } else if (order=='desc') {
          Collections.sort(list, Collections.reverseOrder())
        }
        return list;
      }
      """
        
  @ignore
  Scenario: Verify equipment service is up and running
    * def equipmentServiceInstUrl = replace(karate.get('equipmentServiceUrl'),'equipment','equipment-inst')
    * url equipmentServiceInstUrl
    Given path 'healthz'
    * configure retry = { count: 10, interval: 10000 }
    * retry until responseStatus == 200
    When method get
    Then status 200
    And match response.status == 'ok'

  
  @ignore
  Scenario: Get auth token
    Given path 'token'
    * form field grant_type = 'password'
    * form fields credentials
    * configure retry = { count: 10, interval: 10000 }
    When method post
    Then status 200
    And match response.token_type == 'Bearer'
