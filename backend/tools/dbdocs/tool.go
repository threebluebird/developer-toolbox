// Package dbdocs 内置常用数据库语法速查文档，不连接任何数据库。
package dbdocs

import (
	"developer-toolbox/backend/models"
	"sort"
	"strings"
)

type DatabaseDocsTool struct{}

func NewDatabaseDocsTool() *DatabaseDocsTool { return &DatabaseDocsTool{} }
func (t *DatabaseDocsTool) Info() models.Tool {
	return models.Tool{
		ID:          "database-docs",
		Name:        "Database Syntax Docs",
		Description: "Offline SQL syntax reference for PostgreSQL, MySQL, Oracle, and SQL Server.",
		Category:    "developer",
		Icon:        "database",
		Version:     "0.5.0",
		Keywords:    []string{"postgresql", "mysql", "oracle", "sqlserver", "sql server", "database", "syntax", "文档"},
	}
}

type Entry struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Syntax      string `json:"syntax"`
	Example     string `json:"example"`
}

// Execute 按数据库和关键字过滤离线条目，避免把文档查询误当作数据库连接。
func (t *DatabaseDocsTool) Execute(input models.ToolInput) models.ToolOutput {
	database := strings.ToLower(stringValue(input.Payload["database"], "postgresql"))
	entries, ok := documents[database]
	if !ok {
		return models.Failure("DB_DOCS_ERROR", "不支持的数据库")
	}
	query := strings.ToLower(strings.TrimSpace(stringValue(input.Payload["input"], "")))
	if query != "" {
		filtered := []Entry{}
		for _, entry := range entries {
			if strings.Contains(strings.ToLower(entry.Title+" "+entry.Description+" "+entry.Syntax), query) {
				filtered = append(filtered, entry)
			}
		}
		entries = filtered
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Title < entries[j].Title })
	return models.ToolOutput{Success: true, Data: map[string]any{"database": database, "count": len(entries), "entries": entries}}
}

var common = []Entry{
	{"CREATE TABLE", "创建表", "CREATE TABLE table_name (id INTEGER PRIMARY KEY, name VARCHAR(100));", "CREATE TABLE users (id INTEGER PRIMARY KEY, name VARCHAR(100) NOT NULL);"},
	{"JOIN", "连接查询", "SELECT ... FROM a JOIN b ON a.id = b.a_id;", "SELECT u.name, o.total FROM users u JOIN orders o ON o.user_id = u.id;"},
	{"GROUP BY", "分组聚合", "SELECT column, COUNT(*) FROM table GROUP BY column;", "SELECT status, COUNT(*) FROM orders GROUP BY status;"},
	{"INDEX", "创建索引", "CREATE INDEX index_name ON table_name (column);", "CREATE INDEX idx_users_email ON users (email);"},
	{"TRANSACTION", "事务", "BEGIN; ... COMMIT; -- 或 ROLLBACK", "BEGIN; UPDATE accounts SET balance = balance - 100 WHERE id = 1; COMMIT;"},
}

