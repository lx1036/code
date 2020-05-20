
# for dev in local
# 如果不是 root 用户执行 nginx 进程，会报错：
# nginx: [warn] the "user" directive makes sense only if the master process runs with super-user privileges
sudo openresty -p "$(pwd)" -c nginx.conf


