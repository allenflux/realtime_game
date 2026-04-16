package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	rtconfig "crash/realtime_game/config"
	appctx "crash/realtime_game/context"
	"crash/realtime_game/service"
	rttypes "crash/realtime_game/types"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "realtime_game/etc/realtime-api.docker.yaml", "配置文件")

func main() {
	flag.Parse()

	var c rtconfig.Config
	conf.MustLoad(*configFile, &c)
	ctx := appctx.New(c)
	svc := service.New(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	mux.HandleFunc("/v2/game/current-round", func(w http.ResponseWriter, r *http.Request) {
		channelID, _ := strconv.ParseInt(r.URL.Query().Get("channel_id"), 10, 64)
		resp, err := service.NewCurrentRoundService(svc).Get(r.Context(), channelID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if resp == nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"message": "round not found"})
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})

	mux.HandleFunc("/v2/game/profile", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("api_sys_token")
		if token == "" {
			writeError(w, http.StatusBadRequest, fmt.Errorf("api_sys_token required"))
			return
		}
		userData, err := service.GetApiSysUserData(r.Context(), token, ctx)
		if err != nil {
			writeError(w, http.StatusUnauthorized, err)
			return
		}
		writeJSON(w, http.StatusOK, userData)
	})

	mux.HandleFunc("/v2/game/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		channelID, _ := strconv.ParseInt(r.URL.Query().Get("channel_id"), 10, 64)
		resp, err := service.NewFeatureQueryService(svc).Leaderboard(r.Context(), channelID)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})

	mux.HandleFunc("/v2/game/jackpot", func(w http.ResponseWriter, r *http.Request) {
		channelID, _ := strconv.ParseInt(r.URL.Query().Get("channel_id"), 10, 64)
		currency := r.URL.Query().Get("currency")
		resp, err := service.NewFeatureQueryService(svc).Jackpot(r.Context(), channelID, currency)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if resp == nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"message": "jackpot not found"})
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})

	mux.HandleFunc("/v2/game/bet", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"message": "method not allowed"})
			return
		}
		var req rttypes.CreateBetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		resp, err := service.NewPlaceBetService(svc).Place(r.Context(), &req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})

	mux.HandleFunc("/v2/game/cashout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"message": "method not allowed"})
			return
		}
		var req rttypes.CashoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		resp, err := service.NewCashoutService(svc).Cashout(r.Context(), &req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})

	mux.HandleFunc("/v2/game/my-bets", func(w http.ResponseWriter, r *http.Request) {
		ctxReq := r.Context()
		channelIDStr := r.URL.Query().Get("channel_id")
		channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid channel_id"))
			return
		}
		currency := r.URL.Query().Get("currency")
		token := r.URL.Query().Get("api_sys_token")
		if token == "" {
			writeError(w, http.StatusBadRequest, fmt.Errorf("api_sys_token required"))
			return
		}
		limitStr := r.URL.Query().Get("limit")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 200 {
			limit = 50
		}
		userData, err := service.GetApiSysUserData(ctxReq, token, ctx)
		if err != nil {
			writeError(w, http.StatusUnauthorized, err)
			return
		}
		if userData == nil {
			writeError(w, http.StatusUnauthorized, fmt.Errorf("user not found"))
			return
		}

		userID := userData.ID
		now := time.Now()
		start := now.Add(-24 * time.Hour)

		bets, err := ctx.BetModel.GetUserBets(ctxReq,
			userID,
			channelID,
			0,
			start,
			now,
			limit,
			currency,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, bets)
	})

	addr := c.API.Host + ":" + strconv.Itoa(c.API.Port)
	srv := &http.Server{Addr: addr, Handler: logMiddleware(mux)}
	logx.Infof("realtime-api listen on %s", addr)
	logx.Must(srv.ListenAndServe())
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]any{"message": err.Error()})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logx.Infof("realtime-api %s %s cost=%s", r.Method, r.URL.Path, time.Since(start))
	})
}
