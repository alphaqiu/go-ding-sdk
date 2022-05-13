package sdk

type Lang string
type OrderField string

var (
	ChineseLanguage Lang       = "zh_CN"
	EnglishLanguage Lang       = "en_US"
	EntryAsc        OrderField = "entry_asc"   // 代表按照进入部门的时间升序。
	EntryDesc       OrderField = "entry_desc"  // 代表按照进入部门的时间降序。
	ModifyAsc       OrderField = "modify_asc"  // 代表按照部门信息修改时间升序。
	ModifyDesc      OrderField = "modify_desc" // 代表按照部门信息修改时间降序。
	Custom          OrderField = "custom"      // 代表用户定义(未定义时按照拼音)排序。
)
