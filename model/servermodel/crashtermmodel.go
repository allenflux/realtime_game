package servermodel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	CrashTerm_F_id                   = "id"
	CrashTerm_F_term_id              = "term_id"
	CrashTerm_F_channel_id           = "channel_id"
	CrashTerm_F_multiple             = "multiple"
	CrashTerm_F_is_control           = "is_control"
	CrashTerm_F_is_crashed           = "is_crashed"
	CrashTerm_F_bouns_pool_start     = "bouns_pool_start"
	CrashTerm_F_bouns_pool           = "bouns_pool"
	CrashTerm_F_term_hash            = "term_hash"
	CrashTerm_F_total_bet_amt        = "total_bet_amt"
	CrashTerm_F_fee_amt              = "fee_amt"
	CrashTerm_F_user_profit_correct  = "user_profit_correct"
	CrashTerm_F_rake_amt             = "rake_amt"
	CrashTerm_F_cashed_amt           = "cashed_amt"
	CrashTerm_F_ctrl_cashed_amt      = "ctrl_cashed_amt"
	CrashTerm_F_manual_squib         = "manual_squib"
	CrashTerm_F_manual_squib_state   = "manual_squib_state"
	CrashTerm_F_break_payout_rate    = "break_payout_rate"
	CrashTerm_F_profit_amt           = "profit_amt"
	CrashTerm_F_max_cashedout_bet_id = "max_cashedout_bet_id"
	CrashTerm_F_max_multiple         = "max_multiple"
	CrashTerm_F_create_time          = "create_time"
	CrashTerm_F_update_time          = "update_time"
	CrashTerm_F_pre_start_time       = "pre_start_time"
	CrashTerm_F_starting_time        = "starting_time"
	CrashTerm_F_flying_time          = "flying_time"
	CrashTerm_F_crashed_time         = "crashed_time"
	CrashTerm_F_ctime                = "ctime"
	CrashTerm_F_seed                 = "sha512_seed"
)

const (
	// 手动引爆阶段 0=未引爆 1=准备阶段 2=启动阶段 3=飞行阶段
	CrashTerm_Manual_Squib_State_null     = int64(0)
	CrashTerm_Manual_Squib_State_pre      = int64(1)
	CrashTerm_Manual_Squib_State_starting = int64(2)
	CrashTerm_Manual_Squib_State_flying   = int64(3)

	// 是否控制 1=否 2=是
	CrashTerm_Is_Control_no  = 1
	CrashTerm_Is_Control_yes = 2

	CrashTermIsCrashedNo  = int64(1)
	CrashTermIsCrashedYes = int64(2)

	//是否手动引爆 1=是 2=否
	ManualSquib_yes = 1
	ManualSquib_no  = 2
)

const (
	ChannelDefault    = 0
	ChannelBlockchain = 1

	ChannelBlockchainCrash  = 19998
	ChannelBlockchainRocket = 19999
)

var _ CrashTermModel = (*customCrashTermModel)(nil)

type (
	// CrashTermModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCrashTermModel.
	CrashTermModel interface {
		crashTermModel
		withSession(session sqlx.Session) CrashTermModel
		BatchInsert(ctx context.Context, dataList []*CrashTerm) ([]int64, error)
		BatchInsertIgnoreDupReturnIDs(ctx context.Context, dataList []*CrashTerm) ([]int64, error)
		GetNextTermNo(ctx context.Context, channelID int64) (int64, error)
		GetLatestTerm(ctx context.Context, channelID int64) (*CrashTerm, error)
		GetAllStatusLatestTerm(ctx context.Context, channelID int64) (*CrashTerm, error)
		//CloseTerm(ctx context.Context, termID int64) error
		GetByTermId(ctx context.Context, channelID, termID int64) (*CrashTerm, error)
		GetByIds(ctx context.Context, termIds []int64) ([]*CrashTerm, error)
		UpdateById(ctx context.Context, id int64, clauses map[string]interface{}) error
		UpdateByIds(ctx context.Context, ids []int64, clauses map[string]interface{}) error
		CloseAndRake(ctx context.Context, term *CrashTerm, rake int64, ctrlProfit int64, totalBetAmt, cashedAmt int64) error
		GetLatestTerms(ctx context.Context, channelID int64, pageSize int64) ([]*CrashTerm, error)
		GetPageById(ctx context.Context, channelID int64, posId int64, pageSize int64) ([]*CrashTerm, error)
		GetByHash(ctx context.Context, hash string) (*CrashTerm, error)
		GetAllPageById(ctx context.Context, posId int64, pageSize int64, cutoff int64) ([]*CrashTerm, error)
		GetPreviousIssueId(ctx context.Context, channelId, termId int64) (int, error)
		GetByChannelIdTermIds(ctx context.Context, channelId int64, termIds []int64) (map[int64]*CrashTerm, error)
	}

	customCrashTermModel struct {
		*defaultCrashTermModel
	}
)

