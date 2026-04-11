(() => {
  const state = {
    channelID: window.__APP_INIT__.channelID || 19999,
    round: null,
    pollTimer: null,
    renderTimer: null,
  };

  const $ = (id) => document.getElementById(id);

  const els = {
    channelId: $("channelId"),
    reloadBtn: $("reloadBtn"),
    roundState: $("roundState"),
    currentMultiple: $("currentMultiple"),
    serverTime: $("serverTime"),
    countdown: $("countdown"),
    roundId: $("roundId"),
    version: $("version"),
    canBet: $("canBet"),
    canCashout: $("canCashout"),
    roundSnapshot: $("roundSnapshot"),
    betForm: $("betForm"),
    betResult: $("betResult"),
    betSubmitBtn: $("betSubmitBtn"),
    cashoutForm: $("cashoutForm"),
    cashoutResult: $("cashoutResult"),
    cashoutSubmitBtn: $("cashoutSubmitBtn"),
    queryUserId: $("queryUserId"),
    queryBetsBtn: $("queryBetsBtn"),
    betsTableBody: document.querySelector("#betsTable tbody"),
    logs: $("logs"),
  };

  function log(message, ok = true) {
    const div = document.createElement("div");
    div.className = "log-item";
    const time = new Date().toLocaleTimeString();
    div.innerHTML = `<span class="log-time">${time}</span><span class="${ok ? "status-good" : "status-bad"}">${escapeHtml(message)}</span>`;
    els.logs.prepend(div);
  }

  function escapeHtml(value) {
    return String(value)
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  async function request(url, options = {}) {
    const resp = await fetch(url, {
      headers: { "Content-Type": "application/json" },
      ...options,
    });
    const text = await resp.text();
    let data;
    try {
      data = text ? JSON.parse(text) : null;
    } catch {
      data = text;
    }
    if (!resp.ok) {
      const message = typeof data === "object" && data && data.message ? data.message : `HTTP ${resp.status}`;
      throw new Error(message);
    }
    return data;
  }

  function safeNumber(v, fallback = 0) {
    const n = Number(v);
    return Number.isFinite(n) ? n : fallback;
  }

  function calcMultiple(round, nowMs) {
    if (!round) return 100;
    const roundState = String(round.state || "");
    if (roundState === "pre_start" || roundState === "starting") return 100;
    if (roundState === "crashed" || roundState === "closed") {
      return safeNumber(round.crash_multiple || round.current_multiple || 100, 100);
    }
    const flyingStartAt = safeNumber(round.flying_start_at_ms);
    const incNum = safeNumber(round.inc_num, 1);
    if (!flyingStartAt || incNum <= 1 || nowMs <= flyingStartAt) return 100;
    const elapsedSec = (nowMs - flyingStartAt) / 1000;
    let mul = Math.floor(Math.pow(incNum, elapsedSec) * 100);
    const crashMultiple = safeNumber(round.crash_multiple, 0);
    if (crashMultiple > 0 && mul > crashMultiple) mul = crashMultiple;
    if (mul < 100) mul = 100;
    return mul;
  }

  function formatMultiple100(v) {
    return `${(safeNumber(v) / 100).toFixed(2)}x`;
  }

  function formatTs(ms) {
    if (!ms) return "--";
    return new Date(ms).toLocaleTimeString();
  }

  function formatCountdown(ms) {
    if (ms <= 0) return "0.0s";
    return `${(ms / 1000).toFixed(1)}s`;
  }

  function canBet(round) {
    if (!round) return false;
    const s = String(round.state || "");
    return s === "pre_start" || s === "flying";
  }

  function canCashout(round) {
    if (!round) return false;
    return String(round.state || "") === "flying";
  }

  function renderRound() {
    const round = state.round;
    if (!round) return;

    const nowMs = Date.now();
    const currentMultiple = calcMultiple(round, nowMs);
    const crashAtMs = safeNumber(round.crash_at_ms);
    const closeAtMs = safeNumber(round.close_at_ms);
    const targetAt = ["crashed", "closed"].includes(String(round.state)) ? closeAtMs : crashAtMs;
    const leftMs = targetAt > 0 ? targetAt - nowMs : 0;
    const allowBet = canBet(round);
    const allowCashout = canCashout(round);

    els.roundState.textContent = round.state || "--";
    els.currentMultiple.textContent = formatMultiple100(currentMultiple);
    els.serverTime.textContent = formatTs(round.server_time_ms || nowMs);
    els.countdown.textContent = formatCountdown(leftMs);
    els.roundId.textContent = String(round.round_id ?? round.term_id ?? "--");
    els.version.textContent = String(round.version ?? "--");
    els.canBet.textContent = allowBet ? "是" : "否";
    els.canCashout.textContent = allowCashout ? "是" : "否";
    els.betSubmitBtn.disabled = !allowBet;
    els.cashoutSubmitBtn.disabled = !allowCashout;
    els.roundSnapshot.textContent = JSON.stringify({ ...round, local_current_multiple: currentMultiple }, null, 2);
  }

  async function loadRound() {
    state.channelID = safeNumber(els.channelId.value, 19999);
    const round = await request(`/api/proxy/current-round?channel_id=${state.channelID}`);
    state.round = round;
    renderRound();
  }

  async function loadBets() {
    const userId = safeNumber(els.queryUserId.value, 0);
    if (!userId) {
      log("查询订单失败：用户ID不能为空", false);
      return;
    }
    const list = await request(`/api/proxy/my-bets?channel_id=${state.channelID}&user_id=${userId}&limit=50`);
    const rows = Array.isArray(list) ? list : [];
    els.betsTableBody.innerHTML = rows.map((row) => {
      const orderNo = escapeHtml(row.api_order_no || "");
      return `
        <tr>
          <td><button data-order-no="${orderNo}" class="link-btn" type="button">${orderNo || "--"}</button></td>
          <td>${escapeHtml(row.user_id ?? "")}</td>
          <td>${escapeHtml(row.amount ?? "")}</td>
          <td>${escapeHtml(row.bet_at_multiple ?? "")}</td>
          <td>${escapeHtml(row.auto_cashout_multiple ?? "")}</td>
          <td>${escapeHtml(row.manual_cashout_multiple ?? "")}</td>
          <td>${escapeHtml(row.cashed_out_amount ?? "")}</td>
          <td>${escapeHtml(row.order_status ?? "")}</td>
          <td>${escapeHtml(row.game_play ?? "")}</td>
        </tr>`;
    }).join("");
  }

  async function submitBet(event) {
    event.preventDefault();
    if (!canBet(state.round)) {
      log("当前阶段不允许下注", false);
      return;
    }

    const formData = new FormData(els.betForm);
    const payload = {
      channel_id: state.channelID,
      term_id: safeNumber(state.round?.round_id || state.round?.term_id || 0, 0),
      user_id: safeNumber(formData.get("user_id")),
      user_name: String(formData.get("user_name") || ""),
      amount: String(formData.get("amount") || ""),
      currency: String(formData.get("currency") || "CNY"),
      auto_cashout_multiple: String(formData.get("auto_cashout_multiple") || ""),
      game_play: safeNumber(formData.get("game_play"), 0),
      user_seed: String(formData.get("user_seed") || ""),
    };

    try {
      const resp = await request("/api/proxy/bet", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      els.betResult.textContent = JSON.stringify(resp, null, 2);
      if (resp && resp.api_order_no) {
        els.cashoutForm.elements.order_no.value = resp.api_order_no;
      }
      log(`下注成功，订单号：${resp.api_order_no || "--"}`);
      await loadRound();
      await loadBets();
    } catch (err) {
      els.betResult.textContent = JSON.stringify({ message: err.message }, null, 2);
      log(`下注失败：${err.message}`, false);
    }
  }

  async function submitCashout(event) {
    event.preventDefault();
    if (!canCashout(state.round)) {
      log("当前阶段不允许兑现", false);
      return;
    }

    const formData = new FormData(els.cashoutForm);
    const payload = {
      user_id: safeNumber(formData.get("user_id")),
      order_no: String(formData.get("order_no") || ""),
      game_play: safeNumber(formData.get("game_play"), 0),
      all_settlement: safeNumber(formData.get("all_settlement"), 1),
    };

    try {
      const resp = await request("/api/proxy/cashout", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      els.cashoutResult.textContent = JSON.stringify(resp, null, 2);
      log(`兑现成功，订单号：${payload.order_no}`);
      await loadRound();
      await loadBets();
    } catch (err) {
      els.cashoutResult.textContent = JSON.stringify({ message: err.message }, null, 2);
      log(`兑现失败：${err.message}`, false);
    }
  }

  function bindEvents() {
    els.reloadBtn.addEventListener("click", async () => {
      try {
        await loadRound();
        log("当前局状态刷新成功");
      } catch (err) {
        log(`刷新局状态失败：${err.message}`, false);
      }
    });

    els.queryBetsBtn.addEventListener("click", async () => {
      try {
        await loadBets();
        log("订单列表刷新成功");
      } catch (err) {
        log(`刷新订单失败：${err.message}`, false);
      }
    });

    els.betForm.addEventListener("submit", submitBet);
    els.cashoutForm.addEventListener("submit", submitCashout);

    els.betsTableBody.addEventListener("click", (event) => {
      const target = event.target.closest("button[data-order-no]");
      if (!target) return;
      const orderNo = target.getAttribute("data-order-no") || "";
      els.cashoutForm.elements.order_no.value = orderNo;
      log(`已带入订单号：${orderNo}`);
    });
  }

  function startTimers() {
    state.renderTimer = window.setInterval(renderRound, 100);
    state.pollTimer = window.setInterval(async () => {
      try {
        await loadRound();
      } catch (err) {
        log(`自动轮询失败：${err.message}`, false);
      }
    }, 1000);
  }

  async function boot() {
    bindEvents();
    try {
      await loadRound();
      await loadBets();
      log("页面初始化完成");
    } catch (err) {
      log(`初始化失败：${err.message}`, false);
    }
    startTimers();
  }

  boot();
})();
