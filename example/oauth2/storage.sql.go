package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"go.onebrick.io/brock"
)

func (x _storage_sql) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	return x.ClientRead(ctx, id)
}
func (x _storage_sql) Create(ctx context.Context, info oauth2.TokenInfo) error {
	return x.TokenCreate(ctx, info)
}
func (x _storage_sql) RemoveByCode(ctx context.Context, code string) error {
	return x.TokenDelete(ctx, code, "", "")
}
func (x _storage_sql) RemoveByAccess(ctx context.Context, access string) error {
	return x.TokenDelete(ctx, "", access, "")
}
func (x _storage_sql) RemoveByRefresh(ctx context.Context, refresh string) error {
	return x.TokenDelete(ctx, "", "", refresh)
}
func (x _storage_sql) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	return x.TokenRead(ctx, code, "", "")
}
func (x _storage_sql) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	return x.TokenRead(ctx, "", access, "")
}
func (x _storage_sql) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	return x.TokenRead(ctx, "", "", refresh)
}

var (
	_ oauth2.ClientStore = _storage_sql{}
	_ oauth2.TokenStore  = _storage_sql{}
)

// _storage_sql implements both oauth2.ClientStore and oauth2.TokenStore
type _storage_sql struct {
	brock.SQLConn

	TablenameClient string
	TablenameToken  string
}

// StorageSQL implements both oauth2.ClientStore and oauth2.TokenStore
func (x *instance) StorageSQL() _storage_sql { return _storage_sql{SQLConn: x.SQLConn} }

// =============================================================================
// Client Storage

func (x _storage_sql) ClientCreate(ctx context.Context, info oauth2.ClientInfo) error {
	return nil
}
func (x _storage_sql) ClientRead(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	return nil, nil
}
func (x _storage_sql) ClientDelete(ctx context.Context, id string) error {
	return nil
}

// =============================================================================
// Token Storage

func (x _storage_sql) TokenCreate(ctx context.Context, info oauth2.TokenInfo) error {
	return nil
}
func (x _storage_sql) TokenRead(ctx context.Context, code, access, refresh string) (oauth2.TokenInfo, error) {
	var c, a, r sql.NullString
	switch {
	default:
		return nil, nil
	case code != "" && access == "" && refresh == "":
		c = sql.NullString{String: code, Valid: true}
	case code == "" && access != "" && refresh == "":
		a = sql.NullString{String: access, Valid: true}
	case code == "" && access == "" && refresh != "":
		r = sql.NullString{String: refresh, Valid: true}
	}
	info, t, v, cre, exp, now := tokenInfo{}, 0, "", time.Time{}, time.Time{}, time.Now()
	err := brock.SQL.Box.Query(x.QueryContext(ctx, _sql_select_token, c, a, r)).Scan(func(i int) (pointers []any) {
		return []any{
			&info.ID,
			&info.Token.ClientID,
			&info.Token.UserID,
			&info.Token.RedirectURI,
			&info.Token.Scope,
			&t,
			&v,
			&cre,
			&exp,
		}
	})
	_ = err
	switch t {
	default:
	case 1:
		info.Token.Code = v
		info.Token.CodeCreateAt = cre
		info.Token.CodeExpiresIn = exp.Sub(now)
	case 2:
		info.Token.Access = v
		info.Token.AccessCreateAt = cre
		info.Token.AccessExpiresIn = exp.Sub(now)
	case 3:
		info.Token.Refresh = v
		info.Token.RefreshCreateAt = cre
		info.Token.RefreshExpiresIn = exp.Sub(now)
	}
	return nil, nil
}
func (x _storage_sql) TokenDelete(ctx context.Context, code, access, refresh string) error {
	switch {
	default:
		return ErrUnimplemented
	case code != "" && access == "" && refresh == "":
		return ErrUnimplemented
	case code == "" && access != "" && refresh == "":
		return ErrUnimplemented
	case code == "" && access == "" && refresh != "":
		return ErrUnimplemented
	}
}
