Argo
====

[![Build Status](https://travis-ci.org/aodin/argo.svg)](https://travis-ci.org/aodin/argo)

A REST API in Go.

Given an [aspect](https://github.com/aodin/aspect) schema:

```go
import (
    sql "github.com/aodin/aspect"
)

var usersDB = sql.Table("users",
    sql.Column("id", sql.Integer{}),
    sql.Column("name", sql.String{}),
    sql.Column("age", sql.Integer{}),
    sql.Column("password", sql.String{}),
    sql.PrimaryKey("id"),
)
```

Create a REST resource with the table and a `aspect.Connection`:

```go
users := Resource(
    conn,
    Table(usersDB),
)
```

Happy Hacking,

aodin
