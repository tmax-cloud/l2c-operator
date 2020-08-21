package v1

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
)

const (
	ConditionKeyProjectReady     = status.ConditionType("ProjectReady")
	ConditionKeyProjectRunning   = status.ConditionType("ProjectRunning")
	ConditionKeyProjectSucceeded = status.ConditionType("ProjectSucceeded")

	ConditionKeyPhaseAnalyze   = status.ConditionType("PhaseAnalyze")
	ConditionKeyPhaseDbMigrate = status.ConditionType("PhaseDBMigrate")
	ConditionKeyPhaseBuild     = status.ConditionType("PhaseBuild")
	ConditionKeyPhaseDeploy    = status.ConditionType("PhaseDeploy")
)

const (
	ReasonPhaseRunning     = status.ConditionReason("Running")
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
	SecretKeySourceUser     = "source-user"
	SecretKeySourcePassword = "source-password"
	SecretKeySourceSid      = "source-sid"
	SecretKeyTargetUser     = "target-user"
	SecretKeyTargetPassword = "target-password"
	SecretKeyTargetSid      = "target-sid"
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
