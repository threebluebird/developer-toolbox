import { CodeEditor } from './CodeEditor.js';

function treeNode(key, value) {
    const type = value === null ? 'null' : Array.isArray(value) ? 'array' : typeof value;
    const label = key === null ? '' : `<span class="json-tree-key">${String(key)}:</span> `;
    if (type === 'object' || type === 'array') {
        const entries = Object.entries(value);
        return `<details open><summary>${label}<span class="json-tree-type">${type} · ${entries.length}</span></summary><div>${entries.map(([childKey, child]) => treeNode(childKey, child)).join('')}</div></details>`;
    }
    return `<div class="json-tree-leaf">${label}<span class="json-tree-value ${type}">${escapeHTML(String(value))}</span></div>`;
}

function escapeHTML(value) {
    return value.replace(/[&<>"']/g, char => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[char]));
}

export class JsonEditor extends CodeEditor {
    constructor(options = {}) {
        super({ ...options, language: 'json' });
        this.element.classList.add('json-editor');
        this.mode = options.mode || 'code';
        this.tree = document.createElement('div');
        this.tree.className = 'json-tree hidden';
        this.error = document.createElement('div');
        this.error.className = 'json-editor-status';
        this.element.append(this.tree, this.error);
        const modeButton = document.createElement('button');
        modeButton.type = 'button';
        modeButton.textContent = 'Tree';
        modeButton.title = 'Toggle JSON tree';
        modeButton.onclick = () => { this.setMode(this.mode === 'code' ? 'tree' : 'code'); modeButton.textContent = this.mode === 'code' ? 'Tree' : 'Code'; };
        this.toolbar?.element.querySelector('.editor-toolbar-actions')?.prepend(modeButton);
        this.refreshJSON();
    }

    refresh() { super.refresh(); this.refreshJSON(); }
    refreshJSON() {
        if (!this.tree || !this.error) return;
        if (!this.input.value.trim()) { this.tree.innerHTML = ''; this.error.textContent = ''; return; }
        try {
            const value = JSON.parse(this.input.value);
            this.tree.innerHTML = treeNode(null, value);
            this.error.textContent = 'Valid JSON';
            this.error.className = 'json-editor-status valid';
        } catch (error) {
            this.tree.innerHTML = '';
            this.error.textContent = error.message;
            this.error.className = 'json-editor-status invalid';
        }
    }
    setMode(mode) {
        this.mode = mode === 'tree' ? 'tree' : 'code';
        this.element.querySelector('.editor-surface').classList.toggle('hidden', this.mode === 'tree');
        this.tree.classList.toggle('hidden', this.mode !== 'tree');
        this.refreshJSON();
    }
}