var documents = map[string][]Entry{
	"postgresql": withCommon(
		Entry{"UPSERT", "冲突时更新", "INSERT ... ON CONFLICT (key) DO UPDATE SET ...;", "INSERT INTO users(id,name) VALUES(1,'Tom') ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name;"},
		Entry{"分页", "LIMIT/OFFSET 分页", "SELECT ... LIMIT limit OFFSET offset;", "SELECT * FROM users ORDER BY id LIMIT 20 OFFSET 40;"},
		Entry{"JSONB", "JSONB 查询", "column -> 'key'; column ->> 'key'; column @> jsonb", "SELECT payload->>'name' FROM events WHERE payload @> '{\"active\":true}';"},
		Entry{"函数 · 字符串", "常用字符串拼接、大小写、长度、截取、去空格、替换和聚合函数。LENGTH 按字符计数，OCTET_LENGTH 按字节计数。", "CONCAT / ||; LOWER; UPPER; LENGTH; SUBSTRING; TRIM; REPLACE; STRING_AGG", "SELECT CONCAT(first_name, ' ', last_name), SUBSTRING(code FROM 1 FOR 4), STRING_AGG(tag, ',') FROM users GROUP BY first_name, last_name, code;"},
		Entry{"函数 · 数组", "PostgreSQL 原生数组的创建、长度、展开、追加、删除、包含和重叠判断。数组下标默认从 1 开始。", "ARRAY[...]; array_length(arr, dim); unnest(arr); array_append; array_remove; arr @> other; arr && other", "SELECT array_length(tags, 1), unnest(tags) FROM posts WHERE tags @> ARRAY['go'];"},
		Entry{"函数 · JSON/JSONB", "构造、读取、修改、包含判断和展开 JSON。频繁查询与索引场景通常优先使用 JSONB。", "jsonb_build_object; ->; ->>; jsonb_set; @>; jsonb_array_elements; jsonb_path_query", "SELECT jsonb_set(payload, '{profile,name}', '\"Tom\"'), item FROM events CROSS JOIN LATERAL jsonb_array_elements(payload->'items') AS item;"},
		Entry{"函数 · 日期时间", "获取当前时间、按周期截断、提取字段、计算年龄和使用时间间隔。", "NOW(); CURRENT_DATE; DATE_TRUNC; EXTRACT; AGE; value + INTERVAL '1 day'", "SELECT DATE_TRUNC('month', created_at), EXTRACT(EPOCH FROM NOW()-created_at), created_at + INTERVAL '7 days' FROM orders;"},
	),
	"mysql": withCommon(
		Entry{"UPSERT", "重复键更新", "INSERT ... ON DUPLICATE KEY UPDATE ...;", "INSERT INTO users(id,name) VALUES(1,'Tom') ON DUPLICATE KEY UPDATE name=VALUES(name);"},
		Entry{"分页", "LIMIT 分页", "SELECT ... LIMIT offset, count;", "SELECT * FROM users ORDER BY id LIMIT 40, 20;"},
		Entry{"自增列", "AUTO_INCREMENT", "id BIGINT AUTO_INCREMENT PRIMARY KEY", "CREATE TABLE users (id BIGINT AUTO_INCREMENT PRIMARY KEY);"},
		Entry{"函数 · 字符串", "常用字符串拼接、大小写、字符长度、截取、去空格、替换和分组聚合函数。CHAR_LENGTH 按字符计数，LENGTH 按字节计数。", "CONCAT; CONCAT_WS; LOWER; UPPER; CHAR_LENGTH; SUBSTRING; TRIM; REPLACE; GROUP_CONCAT", "SELECT CONCAT_WS(' ', first_name, last_name), SUBSTRING(code, 1, 4), GROUP_CONCAT(tag ORDER BY tag SEPARATOR ',') FROM users GROUP BY first_name,last_name,code;"},
		Entry{"函数 · 数组/集合", "MySQL 没有独立数组类型，通常使用 JSON 数组；JSON_TABLE 可将数组展开为行。", "JSON_ARRAY; JSON_ARRAY_APPEND; JSON_CONTAINS; JSON_LENGTH; JSON_TABLE", "SELECT jt.value FROM JSON_TABLE('[\"go\",\"sql\"]', '$[*]' COLUMNS(value VARCHAR(20) PATH '$')) AS jt;"},
		Entry{"函数 · JSON", "构造、读取、修改、删除、校验 JSON，并通过 JSON_TABLE 关系化展开。", "JSON_OBJECT; JSON_EXTRACT / ->>; JSON_SET; JSON_REMOVE; JSON_KEYS; JSON_VALID; JSON_TABLE", "SELECT JSON_SET(payload, '$.active', TRUE), payload->>'$.name' FROM events WHERE JSON_VALID(payload);"},
		Entry{"函数 · 日期时间", "获取当前时间、格式化与解析、日期加减以及计算时间差。", "NOW; CURDATE; DATE_FORMAT; STR_TO_DATE; DATE_ADD; DATE_SUB; TIMESTAMPDIFF", "SELECT DATE_FORMAT(created_at,'%Y-%m'), DATE_ADD(created_at, INTERVAL 7 DAY), TIMESTAMPDIFF(HOUR,created_at,NOW()) FROM orders;"},
	),
	"oracle": withCommon(
		Entry{"MERGE", "合并插入和更新", "MERGE INTO target USING source ON (...) WHEN MATCHED THEN UPDATE ... WHEN NOT MATCHED THEN INSERT ...;", "MERGE INTO users u USING new_users n ON (u.id=n.id) WHEN MATCHED THEN UPDATE SET u.name=n.name WHEN NOT MATCHED THEN INSERT(id,name) VALUES(n.id,n.name);"},
		Entry{"分页", "OFFSET/FETCH 分页", "SELECT ... OFFSET n ROWS FETCH NEXT m ROWS ONLY;", "SELECT * FROM users ORDER BY id OFFSET 40 ROWS FETCH NEXT 20 ROWS ONLY;"},
		Entry{"序列", "SEQUENCE", "CREATE SEQUENCE sequence_name START WITH 1 INCREMENT BY 1;", "CREATE SEQUENCE users_seq START WITH 1 INCREMENT BY 1;"},
		Entry{"函数 · 字符串", "常用字符串拼接、大小写、长度、截取、去空格、替换和行转字符串函数。多个字符串拼接推荐使用 ||。", "||; CONCAT; LOWER; UPPER; LENGTH; SUBSTR; TRIM; REPLACE; LISTAGG", "SELECT first_name || ' ' || last_name, SUBSTR(code,1,4), LISTAGG(tag,',') WITHIN GROUP (ORDER BY tag) FROM users GROUP BY first_name,last_name,code;"},
		Entry{"函数 · 数组/集合", "Oracle SQL 常用集合类型、TABLE 展开和 MULTISET 集合运算；普通业务也可使用 JSON 数组。", "SYS.ODCIVARCHAR2LIST; TABLE(collection); MULTISET UNION / INTERSECT / EXCEPT; JSON_ARRAY", "SELECT column_value FROM TABLE(SYS.ODCIVARCHAR2LIST('go','sql'));"},
		Entry{"函数 · JSON", "构造、读取标量、读取对象或数组、判断路径及更新 JSON；部分更新函数取决于 Oracle 版本。", "JSON_OBJECT; JSON_ARRAY; JSON_VALUE; JSON_QUERY; JSON_EXISTS; JSON_TRANSFORM", "SELECT JSON_VALUE(payload,'$.name'), JSON_QUERY(payload,'$.items') FROM events WHERE JSON_EXISTS(payload,'$.active?(@ == true)');"},
		Entry{"函数 · 日期时间", "获取数据库时间与高精度时间戳、截断日期、月份计算、字段提取及格式转换。", "SYSDATE; SYSTIMESTAMP; TRUNC; ADD_MONTHS; MONTHS_BETWEEN; EXTRACT; TO_DATE; TO_CHAR", "SELECT TRUNC(created_at,'MM'), ADD_MONTHS(created_at,1), TO_CHAR(created_at,'YYYY-MM-DD HH24:MI:SS') FROM orders;"},
	),
	"sqlserver": withCommon(
		Entry{"MERGE", "合并插入和更新", "MERGE target USING source ON ... WHEN MATCHED THEN UPDATE WHEN NOT MATCHED THEN INSERT;", "MERGE users AS t USING new_users AS s ON t.id=s.id WHEN MATCHED THEN UPDATE SET name=s.name WHEN NOT MATCHED THEN INSERT(id,name) VALUES(s.id,s.name);"},
		Entry{"分页", "OFFSET/FETCH 分页", "SELECT ... ORDER BY ... OFFSET n ROWS FETCH NEXT m ROWS ONLY;", "SELECT * FROM users ORDER BY id OFFSET 40 ROWS FETCH NEXT 20 ROWS ONLY;"},
		Entry{"TOP", "限制行数", "SELECT TOP (n) ... FROM ...;", "SELECT TOP (10) * FROM users ORDER BY created_at DESC;"},
		Entry{"函数 · 字符串", "常用字符串拼接、大小写、长度、截取、去空格、替换和字符串聚合函数。LEN 不统计尾随空格，字节长度使用 DATALENGTH。", "CONCAT; CONCAT_WS; LOWER; UPPER; LEN; DATALENGTH; SUBSTRING; TRIM; REPLACE; STRING_AGG", "SELECT CONCAT_WS(' ',first_name,last_name), SUBSTRING(code,1,4), STRING_AGG(tag,',') WITHIN GROUP (ORDER BY tag) FROM users GROUP BY first_name,last_name,code;"},
		Entry{"函数 · 数组/集合", "SQL Server 没有原生数组类型，常用 STRING_SPLIT 拆分文本或 OPENJSON 展开 JSON 数组。", "STRING_SPLIT; STRING_AGG; OPENJSON", "SELECT value FROM OPENJSON('[\"go\",\"sql\"]'); SELECT value FROM STRING_SPLIT('go,sql', ',');"},
		Entry{"函数 · JSON", "校验、读取标量或对象、修改 JSON，将数组展开为行，以及把查询结果输出为 JSON。", "ISJSON; JSON_VALUE; JSON_QUERY; JSON_MODIFY; OPENJSON; FOR JSON PATH", "SELECT JSON_VALUE(payload,'$.name'), JSON_MODIFY(payload,'$.active',1) FROM events WHERE ISJSON(payload)=1;"},
		Entry{"函数 · 日期时间", "获取本地或 UTC 时间、日期加减、计算时间差、截断周期及格式转换。DATETRUNC 需要 SQL Server 2022+。", "GETDATE; SYSDATETIME; SYSUTCDATETIME; DATEADD; DATEDIFF; DATETRUNC; CONVERT", "SELECT DATEADD(day,7,created_at), DATEDIFF(hour,created_at,SYSDATETIME()), CONVERT(varchar(19),created_at,120) FROM orders;"},
	),
}

// 为每种数据库复制公共条目，避免后续扩展某一数据库时修改共享底层切片。
func withCommon(entries ...Entry) []Entry {
	result := make([]Entry, 0, len(common)+len(entries))
	result = append(result, common...)
	return append(result, entries...)
}

func stringValue(v any, f string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return f
}
