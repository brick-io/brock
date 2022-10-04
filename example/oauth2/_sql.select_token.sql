SELECT 
    t._id,
    t._client_id,
    t._user_id,
    t._redirect_uri,
    t._scope,
    tk._type,
    tk._value,
    tk._created_at,
    tk._expires_at
FROM
    public.token_keys tk
INNER JOIN
    public.tokens t ON t._id = tk._token_id
WHERE
    (tk._type = 1 AND $1 = tk._value AND $2::TEXT IS NULL AND $3::TEXT IS NULL ) OR
    (tk._type = 2 AND $2 = tk._value AND $1::TEXT IS NULL AND $3::TEXT IS NULL ) OR
    (tk._type = 3 AND $3 = tk._value AND $1::TEXT IS NULL AND $2::TEXT IS NULL )
;