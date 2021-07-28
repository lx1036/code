# air
本地安装 **[air](https://github.com/cosmtrek/air)** 后，执行如下命令可以享受 live reload 功能：
```shell script
curl -fLo ~/.air https://raw.githubusercontent.com/cosmtrek/air/master/bin/darwin/air
chmod +x ~/.air
mv ~/.air /usr/local/bin/air
air -c ./air.conf
```


(1)查看哪个进程占用了指定端口
```shell
lsof -Pni4 | grep :9000
```
