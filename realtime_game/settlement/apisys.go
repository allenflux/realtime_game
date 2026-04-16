package settlement

import (
	"bytes"
	"context"
	"crash/model/gmmodel"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultGameKey        = "CRASH_V2"
	errOrderAlreadySettle = 15029
)

// Adapter 定义资金接口。
type Adapter interface {
	Deduct(ctx context.Context, req DeductRequest) (*DeductResponse, error)
	Refund(ctx context.Context, req RefundRequest) error
	BillRolling(ctx context.Context, req BillRequest) error
	BillPreMatch(ctx context.Context, req BillRequest, partial bool, partialCount int8) error
	BatchBill(ctx context.Context, reqs []BillRequest) ([]string, []string, error)
	ResolveGameKey(ctx context.Context, channelID int64) string
	GetUserInfoByToken(ctx context.Context, token string) (*ApiSysGetUserData, error)
}

type apiSysAdapter struct {
	host  string
	token string
	lang  string
	gm    gmmodel.GameChannelMappingModel
	http  *http.Client
}

func NewApiSysAdapter(host, token, lang string, gm gmmodel.GameChannelMappingModel) Adapter {
	return &apiSysAdapter{
		host:  host,
		token: token,
		lang:  lang,
		gm:    gm,
		http: &http.Client{
			Timeout: 20 * time.Second,
			Transport: &http.Transport{
				Proxy: nil,
			}},
	}
}

type DeductRequest struct {
	ChannelID      int64
	OrderNo        string
	UserID         int64
	Currency       string
	Amount         string
	Metadata       string
	IsSystemReward bool
}

type DeductResponse struct {
	Status        int64  `json:"status"`
	TransactionNo string `json:"transaction_no"`
}

type RefundRequest struct {
	ChannelID int64
	OrderNo   string
	Metadata  string
}

type BillRequest struct {
	ChannelID      int64
	UserID         int64
	OrderNo        string
	Currency       string
	Amount         string
	Metadata       string
	IsSystemReward bool
}

type apiResp struct {
	Code        int             `json:"code"`
	Message     string          `json:"message"`
	MessageCode int             `json:"message_code"`
	Data        json.RawMessage `json:"data"`
}

type apiWagerResp struct {
	Code int `json:"code"`
	Data struct {
		Wagers []struct {
			WagerNo string `json:"wager_no"`
			Code    int    `json:"code"`
		} `json:"wagers"`
	} `json:"data"`
}

type ApiSysGetUserResponse struct {
	Code    int               `json:"code"`
	Data    ApiSysGetUserData `json:"data"`
	Message string            `json:"message"`
}

type ApiSysGetUserData struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

func parseApiSysGetUserResponse(resp string) (*ApiSysGetUserData, error) {
	var result ApiSysGetUserResponse
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		logx.Errorf("parse apisys get user response failed, resp:%s err:%v", resp, err)
		return nil, err
	}

	if result.Code != 200 {
		if result.Message != "" {
			return nil, errors.New(result.Message)
		}
		return nil, fmt.Errorf("apisys code=%d", result.Code)
	}

	return &result.Data, nil
}

func (a *apiSysAdapter) GetUserInfoByToken(ctx context.Context, token string) (*ApiSysGetUserData, error) {
	url := a.host + "/api/internal/v1/user/info"

	respBody, err := a.doRequest(ctx, http.MethodGet, url, nil, a.headersWithToken(token))
	if err != nil {
		return nil, err
	}

	return parseApiSysGetUserResponse(string(respBody))
}

func (a *apiSysAdapter) headersWithToken(token string) map[string]string {
	return map[string]string{
		"Authorization":   token,
		"Accept-Language": a.lang,
		"Content-Type":    "application/json",
	}
}

func (a *apiSysAdapter) doRequest(ctx context.Context, method, url string, body io.Reader, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	dumpReq, _ := httputil.DumpRequestOut(req, true)
	log.Printf("HTTP REQUEST:\n%s", string(dumpReq))

	resp, err := a.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("HTTP RESPONSE: status=%d", resp.StatusCode)
	log.Printf("HTTP RESPONSE BODY: %s", string(respBody))

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("apisys http status=%d body=%s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (a *apiSysAdapter) ResolveGameKey(ctx context.Context, channelID int64) string {
	if channelID == 0 {
		return defaultGameKey
	}
	m, _ := a.gm.FindOneByChannelId(ctx, channelID)
	if m == nil || m.GameKey == "" {
		return defaultGameKey
	}
	return m.GameKey
}

func (a *apiSysAdapter) headers() map[string]string {
	return map[string]string{
		"Authorization":   a.token,
		"Accept-Language": a.lang,
		"Content-Type":    "application/json",
	}
}

//func (a *apiSysAdapter) doJSON(ctx context.Context, method, url string, body any, out any) error {
//	buf, err := json.Marshal(body)
//	if err != nil {
//		return err
//	}
//	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(buf))
//	if err != nil {
//		return err
//	}
//	for k, v := range a.headers() {
//		req.Header.Set(k, v)
//	}
//	resp, err := a.http.Do(req)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//	if resp.StatusCode >= 300 {
//		return fmt.Errorf("apisys http status=%d", resp.StatusCode)
//	}
//	return json.NewDecoder(resp.Body).Decode(out)
//}

func (a *apiSysAdapter) doJSON(ctx context.Context, method, url string, body any, out any) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(buf))
	if err != nil {
		return err
	}

	for k, v := range a.headers() {
		req.Header.Set(k, v)
	}

	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 打印最终请求
	dumpReq, _ := httputil.DumpRequestOut(req, true)
	log.Printf("HTTP REQUEST:\n%s", string(dumpReq))

	resp, err := a.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("HTTP RESPONSE: status=%d", resp.StatusCode)
	log.Printf("HTTP RESPONSE BODY: %s", string(respBody))

	if resp.StatusCode >= 300 {
		return fmt.Errorf("apisys http status=%d body=%s", resp.StatusCode, string(respBody))
	}

	if len(respBody) == 0 {
		return nil
	}

	return json.Unmarshal(respBody, out)
}

