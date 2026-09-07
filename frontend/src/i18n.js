const messages = {
    en: {
        search: 'Search tools...', favorites: 'Favorites', recent: 'Recent', settings: 'Settings',
        noFavorites: 'No favorites yet', noRecent: 'No recent usage yet', allTools: 'All tools',
        subtitle: 'Everything you need for everyday development', tools: 'tools', selectTool: 'Select a tool to get started',
        input: 'Input', output: 'Output', copy: 'Copy', download: 'Download', backHome: 'Back to home',
        emptyInput: 'Enter input to see output…', noMatches: 'No tools match your search',
        appearance: 'Appearance', theme: 'Theme', dark: 'Dark', light: 'Light', system: 'System',
        editor: 'Editor', indent: 'JSON indent', general: 'General', language: 'Language', storage: 'Storage',
        storagePath: 'Storage path', storageHint: 'Leave empty to use the default data directory.',
        save: 'Save changes', reset: 'Reset defaults', saved: 'Settings saved', saveFailed: 'Failed to save settings',
        back: 'Back', selectFile: 'Select file', generate: 'Generate', copied: 'Copied to clipboard', copyFailed: 'Failed to copy',
        downloaded: 'Downloaded', favoriteTitle: 'Toggle favorite', category: 'Category',
    },
    zh: {
        search: '搜索工具…', favorites: '收藏', recent: '最近使用', settings: '设置',
        noFavorites: '暂无收藏', noRecent: '暂无使用记录', allTools: '全部工具',
        subtitle: '日常开发所需的实用工具集合', tools: '个工具', selectTool: '选择一个工具开始使用',
        input: '输入', output: '输出', copy: '复制', download: '下载', backHome: '返回首页',
        emptyInput: '输入内容后将在这里显示结果…', noMatches: '没有符合搜索条件的工具',
        appearance: '外观', theme: '主题', dark: '深色', light: '浅色', system: '跟随系统',
        editor: '编辑器', indent: 'JSON 缩进', general: '通用', language: '语言', storage: '存储',
        storagePath: '存储目录', storageHint: '留空将使用默认数据目录。',
        save: '保存设置', reset: '恢复默认', saved: '设置已保存', saveFailed: '保存设置失败',
        back: '返回', selectFile: '选择文件', generate: '生成', copied: '已复制到剪贴板', copyFailed: '复制失败',
        downloaded: '已下载', favoriteTitle: '切换收藏', category: '分类',
    },
};

