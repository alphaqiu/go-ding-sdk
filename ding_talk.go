package sdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	logging "github.com/ipfs/go-log/v2"
)

const (
	domain             = "https://oapi.dingtalk.com"
	reqAccessToken     = "/gettoken?appkey=%s&appsecret=%s"                               // 获取钉钉企业内部服务的access token
	reqDept            = "/topapi/v2/department/listsub?access_token=%s"                  // 获取组织架构部门
	reqChildrenDept    = "/topapi/v2/department/listsubid?access_token=%s"                // 获取子部门
	reqUser            = "/topapi/user/listsimple?access_token=%s"                        // 获取部门下的用户(simple user)
	reqUserDetail      = "/topapi/v2/user/list?access_token=%s"                           // 获取部门下用户的详细信息
	reqApprovalProcess = "/topapi/processinstance/listids?access_token=%s"                // 获取指定审批流程清单
	reqApprovalDetail  = "/topapi/processinstance/get?access_token=%s"                    // 获取审批流程详细信息
	sendWorkNotify     = "/topapi/message/corpconversation/asyncsend_v2?access_token=%s"  // 发送工作通知
	batchSendAPI       = "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend"      // 发送批量消息
	reqProcessCode     = "/topapi/process/get_by_name?access_token=%s"                    // 获取模板code
	snsReq             = "/sns/getuserinfo_bycode?accessKey=%s&timestamp=%s&signature=%s" // 根据sns临时授权码获取用户信息
	reqUserByUnionID   = "/topapi/user/getbyunionid?access_token=%s"                      // 根据UnionID获取用户信息
)

func NewDingTalkClient(agentId, appKey, appSecret string) *DingTalkClient {
	return &DingTalkClient{
		log:       logging.Logger("dingtalk"),
		agentId:   agentId,
		appKey:    appKey,
		appSecret: appSecret,
		mutex:     new(sync.Mutex),
	}
}

type DingTalkClient struct {
	log         *logging.ZapEventLogger
	agentId     string
	appKey      string
	appSecret   string
	accessToken string
	expireTime  time.Time // 获取到access_token后计算得到的过期时间
	mutex       *sync.Mutex
}

