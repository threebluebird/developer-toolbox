import { TextEditor } from './TextEditor.js';

const LANGUAGES = ['json', 'yaml', 'xml', 'sql', 'javascript', 'typescript', 'go', 'html', 'css', 'text'];
const BRACKETS = { '(': ')', '[': ']', '{': '}', '"': '"', "'": "'", '`': '`' };

function escapeHTML(value) { return value.replace(/[&<>]/g, char => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }[char])); }

function highlightCode(source, language) {
    const keywords = {
        json: ['true', 'false', 'null'], yaml: ['true', 'false', 'null', 'yes', 'no'],
        javascript: ['const', 'let', 'var', 'function', 'return', 'if', 'else', 'class', 'new', 'async', 'await', 'true', 'false', 'null'],
        typescript: ['const', 'let', 'interface', 'type', 'function', 'return', 'if', 'else', 'class', 'new', 'string', 'number', 'boolean'],
        go: ['package', 'import', 'func', 'type', 'struct', 'interface', 'return', 'if', 'else', 'for', 'range', 'go', 'defer', 'var', 'const'],
        sql: ['select', 'from', 'where', 'insert', 'into', 'update', 'delete', 'join', 'group', 'order', 'by', 'as', 'and', 'or', 'null'],
        css: ['important'], html: [], xml: [], text: [],
    }[language] || [];
    const keywordSet = new Set(keywords.map(value => value.toLowerCase()));
    const tokenPattern = /(\/\*[\s\S]*?\*\/|\/\/[^\n]*|#[^\n]*|<!--[\s\S]*?-->|"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|`(?:\\.|[^`\\])*`|\b\d+(?:\.\d+)?\b|\b[A-Za-z_$][\w$]*\b|<\/?[A-Za-z][^>]*>)/g;
    let result = ''; let offset = 0;
    for (const match of source.matchAll(tokenPattern)) {
        result += escapeHTML(source.slice(offset, match.index));
        const token = match[0];
        let kind = '';
        if (/^(\/\/|\/\*|#|<!--)/.test(token)) kind = 'comment';
        else if (/^["'`]/.test(token)) kind = 'string';
        else if (/^\d/.test(token)) kind = 'number';
        else if (/^<\/?/.test(token)) kind = 'tag';
        else if (keywordSet.has(token.toLowerCase())) kind = 'keyword';
        result += kind ? `<span class="syntax-${kind}">${escapeHTML(token)}</span>` : escapeHTML(token);
        offset = match.index + token.length;
    }
    return result + escapeHTML(source.slice(offset)) + (source.endsWith('\n') ? ' ' : '');
}

export class CodeEditor extends TextEditor {
    constructor(options = {}) {
        super(options);
        this.language = LANGUAGES.includes(options.language) ? options.language : 'text';
        this.element.classList.replace('text-editor', 'code-editor');
        this.input.dataset.language = this.language;
        this.input.setAttribute('aria-label', `${this.language} code editor`);
        const surface = this.element.querySelector('.editor-surface');
        this.codeArea = document.createElement('div');
        this.codeArea.className = 'code-editor-area';
        this.highlight = document.createElement('pre');
        this.highlight.className = 'code-highlight';
        this.highlight.setAttribute('aria-hidden', 'true');
        surface.replaceChild(this.codeArea, this.input);
        this.codeArea.append(this.highlight, this.input);
        this.input.addEventListener('scroll', () => {
            this.highlight.scrollTop = this.input.scrollTop;
            this.highlight.scrollLeft = this.input.scrollLeft;
        });
        this.refreshHighlight();
    }

    setLanguage(language) {
        this.language = LANGUAGES.includes(language) ? language : 'text';
        this.input.dataset.language = this.language;
        this.refreshHighlight();
    }

    refresh() { super.refresh(); this.refreshHighlight(); }
    refreshHighlight() { if (this.highlight) this.highlight.innerHTML = highlightCode(this.input.value, this.language); }

    onKeyDown(event) {
        super.onKeyDown(event);
        if (event.defaultPrevented) return;
        if (event.key === 'Enter') {
            event.preventDefault();
            const before = this.input.value.slice(0, this.input.selectionStart);
            const line = before.slice(before.lastIndexOf('\n') + 1);
            const indent = line.match(/^\s*/)?.[0] || '';
            const extra = /[{[(]\s*$/.test(line) ? (this.options.indent || '    ') : '';
            this.input.setRangeText(`\n${indent}${extra}`, this.input.selectionStart, this.input.selectionEnd, 'end');
            this.input.dispatchEvent(new Event('input', { bubbles: true }));
            return;
        }
        if (BRACKETS[event.key] && !event.ctrlKey && !event.metaKey && this.input.selectionStart === this.input.selectionEnd) {
            event.preventDefault();
            const close = BRACKETS[event.key];
            const position = this.input.selectionStart;
            this.input.setRangeText(event.key + close, position, position, 'end');
            this.input.setSelectionRange(position + 1, position + 1);
            this.input.dispatchEvent(new Event('input', { bubbles: true }));
        }
    }
}

export { LANGUAGES, highlightCode };
