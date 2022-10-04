SELECT 
    c._id,
    c._secret,
    c._domain,
    c._user_id
FROM
    public.clients c
WHERE
    c._id = $1
;