const toolMessages = {
    json: ['JSON Formatter', 'JSON 格式化', 'Format, minify, validate, and inspect JSON.', '格式化、压缩、校验并查看 JSON。'],
    base64: ['Base64', 'Base64 编解码', 'Encode and decode Base64 text.', '编码和解码 Base64 文本。'],
    url: ['URL Encoder', 'URL 编解码', 'Encode and decode URLs and query strings.', '编码、解码 URL 与查询参数。'],
    hash: ['Hash', '哈希计算', 'Compute text and file hashes.', '计算文本或文件的哈希值。'],
    uuid: ['UUID', 'UUID 生成器', 'Generate UUID v4 values.', '单个或批量生成 UUID v4。'],
    timestamp: ['Timestamp', '时间戳转换', 'Convert timestamps and dates.', '在时间戳与日期之间转换。'],
    jwt: ['JWT Decoder', 'JWT 解析', 'Decode JWT header and payload.', '解析 JWT Header、Payload 与时间声明。'],
    regex: ['Regex Tester', '正则测试', 'Test regular expressions and replacements.', '测试正则表达式、捕获组与替换。'],
    yaml: ['JSON ↔ YAML', 'JSON ↔ YAML', 'Convert between JSON and YAML.', '在 JSON 与 YAML 之间转换。'],
    xml: ['XML Toolkit', 'XML 工具', 'Format, validate, and convert XML and JSON.', '格式化、校验并转换 XML 与 JSON。'],
    csv: ['CSV Toolkit', 'CSV 工具', 'View, format, and convert CSV and JSON.', '查看、格式化并转换 CSV 与 JSON。'],
    toml: ['TOML Toolkit', 'TOML 工具', 'Format, validate, and convert TOML and JSON.', '格式化、校验并转换 TOML 与 JSON。'],
    text: ['Text Toolkit', '文本工具', 'Format, escape, and inspect text.', '格式化、转义并统计文本。'],
    sql: ['SQL Formatter', 'SQL 格式化', 'Format, beautify, and minify SQL.', '格式化、美化并压缩 SQL。'],
    markdown: ['Markdown', 'Markdown 编辑器', 'Edit, preview, and convert Markdown.', '编辑、预览并转换 Markdown。'],
    html: ['HTML Toolkit', 'HTML 工具', 'Format, minify, escape, and preview HTML.', '格式化、压缩、转义并预览 HTML。'],
    http: ['HTTP Client', 'HTTP 客户端', 'Send HTTP requests and inspect responses.', '发送 HTTP 请求并检查响应。'],
    dns: ['DNS Lookup', 'DNS 查询', 'Resolve common DNS record types.', '查询常用 DNS 记录。'],
    ping: ['Ping', 'Ping 测试', 'Measure host reachability and latency.', '测试主机连通性与延迟。'],
    port: ['Port Checker', '端口检查', 'Check whether a TCP port is open.', '检查 TCP 端口是否开放。'],
    cidr: ['CIDR Calculator', 'CIDR 计算器', 'Calculate IP network ranges.', '计算 IP 网络范围。'],
    'url-parser': ['URL Parser', 'URL 解析器', 'Inspect URL components and query parameters.', '解析 URL 组成与查询参数。'],
    cron: ['Cron Tool', 'Cron 工具', 'Parse and generate cron expressions.', '解析和生成 Cron 表达式。'],
    color: ['Color Converter', '颜色转换', 'Convert HEX, RGB, RGBA, HSL, and HSV.', '转换 HEX、RGB、RGBA、HSL 和 HSV。'],
    'code-generator': ['Code Generator', '代码生成器', 'Generate types from JSON.', '根据 JSON 生成类型定义。'],
    git: ['Git Tools', 'Git 工具', 'Generate gitignore files, commands, and parse URLs.', '生成 gitignore、命令并解析 Git URL。'],
    'amount-cn': ['Chinese Amount', '费用金额中文转换', 'Convert amounts to Chinese financial uppercase.', '将费用金额转换为中文大写或普通中文数字。'],
    tcp: ['TCP Client', 'TCP 调试', 'Send text or hexadecimal data to a TCP endpoint.', '向 TCP 服务端发送文本或十六进制数据。'],
    udp: ['UDP Client', 'UDP 调试', 'Send a UDP datagram and wait for a response.', '发送 UDP 数据报并等待响应。'],
    serial: ['Serial Port', '串口调试', 'List serial ports and exchange text or hex data.', '枚举串口并收发文本或十六进制数据。'],
    pinyin: ['Chinese Pinyin', '文字转拼音', 'Convert Chinese text to full pinyin or initials.', '将中文转换为全拼或拼音首字母。'],
    'database-docs': ['Database Syntax Docs', '数据库语法文档', 'Offline syntax reference for common databases.', 'PostgreSQL、MySQL、Oracle、SQL Server 离线语法速查。'],
    'document-split': ['Document Page Splitter', '文档单双页拆分', 'Split PDF, Word, and PowerPoint files into odd/even PDFs for duplex printing.', '将 PDF、Word、PPT 拆成奇数页和偶数页 PDF，方便正反面打印。'],
    'hiit-timer': ['HIIT Timer', 'HIIT 间歇计时器', 'Configure work, rest, and rounds with breathing-light phase cues.', '配置运动、休息时间和组数，通过呼吸灯圆球提示训练阶段。'],
};

export function translate(language, key) {
	// 缺失翻译自动回退英文，避免界面直接显示 undefined。
    return messages[language]?.[key] || messages.en[key] || key;
}

export function localizeTool(tool, language) {
	// 后端元信息保持稳定英文值，本地化只发生在展示层，不影响工具 ID 和持久化数据。
    const values = toolMessages[tool.id];
    if (!values) return tool;
    return { ...tool, name: language === 'zh' ? values[1] : values[0], description: language === 'zh' ? values[3] : values[2] };
}

export function localizeCategory(category, language) {
    if (language !== 'zh') return category;
    const categories = { Convert: '转换', Crypto: '安全', Encode: '编码', Format: '格式化', Generator: '生成器', Text: '文本', data: '数据', text: '文本', network: '网络', developer: '开发', document: '文档', utility: '实用工具', encoding: '编码', generator: '生成器', security: '安全', time: '时间' };
    return categories[category] || category;
}
