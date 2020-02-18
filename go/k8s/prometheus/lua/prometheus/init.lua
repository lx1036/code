PROJECT_NAME = os.getenv("PROJECT_NAME") or "prometheus"
IDC = os.getenv("IDC") or "beijing"

local log_path = {
    '/api/metrics',
}

local ok = require("wrapper"):init({
    app = PROJECT_NAME,
    idc = IDC,
    buckets = {1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,20,22,24,26,28,30,33,36,39,42,45,49,53,57,62,67,73,79,85,92,100,101}, -- 桶距配置
    monitor_switch = {
        METRIC_COUNTER_RESPONSES = log_path,
        METRIC_COUNTER_SENT_BYTES = log_path,
        METRIC_COUNTER_RECEIVED_BYTES = log_path,
        METRIC_HISTOGRAM_LATENCY = log_path,
        METRIC_COUNTER_EXCEPTION = true,
        METRIC_GAUGE_CONNECTS = true,
    },
    log_method = { "GET", "POST" },
    merge_path = "/gometrics",
    debug = false -- 用于开发环境调试，init 时不 flush 内存。线上请关闭
})

if not ok then
    ngx.log(ngx.ERR, "prometheus init error: ")
end
