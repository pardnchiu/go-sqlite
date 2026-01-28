> [!NOTE]
> 此 README 由 [Claude Code](https://github.com/anthropics/claude-code) 生成，英文版請參閱 [這裡](./README.md)。

![cover](./cover.png)

# go-sqlite

[![pkg](https://pkg.go.dev/badge/github.com/pardnchiu/go-sqlite.svg)](https://pkg.go.dev/github.com/pardnchiu/go-sqlite)
[![tag](https://img.shields.io/github/v/tag/pardnchiu/go-sqlite?label=release)](https://github.com/pardnchiu/go-sqlite/releases)
[![license](https://img.shields.io/github/license/pardnchiu/go-sqlite)](LICENSE)

> 輕量級 Go SQLite 查詢建構器，提供鏈式 API 簡化資料庫 CRUD 操作，支援 Context 傳遞與衝突處理策略。

## 目錄

- [功能特點](#功能特點)
- [安裝](#安裝)
- [使用方法](#使用方法)
- [API 參考](#api-參考)
- [授權](#授權)
- [Author](#author)
- [Stars](#stars)

## 功能特點

- **鏈式 API**：流暢的查詢建構語法，支援 `Table().Where().Get()` 串接
- **多資料庫連線管理**：單例 Connector 統一管理多個 SQLite 資料庫
- **完整 CRUD 支援**：Create、Insert、Select、Update 全方位操作
- **Context 傳遞**：透過 `Context(ctx)` 鏈式呼叫支援 timeout 與 cancellation
- **衝突處理策略**：支援 Ignore、Replace、Abort、Fail、Rollback 五種策略
- **豐富的 WHERE 條件**：Eq、NotEq、Gt、Lt、Ge、Le、In、NotIn、Null、NotNull、Between
- **欄位驗證**：自動檢查 SQL 保留字與欄位名稱格式
- **WAL 模式**：預設啟用 WAL 提升併發效能

## 安裝

```bash
go get github.com/pardnchiu/go-sqlite
```

## 使用方法

### 初始化連線

```go
package main

import (
    "log"
    goSqlite "github.com/pardnchiu/go-sqlite"
)

func main() {
    // 建立連線（單例模式，可多次呼叫註冊不同資料庫）
    conn, err := goSqlite.New(goSqlite.Config{
        Key:      "main",           // 選填，預設使用檔名
        Path:     "./data.db",
        Lifetime: 3600,             // 選填，連線存活時間（秒）
    })
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
}
```

### 建立資料表

```go
builder, _ := conn.DB("main")
err := builder.Table("users").Create(
    goSqlite.Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
    goSqlite.Column{Name: "name", Type: "TEXT", IsNullable: false},
    goSqlite.Column{Name: "email", Type: "TEXT", IsUnique: true},
    goSqlite.Column{Name: "age", Type: "INTEGER", Default: 0},
)
```

### 新增資料

```go
builder, _ := conn.DB("main")

// 基本新增，回傳 LastInsertId
id, err := builder.Table("users").Insert(map[string]any{
    "name":  "Alice",
    "email": "alice@example.com",
    "age":   25,
})

// 使用 Context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

id, err = builder.Table("users").
    Context(ctx).
    Insert(map[string]any{
        "name": "Bob",
    })

// 衝突處理
id, err = builder.Table("users").
    Conflict(goSqlite.Ignore).
    Insert(map[string]any{"email": "alice@example.com"})
```

### 查詢資料

```go
builder, _ := conn.DB("main")

// 查詢多筆
rows, err := builder.Table("users").
    Select("id", "name", "email").
    WhereGt("age", 18).
    OrderBy("name", goSqlite.Asc).
    Limit(10).
    Offset(0).
    Get()

// 查詢單筆
row, err := builder.Table("users").
    WhereEq("id", 1).
    First()

// 計數
count, err := builder.Table("users").
    WhereNotNull("email").
    Count()

// 帶總數查詢（分頁用）
rows, err := builder.Table("users").
    Total().
    Limit(10).
    Get()
```

### 更新資料

```go
builder, _ := conn.DB("main")

// 基本更新，回傳 RowsAffected
affected, err := builder.Table("users").
    WhereEq("id", 1).
    Update(map[string]any{"name": "Alice Updated"})

// 數值增減
_, err = builder.Table("users").
    WhereEq("id", 1).
    Increase("login_count", 1).
    Update()

// 布林切換
_, err = builder.Table("users").
    WhereEq("id", 1).
    Toggle("is_active").
    Update()
```

### 複合條件

```go
builder, _ := conn.DB("main")

// AND 條件
rows, _ := builder.Table("users").
    WhereGe("age", 18).
    WhereLe("age", 30).
    WhereNotNull("email").
    Get()

// OR 條件
rows, _ := builder.Table("users").
    WhereEq("status", "active").
    OrWhereEq("role", "admin").
    Get()

// IN 條件
rows, _ := builder.Table("users").
    WhereIn("id", []any{1, 2, 3}).
    Get()

// BETWEEN 條件
rows, _ := builder.Table("users").
    WhereBetween("age", 20, 30).
    Get()
```

### JOIN 查詢

```go
builder, _ := conn.DB("main")

rows, _ := builder.Table("orders").
    Select("orders.id", "users.name", "orders.total").
    LeftJoin("users", "users.id = orders.user_id").
    WhereGt("orders.total", 100).
    Get()
```

### 原生 SQL

```go
// 透過 Connector 執行
rows, err := conn.Query("main", "SELECT * FROM users WHERE age > ?", 18)

// 或取得 *sql.DB 執行
builder, _ := conn.DB("main")
db := builder.Raw()
rows, err := db.Query("SELECT * FROM users")
```

## API 參考

### Config

| 欄位 | 類型 | 說明 |
|------|------|------|
| `Key` | `string` | 資料庫識別鍵，選填，預設使用檔名 |
| `Path` | `string` | SQLite 檔案路徑 |
| `Lifetime` | `int` | 連線最大存活時間（秒），選填 |

### Connector Methods

| 方法 | 說明 |
|------|------|
| `New(Config) (*Connector, error)` | 建立或取得 Connector 單例 |
| `DB(key string) (*Builder, error)` | 取得指定資料庫的 Builder |
| `Query(key, query string, args ...any)` | 執行原生查詢 |
| `QueryContext(ctx, key, query string, args ...any)` | 帶 Context 執行原生查詢 |
| `Exec(key, query string, args ...any)` | 執行原生命令 |
| `ExecContext(ctx, key, query string, args ...any)` | 帶 Context 執行原生命令 |
| `Close()` | 關閉所有資料庫連線 |

### Builder Methods

| 分類 | 方法 | 說明 |
|------|------|------|
| **基礎** | `Table(name string)` | 指定資料表 |
| | `Context(ctx context.Context)` | 設定 Context |
| | `Raw() *sql.DB` | 取得底層 *sql.DB |
| **建表** | `Create(columns ...Column)` | 建立資料表 |
| **新增** | `Insert(data ...map[string]any) (int64, error)` | 新增資料，回傳 LastInsertId |
| | `Conflict(conflict) *Builder` | 設定衝突處理策略 |
| **查詢** | `Select(columns ...string)` | 指定查詢欄位 |
| | `Get() (*sql.Rows, error)` | 取得多筆結果 |
| | `First() (*sql.Row, error)` | 取得單筆結果 |
| | `Count() (int64, error)` | 取得筆數 |
| | `Total() *Builder` | 啟用總數計算 |
| **條件** | `Where(condition string, args ...any)` | 自訂 WHERE 條件 |
| | `WhereEq / WhereNotEq` | 等於 / 不等於 |
| | `WhereGt / WhereLt` | 大於 / 小於 |
| | `WhereGe / WhereLe` | 大於等於 / 小於等於 |
| | `WhereIn / WhereNotIn` | IN / NOT IN |
| | `WhereNull / WhereNotNull` | IS NULL / IS NOT NULL |
| | `WhereBetween` | BETWEEN |
| | `OrWhere...` | OR 版本的條件方法 |
| **排序分頁** | `OrderBy(column string, direction ...direction)` | 排序（Asc/Desc） |
| | `Limit(num ...int)` | 限制筆數 |
| | `Offset(num int)` | 偏移量 |
| **JOIN** | `Join(table, on string)` | INNER JOIN |
| | `LeftJoin(table, on string)` | LEFT JOIN |
| **更新** | `Update(data ...map[string]any) (int64, error)` | 更新資料，回傳 RowsAffected |
| | `Increase(column string, num ...int)` | 數值遞增 |
| | `Decrease(column string, num ...int)` | 數值遞減 |
| | `Toggle(column string)` | 布林切換 |

### Conflict 策略

| 常數 | 說明 |
|------|------|
| `Ignore` | 忽略衝突，不執行插入 |
| `Replace` | 刪除舊資料並插入新資料 |
| `Abort` | 中止交易並回滾（預設） |
| `Fail` | 中止但保留已執行的變更 |
| `Rollback` | 回滾整個交易 |

### Column 結構

| 欄位 | 類型 | 說明 |
|------|------|------|
| `Name` | `string` | 欄位名稱 |
| `Type` | `string` | SQLite 類型（INTEGER、TEXT、REAL、BLOB） |
| `IsPrimary` | `bool` | 是否為主鍵 |
| `AutoIncrease` | `bool` | 是否自動遞增 |
| `IsNullable` | `bool` | 是否可為 NULL |
| `IsUnique` | `bool` | 是否唯一 |
| `Default` | `any` | 預設值 |
| `ForeignKey` | `*Foreign` | 外鍵設定 |

## 授權

本專案採用 MIT 授權條款 - 詳見 [LICENSE](LICENSE) 檔案。

## Author

<img src="https://avatars.githubusercontent.com/u/25631760" align="left" width="96" height="96" style="margin-right: 0.5rem;">

<h4 style="padding-top: 0">邱敬幃 Pardn Chiu</h4>

<a href="mailto:dev@pardn.io" target="_blank">
<img src="https://pardn.io/image/email.svg" width="48" height="48">
</a> <a href="https://linkedin.com/in/pardnchiu" target="_blank">
<img src="https://pardn.io/image/linkedin.svg" width="48" height="48">
</a>

## Stars

[![Star](https://starchart.cc/pardnchiu/go-sqlite.svg?variant=adaptive)](https://starchart.cc/pardnchiu/go-sqlite)

***

©️ 2026 [邱敬幃 Pardn Chiu](https://linkedin.com/in/pardnchiu)
