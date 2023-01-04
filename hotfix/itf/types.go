package itf

type HotfixEntry interface {
	Apply()
}

type HotfixInterface interface {
	NewFuncEntry(interface{}, interface{}) HotfixEntry
	NewMethodEntry(interface{}, string, interface{}) HotfixEntry
	ApplyHotfix([]HotfixEntry)
}
