# Queries

The `queries` library for Go implements a file based SQL query management system. Only PostgreSQL is supported.

## Installing

```
  go get -u github.com/radim/queries
```

## Usage

Including SQL code directly in Go code does not scale. It limits the benefit of statically typed languages,
has questionable the editor support (highlighting and indentation) and prevents re-usability of the queries
by other tools.


With `queries` you define your SQL code in files like `model/users.sql`

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

Load all the respective files into QueryStore

```go
err = queryStore.LoadFromFile("demo.sql")
if err != nil {
  return err
}

err = queryStore.LoadFromWalk("models/")
if err != nil {
  return err
}
```

Prepare arguments and call the query from you code 

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

The recommended use of the `queries` library is to switch from the default positional parameter notation (a dollar 
sign followed by the digits) to [psql variable definition](https://www.postgresql.org/docs/current/app-psql.html#APP-PSQL-VARIABLES).

The benefit of the variable definition is better visual control. Other aspect is the inter-operability with other PostgreSQL
tools. Notably [regresql](https://github.com/dimitri/regresql).

If you prefer the dolar sign positional parameters, you can skip the argument preparation (`queryStore.Prepare`) and 
use the `query.Raw`.

## Embedding SQL files in the binaries

Being able to distribute only a single binary is one of the benefits of Go. The `queries` library by default uses `os` interface
(specifically `os.Open` and `filepath.Walk`) to locate the files. In order to support embedded assets within Go binaries `queries`
provides integration with those tools via `OpenFunc` and `WalkImplFunc` of the `QueryStore`.

### rice integration

[rice](https://github.com/GeertJohan/go.rice) is prefered way to integrate with queries.

```
func walkTheBox(boxName string) func(string, filepath.WalkFunc) error {
  return func(_ string, walkFn filepath.WalkFunc) error {
    var err error

    for _, file := range embedded.EmbeddedBoxes[boxName].Files {
        walkFn(file.Filename, nil, err)
    }

    return nil
  }
}

  queryStore = queries.NewQueryStore()
  queryStore.WalkImplFunc = walkTheBox("models_sql")
  queryStore.OpenFunc = func(file string, load func(io.Reader) error) error {
    f, err := queryBox.Open(file)
    if err != nil {
      return err
    }

    return load(f)
  }

```

### pkger integration

When using [pkger](https://github.com/markbates/pkger) you have to include the files/locations explicitely as they 
won't get referenced.
```
  pkger -include /query1.sql -include /another
```

Plug in the `pkger` to the `QueryStore`

```go
  queryStore := queries.NewQueryStore()
	queryStore.WalkImplFunc = pkger.Walk
	queryStore.OpenFunc = func(file string, load func(io.Reader) error) error {
		f, err := pkger.Open(file)
		if err != nil {
			return err
		}

		return load(f)
	}
```

### go-bindata

Sample usage with [go-bindata](https://github.com/jteeuwen/go-bindata).

```
  go-bindata -o sql.go another/*.sql query1.sql
```

```
func walkAssets(_ string, walkFn filepath.WalkFunc) error {
  for _, name := range AssetNames() {
    info, err := AssetInfo(name)
    if err != nil {
      return err
    }

    walkFn(name, info, err)
  }

  return nil
}

// ...

queryStore := queries.NewQueryStore()
queryStore.WalkImplFunc = walkAssets
queryStore.OpenFunc = func(file string, load func(io.Reader) error) error {
  asset, err := Asset(file)
  if err != nil {
    return err
  }

  reader := bytes.NewReader(asset)

  return load(reader)
}
```

## Credits

The `queries` library is heavily influenced (and in some cases re-uses part of the logic) by

* [regresql](https://github.com/dimitri/regresql)
* [yesql](https://github.com/krisajenkins/yesql)
