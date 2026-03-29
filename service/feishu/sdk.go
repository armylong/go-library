package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/armylong/go-library/service/conf"
	"github.com/armylong/go-library/service/httpx"
	"github.com/armylong/go-library/service/redis"
)

const (
	TenantAccessTokenUrl = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal/"
	UserAccessTokenUrl   = "https://open.feishu.cn/open-apis/authen/v2/oauth/token/"
)

// 全局单例
var (
	feishuSdk     *FeishuSdk
	initFeishuSdk sync.Once
)

type FeishuSdk struct {
	AppId     string
	AppSecret string
	Ttk       *FsTenantAccessToken
	Utk       *FsUserAccessToken
	mu        sync.Mutex
}

type FsTenantAccessToken struct {
	TenantAccessToken string    `json:"tenant_access_token"`
	ExpireTime        time.Time `json:"expire_time"`
}

type FSRefreshTtkRequest struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

type FSRefreshTtkResponse struct {
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
	Expire            int    `json:"expire"`
}

type FsUserAccessToken struct {
	UserAccessToken        string    `json:"user_access_token"`
	ExpiresSec             int       `json:"expires_sec"`
	ExpireTime             time.Time `json:"expire_time"`
	RefreshToken           string    `json:"refresh_token"`
	RefreshTokenExpiresSec int       `json:"refresh_token_expires_sec"`
	RefreshTokenExpireTime time.Time `json:"refresh_token_expire_time"`
	RefreshTokenCacheKey   string    `json:"refresh_token_cache_key"`
	Code                   string    `json:"code"`
}

type FSRefreshUtkRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	RedirectURI  string `json:"redirect_uri"`
}

