import './style.css';
import './app.css';

import logo from './assets/images/logo-universal.png';
import { appService } from './services/appService';
import { localizeCategory, localizeTool, translate } from './i18n';
import { createAppState } from './stores/appStore';
import { copyText } from './utils/clipboard';
import { errorMessage, formatOutput } from './utils/output';
import { filterTools, groupByCategory } from './utils/tools';
import { createHIITTimeline, formatTimer, resolveHIITProgress } from './utils/hiitTimer.js';
import { CodeEditor, JsonEditor, TextEditor } from './components/editor/index.js';

// Mock data for non-Wails preview
const MOCK_TOOLS = [
    { id: 'json', name: 'JSON Formatter', description: 'Format and validate JSON', category: 'Format', version: '1.0.0' },
    { id: 'base64', name: 'Base64 Encode/Decode', description: 'Encode or decode base64', category: 'Encode', version: '1.0.0' },
    { id: 'hash', name: 'Hash Generator', description: 'Generate hashes (MD5/SHA)', category: 'Crypto', version: '1.0.0' },
    { id: 'url', name: 'URL Encoder', description: 'Encode/decode URL and query strings', category: 'Encode', version: '1.0.0' },
    { id: 'uuid', name: 'UUID Generator', description: 'Generate UUID v4', category: 'Generator', version: '1.0.0' },
    { id: 'timestamp', name: 'Timestamp Converter', description: 'Convert timestamps and dates', category: 'Convert', version: '1.0.0' },
    { id: 'jwt', name: 'JWT Decoder', description: 'Decode JWT tokens', category: 'Crypto', version: '1.0.0' },
    { id: 'regex', name: 'Regex Tester', description: 'Test and debug regular expressions', category: 'Text', version: '1.0.0' },
    { id: 'yaml', name: 'JSON ↔ YAML', description: 'Convert between JSON and YAML', category: 'Convert', version: '1.0.0' },
];

const MOCK_FAVORITES = [];
const MOCK_HISTORY = [];

async function safeListTools() {
    try { return await appService.listTools(); } catch (err) { return MOCK_TOOLS; }
}
async function safeListFavorites() {
    try { return await appService.listFavorites(); } catch (err) { return MOCK_FAVORITES; }
}
async function safeGetHistory() {
    try { return await appService.getHistory(); } catch (err) { return MOCK_HISTORY; }
}

const state = createAppState();
let activeToolCleanup = null;

const appRoot = document.querySelector('#app');
const systemTheme = window.matchMedia('(prefers-color-scheme: light)');
// 所有页面文案和工具元信息均通过当前语言动态解析，切换语言后无需重新请求后端。
const t = key => translate(state.settings.language, key);
const visibleTool = tool => localizeTool(tool, state.settings.language);

// ============== Init & Lifecycle ==============
async function init() {
	// 先渲染可交互骨架，再加载设置并按语言重建骨架，最后加载业务数据。
    appRoot.innerHTML = getAppShell();
    document.getElementById('logo').src = logo;
    setupEventListeners();
    await loadSettings();
    await loadData();
}

function getAppShell() {
    return `
        <div class="app-shell">
            <aside class="sidebar">
                <div class="brand">
                    <img id="logo" class="logo" alt="Logo" />
                    <div>
                        <h1>Developer Toolbox</h1>
                        <p>v0.5.0</p>
                    </div>
                </div>
                <div class="search-box">
                    <span class="search-icon">⌕</span>
                    <input id="searchInput" class="search-input" placeholder="${t('search')}" />
                </div>
                <section class="panel">
                    <h2>${t('favorites')}</h2>
                    <div id="favoriteList" class="tool-list"></div>
                </section>
                <section class="panel">
                    <h2>${t('recent')}</h2>
                    <div id="recentList" class="tool-list"></div>
                </section>
                <div class="sidebar-actions">
                    <button class="btn sidebar-settings" id="settingsBtn" title="${t('settings')}"><span>⚙</span>${t('settings')}</button>
                </div>
            </aside>
            <main class="main-content" id="mainContent"></main>
            <div id="commandPalette" class="command-palette-backdrop hidden" role="dialog" aria-modal="true" aria-label="Command Palette">
                <div class="command-palette"><input id="commandInput" placeholder="Type a command…" autocomplete="off"><div id="commandResults"></div></div>
            </div>
            <div id="toast" class="toast hidden"></div>
        </div>
    `;
}

function setupEventListeners() {
    document.getElementById('searchInput').oninput = onSearch;
    document.getElementById('settingsBtn').onclick = () => goSettings();
    document.getElementById('commandPalette').onclick = event => { if (event.target.id === 'commandPalette') closeCommandPalette(); };
    document.getElementById('commandInput').oninput = renderCommandResults;
}

document.addEventListener('keydown', event => {
    const modifier = event.ctrlKey || event.metaKey;
    if (modifier && event.shiftKey && event.key.toLowerCase() === 'p') { event.preventDefault(); openCommandPalette(); return; }
    if (modifier && event.key.toLowerCase() === 'k') { event.preventDefault(); goHome(); const search=document.getElementById('searchInput'); search?.focus(); search?.select(); return; }
    if (event.key === 'Escape' && state.paletteOpen) { closeCommandPalette(); return; }
    if (modifier && event.key === 'Enter' && state.currentView === 'editor' && !['http', 'document-split', 'hiit-timer'].includes(state.selectedTool?.id)) { event.preventDefault(); executeToolCommand(state.selectedTool.id); }
    if (modifier && event.shiftKey && event.key.toLowerCase() === 'c' && state.currentView === 'editor') { event.preventDefault(); copyOutput(); }
    if (modifier && event.key.toLowerCase() === 'l' && state.currentView === 'editor') { event.preventDefault(); clearCurrentInput(); }
    if (modifier && event.key.toLowerCase() === 'd' && state.selectedTool) { event.preventDefault(); toggleFavorite(state.selectedTool.id); }
    if (modifier && /^[1-4]$/.test(event.key)) { event.preventDefault(); const ids=['json','base64','timestamp','jwt']; const tool=state.allTools.find(item=>item.id===ids[Number(event.key)-1]); if(tool)selectToolAndEdit(visibleTool(tool)); }
});

// ============== Data Loading ==============
async function loadSettings() {
    try {
        const settings = await appService.loadSettings();
        if (settings) {
            state.settings = { ...state.settings, ...settings };
        }
        applyTheme(state.settings.theme);
    } catch (err) {
        console.error('Failed to load settings', err);
        applyTheme('dark'); // Default to dark
    }
    // 语言会影响侧栏静态文案，因此加载设置后重建一次应用骨架。
    appRoot.innerHTML = getAppShell();
    document.getElementById('logo').src = logo;
    setupEventListeners();
}

async function saveSettings() {
    try {
        await appService.saveSettings(state.settings);
        showToast(t('saved'));
    } catch (err) {
        console.error('Failed to save settings', err);
        showToast(t('saveFailed'));
    }
}

