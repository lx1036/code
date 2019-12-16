
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
    self.p = prometheus.init("metrics")
--    self.counter1 = self.p:counter("metric1", "Metric 1")
--    self.counter2 = self.p:counter("metric2", "Metric 2", {"f1", "f2"})
--    self.gauge1 = self.p:gauge("gauge1", "Gauge 1")
--    self.gauge2 = self.p:gauge("gauge2", "Gauge 2", {"f1", "f2"})
--    self.histogram1 = self.p:histogram("l1", "Histogram 1")
--    self.histogram2 = self.p:histogram("l2", "Histogram 2", {"var", "site"})
end

function TestPrometheus:testInit()
    luaunit:assertEquals(self.dict:get("nginx_metric_errors_total"), 0)
    luaunit:assertEquals(ngx.logs, nil)
end






os.exit(luaunit.LuaUnit.run())
