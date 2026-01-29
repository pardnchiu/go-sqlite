> [!NOTE]
> 此 README 由 [Claude Code](https://github.com/pardnchiu/skill-readme-generate) 生成，英文版請參閱 [這裡](./README.md)。

![cover](cover.png)

# go-sqlite

[![pkg](https://pkg.go.dev/badge/github.com/pardnchiu/go-sqlite.svg)](https://pkg.go.dev/github.com/pardnchiu/go-sqlite)
[![card](https://goreportcard.com/badge/github.com/pardnchiu/go-sqlite)](https://goreportcard.com/report/github.com/pardnchiu/go-sqlite)
[![license](https://img.shields.io/github/license/pardnchiu/go-sqlite)](LICENSE)

> 高效能 SQLite 查詢建構器，採用讀寫分離連線池架構，針對高併發場景優化。

## 目錄

- [功能特點](#功能特點)
- [安裝](#安裝)
- [使用方法](#使用方法)
- [API 參考](#api-參考)
- [授權](#授權)
- [Author](#author)
- [Stars](#stars)

## 功能特點

- **讀寫分離架構**：獨立的讀取與寫入連線池，最大化併發效能
- **WAL 模式**：預設啟用 Write-Ahead Logging，提升寫入效能並減少鎖定衝突
- **鏈式查詢建構器**：流暢的 API 設計，支援 SELECT、INSERT、UPDATE、DELETE 操作
- **自動 Struct 綁定**：透過 `Bind()` 方法將查詢結果自動映射至 struct 或 slice
- **SQL Injection 防護**：內建識別符驗證與參數化查詢
- **Context 支援**：所有操作支援 `context.Context` 進行超時與取消控制
- **衝突處理策略**：支援 IGNORE、REPLACE、ABORT、FAIL、ROLLBACK 模式
- **自動 WAL Checkpoint**：背景 goroutine 定期執行 checkpoint

## 安裝

```bash
go get github.com/pardnchiu/go-sqlite
```

## 使用方法

### 初始化連線

```go
package main

import (
    goSqlite "github.com/pardnchiu/go-sqlite"
    "github.com/pardnchiu/go-sqlite/core"
)

func main() {
    conn, err := goSqlite.New(core.Config{
        Path:         "data.db",
        MaxOpenConns: 50,  // 讀取連線池大小
        MaxIdleConns: 25,  // 閒置連線數
        Lifetime:     120, // 連線生命週期（秒）
    })
    if err != nil {
        panic(err)
    }
    defer conn.Close()
}
```

### 建立資料表

```go
err := conn.Write.Table("users").Create(
    core.Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
    core.Column{Name: "name", Type: "TEXT", IsNullable: false},
    core.Column{Name: "email", Type: "TEXT", IsUnique: true},
    core.Column{Name: "created_at", Type: "INTEGER", Default: 0},
)
```

### 插入資料

```go
// 單筆插入
id, err := conn.Write.Table("users").Insert(map[string]any{
    "name":  "Alice",
    "email": "alice@example.com",
})

// 批次插入
affected, err := conn.Write.Table("users").InsertBatch([]map[string]any{
    {"name": "Bob", "email": "bob@example.com"},
    {"name": "Carol", "email": "carol@example.com"},
})

// 衝突處理
id, err := conn.Write.Table("users").
    Conflict(core.Replace).
    Insert(map[string]any{"name": "Alice", "email": "alice@example.com"})
```

### 查詢資料

```go
// 基本查詢
rows, err := conn.Read.Table("users").
    Select("id", "name", "email").
    WhereGt("id", 10).
    OrderBy("created_at", core.Desc).
    Limit(20).
    Get()

// 綁定至 struct
type User struct {
    ID    int64  `db:"id"`
    Name  string `db:"name"`
    Email string `db:"email"`
}

var user User
_, err := conn.Read.Table("users").
    WhereEq("id", 1).
    Bind(&user).
    First()

// 綁定至 slice
var users []User
_, err := conn.Read.Table("users").
    WhereLt("id", 100).
    Bind(&users).
    Get()

// 分頁查詢（含總數）
rows, err := conn.Read.Table("users").
    Total().
    Limit(10).
    Offset(20).
    Get()
```

### 更新資料

```go
// 一般更新
affected, err := conn.Write.Table("users").
    WhereEq("id", 1).
    Update(map[string]any{"name": "Alice Updated"})

// 數值遞增/遞減
affected, err := conn.Write.Table("users").
    WhereEq("id", 1).
    Increase("login_count", 1).
    Update()

// 布林值切換
affected, err := conn.Write.Table("users").
    WhereEq("id", 1).
    Toggle("is_active").
    Update()
```

### 刪除資料

```go
// 條件刪除
affected, err := conn.Write.Table("users").
    WhereEq("id", 1).
    Delete()

// 強制全表刪除
affected, err := conn.Write.Table("users").Delete(true)
```

### JOIN 查詢

```go
rows, err := conn.Read.Table("orders").
    Select("orders.id", "users.name", "orders.amount").
    LeftJoin("users", "orders.user_id = users.id").
    WhereGt("orders.amount", 100).
    Get()
```

### 聚合查詢

```go
// 計數
count, err := conn.Read.Table("users").
    WhereNotNull("email").
    Count()

// GROUP BY + HAVING
rows, err := conn.Read.Table("orders").
    Select("user_id", "SUM(amount) as total").
    GroupBy("user_id").
    HavingGt("total", 1000).
    Get()
```

## API 參考

### 設定

| 欄位 | 類型 | 說明 |
|------|------|------|
| `Path` | `string` | 資料庫檔案路徑 |
| `MaxOpenConns` | `int` | 讀取連線池最大連線數（預設 50） |
| `MaxIdleConns` | `int` | 閒置連線數（預設 25） |
| `Lifetime` | `int` | 連線生命週期秒數（預設 120） |

### Builder 方法

#### 資料表操作

| 方法 | 說明 |
|------|------|
| `Table(name)` | 指定操作的資料表 |
| `Create(columns...)` | 建立資料表 |

#### 查詢建構

| 方法 | 說明 |
|------|------|
| `Select(columns...)` | 指定查詢欄位 |
| `Join(table, on)` | INNER JOIN |
| `LeftJoin(table, on)` | LEFT JOIN |
| `OrderBy(column, direction)` | 排序（`core.Asc` / `core.Desc`） |
| `GroupBy(columns...)` | 分組 |
| `Limit(n)` / `Limit(offset, n)` | 限制筆數 |
| `Offset(n)` | 偏移量 |
| `Total()` | 查詢時包含總筆數 |
| `Context(ctx)` | 設定 context |
| `Bind(target)` | 綁定結果至 struct/slice |

#### WHERE 條件

| 方法 | SQL |
|------|-----|
| `Where(condition, args...)` | 自訂條件 |
| `WhereEq(col, val)` | `col = ?` |
| `WhereNotEq(col, val)` | `col != ?` |
| `WhereGt(col, val)` | `col > ?` |
| `WhereLt(col, val)` | `col < ?` |
| `WhereGe(col, val)` | `col >= ?` |
| `WhereLe(col, val)` | `col <= ?` |
| `WhereIn(col, vals)` | `col IN (?, ...)` |
| `WhereNotIn(col, vals)` | `col NOT IN (?, ...)` |
| `WhereNull(col)` | `col IS NULL` |
| `WhereNotNull(col)` | `col IS NOT NULL` |
| `WhereBetween(col, start, end)` | `col BETWEEN ? AND ?` |
| `OrWhere*(...)` | OR 版本 |

#### HAVING 條件

所有 WHERE 方法皆有對應的 `Having*` 版本。

#### 執行方法

| 方法 | 回傳值 | 說明 |
|------|--------|------|
| `Get()` | `(*sql.Rows, error)` | 執行查詢 |
| `First()` | `(*sql.Row, error)` | 取得第一筆（反向排序） |
| `Last()` | `(*sql.Row, error)` | 取得最後一筆 |
| `Count()` | `(int64, error)` | 計算筆數 |
| `Insert(data, [conflict])` | `(int64, error)` | 插入並回傳 ID |
| `InsertBatch(data)` | `(int64, error)` | 批次插入 |
| `Update([data])` | `(int64, error)` | 更新並回傳影響筆數 |
| `Delete([force])` | `(int64, error)` | 刪除並回傳影響筆數 |

#### 更新輔助

| 方法 | 說明 |
|------|------|
| `Increase(col, [n])` | 數值遞增（預設 +1） |
| `Decrease(col, [n])` | 數值遞減（預設 -1） |
| `Toggle(col)` | 布林值切換 |
| `Conflict(mode)` | 設定衝突處理策略 |

#### 衝突模式

| 常數 | 說明 |
|------|------|
| `core.Ignore` | 忽略衝突 |
| `core.Replace` | 取代既有資料 |
| `core.Abort` | 中止交易 |
| `core.Fail` | 失敗但保留之前變更 |
| `core.Rollback` | 回滾整個交易 |

### Connector 方法

| 方法 | 說明 |
|------|------|
| `Query(key, query, args...)` | 執行原生讀取查詢 |
| `QueryContext(ctx, key, query, args...)` | 含 context 的原生讀取查詢 |
| `Exec(key, query, args...)` | 執行原生寫入操作 |
| `ExecContext(ctx, key, query, args...)` | 含 context 的原生寫入操作 |
| `Close()` | 關閉所有連線 |

## 授權

MIT License

## Author

<img src="https://avatars.githubusercontent.com/u/25631760" align="left" width="96" height="96" style="margin-right: 0.5rem;">

<h4 style="padding-top: 0">邱敬幃 Pardn Chiu</h4>

<a href="mailto:dev@pardn.io" target="_blank">
<img src="https://pardn.io/image/email.svg" width="48" height="48">
</a> <a href="https://linkedin.com/in/pardnchiu" target="_blank">
<img src="https://pardn.io/image/linkedin.svg" width="48" height="48">
</a>

## Stars

[![Star](https://api.star-history.com/svg?repos=pardnchiu/go-sqlite&type=Date)](https://www.star-history.com/#pardnchiu/go-sqlite&Date)

***

©️ 2026 [邱敬幃 Pardn Chiu](https://linkedin.com/in/pardnchiu)
