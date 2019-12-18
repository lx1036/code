-- e.g. module_connections{app="prometheus",state="reading"}
local function full_metric_name(name, label_names, lable_values)
    if not label_names then
        return name
    end

    local label_parts = {}
    for idx, key in ipairs(label_names) do
        local label_value = (string.format("%s", lable_values[idx])
            :gsub("[^\032-\126]", "")
            :gsub("\\", "\\\\")
            :gsub('"', '\\"'))
        table.insert(label_parts, key .. '="' .. label_value .. '"')
    end

    return name .. "{" .. table.concat(label_parts, ",") .. "}"
end

-- e.g. module_connections{app="prometheus",state="reading"} -> module_connections
local function short_metric_name(full_name)
    local labels_start, _ = full_name:find("{")
    if not labels_start then
        return full_name
    end
    local suffix_idx, _ = full_name:find("_bucket{")
    if suffix_idx and full_name:find("le=") then
        -- histogram
        return full_name:sub(1, suffix_idx - 1)
    end
    return full_name:sub(1, labels_start - 1)
end

local Metric = {}
function Metric:new(config)
    config = config or {}
    setmetatable(config, self)
    self.__index = self

    return config
end

function Metric:check_label_values(label_values)
    if label_values == nil and self.label_names == nil then
        return
    elseif label_values == nil and self.label_names ~= nil then
        return "Expected " .. #self.label_names .. "labels for " .. self.name .. ", get none"
    else
        for k, v in ipairs(self.label_names) do
            if label_values[k] == nil then
                return ""
            end
        end
    end
end

local Gauge = Metric:new()

function Gauge:set(value, label_values)
    self.prometheus:set(self.name, self.label_names, label_values, value)
end

local Counter = Metric:new()

-- Args: {value: 1, label_values: {value1, value2}}
function Counter:inc(value, label_values)
    local err = self:check_label_values(label_values)
    if err ~= nil then
        self.prometheus:log_error(err)
    end

    if value ~= nil and value < 0 then
        self.prometheus:log_error_kv(self.name, value, "value can't be negative")
    end

    self.prometheus:inc(self.name, self.label_names, label_values, value or 1)
end

local Histogram = Metric:new()

function Histogram:observe(value, label_values)

    self.prometheus:observe(self.name, self.label_names, label_values, value)
end

local Prometheus = {}
Prometheus.__index = Prometheus
Prometheus.initialized = false

function Prometheus.init(dict_name, prefix)
    local self = setmetatable({}, Prometheus)
    -- https://github.com/openresty/lua-nginx-module#ngxshareddict
    self.dict = ngx.shared[dict_name or "prometheus_metrics"]
    if self.dict == nil then
        ngx.log(ngx.ERR, "Dictionary ", dict_name, "does not exist, define it using `lua_shared_dict` directive.")
        return self
    end

    if prefix then
        self.prefix = prefix
    else
        self.prefix = ''
    end

    self.registered = {}
    self.type = {}
    self.help = {}

    self.buckets = {}
    self.bucket_format = {}
    self.initialized = true

    self:counter("nginx_metric_errors_total", "Number of nginx-lua-prometheus errors")
    -- https://github.com/openresty/lua-nginx-module#ngxshareddictset
    self.dict:set("nginx_metric_errors_total", 0)
    return self
end

-- Register a Gauge object
function Prometheus:gauge(name, description, label_names)
    if self.registered[name] then
        self:log_error("Duplicate metric" .. name)
        return
    end

    self.registered[name] = true
    self.help[name] = description
    self.type[name] = "gauge"

    return Gauge:new{name=name, label_names=label_names, prometheus=self}
end

-- Register a Counter object
-- Args:
-- {name: "name1", label_names: {"label1", "label2"}, description: "description1"}
-- Return: Counter
function Prometheus:counter(name, description, label_names)
    if self.registered[name] then
        self:log_error("Duplicate metric" .. name)
        return
    end

    self.registered[name] = true
    self.help[name] = description
    self.type[name] = "counter"

    return Counter:new{name=name, label_names=label_names, prometheus=self}
