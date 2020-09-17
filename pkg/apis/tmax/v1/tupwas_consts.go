package v1

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
)

const (
	WasTypeWeblogic = "weblogic"
	WasTypeJeus     = "jeus"
)

const (
	WasConditionKeyProjectReady     = status.ConditionType("Ready")
	WasConditionKeyProjectAnalyzing = status.ConditionType("Analyzing")
	WasConditionKeyProjectRunning   = status.ConditionType("Running")
	WasConditionKeyProjectSucceeded = status.ConditionType("Succeeded")
)

// TaskName* : Actual name of Task object
const (
	TaskNameGitClone   = "l2c-git-clone"
	TaskNameAnalyzeWas = "l2c-tup-jeus"

	TaskNameBuild  = "l2c-build"
	TaskNameDeploy = "l2c-deploy"
)

// PipelineTaskName* : Task name written in Pipeline.spec.tasks
type WasPipelineTaskName string

const (
	WasPipelineTaskNameClone   = WasPipelineTaskName("git-clone")
	WasPipelineTaskNameAnalyze = WasPipelineTaskName("analyze")

	WasPipelineTaskNameBuild  = WasPipelineTaskName("build")
	WasPipelineTaskNameDeploy = WasPipelineTaskName("deploy")
)

const (
	WasPipelineParamNameProjectId  = "project-id"
	WasPipelineParamNameGitUrl     = "git-url"
	WasPipelineParamNameGitRev     = "git-rev"
	WasPipelineParamNameSourceType = "source-type"
	WasPipelineParamNameTargetType = "target-type"

	WasPipelineParamNameAppName   = "app-name"
	WasPipelineParamNameDeployCfg = "deploy-cfg-name"
)

const (
	WasPipelineWorkspaceName = "git-report"
)
