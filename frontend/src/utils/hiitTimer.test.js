import assert from 'node:assert/strict';
import test from 'node:test';

import { createHIITTimeline, formatTimer, normalizeHIITConfig, resolveHIITProgress } from './hiitTimer.js';

test('builds a timeline without rest after the final round', () => {
    const timeline = createHIITTimeline({ workSeconds: 20, restSeconds: 10, rounds: 3 });
    assert.equal(timeline.totalMs, 80000);
    assert.deepEqual(timeline.segments.map(value => value.type), ['work', 'rest', 'work', 'rest', 'work']);
});

test('resolves skipped time against the correct phase and round', () => {
    const timeline = createHIITTimeline({ workSeconds: 20, restSeconds: 10, rounds: 3 });
    const rest = resolveHIITProgress(timeline, 25000);
    assert.equal(rest.phase, 'rest');
    assert.equal(rest.round, 1);
    const progress = resolveHIITProgress(timeline, 35000);
    assert.equal(progress.phase, 'work');
    assert.equal(progress.round, 2);
    assert.equal(progress.remainingMs, 15000);
});

test('marks the timer complete and formats remaining time', () => {
    const timeline = createHIITTimeline({ workSeconds: 1, restSeconds: 1, rounds: 1 });
    assert.equal(resolveHIITProgress(timeline, 1000).complete, true);
    assert.equal(formatTimer(61001), '01:02');
    assert.deepEqual(normalizeHIITConfig({ workSeconds: -4, restSeconds: 99999, rounds: 2.6 }), { workSeconds: 1, restSeconds: 3600, rounds: 3 });
});
