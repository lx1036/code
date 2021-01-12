

# 安装gitlab-runner

```shell

 # 1. 下载rpm包，https://gitlab-ci-multi-runner-downloads.s3.amazonaws.com/v1.11.5/index.html
 wget https://gitlab-ci-multi-runner-downloads.s3.amazonaws.com/v1.11.5/rpm/gitlab-ci-multi-runner_amd64.rpm
 
 # 2. 安装rpm
 sudo rpm -ivh gitlab-ci-multi-runner_amd64.rpm


 # 上面会在这台机器上面创建一个用户名为：gitlab-runner的用户，之后执行CI任务，都是这个用户权限下执行的，包括pull docker image，push docker image等。
 # 因此，如果是docker打包，我们必须将这个用户加入到docker组中；并且，gitlab-runner必须有push镜像的权限。


 # 3. 使用有 sudo 权限的账号 将 gitlab-runner 加入到 docker 组
 sudo usermod -a -G docker gitlab-runner


 # 4. gitlab-runner 登录 docker，已经找云平台同事申请了一个用户名专门用于CICD, myusernmae:mypassword
 sudo su gitlab-runner 
 docker login r.a.b.c --username myusername 


 # 5. 注册 gitlab-runner，参考：
 # 打开对应项目 gitlab 右上角 “设置  Runners“
 sudo chmod +x /usr/bin/gitlab-runner
 sudo gitlab-runner register  # 一定要用 sudo
 之后分别输入：
 1. https://xxx.gitlab.com/ci
 2. Runners 页面中的 token 串
 3. 对应描述
 4. 对应标签
 5. true, 否则非tag无法build
 6. shell，这个根据情况选择即可


 # 6. 运行 Runner
 sudo gitlab-runner start # 也要用 sudo


 # 7. 在 gitlab 对应项目 “设置  Runners”已经有配置的 Runner 了。


```




