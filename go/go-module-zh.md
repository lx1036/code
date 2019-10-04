翻译自 [go modules](https://github.com/golang/go/wiki/Modules)

# Go 1.11 Modules
Go 1.11包含[此处]（https://golang.org/design/24301-versioned-go）提出的对版本模块的初步支持。模块是Go 1.11中的一项试验性加入功能，其计划包括反馈并最终确定Go 1.14的功能（https://github.com/golang/go/issues/31857）。即使某些细节可能会更改，将来的发行版也将支持使用Go 1.11、1.12和1.13定义的模块。最初的原型“ vgo”于2018年2月发布（https://research.swtch.com/vgo)。2018年7月，版本化模块[登陆]（https://groups.google.com/d/msg / golang-dev / a5PqQuBljF4 / 61QK4JdtBgAJ）。请通过[现有或新问题]（https://github.com/golang/go/wiki/Modules#github-issues）和[体验报告]（https://github.com/golang/go）提供有关模块的反馈/维基/ ExperienceReports）。

## Recent Changes
Go 1.13中对模块进行了重大改进和更改。如果您使用模块，请务必仔细阅读Go 1.13发行说明中的[modules部分]（https://golang.org/doc/go1.13#modules）。三个值得注意的变化：1.“ go”工具现在默认从https://proxy.golang.org上的公共Go模块镜像下载模块，并且还默认针对公共Go校验和验证下载的模块（无论源如何）数据库位于https://sum.golang.org。 *如果您有私人密码，则很可能应该配置`GOPRIVATE`设置（例如`go env -w GOPRIVATE = *。corp.com，github.com / secret / repo`），或更细粒度的变体支持较少用例的`GONOPROXY`或`GONOSUMDB`。有关更多详细信息，请参见[文档]（https://golang.org/cmd/go/#hdr-Module_configuration_for_non_public_modules）。 2.如果找到任何go.mod，即使在GOPATH内部，`GO111MODULE = auto`也会启用模块模式。 （在Go 1.13之前，`GO111MODULE = auto`永远不会在GOPATH中启用模块模式）。 3.`go get`参数已更改：*`go get -u`（不带任何参数）现在仅升级当前_package_的直接和间接依赖关系，不再检查整个_module_。 *从模块根目录开始`get -u。/ ...`升级模块的所有直接和间接依赖关系，现在不包括测试依赖关系。 *`get -u -t。/ ...`类似，但是也升级了测试依赖项。 *`go get`不再支持`-m`（因为由于其他更改，它会与`go get -d`大部分重叠；您通常可以将`go get -m foo`替换为`go get -d foo` ）。请参阅[发行说明]（https://golang.org/doc/go1.13#modules），以获取有关这些更改和其他更改的更多详细信息。

## Table of Contents
对于开始使用模块的人员，“快速入门”和“新概念”部分特别重要。 “如何...”部分涵盖了有关机械的更多详细信息。此页面上内容最多的是FAQ，回答更具体的问题；至少浏览一下此处列出的FAQ一线可能是值得的。

* [Quick Start](https://github.com/golang/go/wiki/Modules#quick-start)
   * [Example](https://github.com/golang/go/wiki/Modules#example)
   * [Daily Workflow](https://github.com/golang/go/wiki/Modules#daily-workflow)
* [New Concepts](https://github.com/golang/go/wiki/Modules#new-concepts)
   * [Modules](https://github.com/golang/go/wiki/Modules#modules)
   * [go.mod](https://github.com/golang/go/wiki/Modules#gomod)
   * [Version Selection](https://github.com/golang/go/wiki/Modules#version-selection)
   * [Semantic Import Versioning](https://github.com/golang/go/wiki/Modules#semantic-import-versioning)
* [How to Use Modules](https://github.com/golang/go/wiki/Modules#how-to-use-modules)
   * [How to Install and Activate Module Support](https://github.com/golang/go/wiki/Modules#how-to-install-and-activate-module-support)
   * [How to Define a Module](https://github.com/golang/go/wiki/Modules#how-to-define-a-module)
   * [How to Upgrade and Downgrade Dependencies](https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies)
   * [How to Prepare for a Release (All Versions)](https://github.com/golang/go/wiki/Modules#how-to-prepare-for-a-release)
   * [How to Prepare for a Release (v2 or Higher)](https://github.com/golang/go/wiki/Modules#releasing-modules-v2-or-higher)
   * [Publishing a Release](https://github.com/golang/go/wiki/Modules#publishing-a-release)
* [Migrating to Modules](https://github.com/golang/go/wiki/Modules#migrating-to-modules)
* [Additional Resources](https://github.com/golang/go/wiki/Modules#additional-resources)
* [Changes Since the Initial Vgo Proposal](https://github.com/golang/go/wiki/Modules#changes-since-the-initial-vgo-proposal)
* [GitHub Issues](https://github.com/golang/go/wiki/Modules#github-issues)
* [FAQs](https://github.com/golang/go/wiki/Modules#faqs)
  * [How are versions marked as incompatible?](https://github.com/golang/go/wiki/Modules#how-are-versions-marked-as-incompatible)
  * [When do I get old behavior vs. new module-based behavior?](https://github.com/golang/go/wiki/Modules#when-do-i-get-old-behavior-vs-new-module-based-behavior)
  * [Why does installing a tool via 'go get' fail with error 'cannot find main module'?](https://github.com/golang/go/wiki/Modules#why-does-installing-a-tool-via-go-get-fail-with-error-cannot-find-main-module)
  * [How can I track tool dependencies for a module?](https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module)
  * [What is the status of module support in IDEs, editors and standard tools like goimports, gorename, etc.?](https://github.com/golang/go/wiki/Modules#what-is-the-status-of-module-support-in-ides-editors-and-standard-tools-like-goimports-gorename-etc)
* [FAQs — Additional Control](https://github.com/golang/go/wiki/Modules#faqs--additional-control)
  * [What community tooling exists for working with modules?](https://github.com/golang/go/wiki/Modules#what-community-tooling-exists-for-working-with-modules)
  * [When should I use the 'replace' directive?](https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive)
  * [Can I work entirely outside of VCS on my local filesystem?](https://github.com/golang/go/wiki/Modules#can-i-work-entirely-outside-of-vcs-on-my-local-filesystem)
  * [How do I use vendoring with modules? Is vendoring going away?](https://github.com/golang/go/wiki/Modules#how-do-i-use-vendoring-with-modules-is-vendoring-going-away)
  * [Are there "always on" module repositories and enterprise proxies?](https://github.com/golang/go/wiki/Modules#are-there-always-on-module-repositories-and-enterprise-proxies)
  * [Can I control when go.mod gets updated and when the go tools use the network to satisfy dependencies?](https://github.com/golang/go/wiki/Modules#can-i-control-when-gomod-gets-updated-and-when-the-go-tools-use-the-network-to-satisfy-dependencies)
  * [How do I use modules with CI systems such as Travis or CircleCI?](https://github.com/golang/go/wiki/Modules#how-do-i-use-modules-with-ci-systems-such-as-travis-or-circleci)
* [FAQs — go.mod and go.sum](https://github.com/golang/go/wiki/Modules#faqs--gomod-and-gosum)
  * [Why does 'go mod tidy' record indirect and test dependencies in my 'go.mod'?](https://github.com/golang/go/wiki/Modules#why-does-go-mod-tidy-record-indirect-and-test-dependencies-in-my-gomod)
  * [Is 'go.sum' a lock file? Why does 'go.sum' include information for module versions I am no longer using?](https://github.com/golang/go/wiki/Modules#is-gosum-a-lock-file-why-does-gosum-include-information-for-module-versions-i-am-no-longer-using)
  * [Should I still add a 'go.mod' file if I do not have any dependencies?](https://github.com/golang/go/wiki/Modules#should-i-still-add-a-gomod-file-if-i-do-not-have-any-dependencies)
  * [Should I commit my 'go.sum' file as well as my 'go.mod' file?](https://github.com/golang/go/wiki/Modules#should-i-commit-my-gosum-file-as-well-as-my-gomod-file)
* [FAQs — Semantic Import Versioning](https://github.com/golang/go/wiki/Modules#faqs--semantic-import-versioning)
  * [Why must major version numbers appear in import paths?](https://github.com/golang/go/wiki/Modules#why-must-major-version-numbers-appear-in-import-paths)
  * [Why are major versions v0, v1 omitted from import paths?](https://github.com/golang/go/wiki/Modules#why-are-major-versions-v0-v1-omitted-from-import-paths)
  * [What are some implications of tagging my project with major version v0, v1, or making breaking changes with v2+?](https://github.com/golang/go/wiki/Modules#what-are-some-implications-of-tagging-my-project-with-major-version-v0-v1-or-making-breaking-changes-with-v2)
  * [Can a module consume a package that has not opted in to modules?](https://github.com/golang/go/wiki/Modules#can-a-module-consume-a-package-that-has-not-opted-in-to-modules)
  * [Can a module consume a v2+ package that has not opted into modules? What does '+incompatible' mean?](https://github.com/golang/go/wiki/Modules#can-a-module-consume-a-v2-package-that-has-not-opted-into-modules-what-does-incompatible-mean)
  * [How are v2+ modules treated in a build if modules support is not enabled? How does "minimal module compatibility" work in 1.9.7+, 1.10.3+, and 1.11?](https://github.com/golang/go/wiki/Modules#how-are-v2-modules-treated-in-a-build-if-modules-support-is-not-enabled-how-does-minimal-module-compatibility-work-in-197-1103-and-111)
  * [What happens if I create a go.mod but do not apply semver tags to my repository?](https://github.com/golang/go/wiki/Modules#what-happens-if-i-create-a-gomod-but-do-not-apply-semver-tags-to-my-repository)
  * [Can a module depend on a different version of itself?](https://github.com/golang/go/wiki/Modules#can-a-module-depend-on-a-different-version-of-itself)
* [FAQs — Multi-Module Repositories](https://github.com/golang/go/wiki/Modules#faqs--multi-module-repositories)
  * [What are multi-module repositories?](https://github.com/golang/go/wiki/Modules#what-are-multi-module-repositories)
  * [Should I have multiple modules in a single repository?](https://github.com/golang/go/wiki/Modules#should-i-have-multiple-modules-in-a-single-repository)
  * [Is it possible to add a module to a multi-module repository?](https://github.com/golang/go/wiki/Modules#is-it-possible-to-add-a-module-to-a-multi-module-repository)
  * [Is it possible to remove a module from a multi-module repository?](https://github.com/golang/go/wiki/Modules#is-it-possible-to-remove-a-module-from-a-multi-module-repository)
  * [Can a module depend on an internal/ in another?](https://github.com/golang/go/wiki/Modules#can-a-module-depend-on-an-internal-in-another)
  * [Can an additional go.mod exclude unnecessary content? Do modules have the equivalent of a .gitignore file?](https://github.com/golang/go/wiki/Modules#can-an-additional-gomod-exclude-unnecessary-content-do-modules-have-the-equivalent-of-a-gitignore-file)
* [FAQs — Minimal Version Selection](https://github.com/golang/go/wiki/Modules#faqs--minimal-version-selection)
  * [Won't minimal version selection keep developers from getting important updates?](https://github.com/golang/go/wiki/Modules#wont-minimal-version-selection-keep-developers-from-getting-important-updates)
* [FAQs — Possible Problems](https://github.com/golang/go/wiki/Modules#faqs--possible-problems)
  * [What are some general things I can spot check if I am seeing a problem?](https://github.com/golang/go/wiki/Modules#what-are-some-general-things-i-can-spot-check-if-i-am-seeing-a-problem)
  * [What can I check if I am not seeing the expected version of a dependency?](https://github.com/golang/go/wiki/Modules#what-can-i-check-if-i-am-not-seeing-the-expected-version-of-a-dependency)
  * [Why am I getting an error 'cannot find module providing package foo'?](https://github.com/golang/go/wiki/Modules#why-am-i-getting-an-error-cannot-find-module-providing-package-foo)
  * [Why does 'go mod init' give the error 'cannot determine module path for source directory'?](https://github.com/golang/go/wiki/Modules#why-does-go-mod-init-give-the-error-cannot-determine-module-path-for-source-directory)
  * [I have a problem with a complex dependency that has not opted in to modules. Can I use information from its current dependency manager?](https://github.com/golang/go/wiki/Modules#i-have-a-problem-with-a-complex-dependency-that-has-not-opted-in-to-modules-can-i-use-information-from-its-current-dependency-manager)
  * [How can I resolve "parsing go.mod: unexpected module path" and "error loading module requirements" errors caused by a mismatch between import paths vs. declared module identity?](https://github.com/golang/go/wiki/Modules#how-can-i-resolve-parsing-gomod-unexpected-module-path-and-error-loading-module-requirements-errors-caused-by-a-mismatch-between-import-paths-vs-declared-module-identity)
  * [Why does 'go build' require gcc, and why are prebuilt packages such as net/http not used?](https://github.com/golang/go/wiki/Modules#why-does-go-build-require-gcc-and-why-are-prebuilt-packages-such-as-nethttp-not-used)
  * [Do modules work with relative imports like `import "./subdir"`?](https://github.com/golang/go/wiki/Modules#do-modules-work-with-relative-imports-like-import-subdir)
  * [Some needed files may not be present in populated vendor directory](https://github.com/golang/go/wiki/Modules#some-needed-files-may-not-be-present-in-populated-vendor-directory)

## Quick Start

#### Example
详细信息将在本页面的其余部分中介绍，但这是一个从头开始创建模块的简单示例。在GOPATH之外创建目录，并可以选择初始化VCS
```
$ mkdir -p /tmp/scratchpad/repo
$ cd /tmp/scratchpad/repo
$ git init -q
$ git remote add origin https://github.com/my/repo
```

Initialize a new module:
```
$ go mod init github.com/my/repo

go: creating new go.mod: module github.com/my/repo
```

Write your code:
```
$ cat <<EOF > hello.go
package main

import (
    "fmt"
    "rsc.io/quote"
)

func main() {
    fmt.Println(quote.Hello())
}
EOF
```

Build and run:
```
$ go build -o hello
$ ./hello

Hello, world.
```
The `go.mod` file was updated to include explicit versions for your dependencies, where `v1.5.2` here is a [semver](https://semver.org) tag:
```
$ cat go.mod

module github.com/my/repo

require rsc.io/quote v1.5.2
```

#### Daily Workflow

请注意，在上面的示例中不需要“ go get”。典型的日常工作流程可以是：*根据需要将导入语句添加到`.go`代码中。 *标准命令“ go build”或“ go test”将根据需要自动添加新的依赖关系，以实现导入（更新“ go.mod”并下载新的依赖关系）。 *必要时，可以使用诸如go get foo@v1.2.3，go get foo @ master（带有商业性的foo @ tip），go get foo @ e3702bed2等命令选择更具体的依赖版本。 ，或直接编辑`go.mod`。您可能会使用的其他常见功能的简要介绍


* `go list -m all` — View final versions that will be used in a build for all direct and indirect dependencies ([details](https://github.com/golang/go/wiki/Modules#version-selection))
* `go list -u -m all` — View available minor and patch upgrades for all direct and indirect dependencies ([details](https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies))
* `go get -u ./...` or `go get -u=patch ./...` (from module root directory) — Update all direct and indirect dependencies to latest minor or patch upgrades (pre-releases are ignored) ([details](https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies))
* `go build ./...` or `go test ./...` (from module root directory) — Build or test all packages in the module ([details](https://github.com/golang/go/wiki/Modules#how-to-define-a-module))
* `go mod tidy` — Prune any no-longer-needed dependencies from `go.mod` and add any dependencies needed for other combinations of OS, architecture, and build tags ([details](https://github.com/golang/go/wiki/Modules#how-to-prepare-for-a-release))
* `replace` directive or `gohack` — Use a fork, local copy or exact version of a dependency ([details](https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive))
* `go mod vendor` — Optional step to create a `vendor` directory ([details](https://github.com/golang/go/wiki/Modules#how-do-i-use-vendoring-with-modules-is-vendoring-going-away))

在阅读了有关“新概念”的下四个部分之后，您将获得足够的信息来开始使用大多数项目的模块。回顾上面的[目录]（https://github.com/golang/go/wiki/Modules#table-of-contents）（包括那里的FAQ常见问题）也很有用，以使您熟悉更详细的主题列表。 
##新概念
这些部分对主要的新概念进行了高级介绍。有关更多详细信息和原理，请观看此40分钟的介绍性视频[Russ Cox的视频，描述了设计的原理]（https://www.youtube.com/watch?v=F8nrpe0XWRg&amp;list=PLq2Nv-Sh8EbbIjQgDzapOFeVfv5bGOoPE&amp;index=3&amp;t=0s） ，[官方提案文档]（https://golang.org/design/24301-versioned-go）或更详细的初始[vgo博客系列]（https://research.swtch.com/vgo）。 
###模块
* module *是相关的Go软件包的集合，这些版本一起作为一个单元进行版本控制。模块记录精确的依赖要求并创建可复制的构建。通常，版本控制存储库仅包含在存储库根目录中定义的一个模块。 （[单个存储库中支持多个模块]（https://github.com/golang/go/wiki/Modules#faqs--multi-module-repositories），但通常会导致在持续性基础，而不是每个存储库使用单个模块）。总结存储库，模块和软件包之间的关系：*存储库包含一个或多个Go模块。 *每个模块包含一个或多个Go软件包。 *每个软件包在一个目录中包含一个或多个Go源文件。模块必须根据[semver]（https://semver.org/）进行语义版本控制，
通常采用`v（major）。（minor）。（patch）`的形式，例如`v0.1.0`，`v1 .2.3”或“ v1.5.0-rc.1”。前导v是必需的。如果使用Git，则[tag]（https://git-scm.com/book/en/v2/Git-Basics-Tagging）发布的版本及其版本。公共和私有模块存储库和代理变得可用（请参阅常见问题解答[如下]（https://github.com/golang/go/wiki/Modules#are-there-always-on-module-repositories-and-enterprise-proxies ））。 ### go.mod模块由Go源文件树定义，该树的根目录中带有`go.mod`文件。
模块源代码可能位于GOPATH之外。有四个指令：`module`，`require`，`replace`，`exclude`。这是定义模块github.com/my/thing的示例go.mod文件：模块通过提供_module path_的module指令在其go.mod中声明其身份。模块中所有软件包的导入路径将模块路径共享为公共前缀。模块路径和从`go.mod`到软件包目录的相对路径共同决定了软件包的导入路径。例如，如果要为存储库“ github.com/my/repo”创建一个模块，该模块将包含两个带有导入路径“ github.com/my/repo/foo”和“ github.com/my/repo/”的软件包。 ，然后通常在go.mod文件中的第一行将模块路径声明为module github.com/my/repo，并且相应的磁盘结构可能是：在Go源代码中，软件包为使用完整路径（包括模块路径）导入。例如，如果某个模块在其“ go.mod”中声明其身份为“ module example.com/my/module”，则消费者可以这样做：“ exclude”和“ replace”指令仅在当前（“ main” ）模块。构建主模块时，将忽略除主模块以外的其他模块中的`exclude`和`replace`指令。因此，`replace`和`exclude`语句允许主模块完全控制其自身的构建，而不受依赖项的完全控制。 （有关何时使用`replace`指令的讨论，请参见FAQ [以下]（https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive）） 。

### Version Selection
如果您在源代码中添加了一个新的导入，但尚未被go.mod中的&#39;require&#39;覆盖，则大多数go命令（例如“ go build”和“ go test”）将自动查找适当的模块并添加该新直接依赖项的*最高*版本，作为`require`指令对模块的`go.mod`的依赖。例如，如果您的新导入对应于依赖项M，其最新标记的发布版本为“ v1.2.3”，则模块的“ go.mod”将以“ require M v1.2.3”结尾，这表明模块M是具有以下项的依赖项：允许的版本&gt; = v1.2.3（并且&lt;v2，因为v2被认为与v1不兼容）。最小版本选择算法用于选择构建中使用的所有模块的版本。对于构建中的每个模块，通过最小版本选择选择的版本始终是主模块中的“ require”指令或其依赖项之一在语义上*最高*的版本。例如，如果您的模块依赖于具有“ require D v1.0.0”的模块A，而您的模块也依赖于具有“ require D v1.1.1”的模块B，则最小版本选择应选择“ v1”。 D中包含的D的1.1版本（鉴于它是列出的最高版本的require版本）。即使以后一段时间D的“ v1.2.0”可用，对“ v1.1.1”的选择仍保持一致。这是模块系统如何提供100％可复制构建的示例。准备就绪后，模块作者或用户可以选择升级到D的最新可用版本或为D选择一个显式版本。有关最小版本选择算法的简要原理和概述，请参见[高保真度构建]部分]。 （https://github.com/golang/proposal/blob/master/design/24301-versioned-go.md#update-timing--high-fidelity-builds）或参阅[更详细的vgo博客系列]（https://research.swtch.com/vgo）。要查看所选模块版本的列表（包括间接依赖性），请使用“ go list -m all”。另请参见下面的[“如何升级和降级依赖关系”]（https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies）部分和[“如何版本标记为不兼容？“]（https://github.com/golang/go/wiki/Modules#how-are-versions-marked-as-incompatible）以下常见问题解答。多年以来，官方的Go FAQ一直在软件包版本管理中包含以下建议：最后一句话特别重要-如果破坏兼容性，则应更改软件包的导入路径。对于Go 1.11模块，该建议已正式化为_import兼容性规则_：回想[semver]（https://semver.org/），当v1或更高版本的程序包进行向后不兼容的更改时，需要对版本进行重大更改。遵循导入兼容性规则和semver的结果称为_Semantic Import Versioning_，其中主要版本包含在导入路径中-这样可确保在主要版本由于兼容性中断而增加时，导入路径都会更改。由于语义导入版本控制，选择加入Go模块的代码**必须遵守以下规则**：*请遵循[semver]（https://semver.org/）。 （VCS标签示例为“ v1.2.3”）。 *如果模块的版本为v2或更高版本，则模块_must_的主要版本将作为`/ vN`包含在`go.mod&#39;文件中使用的模块路径的末尾（例如，`module github.com/my / mod / v2`，&#39;require github.com/my/mod/v2 v2.0.1&#39;）和包导入路径（例如，“ import” github.com/my/mod/v2/mypkg“`）。这包括`get get&#39;命令中使用的路径（例如，`get get github.com / my / mod / v2 @ v2.0.1`。请注意，其中既有`/ v2`又有`@ v2.0.1`。考虑这个问题的一种方法是，模块名称现在包含`/ v2`，因此在使用模块名称时都包含`/ v2`）。 *如果模块的版本为v0或v1，请_not_在模块路径或导入路径中不包含主版本。

通常，具有不同导入路径的软件包是不同的软件包。例如，“ math / rand”是与“ crypto / rand”不同的软件包。如果不同的导入路径是由于导入路径中出现的主要版本不同而导致的，则也是如此。因此，“ example.com/my/mod/mypkg”与“ example.com/my/mod/v2/mypkg”是一个不同的软件包，并且都可以在单个版本中导入，这不仅有助于解决钻石依赖问题。并且还允许在替换v2方面实施v1模块，反之亦然。

请参阅`go`命令文档的[“模块兼容性和语义版本控制”]（https://golang.org/cmd/go/#hdr-Module_compatibility_and_semantic_versioning）部分，以获取有关语义导入版本控制的更多详细信息，并参见https：/ /semver.org，以获取有关语义版本控制的更多信息。

到目前为止，本节的重点是已选择加入模块并导入其他模块的代码。但是，将主要版本置于v2 +模块的导入路径中可能会与Go的较早版本或尚未选择加入模块的代码产生不兼容性。为了解决这个问题，上述行为和规则有三种重要的过渡性特殊情况或例外。随着越来越多的程序包加入模块，这些过渡性异常将不再重要。

**三个过渡性例外**

1. ** gopkg.in **

    使用以“ gopkg.in”开头的导入路径的现有代码（例如“ gopkg.in/yaml.v1”和“ gopkg.in/yaml.v2”）可以继续将这些格式用于其模块路径，甚至导入路径选择加入模块后

2. **导入非模块v2 +软件包时“ +不兼容” **
模块可以导入尚未选择加入模块的v2 +软件包。具有有效v2 + [semver]（https://semver.org）标记的非模块v2 +软件包将在导入模块的go.mod文件中带有“ + incompatible”后缀记录。 “ + incompatible”后缀表示，即使v2 +软件包具有有效的v2 + [semver]（https://semver.org）标记，例如`v2.0.0`，v2 +软件包仍未主动选择加入模块，因此在理解语义导入版本控制的含义以及如何在导入路径中使用主要版本的前提下，假定未创建v2 +软件包。因此，以[模块模式]（https://github.com/golang/go/wiki/Modules#when-do-i-get-old-behavior-vs-new-module-based-behavior）操作时， “执行”工具会将非模块v2 +软件包视为该软件包的v1版本系列的（不兼容）扩展，并假定该软件包不了解语义导入版本控制，并且后缀“ +不兼容”表示执行工具。

3. **未启用模块模式时，“最小模块兼容性” **
    
    为了帮助向后兼容，对Go版本1.9.7 +，1.10.3 +和1.11进行了更新，以使使用这些发行版构建的代码更容易正确使用v2 +模块，而无需修改现有代码。此行为称为“最小模块兼容性”，并且仅在完整的[模块模式]下有效（https://github.com/golang/go/wiki/Modules#when-do-i-get-old-behavior- vs-new-module-based-behavior）已被`go`工具禁用，例如您在Go 1.11中设置了'GO111MODULE = off'或正在使用Go版本1.9.7+或1.10.3+ 。当依赖于Go 1.9.7 +，1.10.3 +和1.11中的这种“最小模块兼容性”机制时，_not_选择加入模块的软件包将_not_在任何导入的v2 +模块的导入路径中包括主版本。相反，_has_选择加入模块_m​​ust_的软件包在导入路径中包含主版本，以导入任何v2 +模块（以便在“ go”工具在全模块模式下运行且完全了解以下情况时正确导入v2 +模块）。语义导入版本控制）。

有关发布v2 +模块所需的确切机制，请参阅[“发布模块（v2或更高版本）”]（https://github.com/golang/go/wiki/Modules#releasing-modules-v2-or-更高）。

## How to Use Modules

### How to Install and Activate Module Support
要使用模块，两个安装选项是：
* [安装最新的Go 1.11版本]（https://golang.org/dl/）。
* [master]分支上的[从源代码安装Go工具链]（https://golang.org/doc/install/source）。

安装后，您可以通过以下两种方式之一激活模块支持：
*在$ GOPATH / src树之外的目录中调用`go`命令，在当前目录或其任何父目录中使用有效的`go.mod`文件，并且未设置环境变量`GO111MODULE`（或显式设置）设置为“自动”）。
*在环境变量GO111MODULE = on上调用go命令。

###如何定义模块

为现有项目创建`go.mod`：

1.导航到GOPATH之外的模块源代码树的根目录：

   ```
   $ cd <$ GOPATH / src之外的项目路径>＃例如cd〜/ projects / hello
   ```
   请注意，在GOPATH之外，无需设置`GO111MODULE`即可激活模块模式。

   或者，如果要在GOPATH中工作：

   ```
   $ export GO111MODULE = on＃手动激活模块模式
   $ cd $ GOPATH / src / <项目路径>＃例如cd $ GOPATH / src / you / hello
   ```

2.创建初始模块定义并将其写入`go.mod`文件：

   ```
   $ go mod init
   ```
   此步骤从任何现有的[`dep`]（https://github.com/golang/dep）`Gopkg.lock`文件或任何其他[九种支持的依赖性格式]（https：//tip.golang .org / pkg / cmd / go / internal / modconv /？m = all＃pkg-variables），添加require语句以匹配现有配置。

   “ go mod init”通常将能够使用辅助数据（例如VCS元数据）来自动确定适当的模块路径，但是，如果“ go mod init”状态无法自动确定模块路径，或者您是否需要要以其他方式覆盖该路径，您可以提供[模块路径]（https://github.com/golang/go/wiki/Modules#gomod）作为`go mod init`的可选参数，例如：


```
   $ go mod init github.com/my/repo
```
请注意，如果您的依赖项包括v2 +模块，或者正在初始化v2 +模块，则在运行`go mod init`之后，您可能还需要编辑`go.mod`和`.go`代码以添加`/ vN`。 导入路径和模块路径，如上文[“语义导入版本控制”]（https://github.com/golang/go/wiki/Modules#semantic-import-versioning）部分中所述。 即使`go mod init`自动从`dep`或其他依赖项管理器转换了您的依赖项信息，这也适用。 （因此，在运行`go mod init`之后，通常不应该运行`go mod tidy`，直到成功运行`go build。/ ...`或类似的东西，这是本节中显示的顺序）。

3.构建模块。 当从模块的根目录执行时，。/ ...模式匹配当前模块内的所有软件包。 `go build`将根据需要自动添加缺少或未转换的依赖项，以满足对此特定build调用的导入：
```
   $ go build ./...
   ```
4. Test the module as configured to ensure that it works with the selected versions:

   ```
   $ go test ./...
   ```

5. (Optional) Run the tests for your module plus the tests for all direct and indirect dependencies to check for incompatibilities:

   ```
   $ go test all
   ```
   
在标记版本之前，请参见下面的[“如何为版本做准备”]（https://github.com/golang/go/wiki/Modules#how-to-prepare-for-a-release）部分。

有关所有这些主题的更多信息，官方模块文档的主要切入点是[可在golang.org上获得]（https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more）

## How to Upgrade and Downgrade Dependencies

日常的依赖关系升级和降级应该使用“ go get”完成，它将自动更新“ go.mod”文件。另外，您可以直接编辑`go.mod`。

另外，go命令（例如“ go build”，“ go test”或什至“ go list”）将根据需要自动添加新的依赖关系，以实现导入（更新“ go.mod”并下载新的依赖关系）。

要查看所有直接和间接依赖项的可用次要和补丁升级，请运行`go list -u -m all`。

要将当前模块的所有直接和间接依赖关系升级到最新版本，可以在模块根目录中运行以下命令：
 *`get -u。/ ...`以使用最新的* minor或patch *版本（并添加`-t`也可以升级测试依赖项）
 *`get -u = patch。/ ...`以使用最新的* patch *版本（并添加`-t`也可以升级测试依赖项）

go get foo将更新为foo的最新版本。 `go get foo`等同于`go get foo @ latest` —换句话说，如果未指定`@`版本，则默认为@latest。

在本节中，“最新”是带有[semver]（https://semver.org/）标记的最新版本，如果没有semver标记，则是最新的已知提交。除非存储库中没有其他semver标签，否则不要选择预发布标签作为“最新”标签（[详细信息]（https://golang.org/cmd/go/#hdr-Module_aware_go_get））。

一个常见的错误是认为`go get -u foo`仅获得`foo`的最新版本。实际上，`go get -u foo`或`go get -u foo @ latest`中的`-u`意味着_also_获得_all_的最新版本的foo的直接和间接依赖关系。升级`foo`时，通常的出发点是执行`go get foo`或`go get foo @ latest`而没有`-u`（并且在一切正常后，考虑`go get -u = patch foo`， `go get -u = patch`，`go get -u foo`或`go get -u`）。

要将版本升级或降级到更具体的版本，“ go get”允许通过添加@version后缀或[“ module query”]来覆盖版本选择（https://golang.org/cmd/go/#hdr-Module_queries ）到package参数，例如`go get foo @ v1.6.2`，`go get foo @ e3702bed2`或`go get foo @'<v1.6.2'。

使用诸如`go get foo @ master`（带有Mercurial的`foo @ tip`）这样的分支名称是获取最新提交的一种方法，而不管它是否具有semver标签。

一般而言，无法解析为semver标签的模块查询将在[go.mod]文件中记录为[pseudo-versions]（https://tip.golang.org/cmd/go/#hdr-Pseudo_versions） 。

请参阅[“了解模块的获取”]（https://golang.org/cmd/go/#hdr-Module_aware_go_get）和[“模块查询”]（https://golang.org/cmd/go/# go命令文档的hdr-Module_queries）部分，以获取有关此处主题的更多信息。

模块能够使用尚未加入模块的软件包，包括将所有可用的semver标签记录在go.mod中，并使用这些semver标签进行升级或降级。模块也可以使用尚没有适当semver标签的软件包（在这种情况下，它们将使用go.mod中的伪版本进行记录）。

升级或降级任何依赖项之后，您可能希望再次对构建中的所有软件包（包括直接和间接依赖项）运行测试，以检查不兼容性

```
   $ go test all
   ```

## How to Prepare for a Release

### Releasing Modules (All Versions)


创建模块发行版的最佳实践有望作为初始模块实验的一部分出现。其中许多最终可能会由[未来的“发布”工具]（https://github.com/golang/go/issues/26420）自动化。

在标记版本之前，应考虑一些当前建议的最佳做法：

*运行`go mod tidy`来修剪任何多余的要求（如[here]（https://tip.golang.org/cmd/go/#hdr-Maintaining_module_requirements所述）），并确保您当前的go.mod能够反映所有可能的构建标记/ OS /体系结构组合（如[此处]（https://github.com/golang/go/issues/25971#issuecomment-399091682）所述）。
  *相反，其他命令，例如“ go build”和“ go test”不会从不再需要的“ go.mod”中删除依赖项，而是仅基于当前构建调用的标签/ OS /体系结构更新“ go.mod” 。

*运行“全部测试”以测试您的模块（包括针对直接和间接依赖项运行测试），以验证当前所选软件包的版本是否兼容。
  *可能的版本组合的数量与模块的数量成指数关系，因此通常，您不能期望依赖项已经针对其依赖项的所有可能组合进行了测试。
  *作为模块工作的一部分，`go test all`已被[重新定义为更有用]（https://research.swtch.com/vgo-cmd）：包括当前模块中的所有软件包以及通过一系列的一个或多个导入，它们所依赖的所有软件包，但不包括当前模块中无关紧要的软件包。

*确保您的`go.sum`文件与`go.mod`文件一起提交。有关更多详细信息和信息，请参见[常见问题解答]（https://github.com/golang/go/wiki/Modules#should-i-commit-my-gosum-file-as-well-as-my-gomod-file）理。

###发布模块（v2或更高版本）

如果要发布v2或更高版本的模块，请首先查看上述[“语义导入版本控制”]（https://github.com/golang/go/wiki/Modules#semantic-import-versioning）部分中的讨论，其中包括为何主要版本包含在v2 +模块的模块路径和导入路径中，以及如何更新Go版本1.9.7+和1.10.3+以简化该过渡。

请注意，如果您是在采用模块之前第一次为已存在的存储库或已被标记为“ v2.0.0”或更高版本的软件包集采用模块，则[推荐的最佳做法]（https://github.com/github .com / golang / go / issues / 25967＃issuecomment-422828770）将在首次采用模块时增加主要版本。例如，如果您是foo的作者，并且foo存储库的最新标记是v2.2.2，并且foo还没有采用模块，那么最佳实践是使用v3 .0.0`表示采用模块的foo的第一个发行版（因此包含go.mod文件的foo的第一个发行版）。在这种情况下，增加主版本可以使foo的使用者更加清楚，可以在foo的v2系列上使用其他非模块补丁或次要版本，并且可以为基于模块的foo使用者提供强烈的信号。如果您执行`import“ foo”`和相应的`require foo v2.2.2 + incompatible`，则会产生不同的主要版本，而`import“ foo / v3”和相应的`require foo / v3 v3。 0.0`。 （请注意，有关在首次采用模块时增加主要版本的建议确实不适用于最新版本为v0.x.x或v1.x.x的现有存储库或软件包）。

有两种替代机制可以发布v2或更高版本的模块。请注意，使用这两种技术，当模块作者推送新标签时，新模块版本就可以供消费者使用。以创建“ v3.0.0”发行版为例，这两个选项是

1. **主要分支**：更新`go.mod`文件，使其在`module`指令的模块路径末尾包含`/ v3`（例如，`module github.com/my/module/ v3`）。将模块中的import语句更新为也使用/ v3（例如，import import“ github.com/my/module/v3/mypkg”）。用`v3.0.0`标记发行版。
   * Go版本1.9.7 +，1.10.3 +和1.11能够正确使用并构建使用此方法创建的v2 +模块，而无需更新尚未加入模块的使用者代码（如[语义导入版本控制”]（https://github.com/golang/go/wiki/Modules#semantic-import-versioning）部分）。
   *社区工具[github.com/marwan-at-work/mod](https://github.com/marwan-at-work/mod）可帮助实现此过程的自动化。请参阅[存储库]（https://github.com/marwan-at-work/mod）或[社区工具常见问题解答]（https://github.com/golang/go/wiki/Modules#what-c​​ommunity-下面提供了用于模块的工具）以进行概述。
   *为了避免与这种方法混淆，请考虑将模块的`v3。*。*'提交放在单独的v3分支上。
   * **注意：**不需要创建新分支。相反，如果您以前是在master上发布的，并且希望在master上标记“ v3.0.0”，那么这是一个可行的选择。 （但是，请注意，由于`go`工具未意识到[semver]（https：//，在`master`中引入不兼容的API更改可能会给发出`go get -u`的非模块用户带来问题。 semver.org）在Go 1.11之前或在[模块模式]时（https://github.com/golang/go/wiki/Modules#when-do-i-get-old-behavior-vs-new-module-based -行为）在Go 1.11+中未启用）。
   *现有的依赖管理解决方案，例如`dep'，在使用这种方式创建的v2 +模块时可能会遇到问题。参见例如[dep＃1962]（https://github.com/golang/dep/issues/1962）。

2. **主要子目录**：创建一个新的`v3`子目录（例如`my / module / v3`），然后在该子目录中放置一个新的`go.mod`文件。模块路径必须以`/ v3`结尾。将代码复制或移动到“ v3”子目录中。将模块中的import语句更新为也使用/ v3（例如，import import“ github.com/my/module/v3/mypkg”）。用`v3.0.0`标记发行版。
   *这提供了更大的向后兼容性。特别是，低于1.9.7和1.10.3的Go版本也能够正确使用和构建使用此方法创建的v2 +模块。
   *这里一种更复杂的方法可以利用类型别名（在Go 1.9中引入）并在位于不同子目录中的主要版本之间转发垫片。这可以提供额外的兼容性，并允许以另一个主要版本的形式实现一个主要版本，但是对于模块作者而言，这将需要更多的工作。自动执行此操作的正在进行的工具是“ goforward”。请参阅[here]（https://golang.org/cl/137076）了解更多详细信息和基本原理，以及功能正常的`goforward`初始版本。
   *诸如`dep'之类的现有依赖管理解决方案应该能够使用以这种方式创建的v2 +模块。

有关这些替代方案的更深入讨论，请参见https://research.swtch.com/vgo-module。

###发布发行

可以通过将标签推送到包含模块源代码的资源库中来发布新的模块版本。该标签是通过串联两个字符串形成的：*前缀*和*版本*。

* version *是该发行版的语义导入版本。应该遵循[语义导入版本控制]（＃semantic-import-versioning）的规则进行选择。

前缀*指示模块在存储库中的定义位置。如果模块是在存储库的根目录中定义的，则前缀为空，而标记仅为版本。但是，在[多模块存储库]（＃faqs--multi-module-repositories）中，前缀区分不同模块的版本。前缀是存储库中定义模块的目录。如果存储库遵循上述主要子目录模式，则前缀不包括主要版本后缀。

例如，假设我们有一个模块“ example.com/repo/sub/v2”，并且我们要发布版本“ v2.1.6”。仓库根目录对应于“ example.com/repo”，并且模块在仓库内的“ sub / v2 / go.mod”中定义。这个模块的前缀是“ sub /”。此版本的完整标签应为`sub / v2.1.6`。

## Migrating to Modules


本节试图简要列举迁移到模块时要做出的主要决定，并列出其他与迁移相关的主题。通常会提供其他部分的参考，以获取更多详细信息。

该材料主要基于模块实验中社区中出现的最佳实践。因此，这是一个进行中的部分，随着社区获得更多的经验，该部分将有所改善。

摘要：

*模块系统旨在允许整个Go生态系统中的不同软件包以不同的速率选择加入。
*已经在v2或更高版本上的软件包具有更多的迁移注意事项，主要是由于[语义导入版本控制]（https://github.com/golang/go/wiki/Modules#semantic-import-versioning）的影响。
*采用模块时，新软件包和v0或v1上的软件包的考虑要少得多。
*用Go 1.11定义的模块可用于较旧的Go版本（尽管确切的Go版本取决于主模块使用的策略及其依赖项，如下所述）。

迁移主题：

####从先前的依赖管理器自动迁移

  *`go mod init`会自动翻译[dep，glide，govendor，godep和其他5个先前存在的依赖项管理器]中的必需信息（https://tip.golang.org/pkg/cmd/go/internal/modconv /？m = all＃pkg-variables）生成一个生成等效构建的`go.mod`文件。
  *如果要创建v2 +模块，请确保转换后的`go.mod`中的`module`指令包含相应的`/ vN`（例如，module foo / v3`）。
  *请注意，如果要导入v2 +模块，则可能需要在初始转换后进行一些手动调整，以便将`/ vN`添加到`go mod init`从先前的依赖项管理器转换后生成的`require`语句中。有关更多详细信息，请参见上面的[“如何定义模块”]（https://github.com/golang/go/wiki/Modules#how-to-define-a-module）部分。
  *另外，`go mod init`不会编辑您的`.go`代码以在导入语句中添加任何必需的`/ vN`。请参见[“语义导入版本控制”]（https://github.com/golang/go/wiki/Modules#semantic-import-versioning）和[“发行模块（v2或更高版本）”]（https：// github .com / golang / go / wiki / Modules＃releasing-modules-v2-or-higher）部分以获取所需步骤，包括社区工具周围的一些选项，以实现自动转换。

####向Go和非模块消费者的较旧版本提供依赖信息

  *较旧的Go版本了解如何使用由`go mod vendor`创建的供应商目录，禁用模块模式时，Go 1.11和1.12+也是如此。因此，供应是模块提供依赖的一种方式，该依赖提供了对不能完全理解模块的Go的较旧版本以及未启用模块本身的使用者的依赖。请参阅[供应商常见问题解答]（https://github.com/golang/go/wiki/Modules#how-do-i-use-vendoring-with-modules-is-vendoring-going-away）和`go`命令[documentation]（https://tip.golang.org/cmd/go/#hdr-Modules_and_vendoring）了解更多详细信息。

#### Updating Pre-Existing Install Instructions
*前置模块，安装说明通常包含`go get -u foo`。如果要发布模块“ foo”，请考虑在针对基于模块的使用者的说明中删除“ -u”。
     * -u要求`go`工具升级`foo`的所有直接和间接依赖关系。
*模块使用者可以选择稍后再运行`go get -u foo`，但是[[High Fidelity Builds]]有更多好处（https://github.com/golang/proposal/blob/master/design/24301 -versioned-go.md＃update-timing-high-fidelity-builds）（如果-u不属于初始安装说明的一部分）。有关更多详细信息，请参见[“如何升级和降级依赖关系”]（https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies）。
     *`go get -u foo`仍然有效，并且仍然可以作为安装说明的有效选择。
  *另外，对于基于模块的使用者，并不是严格要求`go get foo`。
     *仅添加一个导入语句`import“ foo”`就足够了。 （后续的命令，例如go build或go test将自动下载foo并根据需要更新go.mod）。
  *默认情况下，基于模块的使用者将不使用`vendor`目录。
     *当在`go`工具中启用了模块模式时，使用模块时并不一定要严格要求`vendor'（考虑到`go.mod`中包含的信息和`go.sum`中的加密校验和），但有些现有安装说明假定“ go”工具默认情况下将使用“ vendor”。有关更多详细信息，请参见[供应商常见问题解答]（https://github.com/golang/go/wiki/Modules#how-do-i-use-vendoring-with-modules-is-vendoring-going-away）。
  *在某些情况下，包含`go get foo / ...`的安装说明可能会出现问题（请参阅[＃27215]（https://github.com/golang/go/issues/27215#issuecomment-427672781）中的讨论） 。

####避免破坏现有的导入路径

模块通过`module`指令（例如`module github.com / my / module`）在其go.mod中声明其身份。任何模块支持的使用者都必须使用与模块声明的模块路径匹配的导入路径（确切地说是针对根软件包，或将模块路径作为导入路径的前缀）导入模块内的所有软件包。如果导入路径与相应模块的声明模块路径不匹配，那么“ go”命令将报告“意外模块路径”错误。

在为一组预先存在的软件包采用模块时，应注意避免破坏现有使用者使用的现有导入路径，除非在采用模块时增加主版本。

例如，如果您之前存在的自述文件一直在告诉消费者使用`import“ gopkg.in/foo.v1”`，然后如果您采用的是v1版本的模块，那么您最初的`go.mod`应该几乎可以读为`模块gopkg.in / foo.v1`。如果您想放弃使用`gopkg.in`，那对您当前的消费者来说将是一个巨大的改变。一种方法是，如果您后来转到v2，则更改为类似`module github.com / repo / foo / v2`之类的东西。

请注意，模块路径和导入路径区分大小写。例如，将模块从github.com/Sirupsen/logrus更改为github.com/sirupsen/logrus，对于消费者来说是一项重大更改，即使GitHub自动从一个存储库名称转发到新的存储库名称。

在采用模块之后，更改`go.mod`中的模块路径是一项重大更改。

总的来说，这类似于通过[“导入路径注释”]（https://golang.org/cmd/go/#hdr-Import_path_checking）对规范导入路径的模块前强制实施，有时也称为“导入”实用程序”或“导入路径强制”。例如，软件包go.uber.org/zap当前托管在github.com/uber-go/zap中，但是使用导入路径注释[在软件包声明旁边]（（https：// github.com/uber-go/zap/blob/8a2ee5670ced5d94154bf385dc6a362722945daf/doc.go#L113））使用错误的基于github的导入路径触发任何预模块使用者的错误：

`package zap //导入“ go.uber.org/zap”

go.mod文件的module语句已淘汰了导入路径注释。

#### Incrementing the Major Version When First Adopting Modules with v2+ Packages
*如果您在采用模块之前已将某些软件包标记为v2.0.0或更高版本，那么建议的最佳实践是在首次采用模块时增加主要版本。例如，如果您使用的是“ v2.0.1”，但尚未采用模块，则对于采用模块的第一个发行版，应使用“ v3.0.0”。有关更多详细信息，请参见上面的[“释放模块（v2或更高版本）”]（https://github.com/golang/go/wiki/Modules#releasing-modules-v2-or-higher）部分。

#### v2 +模块允许在一个内部版本中使用多个主要版本

*如果某个模块在v2或更高版本上，则意味着多个主要版本可以位于一个内部版本中（例如，`foo`和`foo / v3`可能最终在单个内部版本中）。
  *这自然源于“具有不同导入路径的包是不同的包”的规则。
  *发生这种情况时，将会有多个软件包级状态的副本（例如，“ foo”的软件包级状态和“ foo / v3”的软件包级状态），并且每个主要版本都将运行其自己的“ init”功能。
  *这种方法有助于解决模块系统的多个方面，包括帮助解决钻石依赖问题，在大型代码库中逐步迁移到新版本，以及允许将主要版本实现为围绕其他主要版本的垫片。
*有关一些相关讨论，请参见https://research.swtch.com/vgo-import或[＃27514]（https://github.com/golang/go/issues/27514）的“避免单例问题”部分。

####使用非模块代码的模块

  *模块可以使用尚未选择加入模块的软件包，并在导入模块的`go.mod`中记录适当的软件包版本信息。模块可以使用尚没有适当的semver标签的软件包。有关更多信息，请参见下面的常见问题解答（https://github.com/golang/go/wiki/Modules#can-a-module-consume-a-package-that-has-not-opted-in-to-modules）细节。
  *模块还可以导入尚未选择模块的v2 +软件包。如果导入的v2 +软件包具有有效的semver标签，则将以“ + incompatible”后缀记录。请参阅常见问题解答[以下]（https://github.com/golang/go/wiki/Modules#can-a-module-consume-a-v2-package-that-has-not-opted-into-modules-what-确实不兼容），以获取更多详细信息。
  
####非模块代码使用模块

  * **消耗v0和v1模块的非模块代码**：
     *尚未选择使用模块的代码可以使用和构建v0和v1模块（与使用的Go版本无关）。

  * **非模块代码消耗v2 +模块**：
  
    * Go版本1.9.7 +，1.10.3 +和1.11已更新，因此使用这些发行版构建的代码可以正确使用v2 +模块，而无需按照[“语义导入版本控制”]（https ：//github.com/golang/go/wiki/Modules#semantic-import-versioning）和[“发布模块（v2或更高版本）”]（https://github.com/golang/go/wiki/Modules#以上的release-modules-v2或更高版本）部分。

    如果1.9.7和1.10.3之前的Go版本可以使用v2 +模块，如果该v2 +模块是按照[“释放模块（v2或更高版本）”]（https://github.com/zh-cn/）中概述的“主要子目录”方法创建的。 com / golang / go / wiki / Modules＃releasing-modules-v2-or-higher）部分。

####预先存在的v2 +软件包作者的策略

对于考虑加入模块的预先存在的v2 +软件包的作者，总结替代方法的一种方法是在三种顶级策略之间进行选择。每个选择都有后续的决定和变化（如上所述）。这些替代的顶级策略是：

1. **要求客户使用Go版本1.9.7 +，1.10.3 +或1.11 + **。

    该方法使用“主要分支”方法，并依赖于“最小模块感知”，该模型被反向移植到1.9.7和1.10.3。请参见[“语义导入版本控制”]（https://github.com/golang/go/wiki/Modules#semantic-import-versioning）和[“发行模块（v2或更高版本）”]（https：// github .com / golang / go / wiki / Modules＃releasing-modules-v2-or-higher）部分以获取更多详细信息。

2. **允许客户端使用甚至更老的Go版本，例如Go 1.8 **。

    这种方法使用“主要子目录”方法，并涉及创建子目录，例如“ / v2”或“ / v3”。请参见[“语义导入版本控制”]（https://github.com/golang/go/wiki/Modules#semantic-import-versioning）和[“发行模块（v2或更高版本）”]（https：// github .com / golang / go / wiki / Modules＃releasing-modules-v2-or-higher）部分以获取更多详细信息。

3. **等待加入模块**。

    在这种策略下，事情继续与选择了模块的客户端代码以及未选择模块的客户端代码一起工作。随着时间的流逝，Go版本1.9.7 +，1.10.3 +和1.11+的发布时间将越来越长，并且在将来的某个时候，要求Go版本变得更加自然或对客户友好1.9.7 + / 1.10.3 + / 1.11 +，此时您可以

## Additional Resources

### Documentation and Proposal
*官方文件：
  *最新的[golang.org上的模块HTML文档]（https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more）
  *运行`go help modules`获取有关模块的更多信息。 （这是通过`go help`进入模块主题的主要切入点）
  *运行`go help mod`，以获得有关`go mod`命令的更多信息。
  *在模块感知模式下，运行`go help module-get`以获得更多关于`go get`行为的信息。
  *运行`go help goproxy`以获得有关模块代理的更多信息，包括通过`file：///`URL的基于纯文件的选项。
* Russ Cox在`vgo`上发布的初始[“ Go＆Versioning”]（https://research.swtch.com/vgo）系列博客文章（首次发布于2018年2月20日）
*正式的[golang.org博客文章，介绍该提案]（https://blog.golang.org/versioning-proposal）（2018年3月26日）
  *与完整的`vgo`博客系列相比，该提案提供了更简洁的概述，以及提案背后的一些历史和流程
*官方[Versioned Go Modules Proposal]（https://golang.org/design/24301-versioned-go）（最新更新于2018年3月20日）

###介绍性材料

* 40分钟的介绍性视频[“ Go中的版本原则”]（https://www.youtube.com/watch?v=F8nrpe0XWRg&list=PLq2Nv-Sh8EbbIjQgDzapOFeVfv5bGOoPE&index=3&t=0s）由Russ Cox提供（5月2日） ，2018）
  *简洁地介绍了版本化Go模块设计背后的哲学，包括“兼容性”，“重复性”和“合作”三个核心原则
*基于示例的35分钟入门视频[“什么是Go模块以及如何使用它们？”]（https://www.youtube.com/watch?v=6MbIzJmLz6Q&list=PL8QGElREVyDA2iDrPNeCe8B1u7li5S6ep&index=5&t=0s）（[幻灯片] Paul Jolly（2018年8月15日）（https://talks.godoc.org/github.com/myitcv/talks/2018-08-15-glug-modules/main.slide#1））
*介绍性博客文章[“采用Go进行旋转模块”]（https://dave.cheney.net/2018/07/14/takeing-go-modules-for-a-spin）（7月14日，Dave Cheney） 2018）
*入门[模块上的聚会聚会幻灯片]（https://docs.google.com/presentation/d/1ansfXN8a_aVL-QuvQNY7xywnS78HE8aG7fPiITNQWMM/edit#slide=id.g3d87f3177d_0_0）（2018年7月16日）
* 30分钟的入门视频[“入门模块和SemVer”]（Francesc Campoy的https://www.youtube.com/watch?v=aeF3l-zmPsY）（2018年11月15日）


### Additional Material
*博客文章[“在Travis CI上使用带有供应商支持的Go模块”]（https://arslan.io/2018/08/26/using-go-modules-with-vendor-support-on-travis-ci/）通过Fatih Arslan（2018年8月26日）
* Todd Keech（2018年7月30日）的博客文章[“ Go Modules and CircleCI”]（https://medium.com/@toddkeech/go-modules-and-circleci-c0d6fac0b000）
*博客文章[“ vgo提案已被接受。现在如何？”]（Russ Cox的https://research.swtch.com/vgo-accepted）（2018年5月29日）
  *概述了版本化模块当前是实验性加入功能的含义
*由Carolyn Van Slyck（2018年7月16日）发表的关于[如何从头开始构建并开始使用go模块的博文]（https://carolynvanslyck.com/blog/2018/07/building-go-from-source/） ）

##自最初的Vgo提案以来的更改

作为提案，原型和Beta版流程的一部分，整个社区创建了400多个问题。请继续提供反馈。

以下是一些较大的更改和改进的部分列表，其中几乎所有更改和改进都主要基于社区的反馈：

*保留了顶级供应商支持，而不是基于vgo的版本，完全忽略了供应商目录（[讨论]（https://groups.google.com/d/msg/golang-dev/FTMScX1fsYk/uEUSjBAHAwAJ），[CL]（ https://go-review.googlesource.com/c/vgo/+/118316））
*向后移植了最小的模块意识，以允许较早的Go版本1.9.7+和1.10.3+可以更轻松地使用v2 +项目的模块（[讨论]（https://github.com/golang/go/issues/24301#issuecomment -371228742），[CL]（https://golang.org/cl/109340））
*默认情况下，允许vgo使用v2 +标签用于预先存在的软件包尚无go.mod（[此处]描述的相关行为的最新更新（此处）（https://github.com/golang/go/issues/25967#issuecomment -407567904））
*通过命令`go get -u = patch`添加了支持，以将所有传递依赖项更新到同一次要版本上的最新可用补丁程序级别版本（[讨论]（https://research.swtch.com/vgo-cmd） ，[文档]（https://tip.golang.org/cmd/go/#hdr-Module_aware_go_get））
*通过环境变量进行其他控制（例如[＃26585]（https://github.com/golang/go/issues/26585）中的GOFLAGS，[CL]（https://go-review.googlesource.com/c / GO / + / 126656））
*关于是否允许更新go.mod，如何使用供应商目录以及是否允许网络访问（例如，-mod = readonly，-mod = vendor，GOPROXY = off；相关[ CL]（https://go-review.googlesource.com/c/go/+/126696），以获取最新更改）
*添加了更灵活的替换指令（[CL]（https://go-review.googlesource.com/c/vgo/+/122400））
*添加了其他查询模块的方式（供人类使用，以及更好的编辑器/ IDE集成）
*到目前为止，go CLI的UX继续根据经验进行改进（例如[＃26581]（https://github.com/golang/go/issues/26581），[CL]（https：// go-review.googlesource.com/c/go/+/126655））
*对通过诸如“ go mod download”之类的CI或docker构建之类的用例预热缓存的附加支持（[＃26610]（https://github.com/golang/go/issues/26610#issuecomment-408654653））
* **最有可能**：更好地支持将特定版本的程序安装到GOBIN（[＃24250]（https://github.com/golang/go/issues/24250#issuecomment-377553022））
## GitHub Issues
* [Currently open module issues](https://golang.org/issues?q=is%3Aopen+is%3Aissue+label:modules)
* [Closed module issues](https://github.com/golang/go/issues?q=is%3Aclosed+is%3Aissue+label%3Amodules+sort%3Aupdated-desc)
* [Closed vgo issues](https://github.com/golang/go/issues?q=-label%3Amodules+vgo+is%3Aclosed+sort%3Aupdated-desc)
* Submit a [new module issue](https://github.com/golang/go/issues/new?title=cmd%2Fgo%3A%20%3Cfill%20this%20in%3E) using 'cmd/go:' as the prefix



## FAQs
###如何将版本标记为不兼容？

`require`指令允许任何模块声明其应使用依赖项D的版本> = x.y.z构建（可能由于与模块D的版本<x.y.z不兼容而指定）。经验数据表明[这是'dep'和'cargo'中使用的约束的主要形式]（https://twitter.com/_rsc/status/1022590868967116800）。另外，构建中的顶层模块可以“排除”特定版本的依赖项，或者“替换”其他模块以不同的代码。请参阅完整的提案以获取[更多详细信息和原理]（https://github.com/golang/proposal/blob/master/design/24301-versioned-go.md）。

版本模块建议的主要目标之一是为工具和开发人员在Go代码的版本周围添加通用词汇和语义。这为将来声明不兼容的其他形式奠定了基础，例如：
*在最初的`vgo`博客系列中将不推荐使用的版本声明为[描述]（https://research.swtch.com/vgo-module）
*声明外部系统中模块之间的成对不兼容性，例如在提案过程中[例如] [https://github.com/golang/go/issues/24301#issuecomment-392111327）
*发布版本后，声明模块的成对不兼容版本或不安全版本。例如，请参见[＃24031]（https://github.com/golang/go/issues/24031#issuecomment-407798552）和[＃26829]（https://github.com/golang/去/问题/ 26829）

###什么时候出现旧行为与新的基于模块的行为？

通常，模块是Go 1.11的可选组件，因此，根据设计，默认情况下会保留旧的行为。

总结何时获得旧的1.10现状行为与新的基于选择加入模块的行为：

*在GOPATH内部-默认为旧的1.10行为（忽略模块）
*在GOPATH外部，而在带有`go.mod`的文件树中-默认为模块行为
* GO111MODULE环境变量：
  *未设置或“自动”-上面的默认行为
  *`on` —强制启用模块支持，而不考虑目录位置
  *`off`-强制关闭模块支持，无论目录位置如何

###为什么通过“去获取”安装工具失败，并显示错误“找不到主模块”？

当您将`GO111MODULE = on'设置为on，但是当您运行`go get`时，它不在带有`go.mod`的文件树中。

最简单的解决方案是不设置“ GO111MODULE”（或等效地显式设置为“ GO111MODULE = auto”），这样可以避免此错误。

回想一下存在的主要原因之一是记录精确的依赖项信息。该依赖项信息将写入您当前的`go.mod`中。如果您不在带有`go.mod`的文件树中，但是通过设置`GO111MODULE = on`告诉`go get`命令以模块模式运行，那么运行`go get`将导致错误。找不到主模块，因为没有可用的go.mod来记录依赖项信息。

解决方案的替代方案包括：

1.保持未设置“ GO111MODULE”（默认设置，或显式设置“ GO111MODULE = auto”），这将导致更友好的行为。当您不在模块内时，这将为您提供Go 1.10的操作，因此将避免“去获取”报告“找不到主模块”。

2.离开GO111MODULE = on，但是根据需要暂时禁用模块并在go get期间启用Go 1.10行为，例如通过GO111MODULE = off go get example.com/cmd。可以将其转换为简单的脚本或shell别名，例如“ alias oldget ='GO111MODULE = off go get'”。

3.创建一个临时的`go.mod`文件，然后将其丢弃。这已由[@rogpeppe]（https://github.com/rogpeppe）的[简单外壳脚本]（https://gist.github.com/rogpeppe/7de05eef4dd774056e9cf175d8e6a168）自动完成。该脚本允许有选择地通过`vgoget example.com/cmd [@version]`提供版本信息。 （这可以避免在GOPATH模式下无法使用path @ version语法的错误）。

4.`gobin`是一个可识别模块的命令，用于安装和运行主软件包。默认情况下，`gobin`无需先手动创建模块即可安装/运行主软件包，但是通过`-m`标志，可以告诉它使用现有模块来解决依赖关系。有关详细信息和其他用途，请参见`gobin'[README]（https://github.com/myitcv/gobin#usage）和[FAQ]（https://github.com/myitcv/gobin/wiki/FAQ）。案例。

5.创建一个用于跟踪全局安装工具的“ go.mod”（例如，在〜/ global-tools / go.mod中），并在运行“ go get”或“ go”之前将“ cd”复制到该目录。 install`来安装任何全球安装的工具。

6.在单独的目录中为每个工具创建一个“ go.mod”，例如“〜/ tools / gorename / go.mod”和“〜/ tools / goimports / go.mod”，并在其相应目录中创建“ cd”在对该工具运行`go get`或`go install`之前。

该当前限制将得到解决。但是，主要问题是模块当前处于启用状态，完整的解决方案可能要等到GO111MODULE = on成为默认行为。有关更多讨论，请参见[＃24250]（https://github.com/golang/go/issues/24250#issuecomment-377553022）：

>这显然必须最终起作用。就该版本而言，我不确定这到底是做什么的：它会创建一个临时模块root和go.mod，执行安装，然后将其丢弃吗？大概。但是我不太确定，就目前而言，我不想让vgo在go.mod树之外做一些事情来使人们感到困惑。当然，最终的go命令集成必须支持这一点。

该常见问题解答一直在讨论跟踪_globally_已安装的工具。

相反，如果要跟踪_specific_模块所需的工具，请参阅下一个FAQ。

###如何跟踪模块的工具依赖关系？

如果你：
 *要在处理模块时使用基于Go的工具（例如`stringer`），并且
 *想要确保每个人都在使用该工具的相同版本，同时在模块的`go.mod`文件中跟踪该工具的版本

那么当前推荐的一种方法是在模块中添加一个“ tools.go”文件，其中包含针对所需工具的导入语句（例如“ import _“ golang.org/x/tools/cmd/stringer”`）带有`// + build tools`构建约束。 import语句允许`go`命令在模块的`go.mod`中精确记录工具的版本信息，而`// + build tools`构建约束阻止您的常规构建实际导入工具。

有关如何执行此操作的具体示例，请参见此[“通过示例进行模块学习”演练]（https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md ）。

[＃25922中的此注释]（https://github.com/golang/go/issues/25922#issuecomment-412992431）中对该方法以及如何执行此方法的较早的具体示例进行了讨论。

简要理由（也来自[＃25922]（https://github.com/golang/go/issues/25922#issuecomment-402918061））：

>实际上，我认为tools.go文件是工具依赖关系的最佳实践，当然对于Go 1.11。
>
>我喜欢它，因为它没有引入新的机制。
>
>它只是重用现有的。

### IDE，编辑器和标准工具（例如goimports，gorename等）中模块支持的状态如何？

对模块的支持已开始在编辑器和IDE中获得。

例如：
* ** GoLand **：目前完全支持GOPATH内外的模块，包括[此处]所述的完成，语法分析，重构，导航（https://blog.jetbrains.com/go/2018/08/24 /的Goland  -  2018年2月2日 - 是 - 这里/）。
* ** VS代码**：工作正在进行中，正在寻找有助于的人。跟踪问题为[＃1532]（https://github.com/Microsoft/vscode-go/issues/1532）。 [VS Code模块状态Wiki页面]（https://github.com/Microsoft/vscode-go/wiki/Go-modules-support-in-Visual-Studio-Code）中描述了初始Beta。
* **带有go +的原子**：跟踪问题为[＃761]（https://github.com/joefitzgerald/go-plus/issues/761）。
*带有vim-go的vim **：最初对语法突出显示和格式`go.mod`的支持[登陆]（https://github.com/fatih/vim-go/pull/1931）。在[＃1906]（https://github.com/fatih/vim-go/issues/1906）中跟踪了更广泛的支持。
*带有go-mode.el的emacs：[＃237]（https://github.com/dominikh/go-mode.el/issues/237）中的跟踪问题。

其他工具（例如goimports，guru，gorename和类似工具）的状态已在总括问题[＃24661]（https://github.com/golang/go/issues/24661）中进行了跟踪。请查看该伞的最新状态。

特定工具的一些跟踪问题包括：
* ** gocode **：[mdempsky / gocode /＃46]（https://github.com/mdempsky/gocode/issues/46）中的跟踪问题。注意，“ nsf / gocode”建议人们从“ nsf / gocode”迁移到“ mdempsky / gocode”。
* ** go-tools **（dominikh提供的工具，例如staticcheck，megacheck，gosimple）：样本跟踪问题[dominikh / go-tools＃328]（https://github.com/dominikh/go-tools/issues/ 328）。

通常，即使您的编辑器，IDE或其他工具尚未被模块识别，如果您在GOPATH内使用模块并执行mod vendor，则它们的许多功能也应能与模块一起使用（因为应该使用适当的依赖项）通过GOPATH提取）。

完整的解决方案是将加载软件包的程序从“ go / build”中移出并移至“ golang.org/x/tools/go/packages”中，该程序了解如何以模块感知的方式查找软件包。这最终可能会变成`go / packages'。


## FAQs — Additional Control

### What community tooling exists for working with modules?

社区开始在模块之上构建工具。例如：

* [github.com/rogpeppe/gohack](https://github.com/rogpeppe/gohack）
  *一种新的社区工具，可自动执行并大大简化“替换”和多模块工作流程，其中包括让您轻松修改其中一个依赖项
  *例如，`gohack example.com / some / dependency`会自动克隆相应的存储库，并将必要的`replace`指令添加到您的`go.mod`中
  *使用`gohack undo`删除所有gohack替换语句
  *该项目正在继续扩展，以简化与模块相关的其他工作流程
* [github.com/marwan-at-work/mod](https://github.com/marwan-at-work/mod）
  *命令行工具可自动升级/降级模块的主要版本
  *在go源代码中自动调整`go.mod`文件和相关的导入语句
  *帮助进行升级，或者在首次选择带有v2 +软件包的模块时提供帮助
* [github.com/akyoto/mgit](https://github.com/akyoto/mgit）
  *让您查看和控制所有本地项目的semver标签
  *显示未标记的提交，并让您一次标记所有提交（`mgit -tag + 0.0.1`）
* [github.com/goware/modvendor](https://github.com/goware/modvendor）
  *有助于将其他文件复制到`vendor /`文件夹中，例如shell脚本，.cpp文件，.proto文件等。
* [github.com/psampaz/go-mod-outdated](https://github.com/psampaz/go-mod-outdated）
  *以人类友好的方式显示过时的依赖关系
  *提供一种过滤间接依赖关系和无需更新的依赖关系的方法
  *提供一种在过时依赖项下中断CI管道的方法

###什么时候应该使用replace指令？

如上面['go.mod'概念部分]（https://github.com/golang/go/wiki/Modules#gomod）所述，`replace`指令在顶层`go.mod'中提供了附加控制实际上用于满足在Go源文件或go.mod文件中找到的依赖关系的`，而在构建主模块时会忽略除主模块以外的模块中的`replace`指令。

“替换”指令允许您提供另一个导入路径，该路径可能是位于VCS（GitHub或其他地方）中或本地文件系统上具有相对或绝对文件路径的另一个模块。使用了来自`replace`指令的新导入路径，而无需更新实际源代码中的导入路径。

 `replace`允许顶层模块控制用于依赖项的确切版本，例如：
  *`替换example.com/some/dependency => example.com/some/dependency v1.2.3`

`replace`还允许使用分叉的依赖项，例如：
  *`替换example.com/some/dependency => example.com/some/dependency-fork v1.2.3`

一个示例用例是，如果您需要修复或研究依赖项中的某些内容，则可以使用本地派生，并在顶级`go.mod`中添加类似以下内容的内容：
  *`替换example.com/original/import/path => / your / forked / import / path`

“替换”还可以用于告知go工具多模块项目中模块的相对或绝对磁盘上位置，例如：
  *`替换example.com/project/foo => ../ foo`

**注意**：如果`replace`指令的右侧是文件系统路径，则目标必须在该位置具有`go.mod`文件。如果`go.mod`文件不存在，则可以使用`go mod init`创建一个文件。

通常，您可以选择在replace指令中在`=>`的左侧指定一个版本，但是通常，如果您忽略此更改，则对更改的敏感度较低（例如，在所有`replace`示例中都已完成）以上）。

在Go 1.11中，对于直接依赖关系，即使执行`replace`也需要`require`指令。例如，如果`foo`是直接依赖项，那么没有`foo`的相应`require`就不能做`replace foo => ../ foo`。如果您不确定在`require`指令中使用哪个版本，可以经常使用`v0.0.0`，例如`require foo v0.0.0`。 Go 1.12中使用[＃26241]（https://golang.org/issue/26241）解决了此问题。

您可以通过运行`go list -m all`来确认您已经获得了期望的版本，它向您显示了将在构建中使用的实际最终版本，其中包括考虑了`replace`语句。

有关更多详细信息，请参见['go mod edit'文档]（https://golang.org/cmd/go/#hdr-Edit_go_mod_from_tools_or_scripts）。

[github.com/rogpeppe/gohack](https://github.com/rogpeppe/gohack）使这些类型的工作流程变得更加容易，尤其是当您的目标是对模块依赖项进行可变签出时。有关概述，请参见[repository]（https://github.com/rogpeppe/gohack）或之前的FAQ。

有关使用`replace`完全在VCS之外工作的详细信息，请参见下一个FAQ。

###我可以完全在本地文件系统上的VCS之外工作吗？

是。不需要VCS。

如果您要一次在VCS之外编辑一个模块，这非常简单（并且您总共只有一个模块，或者其他模块
```
module example.com/me/hello

require (
  example.com/me/goodbye v0.0.0
)

replace example.com/me/goodbye => ../goodbye
```

如本例所示，如果在VCS之外，则可以使用`v0.0.0`作为`require`指令中的版本。请注意，如先前的FAQ中所述，在Go 1.11中必须在此处手动添加`require`指令，但是在Go 1.12+中不再需要手动添加`require`指令（[＃26241]（https：// golang.org/issue/26241））。

此[thread]（https://groups.google.com/d/msg/golang-nuts/1nYoAMFZVVM/eppaRW2rCAAJ）中显示了一个小的可运行示例。

###如何在模块中使用供应商？供应商会消失吗？

最初的“ vgo”博客文章系列确实建议完全放弃供应商，但社区的[feedback]（https://groups.google.com/d/msg/golang-dev/FTMScX1fsYk/uEUSjBAHAwAJ）导致保留了对vendoring。
 
简而言之，要对模块使用供应商：
*`go mod vendor`重置主模块的vendor目录，以包含根据go.mod文件和Go源代码的状态构建和测试所有模块软件包所需的所有软件包。
*默认情况下，在模块模式下，执行“ go build”之类的go命令会忽略供应商目录。
*`-mod = vendor`标志（例如`go build -mod = vendor`）指示go命令使用主模块的顶级供应商目录来满足依赖关系。因此，在此模式下，go命令将忽略go.mod中的依赖项描述，并假定供应商目录包含正确的依赖项副本。请注意，仅使用主模块的顶级供应商目录。其他位置的供应商目录仍然被忽略。
*有些人会希望通过设置`GOFLAGS = -mod = vendor`环境变量来定期选择供应商。

Go的较旧版本（例如1.10）可以理解如何使用“ go mod vendor”创建的供应商目录，在[模块模式]时，Go 1.11和1.12+也可以使用（https://github.com/golang/go/wiki/Modules ＃when-do-i-get-old-behavior-vs-new-module-based-behavior）。因此，供应是模块提供依赖的一种方式，该依赖提供了对不能完全理解模块的Go的较旧版本以及未启用模块本身的使用者的依赖。

如果您正在考虑使用供应商，则值得阅读[“模块和供应商”]（https://tip.golang.org/cmd/go/#hdr-Modules_and_vendoring）和[“制作依赖关系的供应商副本”]提示文档的（https://tip.golang.org/cmd/go/#hdr-Make_vendored_copy_of_dependencies）部分。

###是否有“始终在线”的模块存储库和企业代理？

公共托管的“始终在”不可变模块存储库以及可选的私有托管的代理和存储库正变得可用。

For example:
* [proxy.golang.org](https://proxy.golang.org) - Official project - Run by [Google](https://www.google.com) - The default Go module proxy built by the Go team.
* [gocenter.io](https://gocenter.io) - Commercial project - Run by [JFrog](https://jfrog.com) - The central Go modules repository.
* [mirrors.aliyun.com/goproxy](https://mirrors.aliyun.com/goproxy) - Commercial project - Run by [Alibaba Cloud](https://www.alibabacloud.com) - A Go module proxy alternate.
* [goproxy.cn](https://goproxy.cn) - Open source project - Run by [Qiniu Cloud](https://www.qiniu.com) - The most trusted Go module proxy in China.
* [goproxy.io](https://goproxy.io) - Open source project - Run by China Golang Contributor Club - A global proxy for Go modules.
* [Athens](https://github.com/gomods/athens) - Open source project - Self-hosted - A Go module datastore and proxy.
* [athens.azurefd.net](https://athens.azurefd.net) - Open source project - Run by [Microsoft](https://www.microsoft.com) - A hosted module proxy running Athens.
* [Goproxy](https://github.com/goproxy/goproxy) - Open source project - Self-hosted - A minimalist Go module proxy handler.
* [THUMBAI](https://thumbai.app) - Open source project - Self-hosted - Go mod proxy server and Go vanity import path server.

请注意，您不需要运行代理。相反，1.11中的go工具已通过[GOPROXY]（https://tip.golang.org/cmd/go/#hdr-Module_proxy_protocol）添加了可选的代理支持，以启用更多企业用例（例如更好的控制），并且还可以更好地处理诸如“ GitHub已关闭”或人们删除GitHub存储库之类的情况。

###我可以控制go.mod何时更新以及go工具何时使用网络满足依赖关系吗？

默认情况下，“ go build”之类的命令将根据需要到达网络，以满足输入需求。

一些团队可能希望禁止go工具在某些时候接触网络，或者想要更好地控制go工具何时更新`go.mod`，如何获得依赖关系以及如何使用供应商。

go工具提供了相当大的灵活性来调整或禁用这些默认行为，包括通过-mod = readonly，-mod = vendor，GOFLAGS，GOPROXY = off，GOPROXY = file：// / filesystem / path”，“ go mod供应商”和“ go mod下载”。

这些选项的详细信息遍布整个官方文档。 [这里]（https://github.com/thepudds/go-module-knobs/blob/master/README.md）是一个社区，试图对与这些行为相关的旋钮进行综合概述。其中包括指向官方文档的链接欲获得更多信息。

###如何在Travis或CircleCI等CI系统中使用模块？

最简单的方法可能只是设置环境变量“ GO111MODULE = on”，该变量应适用于大多数CI系统。

但是，由于您的某些用户尚未选择加入模块，因此在启用和禁用模块的Go 1.11上的CI中运行测试可能很有价值。供应商也是要考虑的话题。

以下两个博客文章更具体地介绍了这些主题：

* [“在Travis CI上使用带有供应商支持的Go模块”]（Fatih的https://arslan.io/2018/08/26/using-go-modules-with-vendor-support-on-travis-ci/）阿尔斯兰
* [“ Go模块和CircleCI”]（Todd Keech的https://medium.com/@toddkeech/go-modules-and-circleci-c0d6fac0b000）

##常见问题解答— go.mod和go.sum

###为什么“ go mod tidy”在我的“ go.mod”中记录间接和测试依赖项？

模块系统在您的`go.mod`中记录了精确的依赖要求。 （有关更多详细信息，请参见上面的[go.mod概念]（https://github.com/golang/go/wiki/Modules#gomod）部分或[go.mod技巧文档]（https：// tip。 golang.org/cmd/go/#hdr-The_go_mod_file））。

“ go mod tidy”会更新您当前的“ go.mod”，以在模块中包含测试所需的依赖项-如果测试失败，我们必须知道使用了哪些依赖项来重现失败。

“ go mod tidy”还可以确保您当前的“ go.mod”反映操作系统，架构和构建标记的所有可能组合的依赖性要求（如[此处]所述（https://github.com/golang/go/issues / 25971＃issuecomment-399091682））。相反，诸如go build和go test之类的其他命令仅更新go.mod以在当前的GOOS，GOARCH和build标签下提供由请求的软件包导入的软件包（这是原因之一。 “ go mod tidy”可能会添加“ go build”或类似版本未添加的要求。

如果您模块的依赖项本身不具有`go.mod`（例如，因为该依赖项尚未选择加入模块本身），或者其`go.mod`文件缺少其一个或多个依赖项（例如，由于模块作者未运行`go mod tidy`），则缺少的传递依赖项将被添加到_your_模块的要求中，并带有一个“ //间接”注释，以表明该依赖关系不是来自内部的直接导入您的模块。

注意，这也意味着直接或间接依赖项中缺少的测试依赖项也将记录在您的`go.mod`中。 （当这很重要时的示例：`go test all`运行模块的_all_直接和间接依赖关系的测试，这是验证您当前版本组合可以协同工作的一种方法。如果在其中一个测试中失败当您运行“全部测试”时，必须记录一组完整的测试依赖性信息，以使您具有可重复的“全部测试”行为，这一点很重要。

您的`go.mod`文件中可能具有`//间接`依赖关系的另一个原因是，如果您已经升级（或降级）了其中一个间接依赖关系，超出了直接依赖关系所要求的范围，例如，如果您运行`go get -u`或`go get foo @ 1.2.3`。 go工具需要一个位置来记录这些新版本，并且它会在您的`go.mod`文件中记录（并且不会深入到您的依赖项中，以修改_their_`go.mod`文件）。

通常，上述行为是模块如何通过记录精确的依赖项信息来提供100％可复制的构建和测试的一部分。

如果您对为什么某个特定模块显示在`go.mod`中感到好奇，则可以运行`go mod why -m <module>`来[answer]（https://tip.golang.org/cmd / go /＃hdr-Explain_why_packages_or_modules_are_needed）这个问题。用于检查需求和版本的其他有用工具包括“ go mod graph”和“ go list -m all”。

###'go.sum'是一个锁定文件吗？为什么“ go.sum”包含有关我不再使用的模块版本的信息？

不，`go.sum`不是锁定文件。构建中的`go.mod`文件可为100％可复制的构建提供足够的信息。

为了验证，`go.sum`包含特定模块版本内容的预期密码校验和。有关更多详细信息，请参见下面的[FAQ]（https://github.com/golang/go/wiki/Modules#should-i-commit-my-gosum-file-as-well-as-my-gomod-file）在`go.sum`上（包括为什么通常要在`go.sum`中签入）以及[“模块下载和验证”]（https://tip.golang.org/cmd/go/#hdr-提示文档中的Module_downloading_and_verification）部分。

部分原因是`go.sum`不是锁定文件，即使您停止使用某个模块或特定模块版本，它也会保留模块版本的加密校验和。如果您以后继续使用某些内容，则可以验证校验和，从而提高了安全性。

另外，您模块的`go.sum`记录了构建中使用的所有直接和间接依赖项的校验和（因此，您列出的`go.sum`通常会比`go.mod`列出更多的模块）。

###我应该提交我的“ go.sum”文件还是“ go.mod”文件吗？

通常，模块的`go.sum`文件应与`go.mod`文件一起提交。

*`go.sum`包含特定模块版本内容的预期密码校验和。
*如果有人克隆您的存储库并使用go命令下载了您的依赖项，那么如果他们下载的依赖项副本与您的`go.sum`中的相应条目之间存在任何不匹配，他们就会收到错误消息。
*此外，“ go mod verify”会检查磁盘下载的模块下载在磁盘上的缓存副本是否仍与“ go.sum”中的条目匹配。
*请注意，`go.sum`不是某些替代性依赖项管理系统中使用的锁定文件。 （“ go.mod”为可复制的构建提供了足够的信息）。
*请参阅Filippo Valsorda的非常简短的[rational here]（https://twitter.com/FiloSottile/status/1029404663358087173），以了解为什么要输入“ go.sum”。有关更多详细信息，请参见技巧文档的[“模块下载和验证”]（https://tip.golang.org/cmd/go/#hdr-Module_downloading_and_verification）部分。请参阅[＃24117]（https://github.com/golang/go/issues/24117）和[＃25530]（https://github.com/golang/go/issues/ 25530）。

###如果我没有任何依赖关系，我还应该添加一个“ go.mod”文件吗？

是。这支持在GOPATH之外进行工作，帮助与您选择加入模块的生态系统进行通信，此外，您在`go.mod`中的`module`指令还可以作为代码身份的明确声明（这是一个最终可能不推荐使用导入注释的原因）。当然，模块在Go 1.11中纯粹是可选功能。

## FAQs — Semantic Import Versioning

### Why must major version numbers appear in import paths?

请参阅上面[“语义导入版本控制”]（https://github.com/golang/go/wiki/Modules#semantic-import-versioning）概念部分中有关语义导入版本控制和导入兼容性规则的讨论。另请参阅[宣布该建议的博客文章]（https://blog.golang.org/versioning-proposal），其中更多地介绍了导入兼容性规则的动机和理由。

###为什么从导入路径中省略主要版本v0，v1？”

请参阅问题“为什么导入路径中省略了主要版本v0，v1？”在较早的[来自官方提案讨论的常见问题解答]（https://github.com/golang/go/issues/24301#issuecomment-371228664）中。

###使用主要版本v0，v1标记我的项目，或使用v2 +进行重大更改有什么含义？

为了回应有关“ k8s会发布次要版本，但在每个次要版本中更改Go API” *的评论，Russ Cox做出了以下[响应]（https://github.com/kubernetes/kubernetes/pull/65683#issuecomment -403705882），重点介绍了在您的项目中选择v0，v1与频繁使用v2，v3，v4等进行重大更改的一些含义：

>我不完全了解k8s开发周期等，但我认为通常k8s团队需要决定/确认他们打算向用户保证稳定性的内容，然后相应地使用版本号来表达这一点。
>
> *要保证API兼容性（这似乎是最好的用户体验！），然后开始执行并使用1.X.Y。
> *可以灵活地在每个发行版中进行向后不兼容的更改，但允许大型程序的不同部分按不同的时间表升级其代码，这意味着不同的部分可以在一个程序中使用API​​的不同主要版本，然后使用XY0 ，以及类似k8s.io/client/vX/foo的导入路径。
> *不保证与API兼容，并且无论如何都要求每个构建都只有一个k8s库副本，这意味着即使不是所有构建都已经准备好，构建的所有部分也必须使用相同版本。为此，然后使用0.XY

与此相关的是，Kubernetes具有一些非典型的构建方法（当前在Godep上包括自定义包装脚本），因此，Kubernetes对于许多其他项目来说是不完善的示例，但随着[Kubernetes迈向采用Go， 1.11模块]（https://github.com/kubernetes/kubernetes/pull/64731#issuecomment-407345841）。

###模块可以使用尚未选择加入模块的软件包吗？

是。

如果存储库未选择使用模块，但已使用有效的[semver]（https://semver.org）标签（包括必需的前导v）进行了标记，则可以在go get中使用这些semver标签。 ，相应的semver版本将记录在导入模块的go.mod文件中。如果存储库没有任何有效的semver标签，则将使用[“ pseudo-version”]（https://golang.org/cmd/go/#hdr-Pseudo_versions）（例如`v0）记录存储库的版本。 0.0-20171006230638-a6e239ea1c69`（其中包括时间戳和提交哈希，它们的设计目的是允许对`go.mod`中记录的各个版本进行总体排序，并使其更容易推断出哪个记录​​版本比“ .late”更早另一个录制版本）。

例如，如果软件包foo的最新版本被标记为v1.2.3，但是foo本身尚未选择加入模块，则从内部运行`go get foo`或`go get foo @ v1.2.3`。模块M将记录在模块M的`go.mod`文件中，如下所示：

```
需要foo v1.2.3
```


“ go”工具还将在其他工作流程中为非模块程序包使用可用的semver标签（例如“ go list -u = patch”，它将模块的依赖项升级到可用补丁程序版本，或“ go list -u” -m all`，显示可用的升级等）。

有关尚未选择模块的v2 +软件包的更多详细信息，请参见下一个常见问题解答。

###模块可以使用未选择模块的v2 +软件包吗？ “ +不兼容”是什么意思？
 
是的，模块可以导入尚未选择模块的v2 +软件包，并且如果导入的v2 +软件包具有有效的[semver]（https://semver.org）标记，则会以`+ incompatible'后缀进行记录。

**额外细节**

请熟悉上面[“语义导入版本控制”]（https://github.com/golang/go/wiki/Modules#semantic-import-versioning）部分中的内容。

首先回顾一些通常有用但在考虑本FAQ中描述的行为时要记住的特别重要的核心原则会有所帮助。

当`go`工具在模块模式下运行时（例如，`GO111MODULE = on`），以下核心原则是_always_ true：

1.软件包的导入路径定义了软件包的身份。
   *具有_different_导入路径的软件包被视为_different_软件包。
   *具有_same_导入路径的软件包被视为_same_软件包（即使VCS标签说软件包具有不同的主要版本，也是如此）。
2.不带“ / vN”的导入路径被视为v1或v0模块（即使导入的软件包未选择加入模块且具有表示主要版本大于1的VCS标记，也是如此）。
3.在模块的`go.mod`文件开始处声明的模块路径（例如`module foo / v2`）均为：
   *该模块身份的明确声明
   *关于如何通过使用代码导入该模块的明确声明

正如我们将在下一个FAQ中看到的那样，当`go`工具在模块模式下为_not_时，这些原理并不总是正确的，但是当`go`工具在模块模式下为_is_时，这些原理总是正确的。

简而言之，后缀“ + incompatible”表示当满足以下条件时，以上原则2有效：
*导入的软件包尚未选择加入模块，并且
*它的VCS标签说主要版本大于1，并且
*原则2覆盖了VCS标签-没有`/ vN`的导入路径被视为v1或v0模块（即使VCS标签另有说明）

当`go`工具处于模块模式时，它将假定非模块v2 +软件包不了解语义导入版本控制，并将其视为该软件包的v1版本系列的一个（不兼容）扩展（以及++ compatible）后缀表示“执行”工具正在执行此操作）。

**例**

假设：
*`oldpackage`是一个在引入模块之前的软件包
*`oldpackage`从未选择使用模块（因此本身没有`go.mod`）
*`oldpackage`具有有效的semver标签`v3.0.1`，这是它的最新标签

在这种情况下，例如从模块M内部运行`go get oldpackage @ latest`将在模块M的`go.mod`文件中记录以下内容：

```
需要旧版本v3.0.1 +不兼容
```

请注意，在上面的`go get`命令或已记录的`require`指令的oldpackage`末尾没有使用`/ v3` –在模块路径和导入路径中使用`/ vN`是[[语义导入版本控制]（https://github.com/golang/go/wiki/Modules#semantic-import-versioning），并且由于未选择“ oldpackage”，`oldpackage`并未表示接受并理解了语义导入版本控制。通过在oldpackage本身中包含一个go.mod文件将其分为模块。换句话说，即使`oldpackage`具有[v3.0.1]的[semver]（https://semver.org）标记，也不会授予`oldpackage` [语义导入版本控制]（https： //github.com/golang/go/wiki/Modules#semantic-import-versioning）（例如在导入路径中使用`/ vN`），因为`oldpackage`尚未表示希望这样做。

后缀“ + incompatible”表示“ oldpackage”的“ v3.0.1”版本尚未主动加入模块，因此假定“ oldpackage”的“ v3.0.1”版本是_not_了解语义导入版本控制或如何 在导入路径中使用主要版本。 因此，以[模块模式]（https://github.com/golang/go/wiki/Modules#when-do-i-get-old-behavior-vs-new-module-based-behavior）操作时， “执行”工具会将“ oldpackage”的非模块“ v3.0.1”版本视为“ oldpackage”的v1版本系列的（不兼容）扩展，并假定“ oldpackage”的v3.0.1`版本具有 没有语义导入版本控制的意识，并且后缀+ incompatible表示“ go”工具正在这样做。

根据语义导入版本控制，将“ oldpackage”的“ v3.0.1”版本视为v1发行系列的一部分的事实意味着，例如，版本“ v1.0.0”，“ v2.0.0”和“ v3”。 0.1`总是使用相同的导入路径导入

```
import  "oldpackage"
```

再次注意，在“ oldpackage”末尾没有使用“ / v3”。

通常，具有不同导入路径的软件包是不同的软件包。 在此示例中，给定的“ oldpackage”的版本“ v1.0.0”，“ v2.0.0”和“ v3.0.1”都将使用相同的导入路径导入，因此，它们将被构建视为相同的软件包（ 再次因为`oldpackage`尚未选择使用语义导入版本控制），而`oldpackage`的单个副本最终出现在任何给定的版本中。 （使用的版本将是所有`require`指令中列出的版本的语义上最高的版本；请参见[“版本选择”]（https://github.com/golang/go/wiki/Modules#version-selection））。

如果我们假设稍后创建了新的`oldpackage`的`v4.0.0`版本，该版本采用了模块并因此包含一个`go.mod`文件，则表明'oldpackage'现在已经理解了语义导入的权利和责任。 版本控制，因此基于模块的使用者现在可以在导入路径中使用`/ v4`进行导入：


```
import  "oldpackage/v4"
```

and the version would be recorded as:

```
require  oldpackage/v4  v4.0.0
```

现在，“ oldpackage / v4”与“ oldpackage”是不同的导入路径，因此是不同的软件包。如果构建中的某些使用方具有“ import“ oldpackage / v4”导入，而同一构建中的其他使用方具有import“ oldpackage”，则两个副本（每个导入路径一个）最终将在模块感知的构建中。作为允许逐步采用模块的策略的一部分，这是理想的。另外，即使在模块退出其当前过渡阶段之后，也希望此行为允许随着时间的推移逐步进行代码演化，其中不同的使用者以不同的速率升级到较新版本（例如，允许大型版本中的不同使用者选择升级）以不同的价格从“ oldpackage / v4”到将来的“ oldpackage / v5”）。

###如果未启用模块支持，如何在版本中处理v2 +模块？ 1.9.7 +，1.10.3 +和1.11中的“最小模块兼容性”如何工作？

在考虑尚未加入模块的较旧的Go版本或Go代码时，语义导入版本控制具有与v2 +模块相关的显着向后兼容性含义。

如上文[“语义导入版本控制”]（https://github.com/golang/go/wiki/Modules#semantic-import-versioning）部分中所述：
*版本v2或更高版本的模块必须在其`go.mod`中声明的自己的模块路径中包含`/ vN`。
*基于模块的使用者（即已选择加入模块的代码）必须在导入路径中包含`/ vN`才能导入v2 +模块。

但是，预计生态系统将以不同的采用模块和语义导入版本控制的速度进行。

如[主要版本]中[[如何发布v2 +模块]]（https://github.com/golang/go/wiki/Modules#releasing-modules-v2-or-higher）部分中所述， v2 +模块的作者会创建子目录，例如“ mymodule / v2”或“ mymodule / v3”，并在这些子目录下移动或复制相应的软件包。这意味着传统的导入路径逻辑（即使在旧的Go版本中，如Go 1.8或1.7）也会在看到诸如import“ mymodule / v2 / mypkg”之类的import语句时找到合适的软件包。因此，即使未启用模块支持，也可以找到并使用“主要子目录” v2 +模块中的软件包（这是因为您正在运行Go 1.11且未启用模块，还是因为您正在运行旧版本，如Go）没有完整模块支持的1.7、1.8、1.9或1.10）。有关“主要子目录”的更多详细信息，请参见[“如何发布v2 +模块”]（https://github.com/golang/go/wiki/Modules#releasing-modules-v2-or-higher）部分做法。

该常见问题解答的其余部分着重于[“如何发布v2 +模块”]（https://github.com/golang/go/wiki/Modules#releasing-modules-v2-或更高））。在“主要分支”方法中，不创建`/ vN`子目录，而是通过`go.mod`文件并通过将semver标签应用于提交来传达模块版本信息（通常位于`master`上，但是可能在不同的分支上）。


为了在当前过渡时期提供帮助，将[最小模块兼容性]（https://go-review.googlesource.com/c/go/+/109340）引入到Go 1.11中，以提供对Go代码的更大兼容性尚未选择使用模块，并且“最小模块兼容性”也已向后移植到Go 1.9.7和1.10.3（在这些版本中，鉴于那些较旧的Go版本没有完整版本，这些版本始终在禁用全模块模式的情况下有效运行模块支持）。

“最小模块兼容性”的主要目标是：

1.允许较早的Go版本1.9.7+和1.10.3+能够更轻松地编译在导入路径中使用带有/ vN的语义导入版本控制的模块，并在[模块模式]（https在Go 1.11中已禁用：//github.com/golang/go/wiki/Modules#when-do-i-get-old-behavior-vs-new-module-based-behavior）。

2.允许旧代码能够使用v2 +模块，而在使用v2 +模块时不需要旧的使用者代码立即更改为使用新的`/ vN`导入路径。

3.这样做而无需依赖模块作者来创建/ vN子目录。

**其他详细信息–“最小模块兼容性” **

“最小模块兼容性”仅在完整的[模块模式]下生效（https://github.com/golang/go/wiki/Modules#when-do-i-get-old-behavior-vs-new-module-based -behavior）对于“ go”工具是禁用的，例如，如果您在Go 1.11中设置了“ GO111MODULE = off”，或者使用的是Go 1.9.7+或1.10.3+版本。

当v2 +模块作者创建_not_创建`/ v2`或`/ vN`子目录时，您转而依赖于Go 1.9.7 +，1.10.3 +和1.11中的“最小模块兼容性”机制：

* _not_选择加入模块的软件包将_not_在任何导入的v2 +模块的导入路径中包括主版本。
*相反，_has_选择加入模块_m​​ust_的软件包在导入路径中包括主版本，以导入任何v2 +模块。
  *如果软件包选择了模块，但是在导入v2 +模块时在导入路径中未包含主版本，则当`go`工具在全模块模式下运行时，它将不会导入该模块的v2 +版本。 （假定已选择使用模块的软件包“讲”语义导入版本控制。如果`foo`是具有v2 +版本的模块，则在语义导入版本控制下说`import“ foo”`意味着导入v1语义导入版本控制系列。 foo）。
*用于实现“最小模块兼容性”的机制故意非常狭窄：
  *整个逻辑是–当以GOPATH模式运行时，如果import语句位于选择加入模块的代码内部，则在删除`/ vN`之后，将再次尝试包含`/ vN`的无法解析的import语句。是，将有效语句“ go.mod”导入树中“ .go”文件中的语句）。
  *最终的结果是，位于模块内部代码中的import语句（例如`import“ foo / v2”`）仍将在GOPATH模式下以1.9.7 +，1.10.3 +和1.11正确编译，并且它将就像说“ import“ foo”`（不带/ v2）一样进行解析，这意味着它将使用驻留在您的GOPATH中的foo版本，而不会被多余的/ v2所迷惑。

*“最小模块兼容性”不会影响任何其他内容，包括它不会影响“ go”命令行中使用的路径（例如“ go get”或“ go list”的参数）。
*这种过渡性的“最小模块感知”机制有意打破了“将具有不同导入路径的软件包视为不同的软件包”的规则，以实现非常具体的向后兼容性目标–允许旧代码在使用v2 +模块时进行编译，而无需修改。稍微详细一点：
  *如果旧代码使用v2 +模块的唯一方法是首先更改旧代码，则对整个生态系统来说将是一个更大的负担。
  *如果我们不修改旧代码，则该旧代码必须与v2 +模块的模块前导入路径一起使用。
  *另一方面，选择加入模块的新代码或更新代码必须对v2 +模块使用新的`/ vN`导入。
  *新的导入路径不等于旧的导入路径，但是都允许它们在一个构建中工作，因此，我们有两个不同的功能导入路径可以解析为同一程序包。
  *例如，当在GOPATH模式下操作时，基于模块的代码中出现的“ import” foo / v2“`解析为与” import“ foo”`相同的代码存在于您的GOPATH中，并且构建最终以foo –特别是GOPATH磁盘上的任何版本。这允许带有import“ foo / v2”的基于模块的代码即使在1.9.7 +，1.10.3 +和1.11的GOPATH模式下也可以编译。
*相反，当`go`工具在全模块模式下运行时：
   *“具有不同导入路径的软件包是不同的软件包”规则没有例外（包括在完全模块模式下改进了供应商，以也遵守此规则）。
   *例如，如果`go`工具处于完整模块模式，而`foo`是v2 +模块，则`import“ foo”`要求的是v1版本的`foo`相对于`import“ foo / v2” `要求v2版本的`foo`。

###如果我创建go.mod但不将semver标记应用于存储库，会发生什么情况？

[semver]（https://semver.org）是模块系统的基础。为了向消费者提供最佳体验，鼓励模块作者使用semver VCS标签（例如，“ v0.1.0”或“ v1.2.3-rc.1”），但严格要求不使用semver VCS标签：

1.模块必须遵循_semver规范_，以使`go`命令的行为与所记录的一样。这包括遵循关于如何以及何时允许更改的semver规范。

2.消费者使用semver版本以[pseudo-version]（https://tip.golang.org/cmd/go/#hdr-Pseudo_versions）的形式记录不具有_VCS标签_的模块。通常，这将是v0主版本，除非模块作者在[“主子目录”]之后构造了v2 +模块（https://github.com/golang/go/wiki/Modules#releasing-modules-v2-or-更高的方法。

3.因此，不应用semver VCS标记且未创建“主要子目录”的模块将有效地声明自己属于semver v0主版本系列，并且基于模块的使用者将其视为具有semver v0主版本。版。

###模块可以依赖于其自身的不同版本吗？
一个模块可以依赖于其自身的不同主要版本：总的来说，这相当于依赖于不同的模块。出于各种原因，这可能很有用，包括允许将模块的主要版本实现为围绕其他主要版本的填充程序。

此外，一个模块可以在一个周期中依赖于其自身的不同主要版本，就像两个完全不同的模块可以在一个周期中彼此​​依赖一样。

但是，如果您不希望模块依赖于其自身的不同版本，则可能是错误的征兆。例如，打算从v3模块导入软件包的.go代码可能在import语句中缺少必需的`/ v3`。根据本身的v1版本，该错误可能表现为v3模块。

如果您惊讶地看到某个模块依赖于其自身的不同版本，那么值得回顾一下[“语义导入版本控制”]（https://github.com/golang/go/wiki/Modules#semantic- import-versioning）部分以及常见问题[[“如果我没有看到期望的依赖版本，该怎么检查？”]（https://github.com/golang/go/wiki/Modules#what-c​​an -i  - 检查 - 如果-I-AM-不是视之 - 的 - 预期的版本的-A-依赖）。

两个循环中的每个程序包可能彼此不依赖仍然是一个约束。

## FAQS —多模块存储库

###什么是多模块存储库？

多模块存储库是一个包含多个模块的存储库，每个模块都有自己的go.mod文件。每个模块均从包含其go.mod文件的目录开始，并递归包含该目录及其子目录中的所有程序包，但不包括包含另一个go.mod文件的任何子树。

每个模块都有自己的版本信息。存储库根目录下的模块的版本标签必须包含相对目录作为前缀。例如，考虑以下存储库

```
my-repo
|____foo
| |____rop
| | |____go.mod
```

The tag for version 1.2.3 of module "my-repo/foo/rop" is "foo/rop/v1.2.3".

Typically, the path for one module in the repository will be a prefix of the others. For example, consider this repository:

```
my-repo
|____go.mod
|____bar
|____foo
| |____rop
| |____yut
|____mig
| |____go.mod
| |____vub
```

！[图。 顶级模块的路径是另一个模块的路径的前缀。]（https://github.com/jadekler/module-testing/blob/master/imagery/multi_module_repo.png）

_图。 顶级模块的路径是另一个模块的路径的前缀。

该存储库包含两个模块。 但是，模块“ my-repo”是模块“ my-repo / mig”的路径的前缀。

###我应该在一个存储库中有多个模块吗？

在这样的配置中添加模块，删除模块和版本控制模块需要相当的谨慎和考虑，因此，管理单个模块存储库而不是现有存储库中的多个模块几乎总是更容易，更简单。

拉斯·考克斯（Russ Cox）在[＃26664]（https://github.com/golang/go/issues/26664#issuecomment-455232444）中评论：

>对于除超级用户以外的所有用户，您可能希望采用一种惯例，即一个仓库=一个模块。对于代码存储选项的长期发展很重要，仓库_can_包含多个模块，但是默认情况下您几乎肯定不想这样做。

关于如何使多模块更有效的两个示例：
 *来自存储库根目录的`go test。/ ...`将不再测试存储库中的所有内容
 *您可能需要通过`replace`指令例行管理模块之间的关系。

但是，除了这两个示例之外，还有其他细微差别。如果您正在考虑在单个存储库中包含多个模块，请仔细阅读此[sub-section]（https://github.com/golang/go/wiki/Modules#faqs--multi-module-repository）中的常见问题解答。

在两个示例场景中，一个存储库中可以有多个`go.mod`是有意义的：

1.如果您有一些用法示例，这些示例本身具有一组复杂的依赖关系（例如，也许您的软件包很小，但包括一个将示例软件包与kubernetes结合使用的示例）。在这种情况下，对于您的存储库来说，拥有一个带有自己的`go.mod`的`examples'或`_examples`目录是有意义的，例如[here]（https://godoc.org/github.com/ lOOV的/ hrtime）。

2.如果您的存储库具有一组复杂的依赖关系，但是您的客户端API的依赖关系集较少。在某些情况下，可能有一个`api`或`clientapi`或具有自己的`go.mod`的类似目录，或者将该`clientapi`分离到其自己的存储库中是有意义的。

但是，对于这两种情况，如果您考虑为多组间接依赖项创建性能或下载大小的多模块存储库，则强烈建议您首先尝试使用GOPROXY，它将在Go中默认启用1.13。使用GOPROXY通常等同于可能会因创建多模块存储库而带来的任何性能优势或依赖性下载大小优势。

###是否可以将模块添加到多模块存储库？

是。但是，此问题有两类：

第一类：要添加模块的软件包尚未处于版本控制中（新软件包）。这种情况很简单：将包和go.mod添加到同一提交中，标记该提交，然后推送。

第二类：添加模块的路径在版本控制中，并且包含一个或多个现有软件包。这种情况需要相当多的护理。为了说明，再次考虑以下存储库（现在位于github.com位置，以更好地模拟真实世界）：

```
github.com/my-repo
|____go.mod
|____bar
|____foo
| |____rop
| |____yut
|____mig
| |____vub
```

考虑添加模块“ github.com/my-repo/mig”。 如果要采用与上述相同的方法，则可以通过两个不同的模块提供软件包/ my-repo / mig：旧版本的“ github.com/my-repo”和新的独立模块“ github”。 com / my-repo / mig。如果两个模块都处于活动状态，则导入“ github.com/my-repo/mig”将在编译时导致“模棱两可的导入”错误。

解决此问题的方法是使新添加的模块取决于“雕刻”出的模块，然后再将其雕刻出来。

假设“ github.com/my-repo”当前位于v1.2.3，让我们通过上面的存储库逐步进行操作：

1. Add github.com/my-repo/mig/go.mod:
    
    ```
    cd path-to/github.com/my-repo/mig
    go mod init github.com/my-repo/mig
    
    # Note: if "my-repo/mig" does not actually depend on "my-repo", add a blank
    # import.
    # Note: version must be at or after the carve-out.
    go mod edit -require github.com/myrepo@v1.3
    ```

1. `git commit`
1. `git tag v1.3.0`
1. `git tag mig/v1.0.0`
1. Next, let's test these. We can't `go build` or `go test` naively, since the go commands would try to fetch each dependent module from the module cache. So, we need to use replace rules to cause `go` commands to use the local copies:

    ```    
    cd path-to/github.com/my-repo/mig
    go mod edit -replace github.com/my-repo@v1.3.0=../
    go test ./...
    go mod edit -dropreplace github.com/my-repo@v1.3.0
    ```

1. `git push origin master v1.2.4 mig/v1.0.0` push the commit and both tags

请注意，将来[golang.org/issue/28835](https://github.com/golang/go/issues/28835）应该会使测试步骤变得更简单。

还要注意，在次要版本之间，代码已从模块“ github.com/my-repo”中删除。不将其视为主要更改似乎很奇怪，但是在这种情况下，传递性依存关系继续在其原始导入路径中提供已删除软件包的兼容实现。

###是否可以从多模块存储库中删除模块？

是的，具有与上述相同的两种情况和类似的步骤。

###一个模块可以依赖内部模块吗？

是。一个模块中的程序包可以从另一个模块中导入内部程序包，只要它们共享与内部/路径组件相同的路径前缀即可。例如，考虑以下存储库：

```
我的回购
| ____ go.mod
| ____内部
| ____ FOO
| | ____ go.mod
```

在这里，只要模块“ my-repo / foo”依赖于模块“ my-repo”，软件包foo就可以导入/ my-repo / internal。同样，在以下存储库中：

```
我的回购
| ____内部
| | ____ go.mod
| ____ FOO
| | ____ go.mod
```

在这里，只要模块“ my-repo / foo”依赖于模块“ my-repo / internal”，软件包foo就可以导入my-repo / internal。两者的语义相同：由于my-repo是my-repo / internal和my-repo / foo之间的共享路径前缀，因此允许foo包导入内部包。

###额外的go.mod可以排除不必要的内容吗？模块是否等效于.gitignore文件？

在单个存储库中具有多个`go.mod`文件的另一种用例是，该存储库是否包含应从模块中删除的文件。例如，存储库可能具有Go模块不需要的非常大的文件，或者多语言存储库可能具有许多非Go文件。

目录中的空`go.mod`将导致该目录及其所有子目录从顶级Go模块中排除。

如果排除的目录不包含任何`.go`文件，则除了放置空的`go.mod`文件之外，不需要其他步骤。如果排除的目录中确实包含`.go`文件，请首先仔细查看[此多模块存储库部分]中的其他常见问题解答（https://github.com/golang/go/wiki/Modules#faqs--multi-模块储存库）。

##常见问题解答-最低版本选择

###最少的版本选择是否会使开发人员无法获得重要的更新？

请参阅问题“最小版本选择是否会使开发人员无法获得重要更新？”在较早的[来自官方提案讨论的常见问题解答]（https://github.com/golang/go/issues/24301#issuecomment-371228664）中。

##常见问题解答-可能的问题

###如果发现问题，可以进行哪些常规检查？

*通过运行`go env`来仔细检查模块是否启用，以确认它对只读的`GOMOD`变量不显示空值。
   *注意：永远不要将GOMOD设置为变量，因为它实际上是go env输出的只读调试输出。
   *如果您将`GO111MODULE = on`设置为启用模块，请仔细检查它是否不是偶然的`GO111MODULES = on`。 （人们有时自然会包含`S`，因为该功能通常称为“模块”）。
*如果期望使用供应商，请检查是否将-mod = vendor标志传递给go build或类似名称，或者已设置GOFLAGS = -mod = vendor。
   *默认情况下，模块会忽略`vendor`目录，除非您要求`go`工具使用`vendor`。
*检查`go list -m all'通常会很有帮助，以查看为您的构建选择的实际版本列表
  *`go list -m all`通常会为您提供更多细节，相比之下，您只需要看`go.mod`文件即可。
*如果以某种方式运行`go get foo`失败，或者如果`go build`在特定软件包`foo`上失败，则检查`go get -v foo`或`go get的输出通常会很有帮助。 -v -x foo`：
  *通常，“去获取”通常会比“去构建”提供更多详细的错误消息。
  *进入`get'的`-v`标志要求打印更多详细信息，尽管要记住，根据远程存储库的配置，某些“错误”（例如404错误）可能会发生。
  *如果问题的本质仍然不清楚，您也可以尝试更详细的`go get -v -x foo`，它还会显示git或其他VCS命令。 （如果有保证，通常可以在`go`工具的上下文之外执行相同的git命令，以进行故障排除）。
*您可以查看是否使用了特别旧的git版本
  *较旧版本的git是`vgo`原型和Go 1.11 beta的常见问题根源，但在GA 1.11中却很少出现。
* Go 1.11中的模块缓存有时可能会导致各种错误，主要是如果以前存在网络问题或并行执行多个`go`命令（请参阅[＃26794]（https://github.com/golang/go/issues/ 26794），针对Go 1.12）。作为tro

当前正在检查的错误可能是由于构建中没有特定模块或软件包的预期版本而引起的第二个问题。因此，如果导致特定错误的原因不明显，则可以按照下一个FAQ中的说明对您的版本进行抽查。

###如果没有看到期望的依赖版本，该怎么检查？

第一步是运行`go mod tidy`。有可能解决此问题的可能性，但这也有助于将`go.mod`文件相对于`.go`源代码置于一致的状态，这将使以后的调查更加容易。 （如果`go mod tidy`本身以您不希望的方式更改了依赖项的版本，请先阅读[关于'go mod tidy'的此常见问题解答]]（https://github.com/golang/go/wiki/ ＃为什么要在我的gomod中进行mod-tidy记录间接和测试依赖项的解析。如果那不能解释它，您可以尝试重置您的go.mod然后运行` go list -mod = readonly all`，这可能会提供关于需要更改其版本的更具体的消息）。

2.第二步通常应该检查“ go list -m all”以查看为您的构建选择的实际版本的列表。 `go list -m all'向您显示最终选择的版本，包括用于间接依赖性的版本以及在解决所有共享依赖性的版本之后的版本。它还显示了任何“替换”和“排除”指令的结果。

3.一个好的下一步是检查“ go mod graph”或“ go mod graph |”的输出。 grep <感兴趣模块>`。 go mod graph`打印模块需求图（包括考虑的替换）。输出中的每一行都有两个字段：第一列是使用模块，第二列是该模块的要求之一（包括该使用模块所需的版本）。这是查看哪些模块需要特定依赖项的快速方法，包括当构建的依赖项具有与构建中的不同使用者不同的所需版本时（如果是这种情况，熟悉该模块很重要）。以上[[版本选择]]（https://github.com/golang/go/wiki/Modules#version-selection）部分中所述的行为）。

“ go mod why -m <模块>”在这里也很有用，尽管它通常对于查看为什么完全包含依赖项（而不是为什么依赖项以特定版本结尾）更有用。

“ go list”提供了更多的查询变体，可以在需要时询问模块。以下是一个示例，它将显示构建中使用的确切版本，不包括仅测试依赖项：
```
转到列表-deps -f'{{with .Module}} {{。Path}} {{.Version}} {{end}}'。/ ... |排序-u
```

可以在可运行的“通过示例访问模块”中查看更详细的命令和示例集[walkthough]（https://github.com/go-modules-by-example/index/tree/master/ 018_go_list_mod_graph_why）。

导致意外版本的一种原因可能是由于某人创建了一个无效的或意外的`go.mod`文件，这不是故意的，或者是相关的错误（例如：`v2.0.1`版本的模块可能错误地声明了自己为是在其`go.mod`中的`module foo`中没有必需的`/ v2`；在`.go`代码中用于导入v3模块的import语句可能缺少必需的`/ v3`；`require` v4模块的`go.mod`中的语句可能缺少必需的`/ v4`）。因此，如果您看不到引起特定问题的原因，那么值得首先重新阅读[“ go.mod”]中的资料（https://github.com/golang/go/wiki / Modules＃gomod）和上述[“语义导入版本控制”]（https://github.com/golang/go/wiki/Modules#semantic-import-versioning）部分（鉴于这些内容包括模块必须遵循的重要规则），以及然后花几分钟来检查最相关的`go.mod`文件和导入语句。

###为什么会出现错误“找不到提供软件包foo的模块”？

这是一条常见的错误消息，可能会因几种不同的根本原因而发生。

在某些情况下，此错误仅是由于路径键入错误引起的，因此第一步可能应该是根据错误消息中列出的详细信息再次检查错误的路径。

如果您尚未这样做，那么下一步通常是尝试`go get -v foo`或`go get -v -x foo`：
*通常，“去获取”通常会比“去构建”提供更多详细的错误消息。
*请参阅本节中的第一个故障排除常见问题解答[上方]（https://github.com/golang/go/wiki/Modules#what-are-some-general-things-i-can-spot-check-if-i （请参阅问题）以获取更多详细信息。

其他一些可能的原因：

*如果您已发出`go build`或`go build .`但在当前目录中没有任何`.go`源文件，则可能会看到错误“找不到提供软件包foo的模块”。如果这是您遇到的问题，则解决方案可能是另一种调用，例如`go build。/ ...`（其中`。/ ...`展开以匹配当前模块中的所有软件包）。参见[＃27122]（https://github.com/golang/go/issues/27122）。

* Go 1.11中的模块缓存可能导致此错误，包括面对网络问题或多个并行执行的“ go”命令。这在Go 1.12中已解决。请参阅本节中的第一个故障排除常见问题解答[以上]（https://github.com/golang/go/wiki/Modules#what-are-some-general-things-i-can-spot-check-if-i-看到问题）以获取更多详细信息和可能的纠正步骤。

###为什么'go mod init'给出错误'无法确定源目录的模块路径'？

没有任何参数的`go mod init`会尝试根据不同的提示（例如VCS元数据）猜测正确的模块路径。但是，不希望`go mod init`总是能够猜测正确的模块路径。

如果`go mod init`出现此错误，则这些试探法无法猜测，您必须自己提供模块路径（例如`go mod init github.com / you / hello`）。

###我有一个复杂的依赖性问题，尚未选择模块。我可以使用其当前依赖项管理器中的信息吗？

是。这需要一些手动步骤，但在某些更复杂的情况下可能会有所帮助。

当您在初始化自己的模块时运行`go mod init`时，它将通过将`Gopkg.lock`，`glide.lock`或`vendor.json`之类的配置文件转换成`go'来自动从先前的依赖项管理器转换。包含相应的`require`指令的.mod文件。例如，预先存在的“ Gopkg.lock”文件中的信息通常描述所有直接和间接依赖项的版本信息。

但是，如果改为添加尚未选择加入模块本身的新依赖项，则任何先前的依赖项管理器都不会使用类似的自动转换过程，而新的依赖项可能已经在使用该转换过程。如果该新依赖项本身具有发生了重大更改的非模块依赖项，则在某些情况下可能会导致不兼容问题。换句话说，新依赖项的先前依赖项管理器不会自动使用，在某些情况下，这可能会导致间接依赖项出现问题。

一种方法是对有问题的非模块直接依赖项运行“ go mod init”，以从其当前的依赖项管理器进行转换，然后使用生成的临时“ go.mod”中的“ require”指令填充或更新“ go” .mod`在您的模块中。

例如，如果github.com/some/nonmodule是当前正在使用另一个依赖管理器的模块的直接依赖问题，则可以执行以下操作：


```
$ git clone -b v1.2.3 https://github.com/some/nonmodule /tmp/scratchpad/nonmodule
$ cd /tmp/scratchpad/nonmodule
$ go mod init
$ cat go.mod
```

可以将临时`go.mod`中产生的`require`信息手动移至模块的实际`go.mod`中，或者您可以考虑使用https://github.com/rogpeppe/gomodmerge针对此用例的社区工具。另外，您将需要在实际的go.mod中添加一个require github.com/some/nonmodule v1.2.3以匹配您手动克隆的版本。

在此[＃28489注释]（https://github.com/golang/go/issues/28489#issuecomment-454795390）中，针对Docker使用此技术的一个具体示例说明了如何获取一致的版本集
docker依赖项以避免github.com/sirupsen/logrus与github.com/Sirupsen/logrus之间的区分大小写的问题。

###如何解决由于导入路径与声明的模块标识不匹配而导致的“解析go.mod：意外的模块路径”和“错误加载模块要求”错误？

####为什么会发生此错误？

通常，模块通过“ module”指令（例如“ module example.com/m”）在其“ go.mod”中声明其身份。这是该模块的“模块路径”，“ go”工具在该声明的模块路径与任何使用者使用的导入路径之间强制保持一致性。如果模块的“ go.mod”文件读为“ module example.com/m”，那么使用者必须使用以该模块路径开头的导入路径（例如，“ import“ example.com/m””）从该模块导入软件包。或“导入” example.com/m/sub/pkg”）。

如果使用者使用的导入路径与相应的声明的模块路径不匹配，那么“ go”命令将报告“解析go.mod：意外的模块路径”致命错误。另外，在某些情况下，“执行”命令随后将报告更通用的“错误加载模块要求”错误。

导致此错误的最常见原因是是否更改了名称（例如，将github.com/Sirupsen/logrus更改为github.com/sirupsen/logrus），或者某个模块之前有时通过两个不同的名称使用过由于虚荣导入路径（例如，“ github.com/golang/sync”与推荐的“ golang.org/x/sync”）而导致的模块之间的冲突。

如果您仍然有一个较旧的名称（例如，github.com/Sirupsen/logrus）或非规范名称（例如，github.com/golang/sync）导入的依赖项，那么这可能会导致问题。 ），但该依赖项随后采用了模块，现在在其`go.mod`中声明其规范名称。然后，当发现模块的升级版本声明不再与旧的导入路径匹配的规范模块路径时，在升级期间会触发此错误。




#### Example problem scenario
*您间接取决于`github.com / Quasilyte / go-consistent`。
*该项目采用模块，然后将其名称更改为github.com/quasilyte/go-consistent（将Q更改为小写q），这是一个重大更改。 GitHub从旧名称转发到新名称。
*您运行`go get -u`，它会尝试升级所有直接和间接依赖项。
*试图升级github.com/Quasilyte/go-consistent，但是现在发现的最新的go.mod读为module github.com/quasilyte/go-consistent。
*整体升级操作无法完成，出现以下错误：

>前往：github.com/Quasilyte/go-consistent@v0.0.0-20190521200055-c6f3937de18c：解析go.mod：意外的模块路径“ github.com/quasilyte/go-consistent”
>获取：错误加载模块要求

####解决

错误的最常见形式是：

>转到：example.com/some/OLD/name@vX.Y.Z：解析go.mod：意外的模块路径“ example.com/some/NEW/name”

如果您访问存储库中的“ example.com/some/NEW/name”（位于错误的右侧），则可以检查“ go.mod”文件中的最新版本或“ master”文件，以查看是否在“ go.mod”的第一行将其声明为“ module example.com/some/NEW/name”。如果是这样，则暗示您看到的是“旧模块名称”与“新模块名称”的问题。

本节的其余部分重点在于按顺序执行以下步骤来解决此错误的“旧名称”和“新名称”形式：

1.检查您自己的代码，以查看是否要使用example.com/some/OLD/name进行导入。如果是这样，请使用`example.com / some / NEW / name`更新代码以导入。

2.如果您在升级过程中收到此错误，则应尝试使用Go的提示版本进行升级，该提示版本具有更有针对性的升级逻辑（[＃26902]（https://github.com/golang/go/issues/26902） ），这通常可以避开此问题，并且在这种情况下通常还会提供更好的错误消息。注意，tip / 1.13中的`go get`参数与1.12中的参数不同。获取提示并使用它升级依赖项的示例：
```
前往golang.org/dl/gotip && gotip下载
gotip get -u全部
戈蒂莫整洁
```
由于有问题的旧导入通常是间接依赖的，因此使用tip升级然后运行`go mod tidy`可以经常升级到有问题的版本，然后再从`go.mod`中删除有问题的版本，不再需要，然后，当您返回使用Go 1.12或1.11进行日常使用时，它将进入正常运行状态。例如，请参见[here]（https://github.com/golang/go/issues/30831#issuecomment-489463638）可以将方法升级到`github.com / golang / lint`与`golang.org/ x / lint`问题。

3.如果在执行`go get -u foo`或`go get -u foo @ latest`时收到此错误，请尝试删除`-u`。这将为您提供`foo @ latest`所使用的一组依赖关系，而无需将`foo`的依赖关系升级到超过`foo`的作者在发布`foo`时可能会验证的版本之后。这在过渡时期尤其重要，因为此时foo的某些直接和间接依赖关系可能尚未采用[semver]（https://semver.org）或模块。 （常见的错误是认为`go get -u foo`仅获取`foo`的最新版本。实际上，`go get -u foo`或`go get -u foo @ latest`中的`-u`表示到_all_以获得foo的直接和间接依赖项的所有最新版本；这可能是您想要的，但是如果由于深层间接依赖项而导致失败则可能不是特别的）。

4.如果上述步骤未能解决错误，则下一种方法会稍微复杂一些，但大多数情况下应该可以解决此错误的“旧名称”和“新名称”形式。这仅使用仅来自错误消息本身的信息，并简要介绍了一些VCS历史记录。

   4.1。转到“ example.com/some/NEW/name”存储库

   4.2。确定何时将“ go.mod”文件引入该位置（例如，通过查看“ go.mod”的非常规视图或历史记录视图）。

   4.3。从之前引入`go.mod`文件的地方选择发布或提交。

   4.4。在您的`go.mod`文件中，在`replace`语句的两侧使用旧名称添加`replace`语句：
       ```
       替换example.com/some/OLD/name => example.com/some/OLD/name <version-just-before-go.mod>
       ```
使用我们先前的例子，其中github.com/Quasilyte/go-consistent是旧名称，而github.com/quasilyte/go-consistent是新名称，我们可以看到go.mod最初被引入在那里提交[00c5b0cf371a]（https://github.com/quasilyte/go-consistent/tree/00c5b0cf371a96059852487731370694d75ffacf）。该存储库未使用semver标记，因此我们将立即进行先前的提交[00dd7fb039e]（https://github.com/quasilyte/go-consistent/tree/00dd7fb039e1eff09e7c0bfac209934254254360）并使用旧的大写Quasilyte名称将其添加到替换中在`replace`的两侧：


```
replace github.com/Quasilyte/go-consistent => github.com/Quasilyte/go-consistent 00dd7fb039e
```


然后，通过有效地防止在存在`go.mod`的情况下将旧名称升级为新名称，此`replace`语句使我们能够升级解决有问题的“旧名称”与“新名称”不匹配的问题。通常，通过`go get -u`或类似的升级现在可以避免该错误。如果升级完成，则可以检查是否有人仍在导入旧名称（例如，“ go mod graph | grep github.com/Quasilyte/go-consistent”），如果没有，则可以删除“ replace”。 。 （之所以经常起作用，是因为如果使用了旧的有问题的导入路径，升级本身可能会失败，即使升级完成后也可能不会在最终结果中使用升级路径，该路径在[＃30831]（https： //github.com/golang/go/issues/30831））。

5.如果上述步骤未能解决问题，则可能是因为一个或多个依赖项的最新版本仍在使用有问题的旧导入路径。在这种情况下，重要的是确定谁仍在使用有问题的旧导入路径，并查找或打开一个问题，要求有问题的进口商更改为使用现在的规范导入路径。在上面的第2步中使用`gotip`可能会识别有问题的导入器，但并非在所有情况下都可以识别出，尤其是对于升级（[＃30661]（https://github.com/golang/go/issues/30661# issuecomment-480981833））。如果不清楚是谁在使用有问题的旧导入路径进行导入，通常可以通过创建干净的模块高速缓存，执行触发错误的一个或多个操作，然后在模块高速缓存中查找旧的有问题的导入路径来找出答案。例如：

```
export GOPATH=$(mktemp -d)
go get -u foo               # peform operation that generates the error of interest
cd $GOPATH/pkg/mod
grep -R --include="*.go" github.com/Quasilyte/go-consistent
```

6.如果这些步骤不足以解决问题，或者您是某个项目的维护者，但由于循环引用而似乎无法删除对较旧的问题导入路径的引用，请参阅有关该文档的详细说明。问题在单独的[wiki页面]上（https://github.com/golang/go/wiki/Resolving-Problems-From-Modified-Module-Path）。

最后，上述步骤集中于如何解决潜在的“旧名称”与“新名称”问题。但是，如果将`go.mod`放置在错误的位置或仅具有错误的模块路径，也会出现相同的错误消息。在这种情况下，导入该模块应始终失败。如果要导入刚刚创建且从未成功导入的新模块，则应检查`go.mod`文件是否位于正确位置，并具有与该位置对应的正确模块路径。 （最常见的方法是每个存储库使用单个`go.mod`，将单个`go.mod`文件放置在存储库根目录中，并使用存储库名称作为`module`指令中声明的模块路径）。有关更多详细信息，请参见[“ go.mod”]（https://github.com/golang/go/wiki/Modules#gomod）部分。

###为什么“开始构建”需要gcc，为什么不使用诸如net / http之类的预构建软件包？

简而言之：

>因为预构建的软件包是非模块构建的，因此无法重复使用。抱歉。现在禁用cgo或安装gcc。

这仅是选择模块时的问题（例如，通过`GO111MODULE = on`）。有关其他讨论，请参见[＃26988]（https://github.com/golang/go/issues/26988#issuecomment-417886417）。

###模块是否可以使用相对的导入，例如`import“ ./subdir”`？

否。请参见[＃26645]（https://github.com/golang/go/issues/26645#issuecomment-408572701），其中包括：

>在模块中，最后有一个子目录的名称。如果父目录显示“模块m”，则子目录将导入为“ m / subdir”，而不再是“ ./subdir”。

###某些所需的文件可能未出现在填充的供应商目录中

没有`.go`文件的目录不会被`go mod vendor`复制到`vendor`目录中。这是设计使然。

简而言之，撇开任何特定的供应商行为– Go构建的总体模型是构建软件包所需的文件应位于带有`.go`文件的目录中。

以cgo为例，修改其他目录中的C源代码不会触发重建，而是您的构建将使用陈旧的缓存条目。现在，cgo文档[包括]（https://go-review.googlesource.com/c/go/+/125297/5/src/cmd/cgo/doc.go）：


>请注意，对其他目录中文件的更改不会导致该软件包
要重新编译，因此该软件包的_all non-Go源代码应为
存储在包directory_中，而不是子目录中。

社区工具https://github.com/goware/modvendor允许您轻松地将一整套.c，.h，.s，.proto或其他文件从模块复制到`vendor`导演中。尽管这可能会有所帮助，但是如果您拥有构建`.go`文件目录之外的软件包所需的文件，则必须格外小心以确保您的go构建总体上得到正确处理（与供应商无关） 。

请参阅[＃26366]（https://github.com/golang/go/issues/26366#issuecomment-405683150）中的其他讨论。

传统供应商的另一种方法是检入模块缓存。它最终可能会获得与传统供应商类似的好处，并且在某些方面最终会获得更高的保真度。这种方法被解释为“通过示例执行模块” [演练]（https://github.com/go-modules-by-example/index/blob/master/012_modvendor/README.md）。













