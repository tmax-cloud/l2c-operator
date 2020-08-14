package apis

const (
	DirName = "/wait"

	WaiterContainerName = "waiter"
)

type ScanResult string

const (
	ScanResultOk   = ScanResult("ok")
	ScanResultFail = ScanResult("fail")
)
