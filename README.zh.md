> [!NOTE]
> 此 README 由 [Claude Code](https://github.com/anthropics/claude-code) 生成，英文版請參閱 [這裡](./README.md)。

![cover](./cover.png)

# go-sqlite

[![pkg](https://pkg.go.dev/badge/github.com/pardnchiu/go-sqlite.svg)](https://pkg.go.dev/github.com/pardnchiu/go-sqlite)
[![license](https://img.shields.io/github/license/pardnchiu/go-sqlite)](LICENSE)

> 輕量級 SQLite ORM，基於 [sqlite3 driver](https://github.com/mattn/go-sqlite3) 與 `database/sql` 建構，提供與 [go-mysql](https://github.com/pardnchiu/go-mysql) 一致的 API 介面。

## 目錄

- [功能特點](#功能特點)
- [安裝](#安裝)
- [使用方法](#使用方法)
- [API 參考](#api-參考)
- [授權](#授權)
- [Author](#author)
- [Stars](#stars)

## 功能特點

- **連線池管理**：支援多資料庫連線池，透過 key 管理不同資料庫實例
- **Builder 模式**：鏈式呼叫建構 SQL 查詢，支援 Select/Where/Join/OrderBy/Limit/Offset
- **完整 CRUD 操作**：
  - `Insert` 系列：支援 Conflict 處理策略（Ignore/Replace/Abort/Fail/Rollback）
  - `Update` 系列：支援 Increase/Decrease/Toggle 原子操作
  - `Select` 系列：支援 First/Count/Total 等擴充查詢
- **Where 條件**：提供 Eq/NotEq/Gt/Lt/Ge/Le/In/NotIn/Null/NotNull/Between 等條件方法
- **Context 支援**：所有主要操作皆提供 Context 版本，支援 Timeout 與 Cancellation
- **自動清除狀態**：每次查詢執行後自動重置 Builder 狀態，避免狀態污染
- **SQL Injection 防護**：內建 Column 名稱驗證與參數化查詢

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
)

func main() {
    // 建立資料庫連線
    db, conn, err := goSqlite.New(goSqlite.Config{
        Key:      "main",           // 連線池 key（可選，預設為檔名）
        Path:     "./data.db",      // 資料庫路徑
        Lifetime: 3600,             // 連線存活時間（秒）
    })
    if err != nil {
        panic(err)
    }
    defer db.Close()
}
```

### 建立資料表

```go
builder, _ := db.DB("main")

err := builder.Table("users").Create(
    goSqlite.Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
    goSqlite.Column{Name: "name", Type: "TEXT", IsNullable: false},
    goSqlite.Column{Name: "email", Type: "TEXT", IsUnique: true},
    goSqlite.Column{Name: "age", Type: "INTEGER", Default: 0},
    goSqlite.Column{Name: "is_active", Type: "INTEGER", Default: 1},
)
```

### 新增資料

```go
// 基本新增
err := builder.Table("users").Insert(map[string]any{
    "name":  "John",
    "email": "john@example.com",
    "age":   25,
})

// 新增並取得 ID
id, err := builder.Table("users").InsertReturningID(map[string]any{
    "name":  "Jane",
    "email": "jane@example.com",
})

// 衝突處理
err = builder.Table("users").InsertConflict(goSqlite.Ignore, map[string]any{
    "name":  "John",
    "email": "john@example.com",
})
```

### 查詢資料

```go
// 查詢多筆
rows, err := builder.Table("users").
    Select("id", "name", "email").
    WhereGt("age", 18).
    OrderBy("name", goSqlite.Asc).
    Limit(10).
    Get()

// 查詢單筆
row, err := builder.Table("users").
    WhereEq("id", 1).
    First()

// 計算筆數
count, err := builder.Table("users").
    WhereEq("is_active", 1).
    Count()

// 含總數的分頁查詢（使用 Window Function）
rows, err := builder.Table("users").
    Total().
    Limit(10).
    Offset(20).
    Get()
```

### 更新資料

```go
// 基本更新
result, err := builder.Table("users").
    WhereEq("id", 1).
    Update(map[string]any{"name": "John Doe"})

// 遞增欄位
result, err = builder.Table("users").
    WhereEq("id", 1).
    Increase("login_count", 1).
    Update()

// 切換布林值
result, err = builder.Table("users").
    WhereEq("id", 1).
    Toggle("is_active").
    Update()
```

### JOIN 查詢

```go
rows, err := builder.Table("orders").
    Select("orders.id", "users.name", "orders.total").
    Join("users", "users.id = orders.user_id").
    WhereGt("orders.total", 100).
    Get()

// LEFT JOIN
rows, err = builder.Table("users").
    LeftJoin("orders", "orders.user_id = users.id").
    Get()
```

### Context 支援

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// 查詢
rows, err := builder.Table("users").GetContext(ctx)

// 新增
err = builder.Table("users").InsertContext(ctx, data)

// 更新
result, err := builder.Table("users").
    WhereEq("id", 1).
    UpdateContext(ctx, data)
```

## API 參考

### Config

| 欄位 | 型別 | 說明 |
|------|------|------|
| `Key` | `string` | 連線池識別碼，預設為檔名 |
| `Path` | `string` | SQLite 資料庫檔案路徑 |
| `Lifetime` | `int` | 連線最大存活時間（秒） |

### Column

| 欄位 | 型別 | 說明 |
|------|------|------|
| `Name` | `string` | 欄位名稱 |
| `Type` | `string` | SQLite 資料型別 |
| `IsPrimary` | `bool` | 是否為主鍵 |
| `IsNullable` | `bool` | 是否允許 NULL |
| `AutoIncrease` | `bool` | 是否自動遞增 |
| `IsUnique` | `bool` | 是否唯一 |
| `Default` | `any` | 預設值 |
| `ForeignKey` | `*Foreign` | 外鍵設定 |

### Conflict 策略

| 常數 | 說明 |
|------|------|
| `Ignore` | 忽略衝突，不執行插入 |
| `Replace` | 取代既有資料 |
| `Abort` | 中止交易（預設） |
| `Fail` | 失敗但保留先前變更 |
| `Rollback` | 回滾整個交易 |

### Builder 方法

| 方法 | 說明 |
|------|------|
| `Table(name)` | 設定目標資料表 |
| `Create(columns...)` | 建立資料表 |
| `Select(columns...)` | 設定查詢欄位 |
| `Where(condition, args...)` | 新增 WHERE 條件（AND） |
| `OrWhere(condition, args...)` | 新增 WHERE 條件（OR） |
| `WhereEq/WhereNotEq/WhereGt/WhereLt/WhereGe/WhereLe` | 比較條件 |
| `WhereIn/WhereNotIn` | IN 條件 |
| `WhereNull/WhereNotNull` | NULL 判斷 |
| `WhereBetween` | BETWEEN 範圍 |
| `Join(table, on)` | INNER JOIN |
| `LeftJoin(table, on)` | LEFT JOIN |
| `OrderBy(column, direction)` | 排序 |
| `Limit(num)` / `Offset(num)` | 分頁 |
| `Total()` | 啟用 Window Function 計算總數 |
| `Get()` / `GetContext(ctx)` | 執行查詢，回傳多筆 |
| `First()` / `FirstContext(ctx)` | 執行查詢，回傳單筆 |
| `Count()` / `CountContext(ctx)` | 計算筆數 |
| `Insert(data)` / `InsertContext(ctx, data)` | 新增資料 |
| `InsertReturningID(data)` | 新增並回傳 ID |
| `InsertConflict(conflict, data)` | 新增（含衝突處理） |
| `Update(data)` / `UpdateContext(ctx, data)` | 更新資料 |
| `Increase(column, num)` | 遞增欄位值 |
| `Decrease(column, num)` | 遞減欄位值 |
| `Toggle(column)` | 切換布林值 |

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
