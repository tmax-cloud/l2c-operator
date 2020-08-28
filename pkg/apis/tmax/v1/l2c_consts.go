package v1

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
)

const (
	WasTypeWeblogic = "weblogic"
	WasTypeJeus     = "jeus"
)

const (
	DbTypeOracle = "oracle"
	DbTypeTibero = "tibero"
)

const (
	ConditionKeyProjectReady     = status.ConditionType("Ready")
	ConditionKeyProjectRunning   = status.ConditionType("Running")
	ConditionKeyProjectSucceeded = status.ConditionType("Succeeded")

	ConditionKeyPhaseAnalyze   = status.ConditionType("Analyze")
	ConditionKeyPhaseDbMigrate = status.ConditionType("DBMigrate")
	ConditionKeyPhaseBuild     = status.ConditionType("Build")
	ConditionKeyPhaseDeploy    = status.ConditionType("Deploy")
)

const (
	ReasonPhaseRunning     = status.ConditionReason("Running")
	ReasonPhaseCanceled    = status.ConditionReason("Canceled")
	ReasonPhaseFailed      = status.ConditionReason("Failed")
	ReasonPhaseSucceeded   = status.ConditionReason("Succeeded")
	ReasonPhaseNotExecuted = status.ConditionReason("Not executed yet")
)

const (
	DbConfigMapKeyPvc    = "pvc.yaml"
	DbConfigMapKeySvc    = "svc.yaml"
	DbConfigMapKeySecret = "secret.yaml"
	DbConfigMapKeyDeploy = "deploy.yaml"
)

const (
	DbSecretKeySourceUser     = "source-user"
	DbSecretKeySourcePassword = "source-password"
	DbSecretKeySourceSid      = "source-sid"
	DbSecretKeyTargetUser     = "target-user"
	DbSecretKeyTargetPassword = "target-password"
	DbSecretKeyTargetSid      = "target-sid"
)

// TaskName* : Actual name of Task object
const (
	TaskNameAnalyzeMaven  = "l2c-sonar-scan-java-maven"
	TaskNameAnalyzeGradle = "l2c-sonar-scan-java-gradle"

	TaskNameDbMigration = "l2c-db-migration"

	TaskNameBuild  = "l2c-build"
	TaskNameDeploy = "l2c-deploy"
)

// PipelineResourceName* : Resource name written in Pipeline.spec.resources
type PipelineResourceName string

const (
	PipelineResourceNameGit   = PipelineResourceName("git-source")
	PipelineResourceNameImage = PipelineResourceName("image")
)

// PipelineTaskName* : Task name written in Pipeline.spec.tasks
type PipelineTaskName string

const (
	PipelineTaskNameAnalyze = PipelineTaskName("analyze")
	PipelineTaskNameMigrate = PipelineTaskName("migrate")
	PipelineTaskNameBuild   = PipelineTaskName("build")
	PipelineTaskNameDeploy  = PipelineTaskName("deploy")
)

const (
	PipelineParamNameSonarUrl        = "sonar-url"
	PipelineParamNameSonarToken      = "sonar-token"
	PipelineParamNameSonarProjectKey = "sonar-project-id"
)
