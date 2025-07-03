package bean

const (
	// InstallSucceededReason represents the fact that the Helm install for the
	// HelmRelease succeeded.
	InstallSucceededReason string = "InstallSucceeded"

	// InstallFailedReason represents the fact that the Helm install for the
	// HelmRelease failed.
	InstallFailedReason string = "InstallFailed"

	// UpgradeSucceededReason represents the fact that the Helm upgrade for the
	// HelmRelease succeeded.
	UpgradeSucceededReason string = "UpgradeSucceeded"

	// UpgradeFailedReason represents the fact that the Helm upgrade for the
	// HelmRelease failed.
	UpgradeFailedReason string = "UpgradeFailed"

	// TestSucceededReason represents the fact that the Helm tests for the
	// HelmRelease succeeded.
	TestSucceededReason string = "TestSucceeded"

	// TestFailedReason represents the fact that the Helm tests for the HelmRelease
	// failed.
	TestFailedReason string = "TestFailed"

	// RollbackSucceededReason represents the fact that the Helm rollback for the
	// HelmRelease succeeded.
	RollbackSucceededReason string = "RollbackSucceeded"

	// RollbackFailedReason represents the fact that the Helm test for the
	// HelmRelease failed.
	RollbackFailedReason string = "RollbackFailed"

	// UninstallSucceededReason represents the fact that the Helm uninstall for the
	// HelmRelease succeeded.
	UninstallSucceededReason string = "UninstallSucceeded"

	// UninstallFailedReason represents the fact that the Helm uninstall for the
	// HelmRelease failed.
	UninstallFailedReason string = "UninstallFailed"

	// ArtifactFailedReason represents the fact that the artifact download for the
	// HelmRelease failed.
	ArtifactFailedReason string = "ArtifactFailed"

	// DependencyNotReadyReason represents the fact that
	// one of the dependencies is not ready.
	DependencyNotReadyReason string = "DependencyNotReady"
)
