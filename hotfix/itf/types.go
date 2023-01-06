package itf

type HotfixEntry interface {
	Apply()
}

type HotfixInterface interface {
	NewFuncEntry(interface{}, string, interface{}) HotfixEntry
	ApplyHotfix([]HotfixEntry)
	ResetHotfix()
}
