package tasks

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"sync"
	"time"

	_ "modernc.org/sqlite" // SQLite driver
)

type TaskManagerImpl struct {
	db *sql.DB
	mu sync.RWMutex
}

var _ domain.DecompTaskManager = (*TaskManagerImpl)(nil)

func NewTaskManager(cacheDir string) (*TaskManagerImpl, error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(cacheDir, "tasks.db")+"?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}
	tm := &TaskManagerImpl{db: db}
	if _, err := tm.db.Exec(`CREATE TABLE IF NOT EXISTS task_plans (id TEXT PRIMARY KEY, title TEXT, description TEXT, status TEXT, created_at INTEGER)`); err != nil {
		return nil, err
	}
	if _, err := tm.db.Exec(`CREATE TABLE IF NOT EXISTS tasks (id TEXT PRIMARY KEY, plan_id TEXT, parent_id TEXT, title TEXT, description TEXT, status TEXT, task_order INTEGER, dependencies TEXT, files TEXT, result TEXT, error TEXT, created_at INTEGER, started_at INTEGER, completed_at INTEGER, metadata TEXT)`); err != nil {
		return nil, err
	}
	return tm, nil
}

func (tm *TaskManagerImpl) CreatePlan(title, desc string) (*domain.DecompTaskPlan, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	p := &domain.DecompTaskPlan{ID: genID(), Title: title, Description: desc, Tasks: []*domain.DecompTask{}, CreatedAt: time.Now(), Status: domain.DecompTaskPending}
	if _, err := tm.db.Exec(`INSERT INTO task_plans VALUES (?,?,?,?,?)`, p.ID, p.Title, p.Description, p.Status, p.CreatedAt.Unix()); err != nil {
		return nil, err
	}
	return p, nil
}

func (tm *TaskManagerImpl) AddTask(planID string, t *domain.DecompTask) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if t.ID == "" {
		t.ID = genID()
	}
	t.CreatedAt, t.Status = time.Now(), domain.DecompTaskPending
	d, _ := json.Marshal(t.Dependencies)
	f, _ := json.Marshal(t.Files)
	m, _ := json.Marshal(t.Metadata)
	_, err := tm.db.Exec(`INSERT INTO tasks VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, t.ID, planID, t.ParentID, t.Title, t.Description, t.Status, t.Order, string(d), string(f), t.Result, t.Error, t.CreatedAt.Unix(), nil, nil, string(m))
	return err
}

func (tm *TaskManagerImpl) GetPlan(id string) (*domain.DecompTaskPlan, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	p, err := tm.loadPlanHeader(id)
	if err != nil {
		return nil, err
	}
	p.Tasks, err = tm.loadPlanTasks(id)
	return p, err
}

func (tm *TaskManagerImpl) loadPlanHeader(id string) (*domain.DecompTaskPlan, error) {
	var p domain.DecompTaskPlan
	var ca int64
	err := tm.db.QueryRow(`SELECT id,title,description,status,created_at FROM task_plans WHERE id=?`, id).Scan(&p.ID, &p.Title, &p.Description, &p.Status, &ca)
	if err != nil {
		return nil, err
	}
	p.CreatedAt = time.Unix(ca, 0)
	return &p, nil
}

func (tm *TaskManagerImpl) loadPlanTasks(planID string) ([]*domain.DecompTask, error) {
	rows, err := tm.db.Query(`SELECT id,parent_id,title,description,status,task_order,dependencies,files,result,error,created_at,started_at,completed_at,metadata FROM tasks WHERE plan_id=? ORDER BY task_order`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.DecompTask
	for rows.Next() {
		if t := tm.scanTask(rows); t != nil {
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}

func (tm *TaskManagerImpl) scanTask(rows *sql.Rows) *domain.DecompTask {
	var t domain.DecompTask
	var pid, deps, files, res, er, meta sql.NullString
	var tca, sa, coa sql.NullInt64
	if err := rows.Scan(&t.ID, &pid, &t.Title, &t.Description, &t.Status, &t.Order, &deps, &files, &res, &er, &tca, &sa, &coa, &meta); err != nil {
		return nil
	}
	tm.populateTaskFields(&t, pid, deps, files, res, er, meta, tca, sa, coa)
	return &t
}

func (tm *TaskManagerImpl) populateTaskFields(t *domain.DecompTask, pid, deps, files, res, er, meta sql.NullString, tca, sa, coa sql.NullInt64) {
	if pid.Valid {
		t.ParentID = pid.String
	}
	if deps.Valid {
		_ = json.Unmarshal([]byte(deps.String), &t.Dependencies)
	}
	if files.Valid {
		_ = json.Unmarshal([]byte(files.String), &t.Files)
	}
	if res.Valid {
		t.Result = res.String
	}
	if er.Valid {
		t.Error = er.String
	}
	if meta.Valid {
		_ = json.Unmarshal([]byte(meta.String), &t.Metadata)
	}
	if tca.Valid {
		t.CreatedAt = time.Unix(tca.Int64, 0)
	}
	if sa.Valid {
		tt := time.Unix(sa.Int64, 0)
		t.StartedAt = &tt
	}
	if coa.Valid {
		tt := time.Unix(coa.Int64, 0)
		t.CompletedAt = &tt
	}
}

func (tm *TaskManagerImpl) UpdateTaskStatus(id string, s domain.DecompTaskStatus, res, er string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	now := time.Now().Unix()
	var sa, ca interface{}
	if s == domain.DecompTaskInProgress {
		sa = now
	}
	if s == domain.DecompTaskCompleted || s == domain.DecompTaskFailed {
		ca = now
	}
	_, err := tm.db.Exec(`UPDATE tasks SET status=?,result=?,error=?,started_at=COALESCE(?,started_at),completed_at=? WHERE id=?`, s, res, er, sa, ca, id)
	return err
}

func (tm *TaskManagerImpl) GetNextTask(planID string) (*domain.DecompTask, error) {
	p, err := tm.GetPlan(planID)
	if err != nil {
		return nil, err
	}
	done := map[string]bool{}
	for _, t := range p.Tasks {
		if t.Status == domain.DecompTaskCompleted {
			done[t.ID] = true
		}
	}
	for _, t := range p.Tasks {
		if t.Status != domain.DecompTaskPending {
			continue
		}
		ok := true
		for _, d := range t.Dependencies {
			if !done[d] {
				ok = false
				break
			}
		}
		if ok {
			return t, nil
		}
	}
	return nil, nil
}

func (tm *TaskManagerImpl) GetPlanProgress(id string) (int, int, error) {
	p, err := tm.GetPlan(id)
	if err != nil {
		return 0, 0, err
	}
	c := 0
	for _, t := range p.Tasks {
		if t.Status == domain.DecompTaskCompleted {
			c++
		}
	}
	return c, len(p.Tasks), nil
}

func (tm *TaskManagerImpl) Close() error {
	if tm.db != nil {
		return tm.db.Close()
	}
	return nil
}

func genID() string {
	return time.Now().Format("20060102150405") + rndSfx()
}

func rndSfx() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	// Use crypto/rand for better randomness
	randBytes := make([]byte, 6)
	if _, err := rand.Read(randBytes); err != nil {
		// Fallback to time-based if crypto/rand fails
		for i := range b {
			b[i] = charset[(time.Now().UnixNano()+int64(i))%36]
		}
		return string(b)
	}
	for i := range b {
		b[i] = charset[randBytes[i]%36]
	}
	return string(b)
}
