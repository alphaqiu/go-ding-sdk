package sdk

type CommonDepartmentReq struct {
	DeptID uint64 `json:"dept_id"`
}

type DepartmentReq struct {
	CommonDepartmentReq
	Language Lang `json:"language,omitempty"`
}

type DepartmentChildrenReq struct {
	CommonDepartmentReq
}

type SimpleUserReq struct {
	CommonDepartmentReq
	Cursor             int        `json:"cursor"`
	Size               int        `json:"size"`
	OrderField         OrderField `json:"order_field"`
	ContainAccessLimit bool       `json:"contain_access_limit"`
	Language           Lang       `json:"language"`
}

type ApprovalProcessIDReq struct {
	ProcessCode string `json:"process_code"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
	Size        int    `json:"size"`
	Cursor      int    `json:"cursor"`
	UserIDList  string `json:"userid_list,omitempty"`
}

type ApprovalDetailReq struct {
	ProcessInstanceID string `json:"process_instance_id"`
}

type ProcessCodeReq struct {
	Name string `json:"name"`
}

// SendMsgByRobotReq 批量发送单聊消息的参数
type SendMsgByRobotReq struct {
	RobotCode string   `json:"robotCode"`
	UserIDs   []string `json:"userIds"`
	MsgKey    string   `json:"msgKey"`
	MsgParam  string   `json:"msgParam"`
}

type MsgContent struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type SnsRequest struct {
	TmpAuthCode string `json:"tmp_auth_code"`
}

type UserIDReq struct {
	UnionID string `json:"unionid"`
}
