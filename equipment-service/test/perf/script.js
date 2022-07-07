/*
 * equipment.proto
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * OpenAPI spec version: version not set
 *
 * NOTE: This class is auto generated by OpenAPI Generator.
 * https://github.com/OpenAPITools/openapi-generator
 *
 * OpenAPI generator version: 5.0.1-SNAPSHOT
 */


import http from "k6/http";
import { group, check, sleep } from "k6";

const BASE_URL = "https://optisam-equipment-int.apps.fr01.paas.tech.orange";
// Sleep duration between successive requests.
// You might want to edit the value of this variable or remove calls to the sleep function on the script.
const SLEEP_DURATION = 0.1;
// Global variables should be initialized.

export let options = {

    insecureSkipTLSVerify: true,
    httpDebug: 'full',
    vus: 2,
    iterations: 5,
    thresholds: {
        http_req_duration: ['p(75)<1000'], 
    },
 
};

export function setup() {
    let loginRes = http.post(`https://optisam-auth-int.apps.fr01.paas.tech.orange/api/v1/token`, {
        username: "admin@test.com",
        password: "admin",
        grant_type: "password"
      });
    let authToken = loginRes.json('access_token');
    check(authToken, { 'logged in successfully': () => authToken !== '' });
    return authToken;
}

export default function(authToken) {

    console.log("authToken"+authToken)
    let headers = {
        'Authorization': `Bearer ${authToken}`}
    console.log("headers"+headers)

    group("/api/v1/dashboard/types/equipments", () => {
        let scope = "OFR";
        let url = BASE_URL + `/api/v1/dashboard/types/equipments?scope=${scope}`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);
    });
//     group("/api/v1/equipments", () => {
//         let url = BASE_URL + `/api/v1/equipments`;
//         // Request No. 1
//         // TODO: edit the parameters of the request body.
//         let body = {"scope": "string", "eqType": "string", "eqData": {"fields": "map"}};
//         let params = {headers: {"Content-Type": "application/json", "Accept": "application/json"}};
//         let request = http.post(url, body, params);
//         check(request, {
//             "A successful response.": (r) => r.status === 200
//         });
//         sleep(SLEEP_DURATION);
//     });
    group("/api/v1/equipments/metadata", () => {
        let scopes = "OFR";
        let type = "TODO_EDIT_THE_TYPE";
        let url = BASE_URL + `/api/v1/equipments/metadata?&scopes=${scopes}`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);

//         // Request No. 2
//         // TODO: edit the parameters of the request body.
//         body = {"metadataType": "string", "metadataSource": "string", "metadataAttributes": "list", "scope": "string"};
//         params = {headers: {"Content-Type": "application/json"}};
//         request = http.post(url, body, params);
//         check(request, {
//             "A successful response.": (r) => r.status === 200
//         });
//         sleep(SLEEP_DURATION);
    });
    group("/api/v1/equipments/metadata/{ID}", () => {
        let attributes = "TODO_EDIT_THE_ATTRIBUTES";
        let ID = "0x4e09c";
        let scopes = "FST";
        let url = BASE_URL + `/api/v1/equipments/metadata/${ID}?scopes=${scopes}`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);
    });
    group("/api/v1/equipments/types", () => {
        let scopes = "OFR";
        let url = BASE_URL + `/api/v1/equipments/types?scopes=${scopes}`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);

//         // Request No. 2
//         // TODO: edit the parameters of the request body.
//         body = {"iD": "string", "type": "string", "parentId": "string", "parentType": "string", "metadataId": "string", "metadataSource": "string", "attributes": [{"ID": "string", "name": "string", "dataType": "v1datatypes", "primaryKey": "boolean", "displayed": "boolean", "searchable": "boolean", "parentIdentifier": "boolean", "mappedTo": "string", "simulated": "boolean", "intVal": "integer", "floatVal": "float", "stringVal": "string", "intValOld": "integer", "floatValOld": "float", "stringValOld": "string"}], "scopes": "list"};
//         params = {headers: {"Content-Type": "application/json"}};
//         request = http.post(url, body, params);
//         check(request, {
//             "A successful response.": (r) => r.status === 200
//         });
//         sleep(SLEEP_DURATION);
     });
//     group("/api/v1/equipments/types/{id}", () => {
//         let id = "TODO_EDIT_THE_ID";
//         let url = BASE_URL + `/api/v1/equipments/types/${id}`;
//         // Request No. 1
//         // TODO: edit the parameters of the request body.
//         let body = {"id": "string", "parentId": "string", "attributes": [{"ID": "string", "name": "string", "dataType": "v1datatypes", "primaryKey": "boolean", "displayed": "boolean", "searchable": "boolean", "parentIdentifier": "boolean", "mappedTo": "string", "simulated": "boolean", "intVal": "integer", "floatVal": "float", "stringVal": "string", "intValOld": "integer", "floatValOld": "float", "stringValOld": "string"}], "scopes": "list"};
//         let params = {headers: {"Content-Type": "application/json", "Accept": "application/json"}};
//         let request = http.put(url, body, params);
//         check(request, {
//             "A successful response.": (r) => r.status === 200
//         });
//         sleep(SLEEP_DURATION);

