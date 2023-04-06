# Queries

The `queries` library for Go implements a file based SQL query management system. Only PostgreSQL is supported. 

## Installing

```
  go get -u github.com/radim/queries
```

## Usage

Writing and using SQL code in Go code can be inneficient for several reasons. Not only it diminishes the advantages of a statically typed language, it also worsens editor support for syntax highlighting and indentation, and hinders query reusability with other tools. The `queries` library has been created to address these issues by providing a better way to manage and use SQL queries in Go.


With `queries` you define your SQL code in files like `sql/users.sql`

```sql
-- name: get-user-by-id
SELECT *
FROM users 
WHERE user_id = :user_id AND deleted_at is null

-- name: update-user-last-login
UPDATE users
SET last_login_at = current_timestamp
WHERE user_id = :user_id
```

Load all the respective files into QueryStore using three different ways

```go
err = queryStore.LoadFromFile("sql/users.sql")
if err != nil {
  return err
}

err = queryStore.LoadFromDir("sql/")
if err != nil {
  return err
}

//go:embed sql/*
var sqlFS embed.FS
err = queryStore.LoadFromEmbed("sql/")
if err != nil {
  return err
}

```

Once you get the query loaded you can access them by their name and prepare the named parameter mapping 


```go
getUser := queryStore.MustHaveQuery("get-user-by-id")
args := getUser.Prepare(map[string]interface{}{
  "user_id": 123,
})

err = db.Get(&user, getUser.Query(), args...)
if err != nil {
  return err
}
```

## Query format

The recommende use of the `queries` library is to switch from the default positional parameter notation ($1, $2, etc. - dollar quited sign followed by the parameter position) to [psql variable definition](https://www.postgresql.org/docs/current/app-psql.html#APP-PSQL-VARIABLES).

The benefit of the variable definition is better visual control. Other aspect is the inter-operability with other PostgreSQL tools. Notably [regresql](https://github.com/dimitri/regresql).

If you prefer the default dolar sign positional parameters, you can skip the argument preparation (`queryStore.Prepare`) and use the `query.Raw`.

## Credits

The `queries` library is heavily influenced (and in some cases re-uses part of the logic) by

* [regresql](https://github.com/dimitri/regresql)
* [yesql](https://github.com/krisajenkins/yesql)
