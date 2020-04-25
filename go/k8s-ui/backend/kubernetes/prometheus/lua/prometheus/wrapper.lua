
local function empty(var)
    if type(var) == "table" then
        return next(var) == nil
    end

    return var == nil or var == '' or not var
end

local function inTable(needle, table_name)
    if type(needle) ~= "string" or type(table_name) ~= "table" then
        return false
    end

    for k, v in ipairs(table_name) do
        if v == needle then
            return true
        end
    end

    return false
end

local _M = { _VERSION = "1.1.4" }
_M.CONF = {
    initted = false,
    app = "default",
    idc = "",
    monitor_switch = {
        METRIC_COUNTER_RESPONSES = {},
        METRIC_HISTOGRAM_LATENCY = {},
        METRIC_COUNTER_SENT_BYTES = {},
        METRIC_COUNTER_RECEIVED_BYTES = {},
        METRIC_COUNTER_EXCEPTION = true,
        METRIC_GAUGE_CONNECTS = true,
    },
    log_method = {},
    buckets = {},
    merge_path = false,
    debug = false
}

function _M:init(config)
    for k, v in pairs(config) do
        if k == "app" then
            if type(v) ~= "string" then
                return nil, "'app' must be string"
            end
            self.CONF.app = v
        elseif k == "idc" then
            if type(v) ~= "string" then
                return nil, "'idc' must be string"
            end
            self.CONF.idc = v
        elseif k == "log_method" then
            if type(v) ~= "table" then
                return nil, '"log_method" must be a table'
            end
            self.CONF.log_method = v
        elseif k == "buckets" then
            if type(v) ~= "table" then
                return nil, '"buckets" must be a table'
            end
            self.CONF.buckets = v
        elseif k == "monitor_switch" then
            if type(v) ~= "table" then
                return nil, '"monitor_switch" must be a table'
            end
            for i, j in pairs(v) do
                if type(self.CONF.monitor_switch[i]) == "table" then
                    self.CONF.monitor_switch[i] = j
                end
            end
        elseif k == "merge_path" then
            if type(v) ~= "string" then
                return nil, '"merge_path" must be a string'
            end
            self.CONF.merge_path = v
        elseif k == "debug" then
            if type(v) ~= "boolean" then
                return nil, '"debug" must be a boolean'
            end
            self.CONF.debug = v
        end
    end

    if self.CONF.debug == false then
        local config = ngx.shared.prometheus_metrics
        config:flush_all()
    end

    -- "prometheus_metrics" 必须与 nginx.conf 中的 lua_shared_dict 的map名字相同，不是随便取的
    local prometheus = require("prometheus").init("prometheus_metrics")

    -- module_responses QPS
    -- QPS
    if not empty(self.CONF.monitor_switch.METRIC_COUNTER_RESPONSES) then
        self.metric_requests = prometheus:counter(
            "module_responses",
            "[" .. self.CONF.idc .. "] number of /path",
            {"app", "api", "module", "method", "code"}
        )
    end

    -- response_duration_milliseconds
    -- 延迟
    if not empty(self.CONF.monitor_switch.METRIC_HISTOGRAM_LATENCY) then
        self.metric_latency = prometheus:histogram(
                "response_duration_milliseconds",
                "[" .. self.CONF.idc .. "] http request latency",
                {"app", "api", "module", "method"},
                self.CONF.buckets
        )
    end

    -- status from ngx_http_stub_status_module module
    if not empty(self.CONF.monitor_switch.METRIC_GAUGE_CONNECTS) then
        self.metric_connections = prometheus:gauge("module_connections", "[" .. self.CONF.idc .. "] state",
            {"app", "state"})
    end

    -- module_received_bytes
    if not empty(self.CONF.monitor_switch.METRIC_COUNTER_RECEIVED_BYTES) then
        self.metric_traffic_in = prometheus:counter("module_received_bytes", "[" .. self.CONF.idc .. "] traffic in of /path",
            {"app", "api", "module", "method", "code"})
    end

    -- module_sent_bytes
    if not empty(self.CONF.monitor_switch.METRIC_COUNTER_SENT_BYTES) then
        self.metric_traffic_out = prometheus:counter("module_sent_bytes", "[" .. self.CONF.idc .. "] traffic out of /path",
            {"app", "api", "module", "method", "code"})
    end



    -- nginx_metric_errors_total



    self.CONF.initted = true
    self.prometheus = prometheus

    return self.CONF.initted
end

-- Collect metrics to response from "prometheus_metrics" Lua dictionary object
function _M:metrics()
--    local ip = ngx.var.remote_addr or ""

    if self.metric_connections then
        -- http://nginx.org/en/docs/http/ngx_http_stub_status_module.html
        -- This configuration creates a simple web page with basic status data which may look like as follows:
        -- Active connections: 291
        -- server accepts handled requests
        -- 16630948 16630948 31070465
        -- Reading: 6 Writing: 179 Waiting: 106
        self.metric_connections:set(ngx.var.connections_active, {self.CONF.app, "activing"})
        self.metric_connections:set(ngx.var.connections_reading, {self.CONF.app, "reading"})
        self.metric_connections:set(ngx.var.connections_waiting, {self.CONF.app, "waiting"})
        self.metric_connections:set(ngx.var.connections_writing, {self.CONF.app, "writing"})
    end

    self.prometheus:collect()
end

-- Write metrics data into "prometheus_metrics" Lua dictionary object
function _M:log(app)
    if not self.CONF.initted then
        return nil, "init first"
    end

    local method = ngx.var.request_method or ""
    local request_uri = ngx.var.request_uri or ""
    local status = ngx.var.status or ""

    if inTable(method, self.CONF.log_method) then
        local uri = self:isLogUri(request_uri, "METRIC_COUNTER_RESPONSES")
        if self.metric_requests and uri then
            self.metric_requests:incrBy(1, {app, uri, "self", method, status})
        end

        local uri = self:isLogUri(request_uri, "METRIC_COUNTER_SENT_BYTES")
        if self.metric_traffic_out and uri then
            -- ngx_http_log_module
            -- http://nginx.org/en/docs/http/ngx_http_log_module.html
            -- $bytes_sent: the number of bytes sent to a client
            self.metric_traffic_out:inc(tonumber(ngx.var.bytes_sent), {app, uri, "self", method, status})
        end

        local uri = self:isLogUri(request_uri, "METRIC_COUNTER_RECEIVED_BYTES")
        if self.metric_traffic_in and uri then
            -- ngx_http_log_module
            -- http://nginx.org/en/docs/http/ngx_http_log_module.html
            -- $request_length: request length (including request line, header, and request body)
            self.metric_traffic_in:inc(tonumber(ngx.var.request_length), {app, uri, "self", method, status})
        end

        local uri = self:isLogUri(request_uri, "METRIC_HISTOGRAM_LATENCY")
        if self.metric_latency and uri then
            local time = (ngx.now() - ngx.req.start_time()) * 1000
--            self.metric_latency:observe(time, {app, uri, "self", method})
        end
    end
end

function _M:isLogUri(request_uri, monitor_key)

end



return _M
