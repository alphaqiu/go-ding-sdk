package sdk

type CommonResp struct {
	ErrCode   int    `json:"errcode,omitempty"`
	ErrMsg    string `json:"errmsg,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

type AccessTokenResp struct {
	CommonResp
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type DepartmentResp struct {
	CommonResp
	Result []*DepartmentNameCnf `json:"result"`
}

type DepartmentChildrenResp struct {
	CommonResp
	Result *DeptIDList `json:"result"`
}

type DeptIDList struct {
	DeptIDList []uint64 `json:"dept_id_list"`
}

type DepartmentNameCnf struct {
	AutoAddUser     bool   `json:"auto_add_user"`
	CreateDeptGroup bool   `json:"create_dept_group"`
	DeptID          uint64 `json:"dept_id"`
	Name            string `json:"name"`
	ParentID        uint64 `json:"parent_id"`
}

type SimpleUserResp struct {
	CommonResp
	Result *ListSimpleUserRes
}

type UserDetailResp struct {
	CommonResp
	Result *ListUserDetailRes
}

type ListSimpleUserRes struct {
	HasMore    bool          `json:"has_more"`
	NextCursor int           `json:"next_cursor"`
	List       []*SimpleUser `json:"list"`
}

type SimpleUser struct {
	UserID string   `json:"userid"`
	Name   string   `json:"name"`
	PIDS   []uint64 `json:"pids,omitempty"` //department id
}

type ListUserDetailRes struct {
	HasMore    bool            `json:"has_more"`
	NextCursor int             `json:"next_cursor"`
	List       []*DingDingUser `json:"list"`
}

type DingDingUser struct {
	UserID       string `json:"userid"`
	Name         string `json:"name"`
	UnionID      string `json:"unionid"`
	Avatar       string `json:"avatar"`
	Mobile       string `json:"mobile"`
	HideMobile   bool   `json:"hide_mobile"`
	Title        string `json:"title"`
	Email        string `json:"email"`
	OrgEmail     string `json:"org_email"`
	DepartIDList []int  `json:"dept_id_list"`
}

type DingDingDeptNode struct {
	Info     DingDingDeptInfo   `json:"info"`
	Children []DingDingDeptNode `json:"children"`
}

type DingDingDeptInfo struct {
	DeptID uint64 `json:"dept_id"`
	Name   string `json:"name"`
	PID    uint64 `json:"pid"`
}

type ApprovalProcessIDListResp struct {
	CommonResp
	Result *ApprovalProcessRes
}

type ApprovalProcessRes struct {
	List       []string `json:"list"`
	NextCursor int      `json:"next_cursor"`
}

type ApprovalDetailResp struct {
	CommonResp
	Detail *ApprovalDetail `json:"process_instance"`
}

type ProcessCodeResult struct {
	CommonResp
	Code string `json:"process_code"`
}

type ApprovalDetail struct {
	Title      string               `json:"title"`
	CreateTime string               `json:"create_time"`
	FinishTime string               `json:"finish_time"`
	Uid        string               `json:"originator_userid"`
	UserDeptID string               `json:"originator_dept_id"`
	Status     string               `json:"status"`
	BusinessID string               `json:"business_id"`
	Result     string               `json:"result"`
	Components []*ApprovalComponent `json:"form_component_values,omitempty"`
}

type ApprovalComponent struct {
	ID       string `json:"id"`
	Type     string `json:"component_type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	ExtValue string `json:"ext_value"`
}

type SendMsgByRobotResp struct {
	Code                      string   `json:"code,omitempty"`
	ReqID                     string   `json:"requestid,omitempty"`
	Message                   string   `json:"message,omitempty"`
	ProcessQueryKey           string   `json:"processQueryKey,omitempty"`           // 消息id
	InvalidStaffIdList        []string `json:"invalidStaffIdList,omitempty"`        // 无效的用户userid列表。
	FlowControlledStaffIdList []string `json:"flowControlledStaffIdList,omitempty"` // 被限流的userid列表。
}

type DepartmentNameCnfCollection []*DepartmentNameCnf

func (c DepartmentNameCnfCollection) ForEach(fn func(item *DepartmentNameCnf) error) error {
	for _, item := range c {
		if err := fn(item); err != nil {
			return err
		}
	}
	return nil
}

type SnsResponse struct {
	CommonResp
	UserInfo *SnsUserInfo `json:"user_info"`
}

type SnsUserInfo struct {
	Nick                 string `json:"nick"`
	UnionID              string `json:"unionid"`
	OpenID               string `json:"openid"`
	MainOrgAuthHighLevel bool   `json:"main_org_auth_high_level"`
}

type UserIDResponse struct {
	CommonResp
	Result *UserGetByUnionIdResponse `json:"result"`
}

type UserGetByUnionIdResponse struct {
	UserID      string `json:"userid"`
	ContactType int    `json:"contact_type"` // 联系类型: 0 企业内部员工，1 企业外部联系人
}