// NewCrashTermModel returns a model for the database table.
func NewCrashTermModel(conn sqlx.SqlConn) CrashTermModel {
	return &customCrashTermModel{
		defaultCrashTermModel: newCrashTermModel(conn),
	}
}

func (m *customCrashTermModel) withSession(session sqlx.Session) CrashTermModel {
	return NewCrashTermModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customCrashTermModel) GetByChannelIdTermIds(ctx context.Context, channelId int64, termIds []int64) (map[int64]*CrashTerm, error) {
	// 空参数直接返回
	if len(termIds) == 0 {
		return map[int64]*CrashTerm{}, nil
	}

	// 去重（可选）
	uniq := make([]int64, 0, len(termIds))
	seen := make(map[int64]struct{}, len(termIds))
	for _, id := range termIds {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}

	// placeholder: ?, ?, ?, ...
	placeholders := make([]string, len(uniq))
	args := make([]any, 0, 1+len(uniq))
	args = append(args, channelId)
	for i, id := range uniq {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE `channel_id` = ? AND `term_id` IN (%s)",
		crashTermRows, m.table, strings.Join(placeholders, ","),
	)

	var list []CrashTerm
	if err := m.conn.QueryRowsCtx(ctx, &list, query, args...); err != nil {
		return nil, err
	}

	// 组装成 map[term_id]*CrashTerm
	out := make(map[int64]*CrashTerm, len(list))
	for i := range list {
		ct := list[i] // 取地址要用下标变量
		out[ct.TermId] = &ct
	}
	return out, nil
}

func (m *customCrashTermModel) BatchInsert(ctx context.Context, dataList []*CrashTerm) ([]int64, error) {
	if len(dataList) == 0 {
		return nil, nil
	}

	// 列清单与 vals 的顺序必须严格一致；且绝不包含 `id`
	const cols = "`term_id`,`channel_id`,`multiple`,`is_control`,`is_crashed`,`bouns_pool_start`,`bouns_pool`,`term_hash`,`sha512_seed`,`total_bet_amt`,`fee_amt`,`user_profit_correct`,`rake_amt`,`cashed_amt`,`ctrl_cashed_amt`,`manual_squib`,`manual_squib_state`,`break_payout_rate`,`profit_amt`,`max_cashedout_bet_id`,`max_multiple`,`pre_start_time`,`starting_time`,`flying_time`,`crashed_time`,`ctime`"
	const placeholders = "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)" // 26 个

	vals := make([]any, 0, len(dataList)*26)
	blocks := make([]string, 0, len(dataList))
	for _, d := range dataList {
		blocks = append(blocks, placeholders)
		vals = append(vals,
			d.TermId,
			d.ChannelId,
			d.Multiple,
			d.IsControl,
			d.IsCrashed,
			d.BounsPoolStart,
			d.BounsPool,
			d.TermHash,
			d.Sha512Seed,
			d.TotalBetAmt,
			d.FeeAmt,
			d.UserProfitCorrect,
			d.RakeAmt,
			d.CashedAmt,
			d.CtrlCashedAmt,
			d.ManualSquib,
			d.ManualSquibState,
			d.BreakPayoutRate,
			d.ProfitAmt,
			d.MaxCashedoutBetId,
			d.MaxMultiple,
			d.PreStartTime,
			d.StartingTime,
			d.FlyingTime,
			d.CrashedTime,
			d.Ctime,
		)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", m.table, cols, strings.Join(blocks, ","))
	res, err := m.conn.ExecCtx(ctx, query, vals...)
	if err != nil {
		return nil, err
	}

	firstID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get LastInsertId failed: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get RowsAffected failed: %w", err)
	}

	// 注：单条 INSERT 多值时，MySQL 会为本语句连续分配自增区间；并发/失败行可能造成空洞，但一般仍是连续段
	ids := make([]int64, rows)
	for i := int64(0); i < rows; i++ {
		ids[i] = firstID + i
	}
	return ids, nil
}

