

# backtracking


回溯算法框架(https://mp.weixin.qq.com/s/xzkv1d-BnPzZ2K9oC0enSg):
```shell
func backtrack(选择列表,路径) {
   if 结束条件 {
       得到一种结果
   }
   for i in 选择列表 {
      if 减支条件 {
         continue
      }
      选择列表加入路径
      backtrack(选择列表,路径)
      撤销选择
   }
}
```