function applyTheme(theme) {
    state.settings.theme = theme;
    const root = document.documentElement;
    // system 保存的是用户偏好，data-theme 保存的是当前真正渲染的 light/dark。
    const resolvedTheme = theme === 'system'
        ? (window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark')
        : theme;
    if (resolvedTheme === 'light') {
        root.setAttribute('data-theme', 'light');
    } else {
        root.setAttribute('data-theme', 'dark');
    }
}

systemTheme.addEventListener('change', () => {
    // 仅在“跟随系统”模式响应系统变化，避免覆盖用户手动选择。
    if (state.settings.theme === 'system') applyTheme('system');
});

async function loadData() {
    try {
        const [tools, favorites, history] = await Promise.all([
            safeListTools(),
            safeListFavorites(),
            safeGetHistory(),
        ]);
        state.tools = tools || [];
        // allTools 保存完整集合，搜索结果变化时收藏和最近使用仍可解析工具信息。
        state.allTools = tools || [];
        state.favoriteItems = favorites || [];
        state.favorites = state.favoriteItems.filter(item => !item.kind || item.kind === 'tool').map(item => item.toolId);
        state.history = history || [];
        render();
    } catch (err) {
        console.error('Failed to load data', err);
        showToast('Unable to load tool data.');
    }
}

// ============== Render & Views ==============
function render() {
	// 主区域按 currentView 切换，侧栏保持统一并在每次状态变化后同步刷新。
    const main = document.getElementById('mainContent');
    if (state.currentView === 'editor' && document.getElementById('toolEditor')) {
        activeToolCleanup?.();
        activeToolCleanup = null;
    }
    if (state.currentView === 'home') {
        renderHomeView(main);
    } else if (state.currentView === 'editor') {
        renderEditorView(main);
    } else if (state.currentView === 'settings') {
        renderSettingsView(main);
    }
    renderSidebar();
}

function renderHomeView(main) {
    main.innerHTML = `
        <section class="tool-overview">
            <div class="tool-overview-header">
                <div>
                    <span class="eyebrow">DEVELOPER TOOLBOX</span>
                    <h2>${t('allTools')}</h2>
                    <p>${t('subtitle')}</p>
                </div>
            </div>
        </section>
        <section class="panel">
            <div class="panel-header">
                <span id="toolCount" class="badge"></span>
            </div>
            <div id="categoryList" class="category-list"></div>
            <div id="workflowSearchResults" class="workflow-search-results hidden"></div>
        </section>
    `;
    renderCategories();
    document.getElementById('toolCount').innerText = `${state.tools.length} ${t('tools')}`;
}

function renderEditorView(main) {
    if (!state.selectedTool) return;
    const tool = visibleTool(state.selectedTool);
    
    main.innerHTML = `
        <div class="editor-wrapper">
            <div class="editor-header">
                <div class="editor-title">
                    <h2>${tool.name}</h2>
                    <p>${tool.description}</p>
                </div>
                <button class="btn" id="backBtn">← ${t('backHome')}</button>
            </div>
            <div class="editor-container">
                <div id="toolEditor" class="tool-editor"></div>
            </div>
        </div>
    `;
    
    document.getElementById('backBtn').onclick = () => goHome();
    renderToolEditor(tool);
}

function renderToolEditor(tool) {
    const container = document.getElementById('toolEditor');
    if (tool.id === 'http') { renderHTTPClient(container, tool); return; }
    if (tool.id === 'document-split') { renderDocumentSplitter(container); return; }
    if (tool.id === 'hiit-timer') { renderHIITTimer(container); return; }
    
    // Generic editor layout for all tools
    container.innerHTML = `
        <div class="editor-split">
            <div class="editor-pane input-pane">
                <div class="pane-header">
                    <h3>${t('input')}</h3>
                    <div id="inputOptions" class="editor-options"></div>
                </div>
                <div id="inputEditorHost" class="editor-host"></div>
            </div>
            <div class="editor-pane output-pane">
                <div class="pane-header">
                    <h3>${t('output')}</h3>
                    <div class="pane-actions">
                        <button class="btn small" id="favoriteInputBtn">☆ Save input</button>
                        <button class="btn small" id="copyBtn">${t('copy')}</button>
                        <button class="btn small" id="downloadBtn">${t('download')}</button>
                    </div>
                </div>
                <div id="outputText" class="editor-output"></div>
            </div>
        </div>
    `;

    const Editor = tool.id === 'json' ? JsonEditor : (['yaml', 'jwt', 'regex', 'xml', 'toml', 'sql', 'markdown', 'html', 'code-generator'].includes(tool.id) ? CodeEditor : TextEditor);
    const language = { yaml: 'yaml', jwt: 'json', regex: 'text', xml: 'xml', toml: 'text', sql: 'sql', markdown: 'text', html: 'html', 'code-generator': 'json' }[tool.id] || 'text';
    const editor = new Editor({ id: 'inputText', language, placeholder: t('emptyInput') });
    editor.mount(document.getElementById('inputEditorHost'));
    container.editor = editor;
    
    // Setup tool-specific UI and handlers
    setupToolEditor(tool);
    
    // Setup common actions
    document.getElementById('copyBtn').onclick = () => copyOutput();
    document.getElementById('downloadBtn').onclick = () => downloadOutput(tool.name);
    document.getElementById('favoriteInputBtn').onclick = () => favoriteCurrentInput(tool);
}

function renderDocumentSplitter(container) {
    const copy = state.settings.language === 'zh' ? {
        intro: '生成可直接打印的单双页 PDF', detail: 'PDF 将原样拆页；Word 和 PowerPoint 会先使用 Microsoft Office 或 LibreOffice 按实际版式转换为 PDF。',
        source: '源文件', sourceHint: '支持 PDF、DOC、DOCX、PPT、PPTX', selectFile: '选择文件', outputDir: '输出目录', outputHint: '默认与源文件同目录', selectDir: '选择目录',
        split: '拆分奇数页 / 偶数页', note: '不会修改源文件；若同名文件已存在，会自动添加序号。', result: '处理结果', empty: '选择文件后开始拆分。', working: '正在转换与拆分…', workingHint: '正在处理，请勿关闭应用。Office 文件页数较多时可能需要一些时间…',
    } : {
        intro: 'Create print-ready odd/even PDFs', detail: 'PDF pages are extracted directly. Word and PowerPoint files are first rendered with Microsoft Office or LibreOffice.',
        source: 'Source', sourceHint: 'PDF, DOC, DOCX, PPT, or PPTX', selectFile: 'Select file', outputDir: 'Output', outputHint: 'Defaults to the source folder', selectDir: 'Select folder',
        split: 'Split odd / even pages', note: 'The source is never changed. A number is added if an output already exists.', result: 'Result', empty: 'Select a file to begin.', working: 'Converting and splitting…', workingHint: 'Processing… Keep the app open. Large Office files may take a while.',
    };
    container.innerHTML = `
        <div class="document-splitter">
            <section class="document-split-form">
                <div class="document-split-intro">
                    <strong>${copy.intro}</strong>
                    <p>${copy.detail}</p>
                </div>
                <label class="file-picker-row">
                    <span>${copy.source}</span>
                    <input id="documentFilePath" class="input-field" readonly placeholder="${copy.sourceHint}">
                    <button type="button" class="btn" id="documentFileBtn">${copy.selectFile}</button>
                </label>
                <label class="file-picker-row">
                    <span>${copy.outputDir}</span>
                    <input id="documentOutputDir" class="input-field" readonly placeholder="${copy.outputHint}">
                    <button type="button" class="btn" id="documentOutputBtn">${copy.selectDir}</button>
                </label>
                <div class="document-split-actions">
                    <button type="button" class="btn primary" id="documentSplitBtn" disabled>${copy.split}</button>
                    <span class="hint">${copy.note}</span>
                </div>
            </section>
            <section class="document-split-result">
                <h3>${copy.result}</h3>
                <div id="outputText" class="editor-output"><p class="placeholder">${copy.empty}</p></div>
            </section>
        </div>`;

    const filePath = document.getElementById('documentFilePath');
    const outputDir = document.getElementById('documentOutputDir');
    const splitButton = document.getElementById('documentSplitBtn');
    const output = document.getElementById('outputText');

    document.getElementById('documentFileBtn').onclick = async () => {
        try {
            const selected = await appService.selectDocumentFile();
            if (!selected) return;
            filePath.value = selected;
            splitButton.disabled = false;
        } catch (error) { showToast(`无法选择文件：${error.message}`); }
    };
    document.getElementById('documentOutputBtn').onclick = async () => {
        try {
            const selected = await appService.selectOutputDirectory();
            if (selected) outputDir.value = selected;
        } catch (error) { showToast(`无法选择输出目录：${error.message}`); }
    };
    splitButton.onclick = async () => {
        splitButton.disabled = true;
        splitButton.textContent = copy.working;
        output.innerHTML = `<p class="placeholder">${copy.workingHint}</p>`;
        try {
            const result = await appService.executeTool({ toolId: 'document-split', payload: { filePath: filePath.value, outputDir: outputDir.value } });
            if (!result.success) { renderError(output, errorMessage(result.error)); return; }
            renderDocumentSplitResult(output, result.data);
            state.history = await safeGetHistory();
            renderRecent();
        } catch (error) { renderError(output, `Error: ${error.message}`); }
        finally {
            splitButton.disabled = !filePath.value;
            splitButton.textContent = copy.split;
        }
    };
}

function renderDocumentSplitResult(container, result) {
    const zh = state.settings.language === 'zh';
    container.replaceChildren();
    const summary = document.createElement('div');
    summary.className = 'split-success';
    const heading = document.createElement('strong');
    const stats = document.createElement('span');
    heading.textContent = zh ? '拆分完成' : 'Split complete';
    stats.textContent = zh
        ? `共 ${Number(result.totalPages)} 页 · 奇数页 ${Number(result.oddPages)} 页 · 偶数页 ${Number(result.evenPages)} 页`
        : `${Number(result.totalPages)} pages · ${Number(result.oddPages)} odd · ${Number(result.evenPages)} even`;
    summary.append(heading, stats);
    container.append(summary);
    if (result.converted) {
        const conversion = document.createElement('p');
        conversion.className = 'hint';
        conversion.textContent = zh ? `已通过 ${result.converter || 'Office'} 按实际版式转换为 PDF。` : `Rendered to PDF with ${result.converter || 'Office'}.`;
        container.append(conversion);
    }
    const files = zh ? [['奇数页 PDF', result.oddFile, result.oddPages], ['偶数页 PDF', result.evenFile, result.evenPages]] : [['Odd pages PDF', result.oddFile, result.oddPages], ['Even pages PDF', result.evenFile, result.evenPages]];
    for (const [label, path, count] of files) {
        if (!path) continue;
        const item = document.createElement('div');
        item.className = 'split-output-file';
        const text = document.createElement('div');
        const title = document.createElement('strong');
        const location = document.createElement('code');
        title.textContent = zh ? `${label}（${count} 页）` : `${label} (${count} pages)`;
        location.textContent = path;
        text.append(title, location);
        const copy = document.createElement('button');
        copy.className = 'btn small';
        copy.textContent = zh ? '复制路径' : 'Copy path';
        copy.onclick = async () => { await copyText(path); showToast(t('copied')); };
        item.append(text, copy);
        container.append(item);
    }
}

function renderHIITTimer(container) {
    const zh = state.settings.language === 'zh';
    const copy = zh ? {
        work: '运动', rest: '休息', rounds: '组数', seconds: '秒', sound: '阶段提示音',
        start: '开始训练', pause: '暂停', resume: '继续', restart: '再次开始', reset: '重置',
        ready: '准备开始', complete: '训练完成', round: '第 {current} / {total} 组', total: '总剩余',
        plan: '训练设置', presets: '快捷方案', tabata: 'Tabata 20 / 10 × 8', balanced: '30 / 15 × 10', endurance: '45 / 15 × 8',
    } : {
        work: 'Work', rest: 'Rest', rounds: 'Rounds', seconds: 'sec', sound: 'Sound cues',
        start: 'Start workout', pause: 'Pause', resume: 'Resume', restart: 'Start again', reset: 'Reset',
        ready: 'Ready', complete: 'Workout complete', round: 'Round {current} / {total}', total: 'Total left',
        plan: 'Workout setup', presets: 'Presets', tabata: 'Tabata 20 / 10 × 8', balanced: '30 / 15 × 10', endurance: '45 / 15 × 8',
    };

    container.innerHTML = `
        <div class="hiit-timer">
            <section class="hiit-config-panel">
                <div class="hiit-section-heading"><span>${copy.plan}</span><small id="hiitPlanSummary"></small></div>
                <div class="hiit-presets" aria-label="${copy.presets}">
                    <button class="hiit-preset active" type="button" data-work="20" data-rest="10" data-rounds="8">${copy.tabata}</button>
                    <button class="hiit-preset" type="button" data-work="30" data-rest="15" data-rounds="10">${copy.balanced}</button>
                    <button class="hiit-preset" type="button" data-work="45" data-rest="15" data-rounds="8">${copy.endurance}</button>
                </div>
                <div class="hiit-fields">
                    <label><span>${copy.work}</span><div><input id="hiitWork" type="number" min="1" max="3600" value="20"><small>${copy.seconds}</small></div></label>
                    <label><span>${copy.rest}</span><div><input id="hiitRest" type="number" min="1" max="3600" value="10"><small>${copy.seconds}</small></div></label>
                    <label><span>${copy.rounds}</span><div><input id="hiitRounds" type="number" min="1" max="99" value="8"></div></label>
                </div>
                <label class="hiit-sound"><input id="hiitSound" type="checkbox" checked><span>${copy.sound}</span></label>
                <div class="hiit-actions">
                    <button id="hiitStartPause" type="button" class="btn primary">${copy.start}</button>
                    <button id="hiitReset" type="button" class="btn" disabled>${copy.reset}</button>
                </div>
            </section>
            <section id="hiitStage" class="hiit-stage" data-phase="ready" aria-live="polite">
                <div class="hiit-stage-meta"><span id="hiitRound"></span><span id="hiitTotal"></span></div>
                <div id="hiitRing" class="hiit-ring">
                    <div class="hiit-orb" aria-hidden="true"><span></span></div>
                    <div class="hiit-readout">
                        <span id="hiitPhase">${copy.ready}</span>
                        <strong id="hiitTime">00:20</strong>
                    </div>
                </div>
                <div class="hiit-track"><span id="hiitTrackFill"></span></div>
            </section>
        </div>`;

    const workInput = document.getElementById('hiitWork');
    const restInput = document.getElementById('hiitRest');
    const roundsInput = document.getElementById('hiitRounds');
    const soundInput = document.getElementById('hiitSound');
    const startPause = document.getElementById('hiitStartPause');
    const resetButton = document.getElementById('hiitReset');
    const stage = document.getElementById('hiitStage');
    const ring = document.getElementById('hiitRing');
    const phaseText = document.getElementById('hiitPhase');
    const timeText = document.getElementById('hiitTime');
    const roundText = document.getElementById('hiitRound');
    const totalText = document.getElementById('hiitTotal');
    const trackFill = document.getElementById('hiitTrackFill');
    const planSummary = document.getElementById('hiitPlanSummary');
    const presets = [...container.querySelectorAll('.hiit-preset')];

    let timeline;
    let running = false;
    let elapsedBase = 0;
    let startedAt = 0;
    let frame = 0;
    let lastSegment = -1;
    let lastCountdownSecond = -1;
    let audioContext = null;

    const readTimeline = () => createHIITTimeline({ workSeconds: workInput.value, restSeconds: restInput.value, rounds: roundsInput.value });
    const setText = (element, value) => { if (element.textContent !== value) element.textContent = value; };
    const setConfigEnabled = enabled => {
        for (const control of [workInput, restInput, roundsInput, ...presets]) control.disabled = !enabled;
    };
    const tone = (frequency, offset = 0, duration = .09) => {
        if (!soundInput.checked) return;
        const AudioContext = window.AudioContext || window.webkitAudioContext;
        if (!AudioContext) return;
        audioContext ||= new AudioContext();
        const oscillator = audioContext.createOscillator();
        const gain = audioContext.createGain();
        const start = audioContext.currentTime + offset;
        oscillator.frequency.value = frequency;
        gain.gain.setValueAtTime(.0001, start);
        gain.gain.exponentialRampToValueAtTime(.16, start + .015);
        gain.gain.exponentialRampToValueAtTime(.0001, start + duration);
        oscillator.connect(gain).connect(audioContext.destination);
        oscillator.start(start);
        oscillator.stop(start + duration + .02);
    };
    const announceTransition = snapshot => {
        if (snapshot.complete) {
            tone(660, 0, .12); tone(820, .16, .12); tone(1040, .32, .18);
        } else {
            tone(snapshot.phase === 'work' ? 880 : 520, 0, .14);
        }
    };
    const renderSnapshot = (elapsed, status = running ? 'running' : 'paused') => {
        const snapshot = resolveHIITProgress(timeline, elapsed);
        const displayPhase = elapsed === 0 && !running ? 'ready' : snapshot.phase;
        stage.dataset.phase = displayPhase;
        stage.classList.toggle('paused', status === 'paused' && elapsed > 0 && !snapshot.complete);
        setText(phaseText, displayPhase === 'ready' ? copy.ready : displayPhase === 'complete' ? copy.complete : copy[displayPhase]);
        setText(timeText, displayPhase === 'ready' ? formatTimer(timeline.config.workSeconds * 1000) : formatTimer(snapshot.remainingMs));
        setText(roundText, copy.round.replace('{current}', String(snapshot.round)).replace('{total}', String(timeline.config.rounds)));
        setText(totalText, `${copy.total} ${formatTimer(snapshot.totalRemainingMs)}`);
        ring.style.setProperty('--hiit-progress', `${snapshot.progress * 360}deg`);
        trackFill.style.width = `${Math.min(100, (elapsed / timeline.totalMs) * 100)}%`;
        if (running && !snapshot.complete && snapshot.segmentIndex !== lastSegment) {
            if (lastSegment >= 0) announceTransition(snapshot);
            lastSegment = snapshot.segmentIndex;
            lastCountdownSecond = -1;
        }
        const countdownSecond = Math.ceil(snapshot.remainingMs / 1000);
        if (running && !snapshot.complete && countdownSecond <= 3 && countdownSecond > 0 && countdownSecond !== lastCountdownSecond) {
            tone(700);
            lastCountdownSecond = countdownSecond;
        }
        if (snapshot.complete && running) {
            running = false;
            elapsedBase = timeline.totalMs;
            cancelAnimationFrame(frame);
            announceTransition(snapshot);
            startPause.textContent = copy.restart;
            resetButton.disabled = false;
            setConfigEnabled(true);
        }
        return snapshot;
    };
    const tick = now => {
        if (!running) return;
        const elapsed = Math.min(timeline.totalMs, elapsedBase + now - startedAt);
        const snapshot = renderSnapshot(elapsed, 'running');
        if (!snapshot.complete) frame = requestAnimationFrame(tick);
    };
    const reset = () => {
        running = false;
        cancelAnimationFrame(frame);
        elapsedBase = 0;
        startedAt = 0;
        lastSegment = -1;
        lastCountdownSecond = -1;
        timeline = readTimeline();
        workInput.value = timeline.config.workSeconds;
        restInput.value = timeline.config.restSeconds;
        roundsInput.value = timeline.config.rounds;
        planSummary.textContent = `${timeline.config.workSeconds}s / ${timeline.config.restSeconds}s × ${timeline.config.rounds} · ${formatTimer(timeline.totalMs)}`;
        startPause.textContent = copy.start;
        resetButton.disabled = true;
        setConfigEnabled(true);
        renderSnapshot(0, 'ready');
    };
    const toggle = () => {
        if (elapsedBase >= timeline.totalMs) reset();
        if (running) {
            elapsedBase = Math.min(timeline.totalMs, elapsedBase + performance.now() - startedAt);
            running = false;
            cancelAnimationFrame(frame);
            startPause.textContent = copy.resume;
            resetButton.disabled = false;
            renderSnapshot(elapsedBase, 'paused');
            return;
        }
        running = true;
        startedAt = performance.now();
        startPause.textContent = copy.pause;
        resetButton.disabled = false;
        setConfigEnabled(false);
        if (elapsedBase === 0) tone(880, 0, .14);
        frame = requestAnimationFrame(tick);
    };

    for (const input of [workInput, restInput, roundsInput]) input.onchange = () => {
        presets.forEach(button => button.classList.remove('active'));
        reset();
    };
    for (const preset of presets) preset.onclick = () => {
        workInput.value = preset.dataset.work;
        restInput.value = preset.dataset.rest;
        roundsInput.value = preset.dataset.rounds;
        presets.forEach(button => button.classList.toggle('active', button === preset));
        reset();
    };
    startPause.onclick = toggle;
    resetButton.onclick = reset;
    reset();

    activeToolCleanup = () => {
        running = false;
        cancelAnimationFrame(frame);
        if (audioContext) void audioContext.close();
    };
}

function renderHTTPClient(container, tool) {
    container.innerHTML = `<div class="http-client">
        <div class="http-request-line"><select id="httpMethod" class="select-input">${['GET','POST','PUT','PATCH','DELETE','HEAD','OPTIONS'].map(v=>`<option>${v}</option>`).join('')}</select><input id="httpUrl" class="input-field" placeholder="https://api.example.com/users"><button id="httpSend" class="btn primary">Send</button></div>
        <div class="http-tabs"><button data-http-tab="params" class="active">Params</button><button data-http-tab="headers">Headers</button><button data-http-tab="body">Body</button><button data-http-tab="auth">Auth</button></div>
        <div class="http-config">
            <section data-http-panel="params"><textarea id="httpParams" aria-label="Query parameters" placeholder="key=value&#10;page=1"></textarea></section>
            <section data-http-panel="headers" class="hidden"><textarea id="httpHeaders" aria-label="Request headers" placeholder="Content-Type: application/json&#10;User-Agent: Developer Toolbox"></textarea></section>
            <section data-http-panel="body" class="hidden"><select id="httpBodyType" class="select-input" aria-label="Body type"><option value="none">None</option><option value="json">JSON</option><option value="form">Form</option><option value="raw">Raw</option></select><p id="httpBodyHint" class="http-panel-hint">This request has no body.</p><textarea id="httpBody" class="hidden" aria-label="Request body"></textarea></section>
            <section data-http-panel="auth" class="hidden"><select id="httpAuthType" class="select-input" aria-label="Authentication type"><option value="none">None</option><option value="basic">Basic Auth</option><option value="bearer">Bearer Token</option></select><p id="httpAuthHint" class="http-panel-hint">No authentication will be sent.</p><input id="httpUsername" class="input-field hidden" autocomplete="username" aria-label="Username" placeholder="Username"><input id="httpPassword" type="password" class="input-field hidden" autocomplete="current-password" aria-label="Password" placeholder="Password"><input id="httpToken" type="password" class="input-field hidden" autocomplete="off" aria-label="Bearer token" placeholder="Bearer token"></section>
        </div>
        <div class="http-response-header"><strong id="httpStatus">Response</strong><span><button id="favoriteRequestBtn" class="btn small">☆ Save request</button> <span id="httpMetrics"></span></span></div>
        <div class="http-tabs"><button data-response-tab="pretty" class="active">Pretty JSON</button><button data-response-tab="raw">Raw</button><button data-response-tab="headers">Headers</button><button data-response-tab="preview">Preview</button></div>
        <div id="outputText" class="editor-output http-response"><p class="placeholder">Send a request to see the response.</p></div>
    </div>`;
    let response = null; let responseMode = 'pretty';
    // 请求体和认证字段按当前类型显示，减少无效输入并避免混淆实际会发送的内容。
    const updateBodyControls=()=>{const type=document.getElementById('httpBodyType').value;const body=document.getElementById('httpBody');const enabled=type!=='none';body.classList.toggle('hidden',!enabled);document.getElementById('httpBodyHint').classList.toggle('hidden',enabled);body.placeholder={json:'{\n  "key": "value"\n}',form:'key=value&name=developer',raw:'Request body'}[type]||'';};
    const updateAuthControls=()=>{const type=document.getElementById('httpAuthType').value;const basic=type==='basic';const bearer=type==='bearer';document.getElementById('httpUsername').classList.toggle('hidden',!basic);document.getElementById('httpPassword').classList.toggle('hidden',!basic);document.getElementById('httpToken').classList.toggle('hidden',!bearer);document.getElementById('httpAuthHint').classList.toggle('hidden',type!=='none');};
    document.getElementById('httpBodyType').onchange=updateBodyControls;
    document.getElementById('httpAuthType').onchange=updateAuthControls;
    updateBodyControls();updateAuthControls();
    container.querySelectorAll('[data-http-tab]').forEach(button => button.onclick=()=>{container.querySelectorAll('[data-http-tab]').forEach(v=>v.classList.toggle('active',v===button));container.querySelectorAll('[data-http-panel]').forEach(v=>v.classList.toggle('hidden',v.dataset.httpPanel!==button.dataset.httpTab));});
    container.querySelectorAll('[data-response-tab]').forEach(button => button.onclick=()=>{responseMode=button.dataset.responseTab;container.querySelectorAll('[data-response-tab]').forEach(v=>v.classList.toggle('active',v===button));if(response)renderHTTPResponse(response,responseMode);});
    document.getElementById('httpSend').onclick=async()=>{const output=document.getElementById('outputText');output.textContent='Sending…';const payload={toolId:'http',payload:{url:document.getElementById('httpUrl').value,method:document.getElementById('httpMethod').value,query:parsePairs(document.getElementById('httpParams').value,'='),headers:parsePairs(document.getElementById('httpHeaders').value,':'),bodyType:document.getElementById('httpBodyType').value,body:document.getElementById('httpBody').value,authType:document.getElementById('httpAuthType').value,username:document.getElementById('httpUsername').value,password:document.getElementById('httpPassword').value,token:document.getElementById('httpToken').value}};try{const result=await appService.executeTool(payload);if(!result.success){renderError(output,errorMessage(result.error));return;}response=result.data;document.getElementById('httpStatus').textContent=response.status;document.getElementById('httpMetrics').textContent=`${response.timingMs} ms · ${formatBytes(response.size)}${response.truncated?' · truncated':''}`;renderHTTPResponse(response,responseMode);}catch(error){renderError(output,error.message)}};
    document.getElementById('favoriteRequestBtn').onclick=()=>favoriteRequest(tool);
}

function parsePairs(raw, delimiter) { const result={};for(const line of raw.split('\n')){const index=line.indexOf(delimiter);if(index>0)result[line.slice(0,index).trim()]=line.slice(index+delimiter.length).trim();}return result; }
function formatBytes(value){if(value<1024)return `${value} B`;if(value<1048576)return `${(value/1024).toFixed(1)} KB`;return `${(value/1048576).toFixed(1)} MB`;}
function renderHTTPResponse(response,mode){const output=document.getElementById('outputText');output.replaceChildren();if(mode==='headers'){output.textContent=JSON.stringify(response.headers,null,2);return;}if(mode==='preview'){renderSandboxPreview(output,response.body);return;}if(mode==='pretty'){try{output.textContent=JSON.stringify(JSON.parse(response.body),null,2);return;}catch{}}output.textContent=response.body;}

function setupToolEditor(tool) {
    const inputOptions = document.getElementById('inputOptions');
    const inputText = document.getElementById('inputText');
    
    // Tool-specific input options and behaviors
    // 页面框架由所有工具共享，这里只注入各工具特有的选项和事件。
    switch (tool.id) {
        case 'json':
            inputOptions.innerHTML = `
                <select id="jsonAction" class="select-input">
                    <option value="format">Format</option>
                    <option value="minify">Minify</option>
                    <option value="validate">Validate</option>
                    <option value="tree">Tree</option>
                </select>
                <select id="jsonIndent" class="select-input">
                    <option value="2">2 spaces</option>
                    <option value="4">4 spaces</option>
                </select>
            `;
            document.getElementById('jsonAction').onchange = () => executeToolCommand(tool.id);
            document.getElementById('jsonIndent').value = String(state.settings.jsonIndent || 2);
            document.getElementById('jsonIndent').onchange = () => executeToolCommand(tool.id);
            inputText.oninput = () => executeToolCommand(tool.id);
            break;
        
        case 'base64':
            inputOptions.innerHTML = `
                <select id="base64Action" class="select-input">
                    <option value="encode">Encode</option>
                    <option value="decode">Decode</option>
                </select>
            `;
            document.getElementById('base64Action').onchange = () => executeToolCommand(tool.id);
            inputText.oninput = () => executeToolCommand(tool.id);
            break;
        
        case 'url':
            inputOptions.innerHTML = `
                <select id="urlAction" class="select-input">
                    <option value="encode">Encode</option>
                    <option value="decode">Decode</option>
                    <option value="encodeQuery">Normalize Query String</option>
                    <option value="decodeQuery">Parse Query String</option>
                </select>
            `;
            document.getElementById('urlAction').onchange = () => executeToolCommand(tool.id);
            inputText.oninput = () => executeToolCommand(tool.id);
            break;
        
        case 'hash':
            inputOptions.innerHTML = `
                <select id="hashAlgo" class="select-input">
                    <option value="md5">MD5</option>
                    <option value="sha1">SHA1</option>
                    <option value="sha256">SHA256</option>
                    <option value="sha512">SHA512</option>
                </select>
                <button class="btn small" id="hashFileBtn">Select File</button>
                <span id="hashFileName" class="hint"></span>
            `;
            document.getElementById('hashAlgo').onchange = () => executeToolCommand(tool.id);
            document.getElementById('hashFileBtn').onclick = async () => {
                try {
                    const filePath = await appService.selectFile();
                    if (filePath) {
                        inputText.dataset.filePath = filePath;
                        document.getElementById('hashFileName').textContent = filePath;
                        executeToolCommand(tool.id);
                    }
                } catch (err) {
                    showToast(`Unable to select file: ${err.message}`);
                }
            };
            inputText.oninput = () => executeToolCommand(tool.id);
            break;
        
        case 'uuid':
            inputOptions.innerHTML = `
                <input type="number" id="uuidCount" class="input-field" value="1" min="1" max="100" placeholder="Count" />
                <button class="btn small" id="generateUuidBtn">Generate</button>
            `;
            document.getElementById('generateUuidBtn').onclick = () => executeToolCommand(tool.id);
            break;
        
        case 'timestamp':
            inputOptions.innerHTML = `
                <select id="timestampAction" class="select-input">
                    <option value="toDate">Unix → Date</option>
                    <option value="toTimestamp">Date → Unix</option>
                </select>
                <select id="timestampZone" class="select-input">
                    <option value="local">Local</option>
                    <option value="utc">UTC</option>
                </select>
            `;
            document.getElementById('timestampAction').onchange = () => executeToolCommand(tool.id);
            document.getElementById('timestampZone').onchange = () => executeToolCommand(tool.id);
            inputText.oninput = () => executeToolCommand(tool.id);
            break;
        
        case 'jwt':
            inputOptions.innerHTML = `<p style="color: #999; font-size: 12px;">Paste JWT token to decode</p>`;
            inputText.oninput = () => executeToolCommand(tool.id);
            break;
        
        case 'regex':
            inputOptions.innerHTML = `
                <select id="regexAction" class="select-input">
                    <option value="match">Match</option>
                    <option value="replace">Replace</option>
                </select>
                <input type="text" id="regexPattern" class="input-field" placeholder="Regex pattern" />
                <input type="text" id="regexFlags" class="input-field" placeholder="Flags (gims)" />
                <input type="text" id="regexReplacement" class="input-field" placeholder="Replacement (optional)" />
            `;
            document.getElementById('regexAction').onchange = () => executeToolCommand(tool.id);
            document.getElementById('regexPattern').oninput = () => executeToolCommand(tool.id);
            document.getElementById('regexFlags').oninput = () => executeToolCommand(tool.id);
            document.getElementById('regexReplacement').oninput = () => executeToolCommand(tool.id);
            inputText.oninput = () => executeToolCommand(tool.id);
            break;
        
        case 'yaml':
            inputOptions.innerHTML = `
                <select id="yamlAction" class="select-input">
                    <option value="jsonToYaml">JSON → YAML</option>
                    <option value="yamlToJson">YAML → JSON</option>
                    <option value="validate">Validate YAML</option>
                </select>
            `;
            document.getElementById('yamlAction').onchange = () => executeToolCommand(tool.id);
            inputText.oninput = () => executeToolCommand(tool.id);
            break;

        case 'xml':
            inputOptions.innerHTML = actionSelect('xmlAction', [['format','Format'],['validate','Validate'],['xmlToJson','XML → JSON'],['jsonToXml','JSON → XML']]);
            bindAutoExecute(tool.id, ['xmlAction']);
            break;
        case 'csv':
            inputOptions.innerHTML = `${actionSelect('csvAction', [['view','Viewer'],['format','Format'],['csvToJson','CSV → JSON'],['jsonToCsv','JSON → CSV']])}<select id="csvSeparator" class="select-input"><option value=",">Comma</option><option value=";">Semicolon</option><option value="tab">Tab</option></select><label class="option-check"><input id="csvHeader" type="checkbox" checked> Header</label>`;
            bindAutoExecute(tool.id, ['csvAction','csvSeparator','csvHeader']);
            break;
        case 'toml':
            inputOptions.innerHTML = actionSelect('tomlAction', [['format','Format'],['validate','Validate'],['tomlToJson','TOML → JSON'],['jsonToToml','JSON → TOML']]);
            bindAutoExecute(tool.id, ['tomlAction']);
            break;
        case 'text':
            inputOptions.innerHTML = `${actionSelect('textAction', [['trim','Trim'],['removeEmpty','Remove empty lines'],['deduplicate','Remove duplicate lines'],['sort','Sort lines'],['reverse','Reverse lines'],['prefix','Add prefix'],['suffix','Add suffix'],['jsonEscape','JSON escape'],['htmlEscape','HTML escape'],['urlEscape','URL escape'],['unicodeEscape','Unicode escape'],['statistics','Statistics']])}<input id="textValue" class="input-field" placeholder="Prefix / suffix">`;
            bindAutoExecute(tool.id, ['textAction','textValue']);
            break;
        case 'sql':
            inputOptions.innerHTML = `${actionSelect('sqlAction', [['format','Format'],['beautify','Beautify'],['minify','Minify']])}<select id="sqlIndent" class="select-input"><option value="2">2 spaces</option><option value="4">4 spaces</option></select><select id="sqlKeyword" class="select-input"><option value="upper">UPPERCASE</option><option value="lower">lowercase</option></select><select id="sqlComma" class="select-input"><option value="trailing">Trailing comma</option><option value="leading">Leading comma</option></select>`;
            bindAutoExecute(tool.id, ['sqlAction','sqlIndent','sqlKeyword','sqlComma']);
            break;
        case 'markdown':
            inputOptions.innerHTML = actionSelect('markdownAction', [['toHtml','Markdown → HTML'],['preview','Preview']]);
            bindAutoExecute(tool.id, ['markdownAction']);
            break;
        case 'html':
            inputOptions.innerHTML = actionSelect('htmlAction', [['format','Format'],['minify','Minify'],['escape','Escape'],['preview','Preview']]);
            bindAutoExecute(tool.id, ['htmlAction']);
            break;
        case 'dns':
            inputOptions.innerHTML = `<select id="dnsRecord" class="select-input">${['A','AAAA','CNAME','MX','TXT','NS'].map(v=>`<option>${v}</option>`).join('')}</select><button id="networkRun" class="btn small">Lookup</button>`;
            document.getElementById('networkRun').onclick=()=>executeToolCommand(tool.id); break;
        case 'ping':
            inputOptions.innerHTML = `<input id="pingCount" type="number" min="1" max="10" value="4" class="input-field" title="Count"><input id="networkTimeout" type="number" min="500" max="10000" value="3000" class="input-field" title="Timeout ms"><input id="pingInterval" type="number" min="200" max="5000" value="1000" class="input-field" title="Interval ms"><button id="networkRun" class="btn small">Ping</button>`;
            document.getElementById('networkRun').onclick=()=>executeToolCommand(tool.id); break;
        case 'port':
            inputOptions.innerHTML = `<input id="portNumber" type="number" min="1" max="65535" value="443" class="input-field" title="Port"><input id="networkTimeout" type="number" min="100" max="30000" value="3000" class="input-field" title="Timeout ms"><button id="networkRun" class="btn small">Check</button>`;
            document.getElementById('networkRun').onclick=()=>executeToolCommand(tool.id); break;
        case 'cidr': case 'url-parser':
            inputOptions.innerHTML = `<button id="networkRun" class="btn small">${tool.id==='cidr'?'Calculate':'Parse'}</button>`;
            document.getElementById('networkRun').onclick=()=>executeToolCommand(tool.id); break;
        case 'cron':
            inputOptions.innerHTML = `${actionSelect('cronAction', [['parse','Parse'],['generate','Generate']])}<div id="cronFields" class="cron-fields hidden"><input id="cronMinute" class="input-field" value="0" title="Minute"><input id="cronHour" class="input-field" value="9" title="Hour"><input id="cronDay" class="input-field" value="*" title="Day"><input id="cronMonth" class="input-field" value="*" title="Month"><input id="cronWeekday" class="input-field" value="1-5" title="Weekday"></div><button id="developerRun" class="btn small">Run</button>`;
            document.getElementById('cronAction').onchange=e=>document.getElementById('cronFields').classList.toggle('hidden',e.target.value!=='generate');document.getElementById('developerRun').onclick=()=>executeToolCommand(tool.id);break;
        case 'color':
            inputOptions.innerHTML = `<input id="colorPicker" type="color" value="#1890ff" title="Choose color"><button id="developerRun" class="btn small">Convert</button>`;document.getElementById('colorPicker').oninput=e=>document.getElementById('toolEditor').editor.value=e.target.value;document.getElementById('developerRun').onclick=()=>executeToolCommand(tool.id);break;
        case 'code-generator':
            inputOptions.innerHTML = `<input id="codeName" class="input-field" value="Root" placeholder="Type name"><select id="codeLanguage" class="select-input"><option value="go">Go Struct</option><option value="typescript">TypeScript Interface</option><option value="java">Java Class</option></select><button id="developerRun" class="btn small">Generate</button>`;document.getElementById('developerRun').onclick=()=>executeToolCommand(tool.id);break;
        case 'git':
            inputOptions.innerHTML = `${actionSelect('gitAction', [['gitignore','Gitignore'],['command','Command'],['parseUrl','Parse URL']])}<select id="gitCommand" class="select-input hidden">${[['undo-last-commit','Undo last commit'],['amend-commit','Amend commit'],['create-branch','Create branch'],['delete-branch','Delete branch'],['stash','Stash'],['unstash','Unstash'],['uncommit-file','Unstage file'],['show-history','Show history']].map(([v,l])=>`<option value="${v}">${l}</option>`).join('')}</select><div id="gitTemplates" class="template-checks">${['go','node','python','java','cpp','rust','react','vue','vscode','intellij','windows','macos','linux'].map(v=>`<label><input type="checkbox" value="${v}">${v}</label>`).join('')}</div><button id="developerRun" class="btn small">Generate</button>`;
            document.getElementById('gitAction').onchange=e=>{document.getElementById('gitCommand').classList.toggle('hidden',e.target.value!=='command');document.getElementById('gitTemplates').classList.toggle('hidden',e.target.value!=='gitignore');};document.getElementById('developerRun').onclick=()=>executeToolCommand(tool.id);break;
        case 'amount-cn':
            inputOptions.innerHTML=`${actionSelect('amountAction',[['upper','中文大写'],['lower','普通中文数字']])}<button id="extendedRun" class="btn small">转换</button>`;document.getElementById('extendedRun').onclick=()=>executeToolCommand(tool.id);break;
        case 'pinyin':
            inputOptions.innerHTML=`${actionSelect('pinyinMode',[['full','全拼'],['initial','首字母']])}<input id="pinyinSeparator" class="input-field" value=" " placeholder="分隔符"><button id="extendedRun" class="btn small">转换</button>`;document.getElementById('extendedRun').onclick=()=>executeToolCommand(tool.id);break;
        case 'tcp': case 'udp':
            inputOptions.innerHTML=`<input id="socketHost" class="input-field" value="127.0.0.1" placeholder="主机"><input id="socketPort" type="number" min="1" max="65535" class="input-field" value="9000"><select id="socketEncoding" class="select-input"><option value="text">文本发送</option><option value="hex">HEX 发送</option></select><select id="socketResponseEncoding" class="select-input"><option value="text">文本响应</option><option value="hex">HEX 响应</option></select><input id="socketTimeout" type="number" min="100" max="30000" value="3000" class="input-field"><button id="extendedRun" class="btn small">发送</button>`;document.getElementById('extendedRun').onclick=()=>executeToolCommand(tool.id);break;
        case 'serial':
            inputOptions.innerHTML=`${actionSelect('serialAction',[['list','枚举串口'],['transfer','收发数据']])}<input id="serialName" class="input-field" placeholder="COM3"><select id="serialBaud" class="select-input">${[9600,19200,38400,57600,115200].map(v=>`<option>${v}</option>`).join('')}</select><select id="serialEncoding" class="select-input"><option value="text">文本</option><option value="hex">HEX</option></select><button id="extendedRun" class="btn small">执行</button>`;document.getElementById('extendedRun').onclick=()=>executeToolCommand(tool.id);break;
        case 'database-docs':
            inputOptions.innerHTML=`<select id="databaseType" class="select-input"><option value="postgresql">PostgreSQL</option><option value="mysql">MySQL</option><option value="oracle">Oracle</option><option value="sqlserver">SQL Server</option></select><button id="extendedRun" class="btn small">查询文档</button>`;document.getElementById('extendedRun').onclick=()=>executeToolCommand(tool.id);break;
    }
}

function actionSelect(id, options) { return `<select id="${id}" class="select-input">${options.map(([value,label]) => `<option value="${value}">${label}</option>`).join('')}</select>`; }
function bindAutoExecute(toolId, controlIds) {
    for (const id of controlIds) { const control = document.getElementById(id); if (control) control[control.tagName === 'INPUT' && control.type !== 'checkbox' ? 'oninput' : 'onchange'] = () => executeToolCommand(toolId); }
    document.getElementById('inputText').oninput = () => executeToolCommand(toolId);
}

function renderSettingsView(main) {
    main.innerHTML = `
        <div class="settings-wrapper">
            <div class="settings-header">
                <div><span class="eyebrow">PREFERENCES</span><h2>${t('settings')}</h2></div>
                <button class="btn" id="closeSettingsBtn">← ${t('back')}</button>
            </div>
            <div class="settings-container">
                <div class="settings-section">
                    <h3>${t('appearance')}</h3>
                    <div class="settings-group">
                        <label>${t('theme')}</label>
                        <select id="themeSelect" class="select-input">
                            <option value="dark">${t('dark')}</option>
                            <option value="light">${t('light')}</option>
                            <option value="system">${t('system')}</option>
                        </select>
                    </div>
                </div>

                <div class="settings-section">
                    <h3>${t('editor')}</h3>
                    <div class="settings-group">
                        <label>${t('indent')}</label>
                        <select id="jsonIndentSetting" class="select-input">
                            <option value="2">2 spaces</option>
                            <option value="4">4 spaces</option>
                        </select>
                    </div>
                </div>
                
                <div class="settings-section">
                    <h3>${t('general')}</h3>
                    <div class="settings-group">
                        <label>${t('language')}</label>
                        <select id="languageSelect" class="select-input">
                            <option value="en">English</option>
                            <option value="zh">中文</option>
                        </select>
                    </div>
                </div>
                
                <div class="settings-section">
                    <h3>${t('storage')}</h3>
                    <div class="settings-group">
                        <label>${t('storagePath')}</label>
                        <input type="text" id="storagePathInput" class="input-field" placeholder="${t('storageHint')}" />
                    </div>
                    <p class="hint">${t('storageHint')}</p>
                </div>
                
                <div class="settings-actions">
                    <button class="btn primary" id="saveSettingsBtn">${t('save')}</button>
                    <button class="btn" id="resetSettingsBtn">${t('reset')}</button>
                </div>
            </div>
        </div>
    `;
    
    // Load current settings into form
    document.getElementById('themeSelect').value = state.settings.theme;
    document.getElementById('languageSelect').value = state.settings.language;
    document.getElementById('storagePathInput').value = state.settings.storagePath || '';
    document.getElementById('jsonIndentSetting').value = String(state.settings.jsonIndent || 2);
    
    // Setup event handlers
    document.getElementById('closeSettingsBtn').onclick = () => goHome();
    document.getElementById('themeSelect').onchange = (e) => {
        applyTheme(e.target.value);
    };
    document.getElementById('languageSelect').onchange = (e) => {
        state.settings.language = e.target.value;
        appRoot.innerHTML = getAppShell();
        document.getElementById('logo').src = logo;
        setupEventListeners();
        render();
        goSettings();
    };
    document.getElementById('storagePathInput').oninput = (e) => {
        state.settings.storagePath = e.target.value;
    };
    document.getElementById('jsonIndentSetting').onchange = (e) => {
        state.settings.jsonIndent = Number(e.target.value);
    };
    document.getElementById('saveSettingsBtn').onclick = () => saveSettings();
    document.getElementById('resetSettingsBtn').onclick = () => {
        state.settings = {
            theme: 'dark',
            language: 'en',
            storagePath: '',
            jsonIndent: 2,
        };
        applyTheme('dark');
        appRoot.innerHTML = getAppShell();
        document.getElementById('logo').src = logo;
        setupEventListeners();
        render();
    };
}

async function executeToolCommand(toolId) {
    const inputText = document.getElementById('inputText').value;
    const outputDiv = document.getElementById('outputText');
    
    const selectedHashFile = toolId === 'hash' && document.getElementById('inputText')?.dataset.filePath;
    const allowsEmptyInput = ['uuid', 'serial', 'database-docs'].includes(toolId);
    if (!inputText.trim() && !allowsEmptyInput && !selectedHashFile) {
        outputDiv.innerHTML = '<p class="placeholder">Enter input to see output...</p>';
        return;
    }
    
    // Build tool-specific payload
    // 所有工具统一构造 ToolInput，仅在 switch 中追加工具专属参数。
    let payload = {
        toolId: toolId,
        payload: {
            input: inputText,
        },
    };
    
    // Add tool-specific options
    switch (toolId) {
        case 'json':
            payload.payload.action = document.getElementById('jsonAction')?.value || 'format';
            payload.payload.indent = parseInt(document.getElementById('jsonIndent')?.value || '2');
            break;
        case 'base64':
            payload.payload.action = document.getElementById('base64Action')?.value || 'encode';
            break;
        case 'url':
            payload.payload.action = document.getElementById('urlAction')?.value || 'encode';
            break;
        case 'hash':
            payload.payload.algorithm = document.getElementById('hashAlgo')?.value || 'md5';
            payload.payload.filePath = document.getElementById('inputText')?.dataset.filePath || '';
            break;
        case 'uuid':
            payload.payload.count = parseInt(document.getElementById('uuidCount')?.value || '1');
            break;
        case 'timestamp':
            payload.payload.action = document.getElementById('timestampAction')?.value || 'toDate';
            payload.payload.zone = document.getElementById('timestampZone')?.value || 'local';
            break;
        case 'regex':
            payload.payload.pattern = document.getElementById('regexPattern')?.value || '';
            payload.payload.flags = document.getElementById('regexFlags')?.value || '';
            payload.payload.replacement = document.getElementById('regexReplacement')?.value || '';
            payload.payload.action = document.getElementById('regexAction')?.value || 'match';
            break;
        case 'yaml':
            payload.payload.action = document.getElementById('yamlAction')?.value || 'jsonToYaml';
            break;
        case 'xml': payload.payload.action = document.getElementById('xmlAction').value; break;
        case 'csv': payload.payload.action = document.getElementById('csvAction').value; payload.payload.separator = document.getElementById('csvSeparator').value; payload.payload.header = document.getElementById('csvHeader').checked; break;
        case 'toml': payload.payload.action = document.getElementById('tomlAction').value; break;
        case 'text': payload.payload.action = document.getElementById('textAction').value; payload.payload.value = document.getElementById('textValue').value; break;
        case 'sql': payload.payload.action = document.getElementById('sqlAction').value; payload.payload.indent = Number(document.getElementById('sqlIndent').value); payload.payload.keyword = document.getElementById('sqlKeyword').value; payload.payload.comma = document.getElementById('sqlComma').value; break;
        case 'markdown': payload.payload.action = document.getElementById('markdownAction').value; break;
        case 'html': payload.payload.action = document.getElementById('htmlAction').value; break;
        case 'dns': payload.payload.host=inputText;payload.payload.recordType=document.getElementById('dnsRecord').value;break;
        case 'ping': payload.payload.host=inputText;payload.payload.count=Number(document.getElementById('pingCount').value);payload.payload.timeoutMs=Number(document.getElementById('networkTimeout').value);payload.payload.intervalMs=Number(document.getElementById('pingInterval').value);break;
        case 'port': payload.payload.host=inputText;payload.payload.port=Number(document.getElementById('portNumber').value);payload.payload.timeoutMs=Number(document.getElementById('networkTimeout').value);break;
        case 'cron': payload.payload.action=document.getElementById('cronAction').value;for(const field of ['Minute','Hour','Day','Month','Weekday'])payload.payload[field.toLowerCase()]=document.getElementById(`cron${field}`).value;break;
        case 'code-generator': payload.payload.name=document.getElementById('codeName').value;payload.payload.language=document.getElementById('codeLanguage').value;break;
        case 'git': payload.payload.action=document.getElementById('gitAction').value;payload.payload.command=document.getElementById('gitCommand').value;payload.payload.templates=[...document.querySelectorAll('#gitTemplates input:checked')].map(v=>v.value);break;
        case 'amount-cn':payload.payload.action=document.getElementById('amountAction').value;break;
        case 'pinyin':payload.payload.mode=document.getElementById('pinyinMode').value;payload.payload.separator=document.getElementById('pinyinSeparator').value;break;
        case 'tcp':case 'udp':payload.payload.host=document.getElementById('socketHost').value;payload.payload.port=Number(document.getElementById('socketPort').value);payload.payload.encoding=document.getElementById('socketEncoding').value;payload.payload.responseEncoding=document.getElementById('socketResponseEncoding').value;payload.payload.timeoutMs=Number(document.getElementById('socketTimeout').value);break;
        case 'serial':payload.payload.action=document.getElementById('serialAction').value;payload.payload.portName=document.getElementById('serialName').value;payload.payload.baudRate=Number(document.getElementById('serialBaud').value);payload.payload.encoding=document.getElementById('serialEncoding').value;break;
        case 'database-docs':payload.payload.database=document.getElementById('databaseType').value;break;
    }
    
    try {
        const result = await appService.executeTool(payload);
        if (result.success) {
            const preview = (toolId === 'markdown' && document.getElementById('markdownAction')?.value === 'preview') || (toolId === 'html' && document.getElementById('htmlAction')?.value === 'preview');
            if (preview) renderSandboxPreview(outputDiv, String(result.data)); else if(toolId==='database-docs')renderDatabaseDocs(outputDiv,result.data);else outputDiv.textContent = formatOutput(result.data);
            outputDiv.className = 'editor-output';
            // ExecuteTool 已在后端落盘，成功后立即刷新侧栏，避免必须重启或返回首页才看到最近使用。
            state.history = await safeGetHistory();
            renderRecent();
        } else {
            renderError(outputDiv, errorMessage(result.error));
        }
    } catch (err) {
        console.error('Tool execution failed:', err);
        renderError(outputDiv, `Error: ${err.message}`);
    }
}

function renderSandboxPreview(container, source) {
    container.replaceChildren();
    const frame = document.createElement('iframe');
    frame.className = 'sandbox-preview';
    frame.sandbox = '';
    frame.title = 'Sandboxed preview';
    frame.srcdoc = `<!doctype html><meta charset="utf-8"><style>body{font:14px/1.6 system-ui;padding:20px;color:#172033}code{background:#eef2ff;padding:2px 5px;border-radius:4px}img{max-width:100%}</style>${source}`;
    container.append(frame);
}

// 数据库文档使用 textContent 构造节点，确保内置 SQL 示例不会被当作 HTML 执行。
function renderDatabaseDocs(container,data){container.replaceChildren();const header=document.createElement('div');header.className='docs-summary';header.textContent=`${String(data.database).toUpperCase()} · ${data.count} 条`;container.append(header);for(const entry of data.entries){const article=document.createElement('article');article.className='database-doc-entry';const title=document.createElement('h4');title.textContent=entry.title;const description=document.createElement('p');description.textContent=entry.description;const syntax=document.createElement('pre');syntax.textContent=entry.syntax;const example=document.createElement('pre');example.textContent=entry.example;article.append(title,description,syntax,example);container.append(article)}}

function renderSidebar() {
    renderFavorites();
    renderRecent();
}

function renderFavorites() {
    const container = document.getElementById('favoriteList');
    container.innerHTML = '';
    if (state.favorites.length === 0) {
        container.innerText = t('noFavorites');
        return;
    }

    const favorites = state.favorites
        .map(id => state.tools.find(tool => tool.id === id))
        .filter(Boolean);

    favorites.forEach(tool => {
        const card = renderToolCard(visibleTool(tool), true);
        container.appendChild(card);
    });
    state.favoriteItems.filter(item => item.kind && item.kind !== 'tool').slice(0, 5).forEach(item => {
        const card=document.createElement('div');card.className='tool-card small saved-workflow';card.innerHTML=`<div><strong>${escapeMarkup(item.name || 'Saved input')}</strong><p>${item.kind === 'request' ? 'Saved request' : 'Saved input'}</p></div><button class="favorite-remove" title="Remove">×</button>`;
        card.onclick=event=>{if(event.target.classList.contains('favorite-remove')){event.stopPropagation();removeFavoriteItem(item.id);return;}restoreFavoriteItem(item);};container.appendChild(card);
    });
}

function renderRecent() {
    const container = document.getElementById('recentList');
    container.innerHTML = '';
    if (state.history.length === 0) {
        container.innerText = t('noRecent');
        return;
    }

    const unique = [];
    const seen = new Set();
    for (const item of state.history) {
        if (!seen.has(item.toolId) && unique.length < 5) {
            seen.add(item.toolId);
            unique.push(item);
        }
    }

    unique.forEach(item => {
        const sourceTool = state.allTools.find(tool => tool.id === item.toolId);
        const tool = sourceTool && visibleTool(sourceTool);
        if (!tool) return;
        const card = document.createElement('div');
        card.className = 'tool-card small';
        card.onclick = () => selectToolAndEdit(tool);
        const inputPreview = summarizeSnapshot(item.input);
        const outputPreview = summarizeSnapshot(item.output);
        card.innerHTML = `
            <div>
                <strong>${tool.name}</strong>
                <p>${new Date(item.usedAt).toLocaleString(state.settings.language === 'zh' ? 'zh-CN' : 'en-US')}</p>
                ${inputPreview ? `<pre>Input: ${escapeMarkup(inputPreview)}</pre>` : ''}
                ${outputPreview ? `<pre>Output: ${escapeMarkup(outputPreview)}</pre>` : ''}
            </div>
        `;
        container.appendChild(card);
    });
}

function summarizeSnapshot(value) { if(value == null)return '';const text=typeof value==='string'?value:JSON.stringify(value);return text.length>100?text.slice(0,100)+'…':text; }
function escapeMarkup(value){const element=document.createElement('span');element.textContent=String(value);return element.innerHTML;}

function renderCategories() {
    const container = document.getElementById('categoryList');
    container.innerHTML = '';
    const matched = filterTools(state.tools, state.search);
    const categories = groupByCategory(matched);

    if (matched.length === 0) {
        container.innerHTML = `<div class="empty-state">${t('noMatches')}</div>`;
        return;
    }

    Object.keys(categories).sort().forEach(category => {
        const section = document.createElement('div');
        section.className = 'category-section';
        section.innerHTML = `
            <div class="category-header">
                <h3>${localizeCategory(category, state.settings.language)}</h3>
                <span>${categories[category].length} ${t('tools')}</span>
            </div>
        `;

        const list = document.createElement('div');
        list.className = 'tool-grid';
        categories[category].forEach(tool => {
            const card = renderToolCard(visibleTool(tool), state.favorites.includes(tool.id));
            list.appendChild(card);
        });
        section.appendChild(list);
        container.appendChild(section);
    });
}

function renderToolCard(tool, favorited) {
    const card = document.createElement('div');
    card.className = 'tool-card';

    const favoriteIcon = favorited ? '★' : '☆';
    card.innerHTML = `
        <div class="tool-card-content">
            <div>
                <strong>${tool.name}</strong>
                <p>${tool.description}</p>
            </div>
            <button class="favorite-btn" title="${t('favoriteTitle')}">${favoriteIcon}</button>
        </div>
    `;

    card.onclick = event => {
        if (event.target.classList.contains('favorite-btn')) {
            event.stopPropagation();
            toggleFavorite(tool.id);
            return;
        }
        selectToolAndEdit(tool);
    };
    return card;
}

// ============== User Actions ==============
async function onSearch(event) {
    state.search = event.target.value.trim().toLowerCase();
    if (state.currentView === 'home') {
        try {
            // 正常运行时由后端 SearchService 保持搜索契约；浏览器预览时回退到本地数据。
            state.tools = await appService.searchTools(state.search);
        } catch (err) {
            state.tools = filterTools(MOCK_TOOLS, state.search);
        }
        renderCategories();
        renderWorkflowSearch(state.search);
        document.getElementById('toolCount').innerText = `${state.tools.length} ${t('tools')}`;
    }
}

function renderWorkflowSearch(query) {
    const container=document.getElementById('workflowSearchResults');if(!container)return;if(!query){container.classList.add('hidden');container.replaceChildren();return;}
    const actions=commandDefinitions().filter(command=>!command.label.startsWith('Open ')&&(command.label+' '+command.keywords).toLowerCase().includes(query)).slice(0,5);
    const history=state.history.filter(item=>{const tool=state.allTools.find(value=>value.id===item.toolId);return `${tool?.name||''} ${JSON.stringify(item.input||{})} ${JSON.stringify(item.output||{})}`.toLowerCase().includes(query);}).slice(0,5);
    if(!actions.length&&!history.length){container.classList.add('hidden');return;}container.classList.remove('hidden');container.innerHTML='<h3>Actions & History</h3>';
    for(const action of actions){const button=document.createElement('button');button.textContent=action.label;button.onclick=action.run;container.append(button);}
    for(const item of history){const tool=state.allTools.find(value=>value.id===item.toolId);const button=document.createElement('button');button.textContent=`History · ${visibleTool(tool||{name:item.toolId}).name} · ${summarizeSnapshot(item.input)}`;button.onclick=()=>tool&&selectToolAndEdit(visibleTool(tool));container.append(button);}
}

function selectToolAndEdit(tool) {
    activeToolCleanup?.();
    activeToolCleanup = null;
    state.selectedTool = tool;
    state.currentView = 'editor';
    render();
}

function goHome() {
    activeToolCleanup?.();
    activeToolCleanup = null;
    state.currentView = 'home';
    state.selectedTool = null;
    render();
}

function goSettings() {
    activeToolCleanup?.();
    activeToolCleanup = null;
    state.currentView = 'settings';
    render();
}

async function toggleFavorite(toolId) {
    try {
        if (state.favorites.includes(toolId)) {
            await appService.removeFavorite(toolId);
        } else {
            await appService.addFavorite(toolId);
        }
        await loadData();
    } catch (err) {
        console.error('Failed to toggle favorite', err);
        showToast('Unable to update favorites.');
    }
}

async function favoriteCurrentInput(tool) {
    const input=document.getElementById('inputText')?.value || '';if(!input.trim()){showToast('Nothing to save.');return;}
    await addFavoriteWorkflow({toolId:tool.id,kind:'input',name:`${tool.name} input`,payload:{input}});
}
async function favoriteRequest(tool) {
    const payload={url:document.getElementById('httpUrl').value,method:document.getElementById('httpMethod').value,query:parsePairs(document.getElementById('httpParams').value,'='),headers:parsePairs(document.getElementById('httpHeaders').value,':'),bodyType:document.getElementById('httpBodyType').value,body:document.getElementById('httpBody').value,authType:document.getElementById('httpAuthType').value,username:document.getElementById('httpUsername').value,password:document.getElementById('httpPassword').value,token:document.getElementById('httpToken').value};
    if(!payload.url.trim()){showToast('Enter a URL first.');return;}await addFavoriteWorkflow({toolId:tool.id,kind:'request',name:`${payload.method} ${payload.url}`,payload});
}
async function addFavoriteWorkflow(item){try{await appService.addFavoriteItem(item);await loadData();showToast('Saved to favorites.');}catch(error){console.error(error);showToast('Unable to save favorite.');}}
async function removeFavoriteItem(id){try{await appService.removeFavoriteItem(id);await loadData();}catch(error){showToast('Unable to remove favorite.');}}
function restoreFavoriteItem(item){const source=state.allTools.find(tool=>tool.id===item.toolId);if(!source)return;selectToolAndEdit(visibleTool(source));queueMicrotask(()=>{if(item.kind==='request'){document.getElementById('httpUrl').value=item.payload.url||'';document.getElementById('httpMethod').value=item.payload.method||'GET';document.getElementById('httpParams').value=serializePairs(item.payload.query,'=');document.getElementById('httpHeaders').value=serializePairs(item.payload.headers,': ');document.getElementById('httpBodyType').value=item.payload.bodyType||'none';document.getElementById('httpBody').value=item.payload.body||'';}else{const editor=document.getElementById('toolEditor').editor;if(editor)editor.value=item.payload.input||'';}});}
function serializePairs(value,delimiter){return Object.entries(value||{}).map(([key,item])=>`${key}${delimiter}${item}`).join('\n');}

function commandDefinitions(){const tools=state.allTools.map(tool=>({label:`Open ${visibleTool(tool).name}`,keywords:`open ${tool.id} ${(tool.keywords||[]).join(' ')}`,run:()=>selectToolAndEdit(visibleTool(tool))}));return [...tools,{label:'Clear History',keywords:'clear recent history',run:clearHistory},{label:'Toggle Dark Mode',keywords:'theme dark light',run:toggleTheme},{label:'Copy Output',keywords:'copy result',run:copyOutput},{label:'Clear Input',keywords:'clear editor input',run:clearCurrentInput},{label:'Add Favorite',keywords:'favorite tool',run:()=>state.selectedTool&&toggleFavorite(state.selectedTool.id)}];}
function openCommandPalette(){state.paletteOpen=true;const palette=document.getElementById('commandPalette');palette.classList.remove('hidden');const input=document.getElementById('commandInput');input.value='';renderCommandResults();input.focus();}
function closeCommandPalette(){state.paletteOpen=false;document.getElementById('commandPalette')?.classList.add('hidden');}
function renderCommandResults(){const input=document.getElementById('commandInput');const query=(input?.value||'').toLowerCase();const commands=commandDefinitions().filter(command=>(command.label+' '+command.keywords).toLowerCase().includes(query)).slice(0,12);const results=document.getElementById('commandResults');results.replaceChildren(...commands.map(command=>{const button=document.createElement('button');button.textContent=command.label;button.onclick=()=>{closeCommandPalette();command.run();};return button;}));}
async function clearHistory(){try{await appService.clearHistory();state.history=[];renderSidebar();showToast('History cleared.');}catch(error){showToast('Unable to clear history.');}}
function toggleTheme(){applyTheme(document.documentElement.getAttribute('data-theme')==='dark'?'light':'dark');}
function clearCurrentInput(){const editor=document.getElementById('toolEditor')?.editor;if(editor)editor.clear();else if(document.getElementById('httpUrl')){document.getElementById('httpUrl').value='';document.getElementById('httpBody').value='';}}

async function copyOutput() {
    const output = document.getElementById('outputText').textContent;
    try {
        await copyText(output);
        showToast(t('copied'));
    } catch (err) {
        console.error('Failed to copy:', err);
        showToast(t('copyFailed'));
    }
}

function downloadOutput(toolName) {
    const output = document.getElementById('outputText').textContent;
    const element = document.createElement('a');
    element.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent(output));
    element.setAttribute('download', `${toolName.replace(/\s+/g, '-')}-output.txt`);
    element.style.display = 'none';
    document.body.appendChild(element);
    element.click();
    document.body.removeChild(element);
    showToast(t('downloaded'));
}

function showToast(message) {
    const toast = document.getElementById('toast');
    toast.innerText = message;
    toast.classList.remove('hidden');
    clearTimeout(window.toastTimer);
    window.toastTimer = setTimeout(() => {
        toast.classList.add('hidden');
    }, 3000);
}

// ============== Utils ==============
function renderError(container, message) {
    container.className = 'editor-output error-output';
    const error = document.createElement('div');
    error.className = 'error';
    error.textContent = message;
    container.replaceChildren(error);
}

// ============== Start ==============
init();
