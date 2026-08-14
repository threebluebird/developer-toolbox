import assert from 'node:assert/strict';
import test from 'node:test';

import { diffLines } from './DiffEditor.js';
import { highlightCode, LANGUAGES } from './CodeEditor.js';

test('diffLines identifies unchanged, added, and removed lines', () => {
    const changes = diffLines('one\ntwo\nthree', 'one\nsecond\nthree\nfour');
    assert.deepEqual(changes.map(change => change.type), ['same', 'added', 'removed', 'same', 'added']);
    assert.equal(changes.filter(change => change.type === 'added').length, 2);
    assert.equal(changes.filter(change => change.type === 'removed').length, 1);
});

test('highlightCode escapes source and marks language tokens', () => {
    const output = highlightCode('const value = "<tag>";', 'javascript');
    assert.match(output, /syntax-keyword/);
    assert.match(output, /&lt;tag&gt;/);
    assert.doesNotMatch(output, /<tag>/);
});

test('CodeEditor declares every Phase 2 language', () => {
    for (const language of ['json', 'yaml', 'xml', 'sql', 'javascript', 'typescript', 'go', 'html', 'css']) {
        assert.ok(LANGUAGES.includes(language), `missing ${language}`);
    }
});
