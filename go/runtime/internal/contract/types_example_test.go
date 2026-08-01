package contract

func ExampleUnsupportedImportError_Error() {
	_ = (*UnsupportedImportError).Error
}