// GetAccessToken 在使用access_token时，请注意：
//access_token的有效期为7200秒（2小时），有效期内重复获取会返回相同结果并自动续期，过期后获取会返回新的access_token。
//开发者需要缓存access_token，用于后续接口的调用。因为每个应用的access_token是彼此独立的，所以进行缓存时需要区分应用来进行存储。
//不能频繁调用gettoken接口，否则会受到频率拦截。
func (d *DingTalkClient) GetAccessToken() (string, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.accessToken != "" && time.Now().Before(d.expireTime) {
		return d.accessToken, nil
	}

	resp, err := http.Get(fmt.Sprintf(domain+reqAccessToken, d.appKey, d.appSecret))
	if err != nil {
		return "", fmt.Errorf("请求access_token失败： %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("请求access_token失败: %s(%d)", resp.Status, resp.StatusCode)
	}

	body := resp.Body
	// Output: {"errcode":0,"access_token":"7122c6639d12378195cae4237d5fd61e","errmsg":"ok","expires_in":7200}
	defer func() { _ = body.Close() }()
	var atr AccessTokenResp
	if err = readResult(body, &atr); err != nil {
		return "", fmt.Errorf("读取access_token失败: %v", err)
	}

	if atr.ErrCode != 0 {
		d.accessToken = ""
		d.expireTime = time.Now()
		return "", fmt.Errorf("请求access_token失败: %s(%d)，请检查访问API权限", atr.ErrMsg, atr.ErrCode)
	}

	d.accessToken = atr.AccessToken
	d.expireTime = time.Now().Add(time.Duration(atr.ExpiresIn) * time.Second)

	return atr.AccessToken, nil
}

// GetDepartments 获取部门列表
// 本接口只支持获取当前部门的下一级部门基础信息
func (d *DingTalkClient) GetDepartments(deptID uint64, language Lang) (DepartmentNameCnfCollection, error) {
	accToken, err := d.GetAccessToken()
	if err != nil {
		return nil, err
	}

	var lang = ChineseLanguage
	if language == EnglishLanguage {
		lang = language
	}

	reqUrl := fmt.Sprintf(domain+reqDept, accToken)
	var data DepartmentResp
	err = post(reqUrl, &DepartmentReq{
		CommonDepartmentReq: CommonDepartmentReq{DeptID: deptID},
		Language:            lang,
	}, &data, nil)
	if err != nil {
		return nil, fmt.Errorf("请求部门(%d)清单失败: %v", deptID, err)
	}

	// Output: {"errcode":0,"errmsg":"ok","result":[{"auto_add_user":true,"create_dept_group":true,"dept_id":574367388,"name":"总经办","parent_id":1},{"auto_add_user":true,"create_dept_group":true,"dept_id":574545316,"name":"共","parent_id":1},{"auto_add_user":true,"create_dept_group":true,"dept_id":574575215,"name":"商务部","parent_id":1}],"request_id":"4uqsv89h1x82"}

	if data.ErrCode != 0 {
		return nil, fmt.Errorf("请求部门清单失败: %s(%d)", data.ErrMsg, data.ErrCode)
	}
	return data.Result, nil
}

func (d *DingTalkClient) GetChildrenDepartments(deptID uint64) ([]uint64, error) {
	accToken, err := d.GetAccessToken()
	if err != nil {
		return nil, err
	}

	reqUrl := fmt.Sprintf(domain+reqChildrenDept, accToken)
	var data DepartmentChildrenResp
	err = post(reqUrl, &DepartmentChildrenReq{CommonDepartmentReq{DeptID: deptID}}, &data, nil)
	if err != nil {
		return nil, fmt.Errorf("请求子部门(%d)清单失败: %v", deptID, err)
	}

	if data.ErrCode != 0 {
		return nil, fmt.Errorf("请求子部门清单失败: %s(%d)", data.ErrMsg, data.ErrCode)
	}

	if data.Result == nil {
		return nil, nil
	}

	return data.Result.DeptIDList, nil
}

func (d *DingTalkClient) GetSimpleUsers(reqParams SimpleUserReq) (*ListSimpleUserRes, error) {
	accToken, err := d.GetAccessToken()
	if err != nil {
		return nil, err
	}

	reqUrl := fmt.Sprintf(domain+reqUser, accToken)
	var data SimpleUserResp
	err = post(reqUrl, &reqParams, &data, nil)
	if err != nil {
		return nil, fmt.Errorf("请求部门下(%d)的员工基本信息失败: %v", reqParams.DeptID, err)
	}

	if data.ErrCode != 0 {
		return nil, fmt.Errorf("请求部门员工基本信息失败; %s(%d)", data.ErrMsg, data.ErrCode)
	}

	return data.Result, nil
}

func (d *DingTalkClient) GetUsers(reqParams SimpleUserReq) (*ListUserDetailRes, error) {
	accToken, err := d.GetAccessToken()
	if err != nil {
		return nil, err
	}

	reqUrl := fmt.Sprintf(domain+reqUserDetail, accToken)
	var data UserDetailResp
	err = post(reqUrl, &reqParams, &data, nil)
	if err != nil {
		return nil, fmt.Errorf("请求部门（%d）下的员工详细信息失败: %v", reqParams.DeptID, err)
	}

	if data.ErrCode != 0 {
		return nil, fmt.Errorf("请求部门员工详细信息失败; %s(%d)", data.ErrMsg, data.ErrCode)
	}

	return data.Result, nil
}

func (d *DingTalkClient) GetDepartmentsByParent(ids ...uint64) ([]uint64, error) {
	var data []uint64
	for _, deptId := range ids {
		children, err := d.GetChildrenDepartments(deptId)
		if err != nil {
			return nil, fmt.Errorf("%v, %v", ids, err)
		}

		if len(children) > 0 {
			cc, err := d.GetDepartmentsByParent(children...)
			if err != nil {
				return nil, fmt.Errorf("%v, %v", children, err)
			}

			data = append(data, cc...)
		}
		data = append(data, children...)
	}
	return data, nil
}

func (d *DingTalkClient) GetDepartmentNamesByParent(ids ...uint64) ([]uint64, error) {
	var data []uint64
	for _, deptId := range ids {
		children, err := d.GetChildrenDepartments(deptId)
		if err != nil {
			return nil, fmt.Errorf("%v, %v", ids, err)
		}

		if len(children) > 0 {
			cc, err := d.GetDepartmentsByParent(children...)
			if err != nil {
				return nil, fmt.Errorf("%v, %v", children, err)
			}

			data = append(data, cc...)
		}
		data = append(data, children...)
	}
	return data, nil
}

func (d *DingTalkClient) GetSimpleUserByDeptIDList(depts []uint64) ([]*SimpleUser, error) {
	users := make(map[string]*SimpleUser)
	for _, dept := range depts {
		cursor := 0
		for {
			listRes, err := d.GetSimpleUsers(SimpleUserReq{
				CommonDepartmentReq: CommonDepartmentReq{DeptID: dept},
				Cursor:              cursor,
				Size:                100,
				OrderField:          EntryAsc,
				ContainAccessLimit:  false,
				Language:            ChineseLanguage,
			})

			if err != nil {
				return nil, err
			}

			cursor = listRes.NextCursor
			for _, u := range listRes.List {
				users[u.UserID] = u
			}

			if !listRes.HasMore {
				break
			}
		}
	}

	data := make([]*SimpleUser, 0, len(users))
	for _, item := range users {
		data = append(data, item)
	}
	return data, nil
}

func (d *DingTalkClient) GetUsersByDeptIDList(depts []uint64) ([]*DingDingUser, error) {
	users := make(map[string]*DingDingUser)
	for _, dept := range depts {
		cursor := 0
		for {
			listRes, err := d.GetUsers(SimpleUserReq{
				CommonDepartmentReq: CommonDepartmentReq{DeptID: dept},
				Cursor:              cursor,
				Size:                100,
				OrderField:          EntryAsc,
				ContainAccessLimit:  false,
				Language:            ChineseLanguage,
			})

			if err != nil {
				return nil, err
			}

			cursor = listRes.NextCursor
			for _, u := range listRes.List {
				users[u.UserID] = u
			}

			if !listRes.HasMore {
				break
			}
		}
	}

	data := make([]*DingDingUser, 0, len(users))
	for _, item := range users {
		data = append(data, item)
	}
	return data, nil
}

func (d *DingTalkClient) GetApprovalProcessIDList(params ApprovalProcessIDReq) (*ApprovalProcessRes, error) {
	accToken, err := d.GetAccessToken()
	if err != nil {
		return nil, err
	}

	reqUrl := fmt.Sprintf(domain+reqApprovalProcess, accToken)
	var data ApprovalProcessIDListResp
	err = post(reqUrl, &params, &data, nil)
	if err != nil {
		return nil, fmt.Errorf("请求审批流程(%s)失败: %v", params.ProcessCode, err)
	}

	//fmt.Println(data)
	if data.ErrCode != 0 {
		return nil, fmt.Errorf("请求审批流程失败; %s(%d)", data.ErrMsg, data.ErrCode)
	}

	return data.Result, nil
}

func (d *DingTalkClient) GetApprovalDetail(processID string) (*ApprovalDetail, error) {
	accToken, err := d.GetAccessToken()
	if err != nil {
		return nil, err
	}

	reqUrl := fmt.Sprintf(domain+reqApprovalDetail, accToken)
	var data ApprovalDetailResp
	err = post(reqUrl, &ApprovalDetailReq{ProcessInstanceID: processID}, &data, nil)
	if err != nil {
		return nil, fmt.Errorf("请求审批详情(%s)失败: %v", processID, err)
	}

	if data.ErrCode != 0 {
		return nil, fmt.Errorf("请求审批详情失败: %s(%d)", data.ErrMsg, data.ErrCode)
	}

	return data.Detail, nil
}

func (d *DingTalkClient) SendMessageFromRobot(robotCode, title, content string, to []string) (*SendMsgByRobotResp, error) {
	accToken, err := d.GetAccessToken()
	if err != nil {
		return nil, err
	}

	msg, err := json.Marshal(&MsgContent{Title: title, Text: content})
	if err != nil {
		return nil, fmt.Errorf("生成消息失败: %v", err)
	}

	if len(to) == 0 {
		return nil, nil
	}

	if len(to) > 20 {
		to = to[:20]
	}

	backOff := NewBackoff()
	reqObj := &SendMsgByRobotReq{
		RobotCode: robotCode,
		UserIDs:   to,
		MsgKey:    "officialMarkdownMsg",
		MsgParam:  string(msg),
	}
	header := http.Header{"x-acs-dingtalk-access-token": []string{accToken}}

	var ret SendMsgByRobotResp
	retries := 0
	for {
		if retries > 3 {
			break
		}

		err = post(batchSendAPI, reqObj, &ret, header)
		if err != nil {
			d.log.Errorf("发送消息失败, 重试发送: %v", err)
			time.Sleep(backOff.Duration(retries + 1))
			retries += 1
			continue
		}

		break
	}

	if err != nil {
		return nil, fmt.Errorf("发送批量消息接口失败(Retries: %d): %v", retries, err)
	}

	return &ret, nil
}

func (d *DingTalkClient) GetProcessCode() error {
	accToken, err := d.GetAccessToken()
	if err != nil {
		return err
	}
	reqUrl := fmt.Sprintf(domain+reqProcessCode, accToken)

	var data ProcessCodeResult
	err = post(reqUrl, &ProcessCodeReq{Name: "每日工作结果日志[V]"}, &data, nil)
	if err != nil {
		return fmt.Errorf("请求模版Code失败: %s(%d)", data.ErrMsg, data.ErrCode)
	}

	fmt.Println(data.Code)
	return nil
}

func (d *DingTalkClient) SendWorkNotify() {
	// TODO
}

func (d *DingTalkClient) GetUserIDFromScanQrCode(tmpCode string) (string, error) {
	snsUserInfo, err := d.GetUserUnionIDByCode(tmpCode)
	if err != nil {
		return "", err
	}

	if snsUserInfo == nil {
		return "", fmt.Errorf("无效的UnionID")
	}

	userId, err := d.GetUserIDByUnionID(snsUserInfo.UnionID)
	if err != nil {
		return "", err
	}

	return userId, nil
}

func (d *DingTalkClient) GetUserUnionIDByCode(tmpCode string) (*SnsUserInfo, error) {

	// 根据钉钉OpenAPI设定，通过钉钉扫码登陆过后拿到的临时登陆码换取用户信息步骤如下：
	// 参考：https://open.dingtalk.com/document/orgapp-server/obtain-the-user-information-based-on-the-sns-temporary-authorization
	// 1. 准备三个参数：accessKey (为应用的AppKey，在开发者后台应用详情页查看。)
	// 2. timestamp （当前时间戳，单位毫秒。）
	// 3. 对timestamp做签名后的结果（该结果为HashMacSha256->Base64编码->urlencode编码）
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	hashFn := hmac.New(sha256.New, []byte(d.appSecret))
	hashFn.Write([]byte(timestamp))
	digest := hashFn.Sum(nil)
	sig := url.QueryEscape(base64.StdEncoding.EncodeToString(digest))

	reqUrl := fmt.Sprintf(domain+snsReq, d.appKey, timestamp, sig)
	fmt.Println(reqUrl)
	var data SnsResponse
	err := post(reqUrl, &SnsRequest{TmpAuthCode: tmpCode}, &data, nil)
	if err != nil {
		return nil, fmt.Errorf("根据sns临时授权码获取用户信息失败: %v", err)
	}

	if data.ErrCode > 0 {
		fmt.Println(data)
		return nil, fmt.Errorf("%s(%d)", data.ErrMsg, data.ErrCode)
	}

	fmt.Println(data.UserInfo)
	return data.UserInfo, nil
}

// GetUserIDByUnionID 根据unionid获取用户userid
func (d *DingTalkClient) GetUserIDByUnionID(unionID string) (userId string, err error) {
	accToken, err := d.GetAccessToken()
	if err != nil {
		return "", err
	}

	reqUrl := fmt.Sprintf(domain+reqUserByUnionID, accToken)
	var data UserIDResponse
	if err = post(reqUrl, &UserIDReq{UnionID: unionID}, &data, nil); err != nil {
		return "", err
	}

	if data.ErrCode > 0 {
		fmt.Println(data)
		return "", fmt.Errorf("%s(%d)", data.ErrMsg, data.ErrCode)
	}

	return data.Result.UserID, nil
}

func post(reqUrl string, data interface{}, out interface{}, header http.Header) error {
	param, _ := json.Marshal(data)
	//fmt.Println(string(param))
	reqParams := strings.NewReader(string(param))

	req, err := http.NewRequest(http.MethodPost, reqUrl, reqParams)
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	for key, val := range header {
		for _, item := range val {
			req.Header.Add(key, item)
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}

	body := resp.Body
	defer func() { _ = body.Close() }()
	if err = readResult(body, out); err != nil {
		return err
	}

	return nil
}

func readResult(body io.Reader, out interface{}) error {
	payload, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("读取失败: %v", err)
	}

	//fmt.Println()
	//fmt.Printf("%s\n", payload)
	if out != nil {
		if err = json.Unmarshal(payload, out); err != nil {
			return fmt.Errorf("解析失败: %v", err)
		}
	}
	return nil
}
