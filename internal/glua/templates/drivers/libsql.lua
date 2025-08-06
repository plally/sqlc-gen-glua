---@class LibSQLDriver
---@field uri string
---@field token string
---@field requestQueue {request: LibSQLRequestEntry, callback: fun(results: any[], err: string|nil)}[]
local libsql = {}
libsql.__index = libsql
---@diagnostic disable

local function newLibSQL(opts)
    local tbl = {
        uri = opts.uri,
        token = opts.token,
        requestQueue = {},
    }

    return setmetatable(tbl, libsql)
end

---@class LibSQLRequest
---@field baton string|nil
---@field requests LibSQLRequestEntry[]

---@class LibSQLRequestEntry
---@field type "execute"|"close"
---@field stmt {sql: string, args: LibSQLArg[]|nil, named_args: LibSQLNamedArg[]|nil}|nil

---@class LibSQLArg
---@field type string
---@field value any

---@class LibSQLNamedArg
---@field name string
---@field value LibSQLArg

---@class LibSQLResponse
---@field baton string|nil
---@field base_url string|nil
---@field results LibSQLResponseResult[]

---@class LibSQLResponseResult
---@field type "ok"|"error"
---@field response LibSQLResponseData|nil
---@field error LibSQLResponseError|nil

---@class LibSQLResponseData
---@field type "execute"|"close"
---@field result LibSQLResponseDataResult|nil

---@class LibSQLResponseDataResult
---@field cols {name: string, decltype: string}[]
---@field rows {type: string, value: any}[]
---@field affected_row_count number
---@field last_insert_rowid number
---@field replication_index number
---@field rows_read number
---@field rows_written number
---@field query_duration_ms number

---@class LibSQLResponseError
---@field code string
---@field message string

---@param query string
---@param args any[]
---@param callback fun( results: any[], err: string|nil )
function libsql:Query(query, args, callback)
    local requestArgs = {}
    for _, arg in ipairs(args) do
        if type(arg) == "number" and math.floor(arg) == arg then
            table.insert(requestArgs, { type = "integer", value = tostring(arg) })
        elseif type(arg) == "number" and math.floor(arg) ~= arg then
            table.insert(requestArgs, { type = "float", value = arg })
        elseif type(arg) == "string" then
            table.insert(requestArgs, { type = "text", value = arg })
        end
    end

    ---@type LibSQLRequestEntry
    local req = {
        type = "execute",
        stmt = {
            sql = query,
            args = requestArgs,
        },
    }
    table.insert(self.requestQueue, {
        request = req,
        callback = callback or function() end,
    })
    self:addHook()
end

function libsql:addHook()
    self.hookIdentifier = "libsql_doRequests"
    hook.Add("Think", self.hookIdentifier, function()
        if #self.requestQueue > 0 then
            self:doRequests()
        else
            hook.Remove("Think", self.hookIdentifier)
        end
    end)
end

function libsql:doRequests()
    ---@type LibSQLRequest
    local req = {
        baton = nil,
        requests = {
        }
    }

    local callbacks = {}

    for _, r in ipairs(self.requestQueue) do
        table.insert(req.requests, r.request)
        table.insert(callbacks, r.callback)
    end

    table.insert(req.requests, {
        type = "close",
    })
    table.insert(callbacks, function() end) -- close has no callback

    file.Write("libsql_request.json", util.TableToJSON(req, true))
    self.requestQueue = {}
    HTTP {
        url = self.uri .. "/v2/pipeline",
        method = "POST",
        body = util.TableToJSON(req),
        headers = {
            Authorization = "Bearer " .. self.token,
            ["Content-Type"] = "application/json",
        },
        timeout = 500,
        success = function(code, body, headers)
            if code < 200 or code >= 300 then
                for _, callback in ipairs(callbacks) do
                    -- TODO error parsing for bad request errors
                    callback({}, "HTTP error: " .. tostring(code))
                end
                return
            end

            ---@type LibSQLResponse|nil
            local response = util.JSONToTable(body)
            if not response or not response.results then
                for _, callback in ipairs(callbacks) do
                    callback({}, "Invalid response from libsql: " .. tostring(body))
                end
                return
            end

            local results = response.results
            if not results or #results == 0 or #results ~= #callbacks then
                for _, callback in ipairs(callbacks) do
                    callback({}, "Empty response from libsql")
                end
                return
            end

            for i, result in ipairs(results) do
                callback = callbacks[i]
                local result = results[1]
                if result.type == "error" then
                    callback({}, result.error.code .. ": " .. result.error.message)
                    return
                end

                local out = {}
                for _, row in ipairs(result.response.result.rows) do
                    local newRow = {}
                    for i, col in ipairs(result.response.result.cols) do
                        local value = row[i]
                        if value.type == "integer" then
                            newRow[col.name] = tonumber(value.value)
                        elseif value.type == "float" then
                            newRow[col.name] = value.value
                        elseif value.type == "text" then
                            newRow[col.name] = value.value
                        elseif value.type == "null" then
                            newRow[col.name] = nil
                        else
                            ErrorNoHalt("Unknown value type: " .. tostring(value.type) .. " for column: " .. col.name)
                            newRow[col.name] = value.value -- fallback for other types
                        end
                    end
                    table.insert(out, newRow)
                end
                callback(out, nil)
            end
        end,
        failed = function(reason)
            for _, callback in ipairs(callbacks) do
                callback({}, "HTTP request failed: " .. tostring(reason))
            end
        end,
    }
end

GMAudit.DAL:RegisterDriver("libsql", {
    constructor = newLibSQL,
})
