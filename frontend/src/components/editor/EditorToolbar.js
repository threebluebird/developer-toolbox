export class EditorToolbar {
    constructor(editor, options = {}) {
        this.editor = editor;
        this.options = options;
        this.element = document.createElement('div');
        this.element.className = 'editor-toolbar';
        this.render();
    }

    render() {
        this.element.innerHTML = `
            <div class="editor-toolbar-actions">
                <button type="button" data-action="find" title="Find (Ctrl+F)">Find</button>
                <button type="button" data-action="replace" title="Replace (Ctrl+H)">Replace</button>
                <button type="button" data-action="select">Select all</button>
                <button type="button" data-action="copy">Copy</button>
                <button type="button" data-action="paste">Paste</button>
                <button type="button" data-action="clear">Clear</button>
                <label class="editor-wrap-toggle"><input type="checkbox" data-action="wrap" checked /> Wrap</label>
            </div>
            <div class="editor-find-panel hidden">
                <input class="editor-find-input" aria-label="Find text" placeholder="Find" />
                <input class="editor-replace-input" aria-label="Replace text" placeholder="Replace" />
                <span class="editor-find-count">0/0</span>
                <button type="button" data-action="previous" title="Previous match">↑</button>
                <button type="button" data-action="next" title="Next match">↓</button>
                <button type="button" data-action="replace-one">Replace</button>
                <button type="button" data-action="replace-all">All</button>
                <button type="button" data-action="close-find" title="Close">×</button>
            </div>`;

        this.findPanel = this.element.querySelector('.editor-find-panel');
        this.findInput = this.element.querySelector('.editor-find-input');
        this.replaceInput = this.element.querySelector('.editor-replace-input');
        this.count = this.element.querySelector('.editor-find-count');
        this.element.addEventListener('click', event => {
            const action = event.target.closest('[data-action]')?.dataset.action;
            if (action) this.run(action);
        });
        this.findInput.addEventListener('input', () => this.editor.find(this.findInput.value, false));
        this.findInput.addEventListener('keydown', event => {
            if (event.key === 'Enter') this.editor.find(this.findInput.value, event.shiftKey);
            if (event.key === 'Escape') this.closeFind();
        });
    }

    run(action) {
        const commands = {
            find: () => this.openFind(false), replace: () => this.openFind(true),
            select: () => this.editor.selectAll(), copy: () => this.editor.copy(),
            paste: () => this.editor.paste(), clear: () => this.editor.clear(),
            previous: () => this.editor.find(this.findInput.value, true),
            next: () => this.editor.find(this.findInput.value, false),
            'replace-one': () => this.editor.replace(this.findInput.value, this.replaceInput.value),
            'replace-all': () => this.editor.replaceAll(this.findInput.value, this.replaceInput.value),
            'close-find': () => this.closeFind(),
        };
        if (action === 'wrap') this.editor.setWordWrap(this.element.querySelector('[data-action="wrap"]').checked);
        else commands[action]?.();
    }

    openFind(showReplace) {
        this.findPanel.classList.remove('hidden');
        this.replaceInput.classList.toggle('hidden', !showReplace);
        this.findInput.focus();
        this.findInput.select();
        this.editor.find(this.findInput.value, false, true);
    }

    closeFind() {
        this.findPanel.classList.add('hidden');
        this.editor.focus();
    }

    updateCount(current, total) { this.count.textContent = `${current}/${total}`; }
}