func (m *customCrashTermModel) BatchInsertIgnoreDupReturnIDs(ctx context.Context, dataList []*CrashTerm) ([]int64, error) {
	if len(dataList) == 0 {
		return nil, nil
	}

	const cols = "`term_id`,`channel_id`,`multiple`,`is_control`,`is_crashed`,`bouns_pool_start`,`bouns_pool`,`term_hash`,`sha512_seed`,`total_bet_amt`,`fee_amt`,`user_profit_correct`,`rake_amt`,`cashed_amt`,`ctrl_cashed_amt`,`manual_squib`,`manual_squib_state`,`break_payout_rate`,`profit_amt`,`max_cashedout_bet_id`,`max_multiple`,`pre_start_time`,`starting_time`,`flying_time`,`crashed_time`,`ctime`"
	const placeholders = "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	vals := make([]any, 0, len(dataList)*26)
	blocks := make([]string, 0, len(dataList))
	for _, d := range dataList {
		blocks = append(blocks, placeholders)
		vals = append(vals,
			d.TermId,
			d.ChannelId,
			d.Multiple,
			d.IsControl,
			d.IsCrashed,
			d.BounsPoolStart,
			d.BounsPool,
			d.TermHash,
			d.Sha512Seed,
			d.TotalBetAmt,
			d.FeeAmt,
			d.UserProfitCorrect,
			d.RakeAmt,
			d.CashedAmt,
			d.CtrlCashedAmt,
			d.ManualSquib,
			d.ManualSquibState,
			d.BreakPayoutRate,
			d.ProfitAmt,
			d.MaxCashedoutBetId,
			d.MaxMultiple,
			d.PreStartTime,
			d.StartingTime,
			d.FlyingTime,
			d.CrashedTime,
			d.Ctime,
		)
	}

	// 1. IGNORE 重复插入
	insertSQL := fmt.Sprintf("INSERT IGNORE INTO %s (%s) VALUES %s", m.table, cols, strings.Join(blocks, ","))
	if _, err := m.conn.ExecCtx(ctx, insertSQL, vals...); err != nil {
		return nil, fmt.Errorf("insert ignore failed: %w", err)
	}

	// 2. 查询所有 ID
	tupleBlocks := make([]string, 0, len(dataList))
	args := make([]any, 0, len(dataList)*2)
	for _, d := range dataList {
		tupleBlocks = append(tupleBlocks, "(?,?)")
		args = append(args, d.TermId, d.ChannelId)
	}
	selectSQL := fmt.Sprintf(
		"SELECT id, term_id, channel_id FROM %s WHERE (term_id, channel_id) IN (%s)",
		m.table, strings.Join(tupleBlocks, ","),
	)

	// 用 QueryRowsCtx 查询所有结果
	var rows []struct {
		Id        int64 `db:"id"`
		TermId    int64 `db:"term_id"`
		ChannelId int64 `db:"channel_id"`
	}
	if err := m.conn.QueryRowsCtx(ctx, &rows, selectSQL, args...); err != nil {
		return nil, fmt.Errorf("select ids failed: %w", err)
	}

	// 3. 构建映射表
	idMap := make(map[[2]int64]int64, len(rows))
	for _, r := range rows {
		idMap[[2]int64{r.TermId, r.ChannelId}] = r.Id
	}

	// 4. 按输入顺序返回
	ids := make([]int64, len(dataList))
	for i, d := range dataList {
		key := [2]int64{d.TermId, d.ChannelId}
		if id, ok := idMap[key]; ok {
			ids[i] = id
		} else {
			return nil, fmt.Errorf("missing id for term_id=%d channel_id=%d", d.TermId, d.ChannelId)
		}
	}
	return ids, nil
}

// 获取上一期期数ID
func (m *customCrashTermModel) GetPreviousIssueId(ctx context.Context, channelId, termId int64) (int, error) {
	if channelId <= 0 {
		err := errors.New("empty channel id")
		return 0, err
	}
	sqlStr, sqlParams, err := sq.Select(GameTermId_F_term_id).From(m.table).Where(sq.Eq{
		GameTermId_F_channel_id: channelId,
	}).Where(sq.Lt{GameTermId_F_term_id: termId}).OrderBy(GameTermId_F_term_id + " DESC").Limit(1).ToSql()
	if err != nil {
		return 0, err
	}
	resp := 0
	err = m.conn.QueryRowCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == ErrNotFound {
		err = nil
	}
	return resp, err
}

// 根据id，修改数据
func (m *customCrashTermModel) UpdateById(ctx context.Context, id int64, clauses map[string]interface{}) error {
	sqlStr, sqlParams, err := sq.Update(m.table).SetMap(clauses).Where("id=?", id).ToSql()
	if err != nil {
		return err
	}
	_, err = m.conn.ExecCtx(ctx, sqlStr, sqlParams...)
	return err
}

