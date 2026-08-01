package runtime

func ExampleBackendNotBuiltError_Error() {
	_ = (*BackendNotBuiltError).Error
}

func ExampleNew() {
	_ = New
}

func ExampleSplitKeywordArguments() {
	_ = SplitKeywordArguments
}

func ExampleIsTier2FallbackCandidate() {
	_ = IsTier2FallbackCandidate
}
