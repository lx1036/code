

```shell
# 第一个终端
etcdctl lock mutex1
# 输出 mutex1/694d77b7e8a8cb64

# 第二个终端
etcdctl lock mutex1
# 关闭第一个终端，输出 mutex1/694d77b7e8a8cb64

```


