// UpdateByIds 根据多个id批量更新
func (m *customCrashTermModel) UpdateByIds(ctx context.Context, ids []int64, clauses map[string]interface{}) error {
	if len(ids) == 0 {
		return nil // 没有id直接返回
	}

	sqlStr, sqlParams, err := sq.
		Update(m.table).
		SetMap(clauses).
		Where(sq.Eq{"id": ids}). // id IN (...)
		ToSql()
	if err != nil {
		return err
	}

	_, err = m.conn.ExecCtx(ctx, sqlStr, sqlParams...)
	return err
}

// GetNextTermNo 按渠道获取下一期的期号
func (m *customCrashTermModel) GetNextTermNo(ctx context.Context, channelID int64) (int64, error) {
	sqlStr, sqlArgs, _ := sq.
		Select("COALESCE(MAX(term_id), 0) + 1 AS next_no").
		From(m.table).
		Where(sq.Eq{CrashTerm_F_channel_id: channelID}).
		ToSql()

	var nextNo int64
	err := m.conn.QueryRowCtx(ctx, &nextNo, sqlStr, sqlArgs...)
	if err != nil && (err == sqlx.ErrNotFound || err == sql.ErrNoRows) {
		return 1, nil
	}
	return nextNo, err
}

// GetLatestTerm 按渠道获取最近一期游戏
func (m *customCrashTermModel) GetLatestTerm(ctx context.Context, channelID int64) (*CrashTerm, error) {
	sqlStr, sqlArgs, _ := sq.Select(crashTermFieldNames...).
		From(m.table).
		Where(sq.Eq{
			CrashTerm_F_channel_id: channelID,
			CrashTerm_F_is_crashed: CrashTermIsCrashedNo,
		}).OrderBy("term_id desc").Limit(1).ToSql()
	var crashTerm CrashTerm
	err := m.conn.QueryRowCtx(ctx, &crashTerm, sqlStr, sqlArgs...)
	if err != nil && (err == sqlx.ErrNotFound || err == sql.ErrNoRows) {
		err = nil
	}
	return &crashTerm, err
}

func (m *customCrashTermModel) GetAllStatusLatestTerm(ctx context.Context, channelID int64) (*CrashTerm, error) {
	sqlStr, sqlArgs, err := sq.Select(crashTermFieldNames...).
		From(m.table).
		Where(sq.Eq{
			CrashTerm_F_channel_id: channelID,
		}).OrderBy("term_id desc").Limit(1).ToSql()
	if err != nil {
		return nil, err
	}

	var crashTerm CrashTerm
	err = m.conn.QueryRowCtx(ctx, &crashTerm, sqlStr, sqlArgs...)
	if err != nil && (err == sqlx.ErrNotFound || err == sql.ErrNoRows) {
		err = nil
	}
	return &crashTerm, err
}

// CloseTerm 直接结束该期游戏
func (m *customCrashTermModel) CloseTerm(ctx context.Context, termID int64) error {
	sqlStr, sqlArgs, _ := sq.Update(m.table).Set("is_crashed", CrashTermIsCrashedYes).Where("term_id=?", termID).ToSql()
	_, err := m.conn.ExecCtx(ctx, sqlStr, sqlArgs...)
	return err
}

// 根据id 获取一条数据
func (m *customCrashTermModel) GetByTermId(ctx context.Context, channelID, termID int64) (*CrashTerm, error) {
	sqlStr, sqlArgs, _ := sq.Select(crashTermFieldNames...).From(m.table).Where(sq.Eq{
		CrashTerm_F_channel_id: channelID,
		CrashTerm_F_term_id:    termID,
	}).Limit(1).ToSql()
	var crashTerm CrashTerm
	err := m.conn.QueryRowCtx(ctx, &crashTerm, sqlStr, sqlArgs...)
	if err == sqlx.ErrNotFound || err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	return &crashTerm, nil
}

