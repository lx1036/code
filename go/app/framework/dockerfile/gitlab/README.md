# 在 Raspberry 安装 Gitlab
软件仓库：https://packages.gitlab.com/gitlab/raspberry-pi2
```shell script
curl -o gitlab.sh -sS https://packages.gitlab.com/install/repositories/gitlab/raspberry-pi2/script.deb.sh
# stretch 是 dist 发行版本名称
sudo curl -sS https://packages.gitlab.com/install/repositories/gitlab/raspberry-pi2/script.deb.sh | sudo os=raspbian dist=stretch bash
sudo apt install gitlab-ce
```
