
**[Linux 日志切割神器 logrotate 原理介绍和配置详解](https://wsgzao.github.io/post/logrotate/)**

# logrotate
logrotate 是一个 linux 系统日志的管理工具。
可以对单个日志文件或者某个目录下的文件按时间 / 大小进行切割，压缩操作；指定日志保存数量；还可以在切割之后运行自定义命令。
logrotate 是基于 crontab 运行的，所以这个时间点是由 crontab 控制的，具体可以查询 crontab 的配置文件 /etc/anacrontab。 
系统会按照计划的频率运行 logrotate，通常是每天。在大多数的 Linux 发行版本上，计划每天运行的脚本位于 /etc/cron.daily/logrotate。
如果找不到，可以 `apt/yum install -y logrotate` 。
logrotate的配置文件：
```shell script
# logrotate 定时任务配置
sudo cat /etc/cron.daily/logrotate
# logrotate 配置
sudo cat /etc/logrotate.conf
```
