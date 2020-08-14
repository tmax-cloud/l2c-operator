package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ConditionKeyProjectReady   = metav1.RowConditionType("ProjectReady")
	ConditionKeyProjectRunning = metav1.RowConditionType("ProjectRunning")

	ConditionKeyAnalyze   = metav1.RowConditionType("Analyze")
	ConditionKeyDbMigrate = metav1.RowConditionType("DBMigrate")
	ConditionKeyBuild     = metav1.RowConditionType("Build")
	ConditionKeyDeploy    = metav1.RowConditionType("Deploy")
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
