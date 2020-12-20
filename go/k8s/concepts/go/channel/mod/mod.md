
(1) https://rainbowmango.gitbook.io/go/chapter12/3-foreword/3.5-module-indirect 
间接依赖出现在go.mod文件的情况，可能符合下面所列场景的一种或多种：
* 直接依赖未启用 Go module
* 直接依赖go.mod 文件中缺失部分依赖

若要查看go.mod中某个间接依赖是被哪个依赖引入的，可以使用命令go mod why -m <pkg>来查看。
命令go mod why -m all则可以分析所有依赖的依赖链。
