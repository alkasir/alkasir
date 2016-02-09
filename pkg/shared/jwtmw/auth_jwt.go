// Package jwt provides Json-Web-Token authentication for the go-json-rest framework
package jwtmw

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/dgrijalva/jwt-go"

	"errors"
	"log"
	"strings"
	"time"

	"github.com/alkasir/alkasir/pkg/shared/apierrors"
	"github.com/alkasir/alkasir/pkg/shared/apiutils"
	"github.com/alkasir/alkasir/pkg/shared/fielderrors"
)

// JWTMiddleware provides a Json-Web-Token authentication implementation. On failure, a 401 HTTP response
// is returned. On success, the wrapped middleware is called, and the userId is made available as
// request.Env["REMOTE_USER"].(string).
// Users can get a token by posting a json request to LoginHandler. The token then needs to be passed in
// the Authentication header. Example: Authorization:Bearer XXX_TOKEN_XXX
type JWTMiddleware struct {
	// Realm name to display to the user. Required.
	Realm string

	// signing algorithm - possible values are HS256, HS384, HS512
	// Optional, default is HS256.
	SigningAlgorithm string

	// Secret key used for signing. Required.
	Key []byte

	// Duration that a jwt token is valid. Optional, defaults to one hour.
	Timeout time.Duration

	// This field allows clients to refresh their token until MaxRefresh has passed.
	// Note that clients can refresh their token in the last moment of MaxRefresh.
	// This means that the maximum validity timespan for a token is MaxRefresh + Timeout.
	// Optional, defaults to 0 meaning not refreshable.
	MaxRefresh time.Duration

	// Callback function that should perform the authentication of the user based on userId and
	// password. Must return true on success, false on failure. Required.
	Authenticator func(userId string, password string) bool

	// Callback function that should perform the authorization of the authenticated user. Called
	// only after an authentication success. Must return true on success, false on failure.
	// Optional, default to success.
	Authorizator func(userId string, request *rest.Request) bool

	// Callback function that will be called during login.
	// Using this function it is possible to add additional payload data to the webtoken.
	// The data is then made available during requests via request.Env["JWT_PAYLOAD"].
	// Note that the payload is not encrypted.
	// The attributes mentioned on jwt.io can't be used as keys for the map.
	// Optional, by default no additional data will be set.
	PayloadFunc func(userId string) map[string]interface{}
}

