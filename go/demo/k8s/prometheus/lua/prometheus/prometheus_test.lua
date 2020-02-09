
luaunit = require("luaunit")
prometheus = require("prometheus")


local SimpleDict = {}
SimpleDict.__index = SimpleDict
function SimpleDict:set(k, v)
    if not self.dict then
        self.dict = {}
    end
    self.dict[k] = v
    return true, nil, false
end

function SimpleDict:get(k)
    -- simulate an error
    if k == "gauge2{f2=\"dict_error\",f1=\"dict_error\"}" then
        return nil, 0
    end
    return self.dict[k], 0 -- value, flags
end

function SimpleDict:get_keys(k)
    local keys = {}
    for key, _ in pairs(self.dict) do
        table.insert(keys, key)
    end

    return keys[k]
end

local Nginx = {}
Nginx.__index = Nginx
Nginx.ERR = {}
Nginx.WARN = {}
Nginx.header = {}
function Nginx:log(level, ...)
    if not ngx.logs then
        ngx.logs = {}
    end

    table.insert(ngx.logs, table.concat({...}, " "))
end
function Nginx.print(printed)
    if not ngx.printed then ngx.printed = {} end
    for str in string.gmatch(table.concat(printed, ""), "([^\n]+)") do
        table.insert(ngx.printed, str)
    end
end

TestPrometheus = {}
function TestPrometheus:setUp()
    self.dict = setmetatable({}, SimpleDict)
    ngx = setmetatable({shared={metrics=self.dict}}, Nginx)
    self.prometheus = prometheus.init("metrics")
    self.counter1 = self.prometheus:counter("metric1", "Metric 1")
    self.counter2 = self.prometheus:counter("metric2", "Metric 2", {"f1", "f2"})
    self.gauge1 = self.prometheus:gauge("gauge1", "Gauge 1")
    self.gauge2 = self.prometheus:gauge("gauge2", "Gauge 2", {"f1", "f2"})
    self.histogram1 = self.prometheus:histogram("l1", "Histogram 1")
    self.histogram2 = self.prometheus:histogram("l2", "Histogram 2", {"var", "site"})
end

function TestPrometheus:testInit()
    print(luaunit.prettystr(ngx), luaunit.prettystr(self.dict), luaunit.prettystr(self.dict:get("nginx_metric_errors_total")))
    luaunit.assertEquals(self.dict:get("nginx_metric_errors_total"), 0)
    luaunit.assertEquals(ngx.logs, nil)
end

function TestPrometheus:testErrorInitialized()
--    local p = prometheus
--    p:counter("metric1")
--    p:histogram("metric2")
--    p:gauge("metric3")
--    p:metric_data()
--    luaunit.assertEquals(#ngx.logs, 4)
end

function TestPrometheus:testErrorUnknownDict()
    local p = prometheus.init("nonexist-dict")
    luaunit.assertEquals(p.initialized, false)
    luaunit.assertEquals(#ngx.logs, 1)
    print(luaunit.prettystr(ngx.logs))
    luaunit.assertStrContains(ngx.logs[1], "does not exist")
end

function TestPrometheus:testErrorNoMemory()
    local counter = self.prometheus:counter("notfit")
    self.counter1.inc(5)
    counter:inc(1)

    luaunit.assertEquals(self.dict:get("metric1"), 5)
    luaunit.assertEquals(self.dict:get("nginx_metric_errors_total"), 1)
    luaunit.assertEquals(self.dict:get("willnotfit"), nil)
    luaunit.assertEquals(#ngx.logs, 1)
end


os.exit(luaunit.LuaUnit.run())
