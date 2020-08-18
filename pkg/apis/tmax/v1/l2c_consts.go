package v1

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
)

const (
	ConditionKeyProjectReady   = status.ConditionType("ProjectReady")
	ConditionKeyProjectRunning = status.ConditionType("ProjectRunning")

	ConditionKeyAnalyze   = status.ConditionType("Analyze")
	ConditionKeyDbMigrate = status.ConditionType("DBMigrate")
	ConditionKeyBuild     = status.ConditionType("Build")
	ConditionKeyDeploy    = status.ConditionType("Deploy")
)

const (
	ConfigMapKeyPvc    = "pvc.yaml"
	ConfigMapKeySvc    = "svc.yaml"
	ConfigMapKeySecret = "secret.yaml"
	ConfigMapKeyDeploy = "deploy.yaml"
)

// TaskName* : Actual name of Task object
const (
	TaskNameAnalyzeMaven  = "l2c-sonar-scan-java-maven"
	TaskNameAnalyzeGradle = "l2c-sonar-scan-java-gradle"
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

const (
	PipelineResourceNameGit = "git-source"
)
