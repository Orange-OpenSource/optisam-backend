Feature: Equipment Management test

Background:
  * url equipmentServiceUrl+'/api/v1/equipment'
  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
  * callonce read('common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'

@SmokeTest
Scenario:To get the detials of Equipment Types
    Given path 'types'
    And params { scopes:'#(scope)'}
    When method get 
    Then status 200
  #* match response.equipment_types[*] contains data.Equipment_Type_Detail

Scenario: To get the attributes of a softpartition
  Given path 'metadata',data.metaData.ID
  And params {scopes:'#(scope)'}
  When method get 
  Then status 200
  * match response.attributes[*] contains data.metaData.attributes
 

Scenario Outline: To modify the visibility and serchability of multiple attributed  of softpartition
  Given path 'types',data.equipmentID.softpartition_id
  * set data.Update_Eqp.updattr[*].ID = <id>
  * set data.Update_Eqp.updattr[*].name = <name>
  * set data.Update_Eqp.updattr[*].schema_name = <schema>
  * set data.Update_Eqp.updattr[*].displayed = <displayed>
  * set data.Update_Eqp.updattr[*].searchable = <searchable>
  And request data.Update_Eqp
  When method patch
  Then status 200
  

Examples:    
|    id    |      name          |      schema        |displayed|searchable|
|"0x289fae"|   "environment"    |   "environment"    | false   |  false   |
|"0x289fad"|"softpartition_name"|"Softpartition Name"|  true   |  true   |
|"0x289faf"|     "vcpu"         |    "vcpu"          |  true |  true    |
|"0x289fac"| "softpartition_id" | "Softpartition ID" |  true  |  true  |
# can also change the name of table headersfor this change the schema name 


Scenario Outline: To modify the visibility and serchability of multiple attributed  of softpartition --Negative
  Given path 'types',data.equipmentID.softpartition_id
  * set data.Update_Eqp.updattr[*].ID = <id>
  * set data.Update_Eqp.updattr[*].name = <name>
  * set data.Update_Eqp.updattr[*].schema_name = <schema>
  * set data.Update_Eqp.updattr[*].displayed = <displayed>
  * set data.Update_Eqp.updattr[*].searchable = <searchable>
  And request data.Update_Eqp
  When method patch
  Then status 400
  
Examples:    
|    id    |      name          |      schema        |displayed|searchable|
|"0x289fae"|   "environment"    |   "environment"    | false   |  true   |
|"0x289fad"|"softpartition name"|"Softpartition Name"|  true   |  true   |
|0x289faf|     "vcpu"         |    "vcpu"          |  true |  true    |
|"0x289fac"| "softpartition_id" | "Softpartition ID" |  "true"  |  true|


Scenario: To get the attributes of server
  Given path 'metadata',data.metaData_server.ID
  And params {scopes:'#(scope)'}
  When method get 
  Then status 200
 * match response.attributes[*] contains data.metaData_server.attributes


Scenario Outline: To modify the visibility and serchability of multiple attributed  of Server
  Given path 'types', data.equipmentID.server_id
  * set data.Update_Server.updattr[*].ID = <id>
  * set data.Update_Server.updattr[*].name = <name>
  * set data.Update_Server.updattr[*].schema_name = <schema>
  * set data.Update_Server.updattr[*].displayed = <displayed>
  * set data.Update_Server.updattr[*].searchable = <searchable>
  And request data.Update_Server
  When method patch
  Then status 200
  
Examples:    
|    id    |      name          |      schema        |displayed|searchable|
|"0x2eaf1e"|   "datacenter_name"|   "datacenter_name"| false   |  false   |
|"0x2eaf19"|"cpu_manufacturer"|"cpu_manufacturer"|  true   |  true   |
|"0x2eaf1c"|     "server_type"         |    "server_type" |  true |  true    |
|"0x2eaf1d"| "cpu_model" | "cpu_model" | true |  false  |
|"0x2eaf24"| "oracle_core_factor" | "oracle_core_factor" |  true  |  true   |
|"0x2eaf21"| "server_processors_numbers" | "server_processors_numbers" |  true  |  true   |
|"0x2eaf25"| "sag_uvu" | "sag_uvu" |  true  |  false   |
|"0x2eaf22"| "environment" | "environment" |  true  |  true   |
|"0x2eaf20"| "ibm_pvu" | "ibm_pvu" |  true  |  true   |
|"0x2eaf1f"| 	"server_id"|	"server_id"|true|true|

# can also change the name of table headers for this change the schema name

Scenario Outline: To modify the visibility and serchability of multiple attributed  of Server --Negative
  Given path 'types', data.Update_Eqp.parent_id
  * set data.Update_Server.updattr[*].ID = <id>
  * set data.Update_Server.updattr[*].name = <name>
  * set data.Update_Server.updattr[*].schema_name = <schema>
  * set data.Update_Server.updattr[*].searchable = <searchable>
  And request data.Update_Server
  When method patch
  Then status 400
  
Examples:    
|    id    |      name          |      schema       |displayed | searchable|
|0x289f9a|   "datacenter_name"|   "datacenter_name"| false   |  true   |
|"0x289f9c"|     "server type"    |    "server type"   |  0|  1   |

#showing status 200 for invalid Input

Scenario: To get the attributes of Cluster
  Given path 'metadata',data.metaData_Cluster.ID
  And params {scopes:'#(scope)'}
  When method get 
  Then status 200
 * match response.attributes[*] contains data.metaData_Cluster.attributes

Scenario Outline: To modify the visibility and serchability of multiple attributed  of Cluster
  Given path 'types',data.equipmentID.cluster_id
  * set data.Update_Cluster.updattr[*].displayed = <display>
  * set data.Update_Cluster.updattr[*].searchable = <search>
  * set data.Update_Cluster.updattr[*].schema_name = <schema>
 And request data.Update_Cluster
  When method patch
  Then status 200

  Examples:
  |display|search|schema|
  |true|true|"cluster_name"|

  # can also change the name of table headers for this change the schema name

Scenario Outline: To modify the visibility and serchability of multiple attributed  of Cluster-Negative
  Given path 'types',data.Update_Server.parent_id
  * set data.Update_Cluster.updattr[*].displayed = <display>
  * set data.Update_Cluster.updattr[*].searchable = <search>
  * set data.Update_Cluster.updattr[*].schema_name = <schema>
 And request data.Update_Cluster
  When method patch
  Then status 400

  Examples:
  |display|search|schema|
  |false|true|"cluster_name"|
  |"true"|true|"cluster_name"|


Scenario: To get the atrributes of Vcenter
  Given path 'metadata',data.metadata_Vcenter.ID
  And params { scopes:'#(scope)'}
  When method get 
  Then status 200 
 * match response.attributes[*] contains data.metadata_Vcenter.attributes

Scenario Outline: To modify the visibility and serchability of multiple attributed  of Vcenter
  Given path 'types',data.equipmentID.vcenter_id
  * set data.Update_Vcenter.updattr[*].ID = <id>
  * set data.Update_Vcenter.updattr[*].displayed = <display>
  * set data.Update_Vcenter.updattr[*].searchable = <search>
  * set data.Update_Vcenter.updattr[*].schema_name = <schema>
 And request data.Update_Vcenter
  When method patch
  Then status 200 

  Examples:
  |id|display|search|schema|
  |"0x289f94"|true|false|"vcenter_version"|
  |"0x289f93"|true|true|"vcenter_name"|
   # can also change the name of table headers for this change the schema names


Scenario Outline: To modify the visibility and serchability of multiple attributed  of Vcenter
  Given path 'types',data.equipmentID.vcenter_id
  * set data.Update_Vcenter.updattr[*].ID = <id>
  * set data.Update_Vcenter.updattr[*].displayed = <display>
  * set data.Update_Vcenter.updattr[*].searchable = <search>
  * set data.Update_Vcenter.updattr[*].schema_name = <schema>
 And request data.Update_Vcenter
  When method patch
  Then status 400

  Examples:
  |id|display|search|schema|
  |0x289f94|true|false|"vcenter_version"|
  |"0x289f93"|true|"true"|"vcenter_name"|
  

  