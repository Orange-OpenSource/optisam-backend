/*
 * application.proto
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * OpenAPI spec version: version not set
 *
 * NOTE: This class is auto generated by OpenAPI Generator.
 * https://github.com/OpenAPITools/openapi-generator
 *
 * OpenAPI generator version: 5.0.0-SNAPSHOT
 */


import http from "k6/http";
import { group, check, sleep } from "k6";

const BASE_URL = "https://optisam-application-int.apps.fr01.paas.tech.orange";
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
        http_req_duration: ['p(95)<1000'], 
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
    group("/api/v1/Obsolescence/meta/domaincriticity", () => {
        let url = BASE_URL + `/api/v1/Obsolescence/meta/domaincriticity`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);
    });
     group("/api/v1/Obsolescence/meta/maintenancecriticity", () => {
         let url = BASE_URL + `/api/v1/Obsolescence/meta/maintenancecriticity`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
         check(request, {
             "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);
     });
    group("/api/v1/Obsolescence/meta/risks", () => {
        let url = BASE_URL + `/api/v1/Obsolescence/meta/risks`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);
    });
    group("/api/v1/applications", () => {
        let searchParamsOwnerFilteringkey = "TODO_EDIT_THE_SEARCH_PARAMS.OWNER.FILTERINGKEY";
        let searchParamsObsolescenceRiskFilteringkey = "TODO_EDIT_THE_SEARCH_PARAMS.OBSOLESCENCE_RISK.FILTERINGKEY";
        let searchParamsDomainFilterType = "TODO_EDIT_THE_SEARCH_PARAMS.DOMAIN.FILTER_TYPE";
        let searchParamsProductIdFilteringkey = "TODO_EDIT_THE_SEARCH_PARAMS.PRODUCT_ID.FILTERINGKEY";
        let pageSize = 10;
        let searchParamsNameFilteringkeyMultiple = "TODO_EDIT_THE_SEARCH_PARAMS.NAME.FILTERINGKEY_MULTIPLE";
        let searchParamsDomainFilteringOrder = "TODO_EDIT_THE_SEARCH_PARAMS.DOMAIN.FILTERINGORDER";
        let searchParamsProductIdFilteringkeyMultiple = "TODO_EDIT_THE_SEARCH_PARAMS.PRODUCT_ID.FILTERINGKEY_MULTIPLE";
        let searchParamsNameFilteringOrder = "TODO_EDIT_THE_SEARCH_PARAMS.NAME.FILTERINGORDER";
        let sortBy = "name";
        let searchParamsNameFilterType = "TODO_EDIT_THE_SEARCH_PARAMS.NAME.FILTER_TYPE";
        let searchParamsProductIdFilterType = "TODO_EDIT_THE_SEARCH_PARAMS.PRODUCT_ID.FILTER_TYPE";
        let searchParamsDomainFilteringkeyMultiple = "TODO_EDIT_THE_SEARCH_PARAMS.DOMAIN.FILTERINGKEY_MULTIPLE";
        let searchParamsDomainFilteringkey = "TODO_EDIT_THE_SEARCH_PARAMS.DOMAIN.FILTERINGKEY";
        let searchParamsObsolescenceRiskFilteringkeyMultiple = "TODO_EDIT_THE_SEARCH_PARAMS.OBSOLESCENCE_RISK.FILTERINGKEY_MULTIPLE";
        let searchParamsOwnerFilterType = "TODO_EDIT_THE_SEARCH_PARAMS.OWNER.FILTER_TYPE";
        let searchParamsOwnerFilteringkeyMultiple = "TODO_EDIT_THE_SEARCH_PARAMS.OWNER.FILTERINGKEY_MULTIPLE";
        let searchParamsNameFilteringkey = "TODO_EDIT_THE_SEARCH_PARAMS.NAME.FILTERINGKEY";
        let pageNum = 1;
        let searchParamsObsolescenceRiskFilteringOrder = "TODO_EDIT_THE_SEARCH_PARAMS.OBSOLESCENCE_RISK.FILTERINGORDER";
        let searchParamsOwnerFilteringOrder = "TODO_EDIT_THE_SEARCH_PARAMS.OWNER.FILTERINGORDER";
        let sortOrder = "asc";
        let searchParamsProductIdFilteringOrder = "TODO_EDIT_THE_SEARCH_PARAMS.PRODUCT_ID.FILTERINGORDER";
        let searchParamsObsolescenceRiskFilterType = "TODO_EDIT_THE_SEARCH_PARAMS.OBSOLESCENCE_RISK.FILTER_TYPE";
        let scopes = "OFR";
        // let url = BASE_URL + `/api/v1/applications?page_num=${page_num}&page_size=${page_size}&sort_by=${sort_by}&sort_order=${sort_order}&search_params.name.filteringOrder=${search_params.name.filteringOrder}&search_params.name.filteringkey=${search_params.name.filteringkey}&search_params.name.filter_type=${search_params.name.filter_type}&search_params.name.filteringkey_multiple=${search_params.name.filteringkey_multiple}&search_params.owner.filteringOrder=${search_params.owner.filteringOrder}&search_params.owner.filteringkey=${search_params.owner.filteringkey}&search_params.owner.filter_type=${search_params.owner.filter_type}&search_params.owner.filteringkey_multiple=${search_params.owner.filteringkey_multiple}&search_params.product_id.filteringOrder=${search_params.product_id.filteringOrder}&search_params.product_id.filteringkey=${search_params.product_id.filteringkey}&search_params.product_id.filter_type=${search_params.product_id.filter_type}&search_params.product_id.filteringkey_multiple=${search_params.product_id.filteringkey_multiple}&search_params.domain.filteringOrder=${search_params.domain.filteringOrder}&search_params.domain.filteringkey=${search_params.domain.filteringkey}&search_params.domain.filter_type=${search_params.domain.filter_type}&search_params.domain.filteringkey_multiple=${search_params.domain.filteringkey_multiple}&search_params.obsolescence_risk.filteringOrder=${search_params.obsolescence_risk.filteringOrder}&search_params.obsolescence_risk.filteringkey=${search_params.obsolescence_risk.filteringkey}&search_params.obsolescence_risk.filter_type=${search_params.obsolescence_risk.filter_type}&search_params.obsolescence_risk.filteringkey_multiple=${search_params.obsolescence_risk.filteringkey_multiple}&scopes=${scopes}`;
        // Request No. 1
        let url = BASE_URL + `/api/v1/applications?page_num=1&page_size=10&scopes=${scopes}`
     //   let url = BASE_URL + `/api/v1/applications?page_num=${pageNum}&page_size=${pageSize}&sort_by=${sortBy}&sort_order=${sortOrder}&scopes=${scopes}`
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);

        // Request No. 2
        // TODO: edit the parameters of the request body.
        // let body = {"applicationId": "string", "name": "string", "version": "string", "owner": "string", "domain": "string", "scope": "DMO"};
        // let params = {headers: {"Content-Type": "application/json"}};
        // request = http.post((url,{headers:headers}), body, params);
        // check(request, {
        //     "A successful response.": (r) => r.status === 200
        // });
        // sleep(SLEEP_DURATION);
     });
    group("/api/v1/applications/domains", () => {
        let scope = "OFR";
        let url = BASE_URL + `/api/v1/applications/domains?scope=${scope}`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);
    });
    // group("/api/v1/applications/{application_id}", () => {
    //     let applicationId = "1";
    //     let url = BASE_URL + `/api/v1/applications/${applicationId}`;
    //     // Request No. 1
    //     let request = http.del(url,{headers:headers});
        
    //     check(request, {
    //         "A successful response.": (r) => r.status === 200
    //     });
    //     sleep(SLEEP_DURATION);
    // });
    // group("/api/v1/applications/{application_id}/instances", () => {
    //     let applicationId = "1";
    //     let url = BASE_URL + `/api/v1/applications/${applicationId}/instances`;
    //     // Request No. 1
    //     // TODO: edit the parameters of the request body.
    //     // let body = {"applicationId": "string", "instanceId": "string", "instanceName": "string", "products": {"operation": "string", "productId": "list"}, "equipments": {"operation": "string", "equipmentId": "list"}, "scope": "string"};
    //     // let params = {headers: {"Content-Type": "application/json", "Accept": "application/json"}};
    //     // let request = http.post((url,{headers:headers}), body, params);
    //     // check(request, {
    //     //     "A successful response.": (r) => r.status === 200
    //     // });
    //     // sleep(SLEEP_DURATION);
    // });
    // group("/api/v1/applications/{application_id}/instances/{instance_id}", () => {
    //     let instanceId = "127";
    //     let applicationId = "127";
    //     let url = BASE_URL + `/api/v1/applications/${applicationId}/instances/${instanceId}`;
    //     // Request No. 1
    //     let request = http.del(url,{headers:headers});
    //     check(request, {
    //         "A successful response.": (r) => r.status === 200
    //     });
    //     sleep(SLEEP_DURATION);
    // });
    group("/api/v1/instances", () => {
        let searchParamsApplicationIdFilterType = 1;
        let searchParamsProductIdFilteringkey = "TODO_EDIT_THE_SEARCH_PARAMS.PRODUCT_ID.FILTERINGKEY";
        let pageSize = 10;
        let pageNum = 1;
        let searchParamsApplicationIdFilteringOrder = "TODO_EDIT_THE_SEARCH_PARAMS.APPLICATION_ID.FILTERINGORDER";
        let searchParamsProductIdFilteringkeyMultiple = "TODO_EDIT_THE_SEARCH_PARAMS.PRODUCT_ID.FILTERINGKEY_MULTIPLE";
        let sortOrder = "asc";
        let searchParamsProductIdFilteringOrder = "TODO_EDIT_THE_SEARCH_PARAMS.PRODUCT_ID.FILTERINGORDER";
        let sortBy = "instance_id";
        let searchParamsProductIdFilterType = "TODO_EDIT_THE_SEARCH_PARAMS.PRODUCT_ID.FILTER_TYPE";
        let scopes = "OFR";
        let searchParamsApplicationIdFilteringkey = 22;
        let searchParamsApplicationIdFilteringkeyMultiple = "TODO_EDIT_THE_SEARCH_PARAMS.APPLICATION_ID.FILTERINGKEY_MULTIPLE";
        //let url = BASE_URL + `/api/v1/instances?page_num=${pageNum}&page_size=${pageSize}&scopes=${scopes}`;
        let url = BASE_URL + `/api/v1/instances?page_num=${pageNum}&page_size=${pageSize}&sort_by=${sortBy}&sort_order=${sortOrder}&scopes=${scopes}&search_params.application_id.filter_type=${searchParamsApplicationIdFilterType}&search_params.application_id.filteringkey=${searchParamsApplicationIdFilteringkey}`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);
    });
    group("/api/v1/obsolescence/domains", () => {
        let scope = "OFR";
        let url = BASE_URL + `/api/v1/obsolescence/domains?scope=${scope}`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);

        // Request No. 2
        // TODO: edit the parameters of the request body.
        // let body = {"scope": "string", "domainsCriticity": [{"domainCriticId": "integer", "domainCriticName": "string", "domains": "list"}]};
        // let params = {headers: {"Content-Type": "application/json"}};
        // request = http.post((url,{headers:headers}), body, params);
        // check(request, {
        //     "A successful response.": (r) => r.status === 200
        // });
        // sleep(SLEEP_DURATION);
    });
    group("/api/v1/obsolescence/maintenance", () => {
        let scope = "OFR";
        let url = BASE_URL + `/api/v1/obsolescence/maintenance?scope=${scope}`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);

        // Request No. 2
        // TODO: edit the parameters of the request body.
        // let body = {"scope": "string", "maintenanceCriticy": [{"maintenanceCriticId": "integer", "maintenanceLevelId": "integer", "maintenanceLevelName": "string", "startMonth": "integer", "endMonth": "integer"}]};
        // let params = {headers: {"Content-Type": "application/json"}};
        // request = http.post((url,{headers:headers}), body, params);
        // check(request, {
        //     "A successful response.": (r) => r.status === 200
        // });
        // sleep(SLEEP_DURATION);
    });
    group("/api/v1/obsolescence/matrix", () => {
        let scope = "OFR";
        let url = BASE_URL + `/api/v1/obsolescence/matrix?scope=${scope}`;
        // Request No. 1
        let request = http.get(url,{headers:headers});
        check(request, {
            // "A successful response.": (r) => r.status === 200
            "A successful response.": (r) => r.status === 200
        });
        sleep(SLEEP_DURATION);

        // Request No. 2
        // TODO: edit the parameters of the request body.
        // let body = {"scope": "string", "riskMatrix": [{"configurationId": "integer", "domainCriticId": "integer", "domainCriticName": "string", "maintenanceCriticId": "integer", "maintenanceCriticName": "string", "riskId": "integer", "riskName": "string"}]};
        // let params = {headers: {"Content-Type": "application/json"}};
        // request = http.post((url,{headers:headers}), body, params);
        // check(request, {
        //     "A successful response.": (r) => r.status === 200
        // });
        // sleep(SLEEP_DURATION);
    });
}