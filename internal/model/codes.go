package model

import "github.com/ZanzyTHEbar/faults-go"

var (
	CodeStoreUnread  = faults.Code("ocrecent.store.unreadable")
	CodeNoDB         = faults.Code("ocrecent.store.nodb")
	CodeNotFound     = faults.Code("ocrecent.notfound")
	CodePickerCancel = faults.Code("ocrecent.picker.cancel")
	CodeUsage        = faults.Code("ocrecent.usage")
)
