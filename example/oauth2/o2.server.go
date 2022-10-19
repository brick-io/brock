package main

import (
	"net/http"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/server"
)

func oauth2Server() *server.Server {
	x := new(instance)

	s := server.NewDefaultServer(x.Manager())
	s.SetTokenType("Bearer")
	s.SetAllowGetAccessRequest(true)
	s.SetAllowedResponseType(oauth2.Code, oauth2.Token)
	s.SetAllowedGrantType(
		// AuthorizationCode ===================================================
		// RETURN ACCESS_TOKEN (WITH END-USER DELEGATED AUTHORIZATION, USING AUTHORIZATION_CODE AS INTERMEDIARY)
		//
		// used to get access_token for further request
		//
		// ✅ normal BE to BE
		//   1. get authorization_code (VIA Redirect) : client_id + redirect_uri
		//   2. get access_token : client_id + client_secret + authorization_code + redirect_uri
		//
		// ✅ with PKCE [ONLY ON native/mobile/SPA that end-user distributed]
		//   1. get authorization_code (VIA Redirect) : client_id + redirect_uri + code_challenge + code_challenge_method
		//   2. get access_token : client_id + code_verifier + authorization_code + redirect_uri
		oauth2.AuthorizationCode,

		// PasswordCredentials =================================================
		// RETURN ACCESS_TOKEN (USE INSIDE OUR OWN CLUSTER WITH ADDITIONAL SECURITY CONTEXT FOR FORWARDING USERNAME+PASSWORD)
		//
		// ❌ should be avoided on client side because using username/password combination
		// check the IP, it should be originated from internal WITHOUT EXTERNAL trigger
		// if cron job, should also check the time window of execution
		// might use an encrypted USERNAME & encrypted PASSWORD
		oauth2.PasswordCredentials,

		// ClientCredentials ===================================================
		// RETURN ACCESS_TOKEN (NO END-USER AUTHORIZED THIS FLOW, USED FOR AUTOMATION FROM CLIENT E.G. GENERATE REPORTING)
		//
		// should be used ONLY when the client = resource owner itself
		// the case of not getting any delegated authorization from end-user
		oauth2.ClientCredentials,

		// Refreshing ==========================================================
		// RETURN ACCESS_TOKEN (AS WELL AS INVALIDATE THE PREVIOUS REFRESH_TOKEN)
		//
		// IMO, no need to refresh except on ClientCredentials
		oauth2.Refreshing,

		// Implicit ============================================================
		// RETURN ID_TOKEN ONLY -- NOT ACCESS_TOKEN
		//
		// used to get immediate scope openid return of id_token
		oauth2.Implicit)

	s.Config.AllowedCodeChallengeMethods = []oauth2.CodeChallengeMethod{
		oauth2.CodeChallengeS256,  // use this when challenge via SHA-256
		oauth2.CodeChallengePlain, // else use this
	}

	// ForcePKCE should be use together with AuthorizationCode
	s.Config.ForcePKCE = false

	// srv.SetAccessTokenExpHandler(func(w http.ResponseWriter, r *http.Request) (exp time.Duration, err error) {
	// brock.Printf("\n  SetAccessTokenExpHandler r:[%#v]\n", empty(r))
	// 	return 0, nil
	// })
	// srv.SetAuthorizeScopeHandler(func(w http.ResponseWriter, r *http.Request) (scope string, err error) {
	// brock.Printf("\n  SetAuthorizeScopeHandler r:[%#v]\n", empty(r))
	// 	return "<-->", nil
	// })
	// srv.SetClientAuthorizedHandler(func(clientID string, grant oauth2.GrantType) (allowed bool, err error) {
	// brock.Printf("\n  SetClientAuthorizedHandler clientID:[%s] grant:[%s]", clientID, grant)
	// 	switch grant {
	// 	case
	// 		oauth2.AuthorizationCode,
	// 		oauth2.PasswordCredentials,
	// 		oauth2.ClientCredentials,
	// 		oauth2.Refreshing,
	// 		oauth2.Implicit:
	// 		return true, nil
	// 	}
	// 	return false, nil
	// })
	// srv.SetClientInfoHandler(func(r *http.Request) (clientID string, clientSecret string, err error) {
	// 	if clientID, clientSecret, err = server.ClientBasicHandler(r); err != nil {
	// 		clientID, clientSecret, err = server.ClientFormHandler(r)
	// 	}
	// brock.Printf("\n  SetClientInfoHandler clientID:[%s] clientSecret:[%s] err:[%#v] r:[%#v]\n", clientID, clientSecret, err, r)
	// 	return clientID, clientSecret, err
	// })
	// srv.SetClientScopeHandler(func(tgr *oauth2.TokenGenerateRequest) (allowed bool, err error) {
	// brock.Printf("\n  SetClientScopeHandler tgr:[%#v] \n", tgr)
	// 	return false, nil
	// })
	// srv.SetExtensionFieldsHandler(func(ti oauth2.TokenInfo) (fieldsValue map[string]interface{}) {
	// brock.Printf("\n  SetExtensionFieldsHandler ti:[%#v] \n", ti)
	// 	return nil
	// })
	// srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
	// brock.Printf("\n  SetInternalErrorHandler err:[%s]\n", err)
	// 	if err != nil {
	// 		re = &errors.Response{Error: err}
	// 	}
	// 	return re
	// })
	// srv.SetPasswordAuthorizationHandler(func(ctx context.Context, clientID, username, password string) (userID string, err error) {
	// brock.Printf("\n  SetPasswordAuthorizationHandler clientID:[%s] username:[%s] password:[%s]\n", clientID, username, password)
	// 	return "<-->", nil
	// })
	// srv.SetPreRedirectErrorHandler(func(w http.ResponseWriter, req *server.AuthorizeRequest, err error) error {
	// 	req_ := *req
	// 	req_.Request = empty(req_.Request)
	// brock.Printf("\n  SetPreRedirectErrorHandler err:[%s] req:[%#v]\n", err, &req_)
	// 	return nil
	// })
	// srv.SetRefreshingScopeHandler(func(tgr *oauth2.TokenGenerateRequest, oldScope string) (allowed bool, err error) {
	// brock.Printf("\n  SetRefreshingScopeHandler tgr:[%#v]\n", tgr)
	// 	return false, nil
	// })
	// srv.SetRefreshingValidationHandler(func(ti oauth2.TokenInfo) (allowed bool, err error) {
	// brock.Printf("\n  SetRefreshingValidationHandler ti:[%#v]\n", ti)
	// 	return false, nil
	// })
	// srv.SetResponseErrorHandler(func(re *errors.Response) {
	// brock.Printf("\n  SetResponseErrorHandler re:[%#v]\n", re)
	// })
	// srv.SetResponseTokenHandler(func(w http.ResponseWriter, data map[string]interface{}, header http.Header, statusCode ...int) error {
	// brock.Printf("\n  SetResponseTokenHandler data:[%#v]\n", data)
	// 	return nil
	// })
	s.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
		var u User
		u, err = getUser(r)
		if err != nil || u.ID == "" {
			action := "?a=" + Cipher{"next:" + r.URL.String()}.String()

			// w.WriteHeader(http.StatusFound)
			w.Header().Set("Location", "/authentication"+action)
		}

		return u.ID, err
	})

	return s
}
