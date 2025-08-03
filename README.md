# sqlc-gen-glua
This plugin allows you to generate garrysmod lua code for intereacting with the database with sqlc.

currently all types and sqlc features are not supported if there is a feature you please open an issue.

# Usage

Please follow the [sqlc docs](https://docs.sqlc.dev/en/latest/tutorials/getting-started-sqlite.html) for help defining your queries and schema 

Bellow is an example config ussing the wasm sqlc-gen-glua plugin to generate lua code.
```yaml
version: "2"
plugins:
  - name: glua
    wasm: 
      url: https://github.com/plally/sqlc-gen-glua/releases/download/v0.0.12/sqlc-gen-glua.wasm
      sha256: 9dcc2d75feebc7bbd6c1748c40a0e5202339f5e89be78ab49c666785dc1e5e98

sql:
  - engine: "sqlite"
    queries: "db/queries/"
    schema: "db/migrations/"
    codegen:
    - plugin: glua
      out: lua/myaddon/db
      options:
        global_lua_table: "MyAddon.DAL"
```
