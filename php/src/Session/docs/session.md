**[PHP Session 扩展介绍](http://php.net/manual/zh/intro.session.php)**


**[配置](http://php.net/manual/zh/session.configuration.php)**

# PHP 内置 Session

**状态**: 默认配置是 session.auto_start = 0，session 初始关闭状态，使用 **[session_start](http://php.net/manual/zh/function.session-start.php)**
开启或重用 session。session.save_handler = files 配置默认使用 files 来 store/retrieve data。

**Session ID**: session_id() 获取当前 Session ID。

**开启后，如何获取 session 数据**：当会话自动开始或者通过 session_start() 手动开始的时候， 
PHP 内部会调用会话管理器的 open 和 read 回调函数。 会话管理器可能是 PHP 默认的， 
也可能是扩展提供的（SQLite 或者 Memcached 扩展）， 
也可能是通过 session_set_save_handler() 设定的用户自定义会话管理器。 
通过 read 回调函数返回的现有会话数据（使用特殊的序列化格式存储）， PHP 会自动反序列化数据并且填充 $_SESSION 超级全局变量。