//         // Request No. 2
//         // TODO: edit the parameters of the request body.
//         body = {"id": "string", "parentId": "string", "attributes": [{"ID": "string", "name": "string", "dataType": "v1datatypes", "primaryKey": "boolean", "displayed": "boolean", "searchable": "boolean", "parentIdentifier": "boolean", "mappedTo": "string", "simulated": "boolean", "intVal": "integer", "floatVal": "float", "stringVal": "string", "intValOld": "integer", "floatValOld": "float", "stringValOld": "string"}], "scopes": "list"};
//         params = {headers: {"Content-Type": "application/json"}};
//         request = http.patch(url, body, params);
//         check(request, {
//             "A successful response.": (r) => r.status === 200
//         });
//         sleep(SLEEP_DURATION);
//     });
    group("/api/v1/equipments/{type_id}", () => {
        let filterInstanceIdFilteringkeyMultiple = "TODO_EDIT_THE_FILTER.INSTANCE_ID.FILTERINGKEY_MULTIPLE";
        let pageSize = 10;
        let filterInstanceIdFilteringOrder = "TODO_EDIT_THE_FILTER.INSTANCE_ID.FILTERINGORDER";
        let pageNum = 1;
        let filterProductIdFilteringkey = "TODO_EDIT_THE_FILTER.PRODUCT_ID.FILTERINGKEY";
        let searchParams = "TODO_EDIT_THE_SEARCH_PARAMS";
        let filterProductIdFilteringOrder = "TODO_EDIT_THE_FILTER.PRODUCT_ID.FILTERINGORDER";
        let filterProductIdFilteringkeyMultiple = "TODO_EDIT_THE_FILTER.PRODUCT_ID.FILTERINGKEY_MULTIPLE";
        let filterProductIdFilterType = "TODO_EDIT_THE_FILTER.PRODUCT_ID.FILTER_TYPE";
        let sortOrder = "ASC";
        let filterApplicationIdFilteringkeyMultiple = "TODO_EDIT_THE_FILTER.APPLICATION_ID.FILTERINGKEY_MULTIPLE";
        let typeId = "0x51";
        let sortBy = "datacenter_name";
        let filterApplicationIdFilterType = "TODO_EDIT_THE_FILTER.APPLICATION_ID.FILTER_TYPE";
        let filterApplicationIdFilteringkey = "TODO_EDIT_THE_FILTER.APPLICATION_ID.FILTERINGKEY";
        let scopes = "OFR";
        let filterApplicationIdFilteringOrder = "TODO_EDIT_THE_FILTER.APPLICATION_ID.FILTERINGORDER";
        let filterInstanceIdFilterType = "TODO_EDIT_THE_FILTER.INSTANCE_ID.FILTER_TYPE";
        let filterInstanceIdFilteringkey = "TODO_EDIT_THE_FILTER.INSTANCE_ID.FILTERINGKEY";
        let url = BASE_URL + `/api/v1/equipments/${typeId}?page_num=${pageNum}&page_size=${pageSize}&sort_by=${sortBy}&sort_order=${sortOrder}&scopes=${scopes}`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);
    });
    group("/api/v1/equipments/{type_id}/{equip_id}", () => {
        let equipId = "DT000001";
        let typeId = "0x51";
        let scopes = "OFR";
        let url = BASE_URL + `/api/v1/equipments/${typeId}/${equipId}?scopes=${scopes}`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);
    });
    group("/api/v1/equipments/{type_id}/{equip_id}/childs/{children_type_id}", () => {
        let equipId = "0x30fcf";
        let sortOrder = "ASC";
        let pageSize = 10;
        let typeId = "0x51";
        let sortBy = "vcenter_name";
        let scopes = "OFR";
        let pageNum = "1";
        let childrenTypeId = "0x274";
     //   let searchParams = "TODO_EDIT_THE_SEARCH_PARAMS";
        let url = BASE_URL + `/api/v1/equipments/${typeId}/${equipId}/childs/${childrenTypeId}?page_num=${pageNum}&page_size=${pageSize}&sort_by=${sortBy}&sort_order=${sortOrder}&scopes=${scopes}`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);
    });
    group("/api/v1/equipments/{type_id}/{equip_id}/parents", () => {
        let equipId = "0x30fd0";
        let typeId = "0x863";
        let scopes = "OFR";
        let url = BASE_URL + `/api/v1/equipments/${typeId}/${equipId}/parents?scopes=${scopes}`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);
    });
    group("/api/v1/products/aggregations/{name}/equipments/{eq_type_id}", () => {
        let sortOrder = "asc";
        let name = "Oracle_Database";
        let pageSize = 10;
        let eqTypeId = "0x51";
        let sortBy = "datacenter_name";
        let scopes = "OFR";
        let pageNum = 1;
        let searchParams = "";
        let url = BASE_URL + `/api/v1/products/aggregations/${name}/equipments/${eqTypeId}?page_num=${pageNum}&page_size=${pageSize}&sort_by=${sortBy}&sort_order=${sortOrder}&search_params=${searchParams}&scopes=${scopes}`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);
    });
//     group("/api/v1/products/{swid_tag}/equipments/{eq_type_id}", () => {
//         let sortOrder = "TODO_EDIT_THE_SORT_ORDER";
//         let swidTag = "TODO_EDIT_THE_SWID_TAG";
//         let pageSize = "TODO_EDIT_THE_PAGE_SIZE";
//         let eqTypeId = "TODO_EDIT_THE_EQ_TYPE_ID";
//         let sortBy = "TODO_EDIT_THE_SORT_BY";
//         let scopes = "TODO_EDIT_THE_SCOPES";
//         let pageNum = "TODO_EDIT_THE_PAGE_NUM";
//         let searchParams = "TODO_EDIT_THE_SEARCH_PARAMS";
//         let url = BASE_URL + `/api/v1/products/${swid_tag}/equipments/${eq_type_id}?page_num=${page_num}&page_size=${page_size}&sort_by=${sort_by}&sort_order=${sort_order}&search_params=${search_params}&scopes=${scopes}`;
//         // Request No. 1
//         let request = http.get(url,{headers:headers});
//         check(request, {
//             "A successful response.": (r) => r.status === 200
//         });
//         sleep(SLEEP_DURATION);
//     });
 }