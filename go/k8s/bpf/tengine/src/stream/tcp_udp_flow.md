

访问 `nc localhost 5001` 来访问 tcp server，其流程是：

```
main() -> ngx_single_process_cycle() -> ngx_process_events_and_timers(),ngx_process_events() 
-> ngx_kqueue_process_events(),ev->handler(ev) -> ngx_event_accept(),ls->handler(c)
-> ngx_stream_init_connection(c),event->handler(event) -> ngx_stream_session_handler(event)
-> ngx_stream_core_run_phases(session) -> ngx_stream_core_content_phase(session, phase_handler)
-> ngx_stream_return_handler(session) -> ngx_stream_return_write_handler() -> ngx_stream_write_filter(session)
-> ngx_stream_finalize_session(session)
```




