local gmod = {}
gmod.__index = gmod

local function newGMOD()
    local tbl = {}
    return setmetatable(tbl, gmod)
end

---@param query string
---@param args any[]
---@param callback fun( results: any[], err: string|nil )
function gmod:Query(query, args, callback)
    local rows = sql.QueryTyped(query, unpack(args or {}))
    if rows == false then
        local err = sql.LastError()
        callback({}, err)
        return
    end

    ---@diagnostic disable-next-line: param-type-mismatch
    callback(rows, nil)
end

GMAudit.DAL:RegisterDriver("gmod", {
    constructor = newGMOD,
})
