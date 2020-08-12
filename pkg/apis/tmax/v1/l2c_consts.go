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
