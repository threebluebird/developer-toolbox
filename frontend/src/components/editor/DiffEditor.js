import { TextEditor } from './TextEditor.js';

function diffLines(original, modified) {
    const before = original.split('\n');
    const after = modified.split('\n');
    const rows = before.length + 1;
    const cols = after.length + 1;
    const matrix = Array.from({ length: rows }, () => new Uint32Array(cols));
    for (let i = before.length - 1; i >= 0; i--) {
        for (let j = after.length - 1; j >= 0; j--) matrix[i][j] = before[i] === after[j]
            ? matrix[i + 1][j + 1] + 1 : Math.max(matrix[i + 1][j], matrix[i][j + 1]);
    }
    const changes = [];
    let i = 0; let j = 0;
    while (i < before.length || j < after.length) {
        if (i < before.length && j < after.length && before[i] === after[j]) changes.push({ type: 'same', text: before[i++], oldLine: i, newLine: ++j });
        else if (j < after.length && (i === before.length || matrix[i][j + 1] >= matrix[i + 1][j])) changes.push({ type: 'added', text: after[j++], newLine: j });
        else changes.push({ type: 'removed', text: before[i++], oldLine: i });
    }
    return changes;
}

function escapeHTML(value) { return value.replace(/[&<>]/g, char => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }[char])); }

export class DiffEditor {
    constructor(options = {}) {
        this.options = options;
        this.mode = options.mode || 'side-by-side';
        this.element = document.createElement('div');
        this.element.className = 'diff-editor';
        this.build();
    }
    build() {
        this.element.innerHTML = `<div class="diff-toolbar"><div><strong>Diff</strong><span class="diff-summary"></span></div><select aria-label="Diff mode"><option value="side-by-side">Side by side</option><option value="unified">Unified</option></select></div><div class="diff-inputs"><div class="diff-original"><h4>Original</h4></div><div class="diff-modified"><h4>Modified</h4></div></div><div class="diff-result"></div>`;
        this.original = new TextEditor({ toolbar: false, ariaLabel: 'Original text', onChange: () => this.renderDiff() });
        this.modified = new TextEditor({ toolbar: false, ariaLabel: 'Modified text', onChange: () => this.renderDiff() });
        this.element.querySelector('.diff-original').append(this.original.element);
        this.element.querySelector('.diff-modified').append(this.modified.element);
        const select = this.element.querySelector('select');
        select.value = this.mode;
        select.onchange = () => { this.mode = select.value; this.renderDiff(); };
        this.result = this.element.querySelector('.diff-result');
        this.summary = this.element.querySelector('.diff-summary');
        this.original.value = this.options.original || '';
        this.modified.value = this.options.modified || '';
        this.renderDiff();
    }
    mount(container) { container.replaceChildren(this.element); return this; }
    renderDiff() {
        if (!this.result) return;
        const changes = diffLines(this.original.value, this.modified.value);
        const added = changes.filter(change => change.type === 'added').length;
        const removed = changes.filter(change => change.type === 'removed').length;
        this.summary.textContent = `+${added} −${removed}`;
        this.result.className = `diff-result ${this.mode}`;
        if (this.mode === 'unified') {
            this.result.innerHTML = changes.map(change => `<div class="diff-line ${change.type}"><span>${change.oldLine ?? ''}</span><span>${change.newLine ?? ''}</span><code>${change.type === 'added' ? '+' : change.type === 'removed' ? '-' : ' '} ${escapeHTML(change.text)}</code></div>`).join('');
            return;
        }
        const left = changes.filter(change => change.type !== 'added').map(change => `<div class="diff-line ${change.type}"><span>${change.oldLine ?? ''}</span><code>${escapeHTML(change.text)}</code></div>`).join('');
        const right = changes.filter(change => change.type !== 'removed').map(change => `<div class="diff-line ${change.type}"><span>${change.newLine ?? ''}</span><code>${escapeHTML(change.text)}</code></div>`).join('');
        this.result.innerHTML = `<div>${left}</div><div>${right}</div>`;
    }
}

export { diffLines };
