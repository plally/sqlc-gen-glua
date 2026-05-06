# sqlc-gen-glua
This plugin allows you to generate garrysmod lua code for interacting with the database with sqlc.

Currently all types and sqlc features are not supported. If there is a feature you need, please open an issue.

# Usage

Please follow the [sqlc docs](https://docs.sqlc.dev/en/latest/tutorials/getting-started-sqlite.html) for help defining your queries and schema.

Below is an example config using the wasm sqlc-gen-glua plugin to generate lua code.
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

## Options

| Option | Type | Required | Description |
|---|---|---|---|
| `global_lua_table` | string | yes | The global Lua table the DAL is attached to, e.g. `"MyAddon.DAL"` |
| `rename` | map | no | Override generated field names, e.g. `{"user_id": "userId"}` |
| `include_migrations` | bool | no | Generate a `migrations.lua` file with a migration runner (default: `false`) — **unstable, API may change** |

## Migrations

> **Unstable:** the migrations feature is under active development. The generated API may change between releases.

Setting `include_migrations: true` generates a `migrations.lua` file alongside the DAL. It is automatically included from `dal.lua` and adds a `:Migrate()` method to your DAL table.

### Supported migration formats

| Format | Description |
|---|---|
| `goose` | [pressly/goose](https://github.com/pressly/goose) — parses `-- +goose Up` / `-- +goose StatementBegin` / `-- +goose StatementEnd` markers |

### Example

```yaml
options:
  global_lua_table: "MyAddon.DAL"
  include_migrations: true
```

```lua
-- Run on server startup after the driver is set
MyAddon.DAL:UseDriver("gmod")

MyAddon.DAL:Migrate({
    path = "lua/myaddon/migrations",  -- relative to game root
    appName = "myaddon",              -- names the version table: myaddon_migrations
    migrationLanguage = "goose",
}, function(err)
    if err then
        print("Migration failed: " .. err)
    end
end)
```

Migration files must be named with a leading number followed by an underscore. Goose uses timestamps by default: `20231015120000_create_users.sql`, `20231016090000_add_email.sql`. Sequential numbering (`0001_create_users.sql`) also works. Files are applied in numeric order, each inside a transaction; a failure rolls back that migration and stops the runner.
