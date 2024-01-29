package util

//type GitManagerImpl struct {
//	logger *zap.SugaredLogger
//}
//
//func NewGitManager(logger *zap.SugaredLogger) *GitManagerImpl {
//	return &GitManagerImpl{
//		logger: logger,
//	}
//}

//func (impl *GitManager) Checkout(rootDir string, branch string) (response, errMsg string, err error) {
//	start := time.Now()
//	defer func() {
//		util.TriggerGitOpsMetrics("Checkout", "GitCli", start, err)
//	}()
//	impl.logger.Debugw("git checkout ", "location", rootDir)
//	cmd := exec.Command("git", "-C", rootDir, "checkout", branch, "--force")
//	output, errMsg, err := impl.runCommand(cmd)
//	impl.logger.Debugw("checkout output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
//	return output, errMsg, err
//}

//operations dependent on go-git