// MiddlewareFunc makes JWTMiddleware implement the Middleware interface.
func (mw *JWTMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {

	if mw.Realm == "" {
		log.Fatal("Realm is required")
	}
	if mw.SigningAlgorithm == "" {
		mw.SigningAlgorithm = "HS256"
	}
	if mw.Key == nil {
		log.Fatal("Key required")
	}
	if mw.Timeout == 0 {
		mw.Timeout = time.Hour
	}
	if mw.Authenticator == nil {
		log.Fatal("Authenticator is required")
	}
	if mw.Authorizator == nil {
		mw.Authorizator = func(userId string, request *rest.Request) bool {
			return true
		}
	}

	return func(writer rest.ResponseWriter, request *rest.Request) { mw.middlewareImpl(writer, request, handler) }
}

func (mw *JWTMiddleware) middlewareImpl(writer rest.ResponseWriter, request *rest.Request, handler rest.HandlerFunc) {
	token, err := mw.parseToken(request)

	if err != nil {
		mw.unauthorized(writer)
		return
	}

	id := token.Claims["id"].(string)

	request.Env["REMOTE_USER"] = id
	request.Env["JWT_PAYLOAD"] = token.Claims

	if !mw.Authorizator(id, request) {
		mw.unauthorized(writer)
		return
	}

	handler(writer, request)
}

// Helper function to extract the JWT claims
func ExtractClaims(request *rest.Request) map[string]interface{} {
	if request.Env["JWT_PAYLOAD"] == nil {
		empty_claims := make(map[string]interface{})
		return empty_claims
	}
	jwt_claims := request.Env["JWT_PAYLOAD"].(map[string]interface{})
	return jwt_claims
}

type login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Handler that clients can use to get a jwt token.
// Payload needs to be json in the form of {"username": "USERNAME", "password": "PASSWORD"}.
// Reply will be of the form {"token": "TOKEN"}.
func (mw *JWTMiddleware) LoginHandler(writer rest.ResponseWriter, request *rest.Request) {
	login_vals := login{}
	err := request.DecodeJsonPayload(&login_vals)

	if err != nil {
		apiutils.WriteRestError(writer, apierrors.NewBadRequest(err.Error()))
		return
	}

	if login_vals.Username == "" {
		apiutils.WriteRestError(writer,
			apierrors.NewInvalid("login", "",
				fielderrors.ValidationErrorList{
					fielderrors.NewFieldRequired("username")}))
		return
	}
	if login_vals.Password == "" {
		apiutils.WriteRestError(writer,
			apierrors.NewInvalid("login", "",
				fielderrors.ValidationErrorList{
					fielderrors.NewFieldRequired("password")}))
		return
	}

	if !mw.Authenticator(login_vals.Username, login_vals.Password) {
		mw.unauthorized(writer)
		return
	}

	token := jwt.New(jwt.GetSigningMethod(mw.SigningAlgorithm))

	if mw.PayloadFunc != nil {
		for key, value := range mw.PayloadFunc(login_vals.Username) {
			token.Claims[key] = value
		}
	}

	token.Claims["id"] = login_vals.Username
	token.Claims["exp"] = time.Now().Add(mw.Timeout).Unix()
	if mw.MaxRefresh != 0 {
		token.Claims["orig_iat"] = time.Now().Unix()
	}
	tokenString, err := token.SignedString(mw.Key)

	if err != nil {
		mw.unauthorized(writer)
		return
	}

	writer.WriteJson(&map[string]string{"token": tokenString})
}

func (mw *JWTMiddleware) parseToken(request *rest.Request) (*jwt.Token, error) {
	authHeader := request.Header.Get("Authorization")

	if authHeader == "" {
		return nil, errors.New("Auth header empty")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return nil, errors.New("Invalid auth header")
	}

	return jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(mw.SigningAlgorithm) != token.Method {
			return nil, errors.New("Invalid signing algorithm")
		}
		return mw.Key, nil
	})
}

type token struct {
	Token string `json:"token"`
}

// Handler that clients can use to refresh their token. The token still needs to be valid on refresh.
// Shall be put under an endpoint that is using the JWTMiddleware.
// Reply will be of the form {"token": "TOKEN"}.
func (mw *JWTMiddleware) RefreshHandler(writer rest.ResponseWriter, request *rest.Request) {
	token, err := mw.parseToken(request)

	// Token should be valid anyway as the RefreshHandler is authed
	if err != nil {
		mw.unauthorized(writer)
		return
	}

	origIat := int64(token.Claims["orig_iat"].(float64))

	if origIat < time.Now().Add(-mw.MaxRefresh).Unix() {
		mw.unauthorized(writer)
		return
	}

	newToken := jwt.New(jwt.GetSigningMethod(mw.SigningAlgorithm))

	for key := range token.Claims {
		newToken.Claims[key] = token.Claims[key]
	}

	newToken.Claims["id"] = token.Claims["id"]
	newToken.Claims["exp"] = time.Now().Add(mw.Timeout).Unix()
	newToken.Claims["orig_iat"] = origIat
	tokenString, err := newToken.SignedString(mw.Key)

	if err != nil {
		mw.unauthorized(writer)
		return
	}

	writer.WriteJson(&map[string]string{"token": tokenString})
}

func (mw *JWTMiddleware) unauthorized(writer rest.ResponseWriter) {
	writer.Header().Set("WWW-Authenticate", "JWT realm="+mw.Realm)
	apiutils.WriteRestError(writer, apierrors.NewUnauthorized(""))
}
