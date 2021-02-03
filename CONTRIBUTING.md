# How to Contribute

Devtron is [Apache 2.0 licensed](LICENSE) and accepts contributions via GitHub
pull requests. This document outlines some of the conventions on development
workflow, commit message formatting, contact points and other resources to make
it easier to get your contribution accepted.

We gratefully welcome improvements to issues and documentation as well as to code.

## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution. No action from you is required, but it's a good idea to see the
[DCO](DCO) file for details before you start contributing code to Devtron.

## Communications

The project uses discord for communication:

To join the conversation, simply join the **[discord](https://discord.gg/jsRG5qx2gp)**  and use the __#contrib__ channel.

## Code Structure

Devtron has following components

- [devtron](https://github.com/devtron-labs/devtron.git) main co-ordinating engine
- [git-sensor](https://github.com/devtron-labs/git-sensor.git) microservice for watching and interacting with git
- [ci-runner](https://github.com/devtron-labs/ci-runner.git) Devtron runner for executing jobs
- [guard](https://github.com/devtron-labs/guard.git) A kubernetes validating webhook for policy inforcement
- [imge-scanner](https://github.com/devtron-labs/image-scanner.git) microservice for docker image vulnerability scanning
- [kubewatch](https://github.com/devtron-labs/kubewatch.git) microservice for k8s event filtering and recording 
- [lens](https://github.com/devtron-labs/lens.git) microservice for performing analytical task
- [dashboard](https://github.com/devtron-labs/dashboard.git) UI for devtron written in react js




