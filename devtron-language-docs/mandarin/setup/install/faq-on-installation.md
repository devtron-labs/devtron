
#常见问题

<details>

<summary>1.我如何知道安装何时完成？</summary>

运行以下命令检查安装状态：

\```bash

kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}'

\```

安装过程完成后，上述命令将显示“Applied”。安装过程最多可能需要30分钟。

</details>

<details>

<summary>2.我能如何跟踪安装的进度？</summary>

运行以下命令查看Pod的日志：

\```bash

pod=$(kubectl -n devtroncd get po -l app=inception -o jsonpath='{.items[0].metadata.name}')&& kubectl -n devtroncd logs -f $pod

\```

</details>

<details>

<summary>3.如果Devtron安装程序日志包含错误，我该如何重新启动安装？</summary>

首先，运行以下命令来清理 Devtron 安装程序安装的组件：

\```bash

cd devtron-installation-script/

kubectl delete -n devtroncd -f yamls/

kubectl -n devtroncd patch installer installer-devtron --type json -p '[{"op": "remove", "path": "/status"}]'

\```

接下来，[安装 Devtron]（./install-devtron.md）

</details>


若仍然遇到问题，请通过 [Discord]（https://discord.gg/jsRG5qx2gp）与我们联系。
