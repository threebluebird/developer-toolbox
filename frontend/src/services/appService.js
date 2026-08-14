import {
    AddFavorite,
    AddFavoriteItem,
    ClearHistory,
    ExecuteTool,
    GetHistory,
    ListFavorites,
    ListTools,
    LoadSettings,
    RemoveFavorite,
    RemoveFavoriteItem,
    SaveSettings,
    SearchTools,
    SelectFile,
} from '../../wailsjs/go/main/App.js';

const previewTools = [
    { id: 'json', name: 'JSON Formatter', description: 'Format and validate JSON', category: 'Format', version: '1.0.0' },
    { id: 'base64', name: 'Base64', description: 'Encode and decode Base64 text', category: 'Encode', version: '1.0.0' },
    { id: 'url', name: 'URL Encoder', description: 'Encode and decode URLs', category: 'Encode', version: '1.0.0' },
    { id: 'hash', name: 'Hash', description: 'Compute text and file hashes', category: 'Crypto', version: '1.0.0' },
    { id: 'uuid', name: 'UUID', description: 'Generate UUID v4', category: 'Generator', version: '1.0.0' },
    { id: 'timestamp', name: 'Timestamp', description: 'Convert timestamps and dates', category: 'Convert', version: '1.0.0' },
    { id: 'jwt', name: 'JWT Decoder', description: 'Decode JWT tokens', category: 'Crypto', version: '1.0.0' },
    { id: 'regex', name: 'Regex Tester', description: 'Test regular expressions', category: 'Text', version: '1.0.0' },
    { id: 'yaml', name: 'JSON ↔ YAML', description: 'Convert between JSON and YAML', category: 'Convert', version: '1.0.0' },
    { id: 'xml', name: 'XML Toolkit', description: 'Format and convert XML', category: 'data', version: '0.5.0' },
    { id: 'csv', name: 'CSV Toolkit', description: 'View and convert CSV', category: 'data', version: '0.5.0' },
    { id: 'toml', name: 'TOML Toolkit', description: 'Format and convert TOML', category: 'data', version: '0.5.0' },
    { id: 'text', name: 'Text Toolkit', description: 'Format and inspect text', category: 'text', version: '0.5.0' },
    { id: 'sql', name: 'SQL Formatter', description: 'Format and minify SQL', category: 'data', version: '0.5.0' },
    { id: 'markdown', name: 'Markdown', description: 'Edit and preview Markdown', category: 'text', version: '0.5.0' },
    { id: 'html', name: 'HTML Toolkit', description: 'Format and preview HTML', category: 'text', version: '0.5.0' },
    { id: 'http', name: 'HTTP Client', description: 'Send HTTP requests', category: 'network', version: '0.5.0' },
    { id: 'dns', name: 'DNS Lookup', description: 'Resolve DNS records', category: 'network', version: '0.5.0' },
    { id: 'ping', name: 'Ping', description: 'Measure host latency', category: 'network', version: '0.5.0' },
    { id: 'port', name: 'Port Checker', description: 'Check a TCP port', category: 'network', version: '0.5.0' },
    { id: 'cidr', name: 'CIDR Calculator', description: 'Calculate network ranges', category: 'network', version: '0.5.0' },
    { id: 'url-parser', name: 'URL Parser', description: 'Inspect URL components', category: 'network', version: '0.5.0' },
    { id: 'cron', name: 'Cron Tool', description: 'Parse and generate cron expressions', category: 'developer', version: '0.5.0' },
    { id: 'color', name: 'Color Converter', description: 'Convert color formats', category: 'developer', version: '0.5.0' },
    { id: 'code-generator', name: 'Code Generator', description: 'Generate types from JSON', category: 'developer', version: '0.5.0' },
    { id: 'git', name: 'Git Tools', description: 'Generate Git helpers', category: 'developer', version: '0.5.0' },
    { id: 'amount-cn', name: 'Chinese Amount', description: 'Convert amount to Chinese uppercase', category: 'convert', version: '0.5.0' },
    { id: 'tcp', name: 'TCP Client', description: 'Send TCP data', category: 'network', version: '0.5.0' },
    { id: 'udp', name: 'UDP Client', description: 'Send UDP datagrams', category: 'network', version: '0.5.0' },
    { id: 'serial', name: 'Serial Port', description: 'Exchange serial data', category: 'network', version: '0.5.0' },
    { id: 'pinyin', name: 'Chinese Pinyin', description: 'Convert Chinese to pinyin', category: 'text', version: '0.5.0' },
    { id: 'database-docs', name: 'Database Syntax Docs', description: 'Offline database syntax reference', category: 'developer', version: '0.5.0' },
];

const hasWails = () => Boolean(window.go?.main?.App);
// Web 预览模式没有 Go Repository，使用内存记录模拟桌面端历史行为，便于页面独立调试。
const previewHistory = [];
const previewExecute = input => {
    if (input.toolId === 'amount-cn') return { success: true, data: input.payload.input || '' };
    if (input.toolId === 'database-docs') return {
        success: true,
        data: { database: input.payload.database || 'postgresql', count: 0, entries: [] },
    };
    if (input.toolId === 'markdown') return { success: true, data: input.payload.input.replace(/^# (.+)$/gm, '<h1>$1</h1>').replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>') };
    if (input.toolId === 'html') return { success: true, data: input.payload.input.replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, '') };
    if (input.toolId !== 'json') return { success: true, data: input.payload.input || '' };
    try {
        const value = JSON.parse(input.payload.input);
        const action = input.payload.action || 'format';
        if (action === 'validate') return { success: true, data: 'valid' };
        if (action === 'tree') return { success: true, data: value };
        return { success: true, data: action === 'minify' ? JSON.stringify(value) : JSON.stringify(value, null, input.payload.indent || 2) };
    } catch (error) { return { success: false, error: { message: error.message } }; }
};

const previewExecuteWithHistory = input => {
    const result = previewExecute(input);
    if (result.success) {
        previewHistory.unshift({
            id: crypto.randomUUID(),
            toolId: input.toolId,
            usedAt: new Date().toISOString(),
            input: structuredClone(input.payload || {}),
            output: structuredClone(result.data),
        });
        previewHistory.splice(50);
    }
    return result;
};

const invoke = (fn, fallback) => (...args) => hasWails() ? fn(...args) : Promise.resolve(fallback(...args));

export const appService = {
    addFavorite: invoke(AddFavorite, () => undefined),
    addFavoriteItem: invoke(AddFavoriteItem, item => ({ ...item, id: crypto.randomUUID(), addedAt: new Date().toISOString() })),
    clearHistory: invoke(ClearHistory, () => { previewHistory.length = 0; }),
    executeTool: invoke(ExecuteTool, previewExecuteWithHistory),
    getHistory: invoke(GetHistory, () => structuredClone(previewHistory)),
    listFavorites: invoke(ListFavorites, () => []),
    listTools: invoke(ListTools, () => previewTools),
    loadSettings: invoke(LoadSettings, () => ({ theme: 'dark', language: 'en', storagePath: '', jsonIndent: 2 })),
    removeFavorite: invoke(RemoveFavorite, () => undefined),
    removeFavoriteItem: invoke(RemoveFavoriteItem, () => undefined),
    saveSettings: invoke(SaveSettings, () => undefined),
    searchTools: invoke(SearchTools, query => previewTools.filter(tool => JSON.stringify(tool).toLowerCase().includes(String(query).toLowerCase()))),
    selectFile: invoke(SelectFile, () => ''),
};