func (a *apiSysAdapter) checkCommonError(resp *apiResp) error {
	if resp.Code == 200 {
		return nil
	}
	if resp.MessageCode == errOrderAlreadySettle {
		return nil
	}
	if resp.Message != "" {
		return errors.New(resp.Message)
	}
	return fmt.Errorf("apisys code=%d", resp.Code)
}

func (a *apiSysAdapter) Deduct(ctx context.Context, req DeductRequest) (*DeductResponse, error) {
	gameKey := a.ResolveGameKey(ctx, req.ChannelID)
	payload := map[string]any{
		"game_key":        gameKey,
		"parent_wager_no": req.OrderNo,
		"user_id":         req.UserID,
		"currency":        req.Currency,
		"amount":          mustFloat(req.Amount) * -1,
		"orders": []map[string]any{{
			"wager_no":         req.OrderNo,
			"ticket_no":        req.OrderNo,
			"amount":           req.Amount,
			"effective_amount": req.Amount,
			"metadata":         req.Metadata,
		}},
		"is_system_reward": req.IsSystemReward,
	}
	var resp apiResp
	if err := a.doJSON(ctx, http.MethodPost, a.host+"/api/internal/v1/payment/request", payload, &resp); err != nil {
		return nil, err
	}

	if err := a.checkCommonError(&resp); err != nil {
		return nil, err
	}
	var data struct {
		Status        int64  `json:"status"`
		TransactionNo string `json:"transaction_no"`
	}
	_ = json.Unmarshal(resp.Data, &data)
	if data.Status != 1 {
		return nil, fmt.Errorf("apisys deduct status=%d", data.Status)
	}
	return &DeductResponse{Status: data.Status, TransactionNo: data.TransactionNo}, nil
}

func (a *apiSysAdapter) Refund(ctx context.Context, req RefundRequest) error {
	gameKey := a.ResolveGameKey(ctx, req.ChannelID)
	payload := map[string]any{
		"game_key": gameKey,
		"wager_no": req.OrderNo,
		"metadata": req.Metadata,
	}
	var resp apiResp
	if err := a.doJSON(ctx, http.MethodPost, a.host+"/api/internal/v1/wager/cancel", payload, &resp); err != nil {
		return err
	}
	return a.checkCommonError(&resp)
}

func (a *apiSysAdapter) BillRolling(ctx context.Context, req BillRequest) error {
	return a.bill(ctx, req, false, 0)
}

func (a *apiSysAdapter) BillPreMatch(ctx context.Context, req BillRequest, partial bool, partialCount int8) error {
	return a.bill(ctx, req, partial, partialCount)
}

func (a *apiSysAdapter) bill(ctx context.Context, req BillRequest, partial bool, partialCount int8) error {
	gameKey := a.ResolveGameKey(ctx, req.ChannelID)
	payload := []map[string]any{{
		"game_key":                 gameKey,
		"wager_no":                 req.OrderNo,
		"currency":                 req.Currency,
		"amount":                   mustFloat(req.Amount),
		"metadata":                 req.Metadata,
		"settlement_time":          time.Now().Unix(),
		"is_partial_settlement":    partial,
		"partial_settlement_count": partialCount,
		"idempotency_key":          uuid.NewString(),
		"is_system_reward":         req.IsSystemReward,
	}}
	var resp apiResp
	if err := a.doJSON(ctx, http.MethodPost, a.host+"/api/internal/v1/wager/bulkSettle", payload, &resp); err != nil {
		return err
	}
	return a.checkCommonError(&resp)
}

func (a *apiSysAdapter) BatchBill(ctx context.Context, reqs []BillRequest) ([]string, []string, error) {
	if len(reqs) == 0 {
		return nil, nil, nil
	}
	payload := make([]map[string]any, 0, len(reqs))
	for _, req := range reqs {
		payload = append(payload, map[string]any{
			"game_key":         a.ResolveGameKey(ctx, req.ChannelID),
			"wager_no":         req.OrderNo,
			"currency":         req.Currency,
			"amount":           mustFloat(req.Amount),
			"metadata":         req.Metadata,
			"settlement_time":  time.Now().Unix(),
			"is_system_reward": req.IsSystemReward,
		})
	}
	var resp apiWagerResp
	if err := a.doJSON(ctx, http.MethodPost, a.host+"/api/internal/v1/wager/bulkSettle", payload, &resp); err != nil {
		return nil, nil, err
	}
	var suc []string
	var fail []string
	for _, item := range resp.Data.Wagers {
		if item.Code == 200 {
			suc = append(suc, item.WagerNo)
		} else {
			fail = append(fail, item.WagerNo)
		}
	}
	return suc, fail, nil
}

func mustFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
