> [!NOTE]
> 此 README 由 [Claude Code](https://github.com/anthropics/claude-code) 生成，英文版請參閱 [這裡](./README.md)。

![cover](cover.png)

# go-sqlite

[![pkg](https://pkg.go.dev/badge/github.com/pardnchiu/go-sqlite.svg)](https://pkg.go.dev/github.com/pardnchiu/go-sqlite)
[![license](https://img.shields.io/github/license/pardnchiu/go-sqlite)](LICENSE)

> 輕量級 Go SQLite 封裝，提供讀寫分離連線池與鏈式 Query Builder。

## 目錄

- [功能特點](#功能特點)
- [安裝](#安裝)
- [使用方法](#使用方法)
- [API 參考](#api-參考)
- [授權](#授權)
- [Author](#author)
- [Stars](#stars)

## 功能特點

- **讀寫分離**：獨立的讀取與寫入連線池，優化高併發場景
- **WAL 模式**：預設啟用 Write-Ahead Logging，提升寫入效能
- **鏈式 API**：流暢的 Query Builder 語法，支援 SELECT、INSERT、UPDATE、DELETE
- **完整條件支援**：WHERE、OR WHERE、HAVING、OR HAVING 等多樣條件組合
- **JOIN 支援**：INNER JOIN、LEFT JOIN 跨表查詢
- **聚合查詢**：GROUP BY、HAVING、COUNT 等聚合操作
- **衝突處理**：INSERT OR IGNORE/REPLACE/ABORT/FAIL/ROLLBACK
- **批次插入**：`InsertBatch` 支援多筆資料一次寫入
- **Context 支援**：所有操作支援 `context.Context` 傳遞
- **自動 Checkpoint**：背景執行 WAL checkpoint，維持資料庫效能

## 安裝

```bash
go get github.com/pardnchiu/go-sqlite
```

## 使用方法

### 初始化連線

```go
import (
    goSqlite "github.com/pardnchiu/go-sqlite"
    "github.com/pardnchiu/go-sqlite/core"
)

conn, err := goSqlite.New(core.Config{
    Path:         "./data.db",
    MaxOpenConns: 50,  // 讀取連線池上限
    MaxIdleConns: 25,  // 閒置連線上限
    Lifetime:     120, // 連線存活秒數
})
if err != nil {
    log.Fatal(err)
}
defer conn.Close()
```

### 建立資料表

```go
err := conn.Write.Table("users").Create(
    core.Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
    core.Column{Name: "name", Type: "TEXT", IsNullable: false},
    core.Column{Name: "email", Type: "TEXT", IsUnique: true},
    core.Column{Name: "age", Type: "INTEGER", Default: 0},
)
```

### INSERT

```go
// 單筆插入
id, err := conn.Write.Table("users").Insert(map[string]any{
    "name":  "Alice",
    "email": "alice@example.com",
})

// 帶衝突處理
id, err := conn.Write.Table("users").
    Conflict(core.Replace).
    Insert(map[string]any{"name": "Bob"})

// 批次插入
affected, err := conn.Write.Table("users").InsertBatch([]map[string]any{
    {"name": "User1", "email": "u1@example.com"},
    {"name": "User2", "email": "u2@example.com"},
})
```

### SELECT

```go
// 查詢全部
rows, err := conn.Read.Table("users").Get()

// 指定欄位與條件
rows, err := conn.Read.Table("users").
    Select("id", "name", "email").
    WhereEq("status", 1).
    WhereGt("age", 18).
    OrderBy("created_at", core.Desc).
    Limit(10).
    Get()

// 單筆查詢
row, err := conn.Read.Table("users").WhereEq("id", 1).First()

// 計數
count, err := conn.Read.Table("users").WhereEq("status", 1).Count()

// 帶總數的分頁查詢
rows, err := conn.Read.Table("users").
    Total().
    Limit(20, 10). // offset=20, limit=10
    Get()
```

### UPDATE

```go
// 一般更新
affected, err := conn.Write.Table("users").
    WhereEq("id", 1).
    Update(map[string]any{"name": "NewName"})

// 數值遞增/遞減
affected, err := conn.Write.Table("users").
    WhereEq("id", 1).
    Increase("login_count").
    Update()

affected, err := conn.Write.Table("users").
    WhereEq("id", 1).
    Decrease("credits", 10).
    Update()

// 布林值切換
affected, err := conn.Write.Table("users").
    WhereEq("id", 1).
    Toggle("is_active").
    Update()
```

### DELETE

```go
// 條件刪除
affected, err := conn.Write.Table("users").
    WhereEq("status", 0).
    Delete()

// 全表刪除需強制確認
affected, err := conn.Write.Table("users").Delete(true)
```

### JOIN

```go
rows, err := conn.Read.Table("orders").
    Select("orders.id", "users.name", "orders.amount").
    Join("users", "users.id = orders.user_id").
    WhereGt("orders.amount", 100).
    Get()

rows, err := conn.Read.Table("posts").
    LeftJoin("comments", "comments.post_id = posts.id").
    Get()
```

### GROUP BY & HAVING

```go
rows, err := conn.Read.Table("orders").
    Select("user_id", "SUM(amount) as total").
    GroupBy("user_id").
    HavingGt("total", 1000).
    Get()
```

### Context 支援

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

rows, err := conn.Read.Table("users").
    Context(ctx).
    WhereEq("status", 1).
    Get()
```

### 原始 SQL

```go
// 查詢
rows, err := conn.Query("", "SELECT * FROM users WHERE id = ?", 1)

// 執行
result, err := conn.Exec("", "UPDATE users SET name = ? WHERE id = ?", "Bob", 1)

// 直接存取底層 *sql.DB
db := conn.Read.Raw()
```

## API 參考

### Config

| 欄位 | 類型 | 說明 | 預設值 |
|------|------|------|--------|
| `Path` | `string` | 資料庫檔案路徑 | - |
| `MaxOpenConns` | `int` | 讀取連線池最大連線數 | `50` |
| `MaxIdleConns` | `int` | 閒置連線數上限 | `25` |
| `Lifetime` | `int` | 連線存活時間（秒） | `120` |

### Connector

| 方法 | 說明 |
|------|------|
| `Query(key, query, args...)` | 執行原始 SELECT 查詢 |
| `QueryContext(ctx, key, query, args...)` | 帶 Context 的 SELECT |
| `Exec(key, query, args...)` | 執行原始寫入操作 |
| `ExecContext(ctx, key, query, args...)` | 帶 Context 的寫入 |
| `Close()` | 關閉所有連線 |

### Builder - 鏈式方法

| 方法 | 說明 |
|------|------|
| `Table(name)` | 指定資料表 |
| `Select(columns...)` | 指定查詢欄位 |
| `Join(table, on)` | INNER JOIN |
| `LeftJoin(table, on)` | LEFT JOIN |
| `Where(condition, args...)` | WHERE 條件 |
| `WhereEq/WhereNotEq/WhereGt/WhereLt/WhereGe/WhereLe` | 比較條件 |
| `WhereIn/WhereNotIn` | IN 條件 |
| `WhereNull/WhereNotNull` | NULL 判斷 |
| `WhereBetween` | BETWEEN 條件 |
| `OrWhere...` | OR 條件系列 |
| `GroupBy(columns...)` | GROUP BY |
| `Having...` | HAVING 條件系列 |
| `OrHaving...` | OR HAVING 條件系列 |
| `OrderBy(column, direction)` | 排序 |
| `Limit(num...)` | 限制筆數 |
| `Offset(num)` | 偏移量 |
| `Total()` | 啟用總數計算 |
| `Context(ctx)` | 設定 Context |
| `Conflict(mode)` | 設定衝突處理模式 |
| `Increase/Decrease(column, num)` | 數值遞增/遞減 |
| `Toggle(column)` | 布林值切換 |

### Builder - 執行方法

| 方法 | 回傳 | 說明 |
|------|------|------|
| `Create(columns...)` | `error` | 建立資料表 |
| `Insert(data, conflictData?)` | `(int64, error)` | 插入並回傳 LastInsertId |
| `InsertBatch(data)` | `(int64, error)` | 批次插入並回傳影響筆數 |
| `Get()` | `(*sql.Rows, error)` | 查詢多筆 |
| `First()` | `(*sql.Row, error)` | 查詢單筆 |
| `Count()` | `(int64, error)` | 計數 |
| `Update(data?)` | `(int64, error)` | 更新並回傳影響筆數 |
| `Delete(force?)` | `(int64, error)` | 刪除並回傳影響筆數 |
| `Raw()` | `*sql.DB` | 取得底層連線 |

### Conflict 模式

| 常數 | 說明 |
|------|------|
| `core.Ignore` | INSERT OR IGNORE |
| `core.Replace` | INSERT OR REPLACE |
| `core.Abort` | INSERT OR ABORT |
| `core.Fail` | INSERT OR FAIL |
| `core.Rollback` | INSERT OR ROLLBACK |

### Column 定義

| 欄位 | 類型 | 說明 |
|------|------|------|
| `Name` | `string` | 欄位名稱 |
| `Type` | `string` | SQLite 類型 |
| `IsPrimary` | `bool` | 主鍵 |
| `AutoIncrease` | `bool` | 自動遞增 |
| `IsNullable` | `bool` | 允許 NULL |
| `IsUnique` | `bool` | 唯一約束 |
| `Default` | `any` | 預設值 |
| `ForeignKey` | `*Foreign` | 外鍵參照 |

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
