package tier2

func ExampleResult_OK() {
	_ = (*Result).OK
}

func ExampleExitError_Error() {
	_ = (*ExitError).Error
}

func ExampleExitError_Unwrap() {
	_ = (*ExitError).Unwrap
}

func ExampleNewRunner() {
	_ = NewRunner
}

func ExampleResolvePython() {
	_ = ResolvePython
}

func ExampleRunner_RunSource() {
	_ = (*Runner).RunSource
}

func ExampleRunner_RunFile() {
	_ = (*Runner).RunFile
}

func ExampleLocalPythonPath() {
	_ = LocalPythonPath
}