type FSRefreshUtkResponse struct {
	Code                  int    `json:"code"`
	AccessToken           string `json:"access_token"`
	ExpiresIn             int    `json:"expires_in"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
	TokenType             int    `json:"token_type"`
	Scope                 string `json:"scope"`
	ErrorDescription      string `json:"error_description"`
}

func GetFeishuSDK() *FeishuSdk {
	_initFeishuSDK()
	return feishuSdk
}

func _initFeishuSDK() {
	initFeishuSdk.Do(func() {
		feishuConfig := conf.GetFsConfig()
		feishuSdk = &FeishuSdk{
			AppId:     feishuConfig.AppId,
			AppSecret: feishuConfig.AppSecret,
			Utk: &FsUserAccessToken{
				RefreshTokenCacheKey: feishuConfig.UserAccessTokenRefreshCacheKey,
			},
		}
	})
}

func GetTenantAccessTokenHeader() string {
	sdk := GetFeishuSDK()
	sdk.refreshTenantAccessToken()
	return `Bearer ` + sdk.Ttk.TenantAccessToken
}

func (t *FeishuSdk) refreshTenantAccessToken() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Ttk == nil {
		t.Ttk = &FsTenantAccessToken{}
	}

	if time.Now().Before(t.Ttk.ExpireTime.Add(-5 * time.Minute)) {
		return
	}

	req := &FSRefreshTtkRequest{
		AppID:     t.AppId,
		AppSecret: t.AppSecret,
	}
	b, _ := json.Marshal(req)
	res, err := httpx.Post(TenantAccessTokenUrl, b)
	if err != nil {
		return
	}

	data := &FSRefreshTtkResponse{}
	err = json.Unmarshal(res, &data)
	if err != nil {
		return
	}

	if data.Code == 0 {
		t.Ttk.TenantAccessToken = data.TenantAccessToken
		t.Ttk.ExpireTime = time.Now().Add(time.Duration(data.Expire) * time.Second)
	}

}

func GetUserAccessTokenHeader(code, redirectUri string) string {
	sdk := GetFeishuSDK()
	sdk.Utk.Code = code
	sdk.initUserAccessToken(code, redirectUri)
	if sdk.Utk == nil || sdk.Utk.UserAccessToken == "" {
		fmt.Println("UserAccessToken is empty")
		return ""
	}
	return `Bearer ` + sdk.Utk.UserAccessToken
}

func (t *FeishuSdk) initUserAccessToken(code, redirectUri string) {

	if code == "" {
		t.refreshUserAccessToken()
		return
	}

	// 若有code, 证明需要用户手动授权了, 授权完成后需要拿着code手动调用
	req := &FSRefreshUtkRequest{
		GrantType:    "authorization_code",
		ClientID:     t.AppId,
		ClientSecret: t.AppSecret,
		Code:         code,
		RedirectURI:  redirectUri,
	}
	b, _ := json.Marshal(req)
	res, err := httpx.Post(UserAccessTokenUrl, b)
	fmt.Println(string(res))
	if err != nil {
		fmt.Println("Post authorization_code error", err.Error())
		return
	}

	data := &FSRefreshUtkResponse{}
	err = json.Unmarshal(res, &data)
	if err != nil {
		fmt.Println("Unmarshal authorization_code error", err.Error())
		return
	}

	if data.Code == 0 {
		t.Utk = &FsUserAccessToken{
			UserAccessToken:        data.AccessToken,
			ExpiresSec:             data.ExpiresIn,
			ExpireTime:             time.Now().Add(time.Duration(data.ExpiresIn) * time.Second),
			RefreshToken:           data.RefreshToken,
			RefreshTokenExpiresSec: data.RefreshTokenExpiresIn,
			RefreshTokenExpireTime: time.Now().Add(time.Duration(data.RefreshTokenExpiresIn) * time.Second),
		}

		t.updateUserAccessRefreshTokenCache()
	}
}

func (t *FeishuSdk) refreshUserAccessToken() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Utk == nil {
		t.Utk = &FsUserAccessToken{}
	}

	if time.Now().Before(t.Utk.ExpireTime.Add(-5 * time.Minute)) {
		return
	}

	if t.Utk.RefreshToken == "" || time.Now().Before(t.Utk.RefreshTokenExpireTime.Add(-5*time.Minute)) {
		t.Utk.RefreshToken = t.GetUserAccessRefreshTokenCache()
		if t.Utk.RefreshToken == "" {
			fmt.Println("RefreshToken is empty")
			return
		}
	}

	req := &FSRefreshUtkRequest{
		GrantType:    "refresh_token",
		ClientID:     t.AppId,
		ClientSecret: t.AppSecret,
		RefreshToken: t.Utk.RefreshToken,
	}
	b, _ := json.Marshal(req)
	res, err := httpx.Post(UserAccessTokenUrl, b)
	if err != nil {
		fmt.Println("Post refresh_token error", err.Error())
		return
	}

	data := &FSRefreshUtkResponse{}
	err = json.Unmarshal(res, &data)
	if err != nil {
		fmt.Println("Unmarshal refresh_token error", err.Error())
		return
	}

	if data.Code == 0 {
		t.Utk = &FsUserAccessToken{
			UserAccessToken:        data.AccessToken,
			ExpiresSec:             data.ExpiresIn,
			ExpireTime:             time.Now().Add(time.Duration(data.ExpiresIn) * time.Second),
			RefreshToken:           data.RefreshToken,
			RefreshTokenExpiresSec: data.RefreshTokenExpiresIn,
			RefreshTokenExpireTime: time.Now().Add(time.Duration(data.RefreshTokenExpiresIn) * time.Second),
		}
		t.updateUserAccessRefreshTokenCache()
	} else {
		fmt.Println(data)
	}
}

func (t *FeishuSdk) updateUserAccessRefreshTokenCache() {
	redisClient := redis.GetClient(`default`)
	redisClient.Set(context.Background(), t.Utk.RefreshTokenCacheKey, t.Utk.RefreshToken, time.Duration(t.Utk.ExpiresSec)*time.Second)
}

func (t *FeishuSdk) GetUserAccessRefreshTokenCache() string {
	redisClient := redis.GetClient(`default`)
	refreshToken, err := redisClient.Get(context.Background(), t.Utk.RefreshTokenCacheKey).Result()
	if err != nil {
		return ""
	}
	return refreshToken
}