// clauses := map[string]interface{}{"multiple": term.Multiple, "is_crashed": term.IsCrashed, "term_hash": term.TermHash}
// CloseAndRake 关闭该期游戏，并抽水
func (m *customCrashTermModel) CloseAndRake(ctx context.Context, term *CrashTerm, rake int64, ctrlProfit int64, totalBetAmt, cashedAmt int64) error {
	clauses := map[string]interface{}{
		"multiple":      term.Multiple,
		"is_crashed":    term.IsCrashed,
		"term_hash":     term.TermHash,
		"total_bet_amt": totalBetAmt,
		"cashed_amt":    cashedAmt,
	}
	sqlTermStr, sqlTermArgs, _ := sq.Update(m.table).SetMap(clauses).Where("id=?", term.Id).ToSql()
	sqlChanStr, sqlChanArgs, _ := sq.Update("channel").SetMap(map[string]interface{}{
		Channel_F_total_profit:     sq.Expr(Channel_F_total_profit+"+ ?", rake),
		Channel_F_ctrl_profit:      sq.Expr(Channel_F_ctrl_profit+"+ ?", ctrlProfit),
		Channel_F_total_bet_amt:    sq.Expr(Channel_F_total_bet_amt+" + ?", totalBetAmt),
		Channel_F_total_cashed_amt: sq.Expr(Channel_F_total_cashed_amt+" + ?", cashedAmt),
	}).Where("id=?", term.ChannelId).ToSql()
	return m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		if _, err := session.ExecCtx(ctx, sqlTermStr, sqlTermArgs...); err != nil {
			return err
		}
		_, err := session.ExecCtx(ctx, sqlChanStr, sqlChanArgs...)
		return err
	})
}

// 获取最近的多局游戏信息
func (m *customCrashTermModel) GetLatestTerms(ctx context.Context, channelID int64, pageSize int64) (resp []*CrashTerm, err error) {
	resp = make([]*CrashTerm, 0)
	if channelID <= 0 || pageSize <= 0 {
		return
	}

	sqlStr, sqlArgs, _ := sq.Select(crashTermFieldNames...).From(m.table).Where(sq.Eq{
		CrashTerm_F_channel_id: channelID,
		CrashTerm_F_is_crashed: CrashTermIsCrashedYes,
	}).OrderBy(CrashTerm_F_term_id + " desc").Limit(uint64(pageSize)).ToSql()

	err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlArgs...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	return
}

// 获取id小于posid 的一页已完成的游戏数据
func (m *customCrashTermModel) GetPageById(ctx context.Context, channelID int64, posId int64, pageSize int64) (resp []*CrashTerm, err error) {
	resp = make([]*CrashTerm, 0)
	if channelID <= 0 || pageSize <= 0 || posId < 0 {
		return
	}

	sqb := sq.Select(crashTermFieldNames...).From(m.table).Where(sq.Eq{
		CrashTerm_F_channel_id: channelID,
		CrashTerm_F_is_crashed: CrashTermIsCrashedYes,
	})
	if posId > 0 {
		sqb = sqb.Where(sq.Lt{
			CrashTerm_F_term_id: posId,
		})
	}

	sqlStr, sqlArgs, _ := sqb.OrderBy(CrashTerm_F_term_id + " desc").Limit(uint64(pageSize)).ToSql()
	err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlArgs...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	return
}

// 根据hash 获取一条数据
func (m *customCrashTermModel) GetByHash(ctx context.Context, hash string) (*CrashTerm, error) {
	sqlStr, sqlArgs, _ := sq.Select(crashTermFieldNames...).From(m.table).Where(sq.Eq{
		CrashTerm_F_term_hash: hash,
	}).Limit(1).ToSql()
	var crashTerm CrashTerm
	err := m.conn.QueryRowCtx(ctx, &crashTerm, sqlStr, sqlArgs...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	return &crashTerm, err
}

// 根据ids获取多条数据
func (m *customCrashTermModel) GetByIds(ctx context.Context, termIds []int64) (resp []*CrashTerm, err error) {
	resp = make([]*CrashTerm, 0)
	if len(termIds) == 0 {
		return
	}
	sqlStr, sqlArgs, _ := sq.Select(crashTermFieldNames...).From(m.table).Where(sq.Eq{
		CrashTerm_F_id: termIds,
	}).ToSql()
	err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlArgs...)
	if err == sql.ErrNoRows || err == sqlx.ErrNotFound {
		err = nil
	}
	return
}

// 获取全部渠道的一页数据
func (m *customCrashTermModel) GetAllPageById(ctx context.Context, posId int64, pageSize int64, cutoff int64) (resp []*CrashTerm, err error) {
	resp = make([]*CrashTerm, 0)
	if posId < 0 || pageSize <= 0 {
		return
	}

	sqlStr, sqlArgs, _ := sq.
		Select(crashTermFieldNames...).
		From(m.table).
		Where(
			sq.Gt{"id": posId},
		).
		Where(
			sq.Eq{"is_crashed": 2},
		).
		Where(
			sq.LtOrEq{"crashed_time": cutoff},
		).
		OrderBy("id ASC").
		Limit(uint64(pageSize)).
		ToSql()

	err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlArgs...)
	if err != nil && (err == sqlx.ErrNotFound || err == sql.ErrNoRows) {
		err = nil
	}
	return
}
