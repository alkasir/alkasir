package jwtmw

import (
	"testing"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	jwt "gopkg.in/dgrijalva/jwt-go.v2"

)

var (
	key = []byte("secret key")
)

type DecoderToken struct {
	Token string `json:"token"`
}

func makeTokenString(username string, key []byte) string {
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	token.Claims["id"] = "admin"
	token.Claims["exp"] = time.Now().Add(time.Hour).Unix()
	token.Claims["orig_iat"] = time.Now().Unix()
	tokenString, _ := token.SignedString(key)
	return tokenString
}

func TestAuthJWT(t *testing.T) {

	// the middleware to test
	authMiddleware := &JWTMiddleware{
		Realm:      "test zone",
		Key:        key,
		Timeout:    time.Hour,
		MaxRefresh: time.Hour * 24,
		Authenticator: func(userId string, password string) bool {
			if userId == "admin" && password == "admin" {
				return true
			}
			return false
		},
		Authorizator: func(userId string, request *rest.Request) bool {
			if request.Method == "GET" {
				return true
			}
			return false
		},
	}

	// api for testing failure
	apiFailure := rest.NewApi()
	apiFailure.Use(authMiddleware)
	apiFailure.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		t.Error("Should never be executed")
	}))
	handler := apiFailure.MakeHandler()

	// simple request fails
	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/", nil))
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// auth with right cred and wrong method fails
	wrongMethodReq := test.MakeSimpleRequest("POST", "http://localhost/", nil)
	wrongMethodReq.Header.Set("Authorization", "Bearer "+makeTokenString("admin", key))
	recorded = test.RunRequest(t, handler, wrongMethodReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// wrong Auth format - bearer lower case
	wrongAuthFormat := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	wrongAuthFormat.Header.Set("Authorization", "bearer "+makeTokenString("admin", key))
	recorded = test.RunRequest(t, handler, wrongAuthFormat)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// wrong Auth format - no space after bearer
	wrongAuthFormat = test.MakeSimpleRequest("GET", "http://localhost/", nil)
	wrongAuthFormat.Header.Set("Authorization", "bearer"+makeTokenString("admin", key))
	recorded = test.RunRequest(t, handler, wrongAuthFormat)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// wrong Auth format - empty auth header
	wrongAuthFormat = test.MakeSimpleRequest("GET", "http://localhost/", nil)
	wrongAuthFormat.Header.Set("Authorization", "")
	recorded = test.RunRequest(t, handler, wrongAuthFormat)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// right credt, right method but wrong priv key
	wrongPrivKeyReq := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	wrongPrivKeyReq.Header.Set("Authorization", "Bearer "+makeTokenString("admin", []byte("sekret key")))
	recorded = test.RunRequest(t, handler, wrongPrivKeyReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// right credt, right method, right priv key but timeout
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	token.Claims["id"] = "admin"
	token.Claims["exp"] = 0
	tokenString, _ := token.SignedString(key)

	expiredTimestampReq := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	expiredTimestampReq.Header.Set("Authorization", "Bearer "+tokenString)
	recorded = test.RunRequest(t, handler, expiredTimestampReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// right credt, right method, right priv, wrong signing method on request
	tokenBadSigning := jwt.New(jwt.GetSigningMethod("HS384"))
	tokenBadSigning.Claims["id"] = "admin"
	tokenBadSigning.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	tokenBadSigningString, _ := tokenBadSigning.SignedString(key)

	BadSigningReq := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	BadSigningReq.Header.Set("Authorization", "Bearer "+tokenBadSigningString)
	recorded = test.RunRequest(t, handler, BadSigningReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// api for testing success
	apiSuccess := rest.NewApi()
	apiSuccess.Use(authMiddleware)
	apiSuccess.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		if r.Env["REMOTE_USER"] == nil {
			t.Error("REMOTE_USER is nil")
		}
		user := r.Env["REMOTE_USER"].(string)
		if user != "admin" {
			t.Error("REMOTE_USER is expected to be 'admin'")
		}
		w.WriteJson(map[string]string{"Id": "123"})
	}))

	// auth with right cred and right method succeeds
	validReq := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	validReq.Header.Set("Authorization", "Bearer "+makeTokenString("admin", key))
	recorded = test.RunRequest(t, apiSuccess.MakeHandler(), validReq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	// login tests
	loginApi := rest.NewApi()
	loginApi.SetApp(rest.AppSimple(authMiddleware.LoginHandler))

	// wrong login
	wrongLoginCreds := map[string]string{"username": "admin", "password": "admIn"}
	wrongLoginReq := test.MakeSimpleRequest("POST", "http://localhost/", wrongLoginCreds)
	recorded = test.RunRequest(t, loginApi.MakeHandler(), wrongLoginReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// empty login
	emptyLoginCreds := map[string]string{}
	emptyLoginReq := test.MakeSimpleRequest("POST", "http://localhost/", emptyLoginCreds)
	recorded = test.RunRequest(t, loginApi.MakeHandler(), emptyLoginReq)
	recorded.CodeIs(422)
	recorded.ContentTypeIsJson()

	// correct login
	before := time.Now().Unix()
	loginCreds := map[string]string{"username": "admin", "password": "admin"}
	rightCredReq := test.MakeSimpleRequest("POST", "http://localhost/", loginCreds)
	recorded = test.RunRequest(t, loginApi.MakeHandler(), rightCredReq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	nToken := DecoderToken{}
	test.DecodeJsonPayload(recorded.Recorder, &nToken)
	newToken, err := jwt.Parse(nToken.Token, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		t.Error("Received new token with wrong signature", err)
	}

	if newToken.Claims["id"].(string) != "admin" ||
		int64(newToken.Claims["exp"].(float64)) < before {
		t.Error("Received new token with wrong data")
	}

	refreshApi := rest.NewApi()
	refreshApi.Use(authMiddleware)
	refreshApi.SetApp(rest.AppSimple(authMiddleware.RefreshHandler))

	// refresh with expired max refresh
	unrefreshableToken := jwt.New(jwt.GetSigningMethod("HS256"))
	unrefreshableToken.Claims["id"] = "admin"
	// the combination actually doesn't make sense but is ok for the test
	unrefreshableToken.Claims["exp"] = time.Now().Add(time.Hour).Unix()
	unrefreshableToken.Claims["orig_iat"] = 0
	tokenString, _ = unrefreshableToken.SignedString(key)

	unrefreshableReq := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	unrefreshableReq.Header.Set("Authorization", "Bearer "+tokenString)
	recorded = test.RunRequest(t, refreshApi.MakeHandler(), unrefreshableReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// valid refresh
	refreshableToken := jwt.New(jwt.GetSigningMethod("HS256"))
	refreshableToken.Claims["id"] = "admin"
	// we need to substract one to test the case where token is being created in
	// the same second as it is checked -> < wouldn't fail
	refreshableToken.Claims["exp"] = time.Now().Add(time.Hour).Unix() - 1
	refreshableToken.Claims["orig_iat"] = time.Now().Unix() - 1
	tokenString, _ = refreshableToken.SignedString(key)

	validRefreshReq := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	validRefreshReq.Header.Set("Authorization", "Bearer "+tokenString)
	recorded = test.RunRequest(t, refreshApi.MakeHandler(), validRefreshReq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	rToken := DecoderToken{}
	test.DecodeJsonPayload(recorded.Recorder, &rToken)
	refreshToken, err := jwt.Parse(rToken.Token, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		t.Error("Received refreshed token with wrong signature", err)
	}

	if refreshToken.Claims["id"].(string) != "admin" ||
		int64(refreshToken.Claims["orig_iat"].(float64)) != refreshableToken.Claims["orig_iat"].(int64) ||
		int64(refreshToken.Claims["exp"].(float64)) < refreshableToken.Claims["exp"].(int64) {
		t.Error("Received refreshed token with wrong data")
	}
}

func TestAuthJWTPayload(t *testing.T) {
	authMiddleware := &JWTMiddleware{
		Realm:            "test zone",
		SigningAlgorithm: "HS256",
		Key:              key,
		Timeout:          time.Hour,
		MaxRefresh:       time.Hour * 24,
		Authenticator: func(userId string, password string) bool {
			if userId == "admin" && password == "admin" {
				return true
			}
			return false
		},
		PayloadFunc: func(userId string) map[string]interface{} {
			// tests normal value
			// tests overwriting of reserved jwt values should have no effect
			return map[string]interface{}{"testkey": "testval", "exp": 0}
		},
	}

	loginApi := rest.NewApi()
	loginApi.SetApp(rest.AppSimple(authMiddleware.LoginHandler))

	// correct payload
	loginCreds := map[string]string{"username": "admin", "password": "admin"}
	rightCredReq := test.MakeSimpleRequest("POST", "http://localhost/", loginCreds)
	recorded := test.RunRequest(t, loginApi.MakeHandler(), rightCredReq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	nToken := DecoderToken{}
	test.DecodeJsonPayload(recorded.Recorder, &nToken)
	newToken, err := jwt.Parse(nToken.Token, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		t.Error("Received new token with wrong signature", err)
	}

	if newToken.Claims["testkey"].(string) != "testval" || newToken.Claims["exp"].(float64) == 0 {
		t.Error("Received new token without payload")
	}

	// correct payload after refresh
	refreshApi := rest.NewApi()
	refreshApi.Use(authMiddleware)
	refreshApi.SetApp(rest.AppSimple(authMiddleware.RefreshHandler))

	refreshableToken := jwt.New(jwt.GetSigningMethod("HS256"))
	refreshableToken.Claims["id"] = "admin"
	refreshableToken.Claims["exp"] = time.Now().Add(time.Hour).Unix()
	refreshableToken.Claims["orig_iat"] = time.Now().Unix()
	refreshableToken.Claims["testkey"] = "testval"
	tokenString, _ := refreshableToken.SignedString(key)

	validRefreshReq := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	validRefreshReq.Header.Set("Authorization", "Bearer "+tokenString)
	recorded = test.RunRequest(t, refreshApi.MakeHandler(), validRefreshReq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	rToken := DecoderToken{}
	test.DecodeJsonPayload(recorded.Recorder, &rToken)
	refreshToken, err := jwt.Parse(rToken.Token, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		t.Error("Received refreshed token with wrong signature", err)
	}

	if refreshToken.Claims["testkey"].(string) != "testval" {
		t.Error("Received new token without payload")
	}

	// payload is accessible in request
	payloadApi := rest.NewApi()
	payloadApi.Use(authMiddleware)
	payloadApi.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		testval := r.Env["JWT_PAYLOAD"].(map[string]interface{})["testkey"].(string)
		w.WriteJson(map[string]string{"testkey": testval})
	}))

	payloadToken := jwt.New(jwt.GetSigningMethod("HS256"))
	payloadToken.Claims["id"] = "admin"
	payloadToken.Claims["exp"] = time.Now().Add(time.Hour).Unix()
	payloadToken.Claims["orig_iat"] = time.Now().Unix()
	payloadToken.Claims["testkey"] = "testval"
	payloadTokenString, _ := payloadToken.SignedString(key)

	payloadReq := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	payloadReq.Header.Set("Authorization", "Bearer "+payloadTokenString)
	recorded = test.RunRequest(t, payloadApi.MakeHandler(), payloadReq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	payload := map[string]string{}
	test.DecodeJsonPayload(recorded.Recorder, &payload)

	if payload["testkey"] != "testval" {
		t.Error("Received new token without payload")
	}

}

func TestClaimsDuringAuthorization(t *testing.T) {
	authMiddleware := &JWTMiddleware{
		Realm:            "test zone",
		SigningAlgorithm: "HS256",
		Key:              key,
		Timeout:          time.Hour,
		MaxRefresh:       time.Hour * 24,
		PayloadFunc: func(userId string) map[string]interface{} {
			// Set custom claim, to be checked in Authorizator method
			return map[string]interface{}{"testkey": "testval", "exp": 0}
		},
		Authenticator: func(userId string, password string) bool {
			// Not testing authentication, just authorization, so always return true
			return true
		},
		Authorizator: func(userId string, request *rest.Request) bool {
			jwt_claims := ExtractClaims(request)

			// Check the actual claim, set in PayloadFunc
			return (jwt_claims["testkey"] == "testval")
		},
	}

	// Simple endpoint
	endpoint := func(w rest.ResponseWriter, r *rest.Request) {
		// Dummy endpoint, output doesn't really matter, we are checking
		// the code returned
		w.WriteJson(map[string]string{"Id": "123"})
	}

	// Setup simple app structure
	loginApi := rest.NewApi()
	loginApi.SetApp(rest.AppSimple(authMiddleware.LoginHandler))
	loginApi.Use(&rest.IfMiddleware{
		// Only authenticate non /login requests
		Condition: func(request *rest.Request) bool {
			return request.URL.Path != "/login"
		},
		IfTrue: authMiddleware,
	})
	api_router, _ := rest.MakeRouter(
		rest.Post("/login", authMiddleware.LoginHandler),
		rest.Get("/", endpoint),
	)
	loginApi.SetApp(api_router)

	// Authenticate
	loginCreds := map[string]string{"username": "admin", "password": "admin"}
	rightCredReq := test.MakeSimpleRequest("POST", "http://localhost/login", loginCreds)
	recorded := test.RunRequest(t, loginApi.MakeHandler(), rightCredReq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	// Decode received token, to be sent with endpoint request
	nToken := DecoderToken{}
	test.DecodeJsonPayload(recorded.Recorder, &nToken)

	// Request endpoint, triggering Authorization.
	// If we get a 200 then the claims were available in Authorizator method
	req := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	req.Header.Set("Authorization", "Bearer "+nToken.Token)
	recorded = test.RunRequest(t, loginApi.MakeHandler(), req)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
}
