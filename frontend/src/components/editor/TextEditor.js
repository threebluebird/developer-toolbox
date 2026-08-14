import { EditorToolbar } from './EditorToolbar.js';

export class TextEditor {
    constructor(options = {}) {
        this.options = { id: '', value: '', placeholder: '', readOnly: false, toolbar: true, ...options };
        this.currentMatch = -1;
        this.matches = [];
        this.element = document.createElement('div');
        this.element.className = 'unified-editor text-editor';
        this.build();
    }

    build() {
        if (this.options.toolbar && !this.options.readOnly) {
            this.toolbar = new EditorToolbar(this, this.options);
            this.element.append(this.toolbar.element);
        }
        const surface = document.createElement('div');
        surface.className = 'editor-surface';
        this.gutter = document.createElement('div');
        this.gutter.className = 'editor-line-numbers';
        this.gutter.setAttribute('aria-hidden', 'true');
        this.input = document.createElement('textarea');
        this.input.className = 'editor-input';
        this.input.id = this.options.id;
        this.input.value = this.options.value;
        this.input.placeholder = this.options.placeholder;
        this.input.readOnly = this.options.readOnly;
        this.input.spellcheck = false;
        this.input.setAttribute('aria-label', this.options.ariaLabel || 'Text editor');
        surface.append(this.gutter, this.input);
        this.element.append(surface);
        this.input.addEventListener('input', () => this.refresh());
        this.input.addEventListener('scroll', () => { this.gutter.scrollTop = this.input.scrollTop; });
        this.input.addEventListener('keydown', event => this.onKeyDown(event));
        this.refresh();
    }

    mount(container) { container.replaceChildren(this.element); return this; }
    get value() { return this.input.value; }
    set value(value) { this.input.value = value ?? ''; this.refresh(); }
    focus() { this.input.focus(); }
    refresh() {
        const lines = Math.max(1, this.input.value.split('\n').length);
        this.gutter.innerHTML = Array.from({ length: lines }, (_, index) => `<span>${index + 1}</span>`).join('');
        this.options.onChange?.(this.input.value);
    }
    setWordWrap(enabled) {
        this.input.wrap = enabled ? 'soft' : 'off';
        this.element.classList.toggle('no-wrap', !enabled);
    }
    selectAll() { this.input.focus(); this.input.select(); }
    async copy() {
        const text = this.input.selectionStart === this.input.selectionEnd
            ? this.input.value : this.input.value.slice(this.input.selectionStart, this.input.selectionEnd);
        if (text) await navigator.clipboard.writeText(text);
    }
    async paste() {
        const text = await navigator.clipboard.readText();
        this.input.setRangeText(text, this.input.selectionStart, this.input.selectionEnd, 'end');
        this.input.dispatchEvent(new Event('input', { bubbles: true }));
    }
    clear() { this.value = ''; this.input.dispatchEvent(new Event('input', { bubbles: true })); this.focus(); }
    find(query, backwards = false, keepIndex = false) {
        this.matches = [];
        if (!query) { this.currentMatch = -1; this.toolbar?.updateCount(0, 0); return; }
        const source = this.input.value.toLocaleLowerCase();
        const target = query.toLocaleLowerCase();
        let offset = 0;
        while ((offset = source.indexOf(target, offset)) !== -1) { this.matches.push(offset); offset += Math.max(1, target.length); }
        if (!this.matches.length) { this.currentMatch = -1; this.toolbar?.updateCount(0, 0); return; }
        if (!keepIndex) this.currentMatch = backwards
            ? (this.currentMatch <= 0 ? this.matches.length - 1 : this.currentMatch - 1)
            : (this.currentMatch + 1) % this.matches.length;
        if (this.currentMatch < 0 || this.currentMatch >= this.matches.length) this.currentMatch = 0;
        const start = this.matches[this.currentMatch];
        this.input.focus();
        this.input.setSelectionRange(start, start + query.length);
        this.toolbar?.updateCount(this.currentMatch + 1, this.matches.length);
    }
    replace(query, replacement) {
        if (!query) return;
        const selected = this.input.value.slice(this.input.selectionStart, this.input.selectionEnd);
        if (selected.toLocaleLowerCase() === query.toLocaleLowerCase()) {
            this.input.setRangeText(replacement, this.input.selectionStart, this.input.selectionEnd, 'end');
            this.input.dispatchEvent(new Event('input', { bubbles: true }));
        }
        this.find(query);
    }
    replaceAll(query, replacement) {
        if (!query) return;
        this.value = this.input.value.replaceAll(query, replacement);
        this.input.dispatchEvent(new Event('input', { bubbles: true }));
        this.find(query, false, true);
    }
    onKeyDown(event) {
        if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'f') { event.preventDefault(); this.toolbar?.openFind(false); }
        if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'h') { event.preventDefault(); this.toolbar?.openFind(true); }
        if (event.key === 'Tab') {
            event.preventDefault();
            this.input.setRangeText('    ', this.input.selectionStart, this.input.selectionEnd, 'end');
            this.input.dispatchEvent(new Event('input', { bubbles: true }));
        }
    }
}