end

-- Default set of latency buckets, 5ms to 10s:
local DEFAULT_BUCKETS = {0.005, 0.01, 0.02, 0.03, 0.05, 0.075, 0.1, 0.2, 0.3,
    0.4, 0.5, 0.75, 1, 1.5, 2, 3, 4, 5, 10}

function Prometheus:histogram(name, description, label_names, buckets)
    if self.registered[name] then
        self:log_error("Duplicate metric" .. name)
        return
    end

    self.registered[name] = true
    self.help[name] = description
    self.type[name] = "histogram"
    self.buckets[name] = buckets or DEFAULT_BUCKETS

    return Histogram:new{name=name, label_names=label_names, prometheus=self}
end

function Prometheus:metric_data()
    if not self.initialized then
        ngx.log(ngx.ERR, "Prometheus module has not been initialized")
    end

    print(self.dict)
    local keys = self.dict:get_keys(0)
    table.sort(keys)

    print(keys)
end

function Prometheus.log_error(...)
    ngx.log(ngx.ERR, ...)
    self.dict:incr("nginx_metric_errors_total", 1)
end

function Prometheus.log_error_kv(key, value, err)
    self:log_error("Error while setting ", key, " to ", value, ": ", err)
end

function Prometheus:set(name, label_names, label_values, value)
    local key = full_metric_name(name, label_names, label_values)
    self:set_key(key, value)
end

function Prometheus:set_key(key, value)
    local ok, err = self.dict:safe_set(key, value)
    if not ok then
        self:log_error_kv(key, value, err)
    end
end

function Prometheus:inc(name, label_names, label_values, value)
    local key = full_metric_name(name, label_names, label_values)
    -- https://github.com/openresty/lua-nginx-module#ngxshareddictincr
    local newValue, err = self.dict:incr(key, value)
    if newValue then
        return
    end
end

local function copy_table(table)
    local copy = {}
    if table ~= nil then
        for k, v in ipairs(table) do
            copy[k] = v
        end
    end

    return copy
end

function Prometheus:observe(name, label_names, label_values, value)
    self:inc(name .. "_count", label_names, label_values, 1)
    self:inc(name .. "_sum", label_names, label_values, value)

    local l_names = copy_table(label_names)
    local l_values = copy_table(label_values)

    table.insert(l_names, "le")
    table.insert(l_values, "Inf")

    for _, bucket in ipairs(self.buckets[name]) do
        if value <= bucket then
            self:inc(name .. "_bucket", l_names, l_values, 1)
        end
    end
end

function Prometheus:collect()
    ngx.header.content_type = "text/plain"
    -- By default, only the first 1024 keys (if any) are returned. When the <max_count> argument is given the value 0, then all the keys will be returned even there is more than 1024 keys in the dictionary.
    -- https://github.com/openresty/lua-nginx-module#ngxshareddictget_keys
    local keys = self.dict:get_keys(0)
    table.sort(keys)

    local seen_metrics = {}
    for _, key in ipairs(keys) do
        local value, err = self.dict:get(key)
        if value then
            local short_name = short_metric_name(key)
            -- ngx.log(ngx.ERR, short_name)
            if not seen_metrics[short_name] then
                if self.help[short_name] then
                    ngx.say("# HELP " .. self.prefix .. short_name .. " " .. self.help[short_name])
                end
                if self.type[short_name] then
                    ngx.say("# TYPE " .. self.prefix .. short_name .. " " .. self.type[short_name])
                end

                seen_metrics[short_name] = true
            end

            ngx.say(self.prefix ..key:gsub('le="Inf"', 'le="+Inf"'), " ", value)
        else
            self:log_error("Error getting ", key, ": ", err)
        end
    end
end

return Prometheus
