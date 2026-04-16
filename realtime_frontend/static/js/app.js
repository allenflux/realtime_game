(() => {
  const params = new URLSearchParams(window.location.search);

  const state = {
    channelID: Number(window.__APP_INIT__.channelID || params.get("channel_id") || 1001),
    apiSysToken: String(window.__APP_INIT__.apiSysToken || params.get("api_sys_token") || "token-demo-1"),
    round: null,
    profile: null,
    leaderboard: [],
    jackpot: null,
    bets: [],
    roundHistory: [],
    stars: [],
    pollTimer: null,
    renderTimer: null,
  };

  const $ = (id) => document.getElementById(id);

  const els = {
    channelId: $("channelId"),
    apiSysToken: $("apiSysToken"),
    applyUrlBtn: $("applyUrlBtn"),
    reloadBtn: $("reloadBtn"),
    profileBadge: $("profileBadge"),
    profileName: $("profileName"),
    profileId: $("profileId"),
    roundId: $("roundId"),
    roundState: $("roundState"),
    countdown: $("countdown"),
    serverTime: $("serverTime"),
    version: $("version"),
    canBet: $("canBet"),
    canCashout: $("canCashout"),
    currentMultiple: $("currentMultiple"),
    statusLine: $("statusLine"),
    roundSnapshot: $("roundSnapshot"),
    betFormA: $("betFormA"),
    betFormB: $("betFormB"),
    betResultA: $("betResultA"),
    betResultB: $("betResultB"),
    betSubmitButtons: document.querySelectorAll(".bet-submit-btn"),
    cashoutForm: $("cashoutForm"),
    cashoutResult: $("cashoutResult"),
    cashoutSubmitBtn: $("cashoutSubmitBtn"),
    queryBetsBtn: $("queryBetsBtn"),
    copyTopOrderBtn: $("copyTopOrderBtn"),
    syncAllBtn: $("syncAllBtn"),
    opsResult: $("opsResult"),
    betsTableBody: document.querySelector("#betsTable tbody"),
    logs: $("logs"),
    historyStrip: $("historyStrip"),
    rocketShip: $("rocketShip"),
    flightTrail: $("flightTrail"),
    flightTrailGlow: $("flightTrailGlow"),
    flightCanvas: $("flightCanvas"),
    leaderboard: $("leaderboard"),
    jackpotBalance: $("jackpotBalance"),
    jackpotPrize1: $("jackpotPrize1"),
    jackpotPrize2: $("jackpotPrize2"),
    jackpotPrize3: $("jackpotPrize3"),
    jackpotIn: $("jackpotIn"),
  };

  const canvasCtx = els.flightCanvas.getContext("2d");

  function syncInputs() {
    els.channelId.value = String(state.channelID);
    els.apiSysToken.value = state.apiSysToken;
  }

  function log(message, ok = true) {
    const item = document.createElement("div");
    item.className = "log-item";
    item.innerHTML = `<span class="log-time">${new Date().toLocaleTimeString()}</span><span class="${ok ? "status-good" : "status-bad"}">${escapeHtml(message)}</span>`;
    els.logs.prepend(item);
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
    const response = await fetch(url, {
      headers: { "Content-Type": "application/json" },
      ...options,
    });
    const text = await response.text();
    let data;
    try {
      data = text ? JSON.parse(text) : null;
    } catch {
      data = text;
    }
    if (!response.ok) {
      const message = typeof data === "object" && data && data.message ? data.message : `HTTP ${response.status}`;
      throw new Error(message);
    }
    return data;
  }

  function safeNumber(value, fallback = 0) {
    const number = Number(value);
    return Number.isFinite(number) ? number : fallback;
  }

  function formatMoney(value) {
    const amount = Number(value || 0);
    return Number.isFinite(amount) ? amount.toFixed(2) : "--";
  }

  function formatMultiple100(value) {
    return `${(safeNumber(value, 100) / 100).toFixed(2)}x`;
  }

  function formatServerTime(ms) {
    if (!ms) return "--";
    return new Date(ms).toLocaleTimeString();
  }

  function countdownTarget(round) {
    if (!round) return 0;
    const currentState = String(round.state || "");
    if (currentState === "pre_start") return safeNumber(round.bet_close_at_ms);
    if (currentState === "starting") return safeNumber(round.flying_start_at_ms);
    if (currentState === "flying") return safeNumber(round.crash_at_ms);
    return safeNumber(round.close_at_ms);
  }

  function calcCurrentMultiple(round, nowMs) {
    if (!round) return 100;
    const currentState = String(round.state || "");
    if (currentState === "pre_start" || currentState === "starting") return 100;
    if (currentState === "crashed" || currentState === "closed") return safeNumber(round.crash_multiple, 100);
    const flyingStart = safeNumber(round.flying_start_at_ms);
    const incNum = safeNumber(round.inc_num, 1);
    if (!flyingStart || incNum <= 1 || nowMs <= flyingStart) return 100;
    const elapsedSeconds = (nowMs - flyingStart) / 1000;
    let multiple = Math.floor(Math.pow(incNum, elapsedSeconds) * 100);
    const crashMultiple = safeNumber(round.crash_multiple, 100);
    if (multiple > crashMultiple) multiple = crashMultiple;
    if (multiple < 100) multiple = 100;
    return multiple;
  }

  function roundStatusLine(round, currentMultiple) {
    if (!round) return "Waiting for telemetry";
    const currentState = String(round.state || "");
    if (currentState === "pre_start") return "Fueling engines. Bets are open.";
    if (currentState === "starting") return "Ignition sequence in progress.";
    if (currentState === "flying") return `Ship is live at ${formatMultiple100(currentMultiple)}. Cash out before impact.`;
    if (currentState === "crashed") return `Ship crashed at ${formatMultiple100(round.crash_multiple)}.`;
    if (currentState === "closed") return "Round closed. Next ship loading.";
    return "Telemetry acquired.";
  }

  function canBet(round) {
    if (!round) return false;
    const currentState = String(round.state || "");
    return currentState === "pre_start" || currentState === "flying";
  }

  function canCashout(round) {
    return Boolean(round) && String(round.state || "") === "flying";
  }

  function buildFlightPath(progress) {
    const clamped = Math.max(0, Math.min(progress, 1));
    const endX = 130 + clamped * 720;
    const endY = 455 - Math.pow(clamped, 0.8) * 310;
    return `M 90 455 C 180 445, 280 420, ${Math.max(260, endX - 180)} ${Math.max(180, endY + 120)} S ${Math.max(420, endX - 80)} ${Math.max(120, endY + 30)}, ${endX} ${endY}`;
  }

  function setRocketPosition(progress) {
    const clamped = Math.max(0, Math.min(progress, 1));
    const x = 56 + clamped * 70;
    const y = -clamped * 290;
    const rotate = -14 - clamped * 12;
    const scale = 1 + clamped * 0.18;
    els.rocketShip.style.transform = `translate(${x}%, ${y}%) rotate(${rotate}deg) scale(${scale})`;
  }

  function renderHistory() {
    const history = state.roundHistory.slice(0, 12);
    els.historyStrip.innerHTML = history.map((value) => {
      const hot = Number(value) >= 2 ? "history-hot" : "history-cold";
      return `<div class="history-item ${hot}">${escapeHtml(value)}</div>`;
    }).join("");
  }

  function resizeCanvas() {
    const rect = els.flightCanvas.getBoundingClientRect();
    const ratio = Math.min(window.devicePixelRatio || 1, 2);
    els.flightCanvas.width = Math.floor(rect.width * ratio);
    els.flightCanvas.height = Math.floor(rect.height * ratio);
    canvasCtx.setTransform(ratio, 0, 0, ratio, 0, 0);
  }

  function initStars() {
    const rect = els.flightCanvas.getBoundingClientRect();
    state.stars = Array.from({ length: 70 }, () => ({
      x: Math.random() * rect.width,
      y: Math.random() * rect.height,
      r: Math.random() * 1.8 + 0.4,
      alpha: Math.random() * 0.7 + 0.2,
      speed: Math.random() * 0.4 + 0.1,
    }));
  }

  function drawCanvas(progress) {
    const rect = els.flightCanvas.getBoundingClientRect();
    canvasCtx.clearRect(0, 0, rect.width, rect.height);

    for (const star of state.stars) {
      star.y += star.speed;
      if (star.y > rect.height) {
        star.y = -4;
        star.x = Math.random() * rect.width;
      }
      canvasCtx.beginPath();
      canvasCtx.fillStyle = `rgba(255,255,255,${star.alpha})`;
      canvasCtx.arc(star.x, star.y, star.r, 0, Math.PI * 2);
      canvasCtx.fill();
    }

    const clamped = Math.max(0, Math.min(progress, 1));
    const rocketX = 120 + clamped * (rect.width * 0.72);
    const rocketY = rect.height - 96 - Math.pow(clamped, 0.82) * (rect.height * 0.55);
    const flameLen = 40 + clamped * 36 + Math.sin(Date.now() / 70) * 8;

    const gradient = canvasCtx.createLinearGradient(rocketX - flameLen, rocketY, rocketX + 16, rocketY);
    gradient.addColorStop(0, "rgba(249,115,22,0)");
    gradient.addColorStop(0.35, "rgba(249,115,22,0.65)");
    gradient.addColorStop(1, "rgba(56,189,248,0)");
    canvasCtx.beginPath();
    canvasCtx.strokeStyle = gradient;
    canvasCtx.lineWidth = 10;
    canvasCtx.lineCap = "round";
    canvasCtx.moveTo(rocketX - flameLen, rocketY + 4);
    canvasCtx.lineTo(rocketX - 10, rocketY + 2);
    canvasCtx.stroke();
  }

  function pushHistory(value) {
    if (!value) return;
    if (state.roundHistory[0] === value) return;
    state.roundHistory.unshift(value);
    state.roundHistory = state.roundHistory.slice(0, 20);
    renderHistory();
  }

  function renderRound() {
    const round = state.round;
    if (!round) return;

    const nowMs = Date.now();
    const currentMultiple = calcCurrentMultiple(round, nowMs);
    const targetMs = countdownTarget(round);
    const leftMs = targetMs > 0 ? Math.max(0, targetMs - nowMs) : 0;
    const allowBet = canBet(round);
    const allowCashout = canCashout(round);

    els.roundId.textContent = String(round.term_id || "--");
    els.roundState.textContent = String(round.state || "--");
    els.countdown.textContent = `${(leftMs / 1000).toFixed(1)}s`;
    els.serverTime.textContent = formatServerTime(round.server_time_ms || nowMs);
    els.version.textContent = String(round.version || "--");
    els.canBet.textContent = allowBet ? "Open" : "Closed";
    els.canCashout.textContent = allowCashout ? "Ready" : "Locked";
    els.currentMultiple.textContent = formatMultiple100(currentMultiple);
    els.statusLine.textContent = roundStatusLine(round, currentMultiple);
    els.betSubmitButtons.forEach((button) => {
      button.disabled = !allowBet;
    });
    els.cashoutSubmitBtn.disabled = !allowCashout;
    els.roundSnapshot.textContent = JSON.stringify({ ...round, local_current_multiple: currentMultiple }, null, 2);

    if (String(round.state) === "crashed" || String(round.state) === "closed") {
      pushHistory(formatMultiple100(round.crash_multiple));
    }

    let progress = 0;
    if (String(round.state) === "starting") {
      const startAt = safeNumber(round.bet_close_at_ms);
      const endAt = safeNumber(round.flying_start_at_ms);
      progress = endAt > startAt ? (nowMs - startAt) / (endAt - startAt) * 0.12 : 0.05;
    } else if (String(round.state) === "flying") {
      const startAt = safeNumber(round.flying_start_at_ms);
      const endAt = safeNumber(round.crash_at_ms);
      progress = endAt > startAt ? (nowMs - startAt) / (endAt - startAt) : 0;
    } else if (String(round.state) === "crashed" || String(round.state) === "closed") {
      progress = 1;
    }
    const path = buildFlightPath(progress);
    els.flightTrail.setAttribute("d", path);
    els.flightTrailGlow.setAttribute("d", path);
    setRocketPosition(progress);
    drawCanvas(progress);
  }

  async function loadProfile() {
    state.profile = await request(`/api/proxy/profile?api_sys_token=${encodeURIComponent(state.apiSysToken)}`);
    els.profileBadge.textContent = `ID ${state.profile.id}`;
    els.profileName.textContent = state.profile.username || "Unknown Pilot";
    els.profileId.textContent = `Token bound user`;
  }

  async function loadRound() {
    const round = await request(`/api/proxy/current-round?channel_id=${state.channelID}`);
    state.round = round;
    renderRound();
  }

  async function loadLeaderboard() {
    state.leaderboard = (await request(`/api/proxy/leaderboard?channel_id=${state.channelID}`)).items || [];
    els.leaderboard.innerHTML = state.leaderboard.map((item, index) => `
      <div class="leaderboard-item">
        <div class="leader-rank">#${index + 1}</div>
        <div class="leader-main">
          <strong>${escapeHtml(item.user_name || "Pilot")}</strong>
          <span>${escapeHtml(item.order_no || "--")}</span>
        </div>
        <div class="leader-score">
          <strong>${escapeHtml(item.payout || "--")}</strong>
          <span>${escapeHtml(item.multiplier || "--")}</span>
        </div>
      </div>`).join("") || `<div class="leaderboard-item"><div class="leader-rank">-</div><div class="leader-main"><strong>No winners yet</strong><span>Round is warming up</span></div><div class="leader-score"><strong>--</strong><span>--</span></div></div>`;
  }

  async function loadJackpot() {
    try {
      state.jackpot = await request(`/api/proxy/jackpot?channel_id=${state.channelID}&currency=USD`);
      els.jackpotBalance.textContent = state.jackpot.jackpot_balance || "--";
      els.jackpotPrize1.textContent = state.jackpot.jackpot_prize_1 || "--";
      els.jackpotPrize2.textContent = state.jackpot.jackpot_prize_2 || "--";
      els.jackpotPrize3.textContent = state.jackpot.jackpot_prize_3 || "--";
      els.jackpotIn.textContent = state.jackpot.jackpot_in || "--";
    } catch {
      els.jackpotBalance.textContent = "--";
      els.jackpotPrize1.textContent = "--";
      els.jackpotPrize2.textContent = "--";
      els.jackpotPrize3.textContent = "--";
      els.jackpotIn.textContent = "--";
    }
  }

  function orderStatusText(status) {
    const mapping = {
      1000: "Creating",
      2000: "Created",
      3000: "Cashing",
      4000: "Cashed",
      5000: "Refunding",
      6000: "Refunded",
      9999: "Retry",
      10100: "Failed",
    };
    return mapping[status] || String(status || "--");
  }

  function gamePlayText(value) {
    return Number(value) === 1 ? "Pre Match" : "Rolling";
  }

  async function loadBets() {
    state.bets = await request(`/api/proxy/my-bets?channel_id=${state.channelID}&api_sys_token=${encodeURIComponent(state.apiSysToken)}&limit=50&currency=USD`);
    els.betsTableBody.innerHTML = (state.bets || []).map((row) => {
      const orderNo = escapeHtml(row.api_order_no || "");
      return `
        <tr>
          <td><button data-order-no="${orderNo}" class="link-btn" type="button">${orderNo || "--"}</button></td>
          <td>${formatMoney(row.amount / 10000)}</td>
          <td>${escapeHtml(row.bet_at_multiple || "--")}</td>
          <td>${escapeHtml(row.auto_cashout_multiple || "--")}</td>
          <td>${escapeHtml(row.manual_cashout_multiple || "--")}</td>
          <td>${formatMoney(row.cashed_out_amount / 10000)}</td>
          <td>${escapeHtml(orderStatusText(row.order_status))}</td>
          <td>${escapeHtml(gamePlayText(row.game_play))}</td>
        </tr>`;
    }).join("");
  }

  async function submitBet(event) {
    event.preventDefault();
    if (!state.profile) {
      log("用户信息未加载完成", false);
      return;
    }
    const form = event.currentTarget;
    const panel = String(form.dataset.panel || "A");
    const resultBox = panel === "B" ? els.betResultB : els.betResultA;
    const formData = new FormData(form);
    const payload = {
      channel_id: state.channelID,
      api_sys_token: state.apiSysToken,
      amount: String(formData.get("amount") || ""),
      currency: String(formData.get("currency") || "USD"),
      auto_cashout_multiple: String(formData.get("auto_cashout_multiple") || ""),
      game_play: safeNumber(formData.get("game_play"), 0),
      user_seed: String(formData.get("user_seed") || ""),
    };

    try {
      const response = await request("/api/proxy/bet", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      resultBox.textContent = JSON.stringify(response, null, 2);
      if (response && response.api_order_no) {
        els.cashoutForm.elements.order_no.value = response.api_order_no;
      }
      els.opsResult.textContent = JSON.stringify({ panel, last_order_no: response.api_order_no || "--" }, null, 2);
      log(`下注成功 ${panel}：${response.api_order_no || "--"}`);
      await refreshData();
    } catch (error) {
      resultBox.textContent = JSON.stringify({ message: error.message }, null, 2);
      log(`下注失败 ${panel}：${error.message}`, false);
    }
  }

  async function submitCashout(event) {
    event.preventDefault();
    const formData = new FormData(els.cashoutForm);
    const payload = {
      api_sys_token: state.apiSysToken,
      order_no: String(formData.get("order_no") || ""),
      game_play: safeNumber(formData.get("game_play"), 0),
      settlement_mode: safeNumber(formData.get("settlement_mode"), 1),
    };

    try {
      const response = await request("/api/proxy/cashout", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      els.cashoutResult.textContent = JSON.stringify(response, null, 2);
      log(`兑现成功：${payload.order_no}`);
      await refreshData();
    } catch (error) {
      els.cashoutResult.textContent = JSON.stringify({ message: error.message }, null, 2);
      log(`兑现失败：${error.message}`, false);
    }
  }

  function updateUrlState() {
    state.channelID = safeNumber(els.channelId.value, 1001);
    state.apiSysToken = String(els.apiSysToken.value || "token-demo-1").trim();
    const url = new URL(window.location.href);
    url.searchParams.set("channel_id", String(state.channelID));
    url.searchParams.set("api_sys_token", state.apiSysToken);
    window.history.replaceState({}, "", url);
    log("URL 参数已更新");
  }

  async function refreshData() {
    await Promise.allSettled([loadRound(), loadProfile(), loadBets(), loadLeaderboard(), loadJackpot()]);
  }

  function bindEvents() {
    els.applyUrlBtn.addEventListener("click", async () => {
      updateUrlState();
      await refreshData();
    });

    els.reloadBtn.addEventListener("click", async () => {
      try {
        updateUrlState();
        await refreshData();
        log("状态已同步");
      } catch (error) {
        log(`同步失败：${error.message}`, false);
      }
    });

    els.betFormA.addEventListener("submit", submitBet);
    els.betFormB.addEventListener("submit", submitBet);
    els.cashoutForm.addEventListener("submit", submitCashout);
    els.queryBetsBtn.addEventListener("click", async () => {
      try {
        await loadBets();
        log("订单已刷新");
      } catch (error) {
        log(`订单刷新失败：${error.message}`, false);
      }
    });
    els.betsTableBody.addEventListener("click", (event) => {
      const button = event.target.closest("button[data-order-no]");
      if (!button) return;
      els.cashoutForm.elements.order_no.value = button.getAttribute("data-order-no") || "";
      log(`已带入订单号 ${els.cashoutForm.elements.order_no.value}`);
    });
    document.querySelectorAll(".chip-btn").forEach((button) => {
      button.addEventListener("click", () => {
        const target = document.getElementById(button.dataset.targetForm);
        if (!target) return;
        target.elements.amount.value = button.dataset.amount || "10.00";
      });
    });
    els.copyTopOrderBtn.addEventListener("click", () => {
      const first = state.bets[0];
      if (!first) {
        log("暂无可用订单", false);
        return;
      }
      els.cashoutForm.elements.order_no.value = first.api_order_no || "";
      els.opsResult.textContent = JSON.stringify({ copied_order_no: first.api_order_no || "" }, null, 2);
      log(`已使用最新订单 ${first.api_order_no || "--"}`);
    });
    els.syncAllBtn.addEventListener("click", async () => {
      await refreshData();
      log("全量数据已刷新");
    });
    window.addEventListener("resize", () => {
      resizeCanvas();
      initStars();
      renderRound();
    });
  }

  function startTimers() {
    state.renderTimer = window.setInterval(renderRound, 50);
    state.pollTimer = window.setInterval(async () => {
      try {
        await Promise.allSettled([loadRound(), loadLeaderboard(), loadJackpot()]);
      } catch (error) {
        log(`轮询失败：${error.message}`, false);
      }
    }, 1000);
  }

  async function boot() {
    syncInputs();
    resizeCanvas();
    initStars();
    bindEvents();
    renderHistory();
    try {
      await refreshData();
      log("Crash 控制台已就绪");
    } catch (error) {
      log(`初始化失败：${error.message}`, false);
    }
    startTimers();
  }

  boot();
})();
