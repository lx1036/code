
local log_path = {
    "/api/v1/hello"
}

local ok, err = require("wrapper"):init({
    app = os.getenv('PROJECT_NAME') or "meetup-2020-04-18",
    idc = os.getenv('IDC') or "dev",
    monitor_switch = {
        METRIC_COUNTER_RESPONSES = log_path, -- QPS
        METRIC_HISTOGRAM_LATENCY = log_path, -- 延迟 P95/P99
        METRIC_COUNTER_SENT_BYTES = log_path, -- 流量 out
        METRIC_COUNTER_REVD_BYTES = log_path, -- 流量 in
        METRIC_COUNTER_EXCEPTION = true, -- 程序异常计数器
        METRIC_GAUGE_CONNECTS = true, -- 程序状态 nginx connections
    },
    log_method = { "GET", "POST", "PUT" },
    merge_path = "/gometrics",
    buckets = { 10, 11, 13, 15, 17, 19, 22, 25, 28, 32, 36, 41, 47, 54, 62, 71, 81, 92, 105, 120, 137, 156, 178, 203, 231, 263, 299, 340, 387, 440, 500 }, -- 桶距配置
    debug = false -- 用于开发环境调试，init 时不 flush 内存。线上请关闭
})
if not ok then
    ngx.log(ngx.ERR, "prometheus init error: ", err)
end
