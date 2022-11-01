package main

import (
	"context"
	"os"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"

	"go.onebrick.io/brock"
)

var _ = a()

func a() int {
	_ = store.NewClientStore()
	_, _ = store.NewMemoryTokenStore()
	_ = generates.NewAccessGenerate()
	_ = generates.NewAuthorizeGenerate()

	return 0
}

type instance struct {
	brock.SQLConn
}

func (x *instance) Manager() oauth2.Manager {
	if x.SQLConn == nil {
		x.SQLConn, _ = brock.SQL.Open(os.Getenv("OAUTH2_DSN"))
	}

	m := manage.NewDefaultManager()
	m.MapTokenStorage(x.StorageSQL())
	m.MapClientStorage(x.StorageSQL())
	m.MapAccessGenerate(x.AccessGenerate())
	m.MapAuthorizeGenerate(x.AuthorizeGenerate())
	m.SetAuthorizeCodeExp(manage.DefaultCodeExp)
	m.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	m.SetClientTokenCfg(manage.DefaultClientTokenCfg)
	m.SetImplicitTokenCfg(manage.DefaultImplicitTokenCfg)
	m.SetPasswordTokenCfg(manage.DefaultPasswordTokenCfg)
	m.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)
	m.SetValidateURIHandler(manage.DefaultValidateURI)

	return m
}

func (x *instance) AccessGenerate() oauth2.AccessGenerate {
	return _generate_access_and_refresh_token{x}
}

func (x *instance) AuthorizeGenerate() oauth2.AuthorizeGenerate {
	return _generate_authorization_code{x}
}

type (
	_gaart = _generate_access_and_refresh_token
	_gac   = _generate_authorization_code
)

type _generate_access_and_refresh_token struct{ *instance }

//nolint:lll
func (x _gaart) Token(ctx context.Context, data *oauth2.GenerateBasic, isRefresh bool) (access, refresh string, err error) {
	_ = data.Client.GetID()
	_ = data.UserID
	_ = data.TokenInfo.GetAccessCreateAt().Add(data.TokenInfo.GetAccessExpiresIn()).Unix()

	return "<-1->", "<-2->", nil
}

type _generate_authorization_code struct{ *instance }

func (x _gac) Token(ctx context.Context, data *oauth2.GenerateBasic) (code string, err error) {
	return "<-3->", nil
}

var (
	_ oauth2.ClientInfo = (*clientInfo)(nil)
	_ oauth2.TokenInfo  = (*tokenInfo)(nil)
)

type clientInfo struct { // extend the client info
	*models.Client

	Name string
}

type tokenInfo struct { // extend the token info
	*models.Token

	ID string
}
