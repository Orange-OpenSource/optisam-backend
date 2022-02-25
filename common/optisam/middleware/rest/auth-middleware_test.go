package rest

// import (
// 	jwt "github.com/dgrijalva/jwt-go"
// 	"io/ioutil"
// 	"net/http"
// 	"net/http/httptest"
// 	"net/url"
// 	"strings"
// 	"testing"
// )

// func TestValidateAuth(t *testing.T) {
// 	// Create a request to pass to our handler.
// 	data := url.Values{}
// 	data.Add("firstName", "dharmjit")
// 	data.Add("lastName", "singh")
// 	data.Add("locale", "en")
// 	req, err := http.NewRequest("PUT", "/api/v1/account/", strings.NewReader(data.Encode()))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJPcHRpc2FtQ2xhaW1zIjp7IlVzZXJJRCI6InVzZXIxQHRlc3QuY29tIiwiRW50aXR5IjoiIiwiTG9jYWxlIjoiIn0sImV4cCI6MTU1MTk1ODI2NiwiaWF0IjoxNTUxOTUxMDY2LCJpc3MiOiJPcmFuZ2UiLCJzdWIiOiJBY2Nlc3MgVG9rZW4ifQ.TpgJG9XbEZhPjjN04NGdCiz2v32pdfDEtn2iZNU8Z6lBEdsHO-_iZfDZULe1nkq27OLUv0xzyMxeZHveieyoN6roqfXlpDWrK1s3vAvL1cqatxPvEO4UZwLML0rMFy-ebwcbP-beFEPCaHXTQZw0B-yRgPKb_esv4kT2vx59qVbsTlL_OMqAB_l2nJP5Zkm8IeJegJWkyK2DrQTprqy-c_p5WfKLUgJDGBgkZweXD9km4bl6jSiqvO_7sC0GnK-8nWJBr1dFWy-naH_SM-Gz2Z51BYjd1hXyyo-4BP1QorY4Z53dH0RCoRf365aXZR_u1nZBzSAiiMa1fGt51vG9YQ")

// 	// Get the verify key

// 	verifyBytes, err := ioutil.ReadFile("./../../../../auth-service/cmd/server/cert.pem")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if val, ok := r.Context().Value(keyUserID).(string); !ok {
// 			t.Errorf("keyUserID not in request context: got %q", val)
// 		}
// 	})

// 	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
// 	rr := httptest.NewRecorder()
// 	handler := ValidateAuth(verifyKey, testHandler)

// 	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
// 	// directly and pass in our Request and ResponseRecorder.
// 	handler.ServeHTTP(rr, req)

// 	// Check the status code is what we expect.
// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("handler returned wrong status code: got %v want %v",
// 			status, http.StatusOK)
// 	}
// 	// type args struct {
// 	// 	privateKey string
// 	// 	h          http.Handler
// 	// }
// 	// tests := []struct {
// 	// 	name string
// 	// 	args args
// 	// 	want http.Handler
// 	// }{
// 	// 	// TODO: Add test cases.
// 	// }
// 	// for _, tt := range tests {
// 	// 	t.Run(tt.name, func(t *testing.T) {
// 	// 		if got := ValidateAuth(tt.args.privateKey, tt.args.h); !reflect.DeepEqual(got, tt.want) {
// 	// 			t.Errorf("ValidateAuth() = %v, want %v", got, tt.want)
// 	// 		}
// 	// 	})
// 	// }
// }
