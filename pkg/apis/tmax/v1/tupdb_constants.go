package v1

import "github.com/operator-framework/operator-sdk/pkg/status"

const (
	DBConditionKeyDBAnalyzing = status.ConditionType("Analyzing")
	DBConditionKeyDBMigrating = status.ConditionType("Migrating")
	DBConditionKeyDBSucceed   = status.ConditionType("Succeeded")
)

const (
	TaskNameAnalyzeDB = "l2c-tup-db"
	TaskNameMigrateDB = "l2c-migration-db"
)

const (
	DBPipelineTaskNameAnalyzeDB = "analyze"
	DBPipelineTaskNameMigrateDB = "migrate"
)

// Params for migrate
const (
	DBPipelineParamNameSourceUserName = "source-username"
	DBPipelineParamNameSourcePassword = "source-password"
	DBPipelineParamNameSourceType     = "source-type"
	DBPipelineParamNameSourceSID      = "source-sid"
	DBPipelineParamNameSourceAs       = "source-as"
	DBPipelineParamNameSourcePort     = "source-port"
	DBPipelineParamNameSourceIP       = "source-ip"
	DBPipelineParamNameTargetIP       = "target-ip"
	DBPipelineParamNameTargetPort     = "target-port"
	DBPipelineParamNameTargetUserName = "target-username"
	DBPipelineParamNameTargetUser     = "target-user" //[TODO] Figure out what it is
	DBPipelineParamNameTargetSID      = "target-sid"
	DBPipelineParamNameTargetType     = "target-type"
	DBPipelineParamNameTargetPassword = "target-password"
	DBPipelineParamNameFull           = "full"
)

// Params for analyze
const (
	DBAnalyzePipelineParamTarget        = "analyze-target"
	DBAnalyzePipelineParamFileType      = "analyze-file-type"
	DBAnalyzePipelineParamFileSyntax    = "analyze-file-syntax"
	DBAnalyzePipelineParamFileExtension = "analyze-file-extension"
	DBAnalyzePipelineParamFileSearch    = "analyze-file-search"
	DBAnalyzePipelineParamFileLocation  = "analyze-file-location"
	DBAnalyzePipelineParamFileCharset   = "analyze-file-charset"
	DBAnalyzePipelineParamReportOptions = "analyze-report-options"
)
