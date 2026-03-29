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
	TokenType             string `json:"token_type"`
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
		fmt.Println("未获取到user_access_token")
		return ""
	}
	return `Bearer ` + sdk.Utk.UserAccessToken
}

func (t *FeishuSdk) initUserAccessToken(code, redirectUri string) {

	if code == "" {
		t.refreshUserAccessToken()
		return
	}

	fmt.Println(`初始化 user_access_token`)
	// code获取教程(飞书官方文档): https://open.feishu.cn/document/authentication-management/access-token/obtain-oauth-code#15391e22
	// (下面是防止踩坑步骤)
	// 1. 先在开发者后台配置你的应用回调地址(随意一个地址都行, 不用非得是自己的服务地址)
	// 2. 编写授权链接, 例如: https://accounts.feishu.cn/open-apis/authen/v1/authorize?client_id=cli_a94dc0fc84f6dbdd&redirect_uri=https://olfrzjwptnle.ap-northeast-1.clawcloudrun.com&scope=offline_access bitable:app
	//    a. url 是飞书官方提供,写死https://accounts.feishu.cn/open-apis/authen/v1/authorize
	//    b. client_id: 你的app_id
	//    c. redirect_uri: 你前面填写的回调地址
	//    d. scope: 你的应用权限范围, 空格分隔(需要在应用-权限处先配置好, offline_access是离线获取数据权限可以写死, bitable:app是获取多维表格权限)
	// 3. 将授权链接url编码(url快捷编码方式: 可以粘贴到浏览器地址栏上再复制回来)
	// 4. 将url编码后的链接发送到飞书任一聊天内(可以发给自己)
	// 5. 在飞书里点击这个链接, 就会跳转授权页面, 点击同意授权
	// 6. 授权完成后, 会跳转回你的回调地址, 并在回调地址后添加code参数, 将code值粘贴出来就可以用了

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

	if err != nil {
		fmt.Println("Post authorization_code error", err.Error())
		return
	}

	data := &FSRefreshUtkResponse{}
	err = json.Unmarshal(res, &data)
	if err != nil {
		fmt.Println(string(res))
		fmt.Println("Unmarshal authorization_code error", err.Error())
		return
	}

	if data.Code == 0 {
		t.Utk.UserAccessToken = data.AccessToken
		t.Utk.ExpiresSec = data.ExpiresIn
		t.Utk.ExpireTime = time.Now().Add(time.Duration(data.ExpiresIn) * time.Second)
		t.Utk.RefreshToken = data.RefreshToken
		t.Utk.RefreshTokenExpiresSec = data.RefreshTokenExpiresIn
		t.Utk.RefreshTokenExpireTime = time.Now().Add(time.Duration(data.RefreshTokenExpiresIn) * time.Second)

		t.updateUserAccessRefreshTokenCache()
	} else {
		fmt.Println(string(res))
	}
}

func (t *FeishuSdk) refreshUserAccessToken() {

	fmt.Println(`刷新 user_access_token`)

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Utk == nil {
		t.Utk = &FsUserAccessToken{}
	}

	// user_access_token没过期的话, 则直接返回
	if time.Now().Before(t.Utk.ExpireTime.Add(-5 * time.Minute)) {
		fmt.Println("UserAccessToken is not expired")
		return
	}

	// refresh_token过期了, 则直接用此refresh_token刷新user_access_token, 否则要从redis中获取refresh_token再刷新
	if time.Now().After(t.Utk.RefreshTokenExpireTime.Add(-5 * time.Minute)) {
		t.Utk.RefreshToken = t.GetUserAccessRefreshTokenCache()
		if t.Utk.RefreshToken == "" {
			fmt.Println("redis里的refresh_token值为空, 需要重新初始化")
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
		t.Utk.UserAccessToken = data.AccessToken
		t.Utk.ExpiresSec = data.ExpiresIn
		t.Utk.ExpireTime = time.Now().Add(time.Duration(data.ExpiresIn) * time.Second)
		t.Utk.RefreshToken = data.RefreshToken
		t.Utk.RefreshTokenExpiresSec = data.RefreshTokenExpiresIn
		t.Utk.RefreshTokenExpireTime = time.Now().Add(time.Duration(data.RefreshTokenExpiresIn) * time.Second)

		t.updateUserAccessRefreshTokenCache()
	} else {
		fmt.Println(string(res))
	}
}

func (t *FeishuSdk) updateUserAccessRefreshTokenCache() {
	if t.Utk == nil {
		fmt.Println("updateUserAccessRefreshTokenCache Utk is nil")
		return
	}
	redisClient := redis.GetClient(`default`)
	_, err := redisClient.Set(context.Background(), t.Utk.RefreshTokenCacheKey, t.Utk.RefreshToken, time.Duration(t.Utk.ExpiresSec)*time.Second).Result()
	if err != nil {
		fmt.Println("updateUserAccessRefreshTokenCache error", err.Error())
		return
	}
}

func (t *FeishuSdk) GetUserAccessRefreshTokenCache() string {
	ctx := context.Background()
	redisClient := redis.GetClient(`default`)
	refreshToken, err := redisClient.Get(ctx, t.Utk.RefreshTokenCacheKey).Result()
	if err != nil && err.Error() != "redis: nil" {
		fmt.Println("GetUserAccessRefreshTokenCache error", err.Error())
		return ""
	}
	redisClient.Del(ctx, t.Utk.RefreshTokenCacheKey)
	return refreshToken
}
