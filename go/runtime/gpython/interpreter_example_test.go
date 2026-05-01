package gpython

func ExampleNew() {
	_ = New
}

func ExampleInterpreter_Run() {
	_ = (*Interpreter).Run
}

func ExampleInterpreter_RegisterModule() {
	_ = (*Interpreter).RegisterModule
}

func ExampleInterpreter_NewSession() {
	_ = (*Interpreter).NewSession
}

func ExampleSession_Run() {
	_ = (*Session).Run
}

func ExampleInterpreter_Close() {
	_ = (*Interpreter).Close
}

func ExampleInterpreter_Modules() {
	_ = (*Interpreter).Modules
}

func ExampleInterpreter_Call() {
	_ = (*Interpreter).Call
}
