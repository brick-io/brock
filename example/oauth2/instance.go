package main

import (
	"context"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
	"go.onebrick.io/brock"
)

func a() {
	_ = store.NewClientStore()
	_, _ = store.NewMemoryTokenStore()
	_ = generates.NewAccessGenerate()
	_ = generates.NewAuthorizeGenerate()
}

type instance struct {
	brock.SQLConn
}

func (x *instance) Manager() oauth2.Manager {
	if x.SQLConn == nil {
		brock.SQL.Open("postgres://oauth_user:oauth_password@localhost:5432/oauth2_db")
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

type _generate_access_and_refresh_token struct{ *instance }

func (x _generate_access_and_refresh_token) Token(ctx context.Context, data *oauth2.GenerateBasic, isGenRefresh bool) (access, refresh string, err error) {
	_ = data.Client.GetID()
	_ = data.UserID
	_ = data.TokenInfo.GetAccessCreateAt().Add(data.TokenInfo.GetAccessExpiresIn()).Unix()
	return "<-->", "<-->", nil
}

type _generate_authorization_code struct{ *instance }

func (x _generate_authorization_code) Token(ctx context.Context, data *oauth2.GenerateBasic) (code string, err error) {
	return "<-->", nil
}

var (
	_ oauth2.ClientInfo = (*client_info)(nil)
	_ oauth2.TokenInfo  = (*token_info)(nil)
)

type client_info struct { // extend the client info
	*models.Client

	Name string
}

type token_info struct { // extend the token info
	*models.Token

	ID string
}
