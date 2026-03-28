package feishu

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/armylong/go-library/service/httpx"
)

const TokenUrl = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal/"

const (
	//告警机器人
	AppIdMz     = `cli_a0b61445cf78d00c`
	AppSecretMz = `fuJ4u8qVYubqh1ltdbxJOgGluDX1v1Ah`
)

// 全局单例
var (
	feishuSdk     *FeishuSdk
	initFeishuSdk sync.Once
)

type FeishuSdk struct {
	AppId     string
	AppSecret string
	Tk        *FSToken
	mu        sync.Mutex
}

type FSRefreshTkRequest struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

type FSRefreshTkResponse struct {
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
	Expire            int    `json:"expire"`
}

type FSToken struct {
	TenantAccessToken string    `json:"tenant_access_token"`
	Expire            time.Time `json:"expire"`
}

func GetFeishuSdk() *FeishuSdk {
	_initFeishuSdk()
	return feishuSdk
}

func _initFeishuSdk() {
	initFeishuSdk.Do(func() {
		feishuSdk = &FeishuSdk{
			AppId:     AppIdMz,
			AppSecret: AppSecretMz,
		}
	})
}

func (t *FeishuSdk) GetTenantAccessToken() string {
	t.refreshToken()
	return t.Tk.TenantAccessToken
}

func (t *FeishuSdk) GetAuthorizationHeader() string {
	return `Bearer ` + t.GetTenantAccessToken()
}

func (t *FeishuSdk) refreshToken() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Tk == nil {
		t.Tk = &FSToken{}
	}

	if time.Now().Before(t.Tk.Expire.Add(-5 * time.Minute)) {
		return
	}

	req := &FSRefreshTkRequest{
		AppID:     t.AppId,
		AppSecret: t.AppSecret,
	}
	b, _ := json.Marshal(req)
	res, err := httpx.Post(TokenUrl, b)
	if err != nil {
		return
	}

	data := &FSRefreshTkResponse{}
	err = json.Unmarshal(res, &data)
	if err != nil {
		return
	}

	if data.Code == 0 {
		t.Tk.TenantAccessToken = data.TenantAccessToken
		t.Tk.Expire = time.Now().Add(time.Duration(data.Expire) * time.Second)
	}

}
