export function normalizeHIITConfig(value = {}) {
    return {
        workSeconds: clampInteger(value.workSeconds, 30, 1, 3600),
        restSeconds: clampInteger(value.restSeconds, 15, 1, 3600),
        rounds: clampInteger(value.rounds, 8, 1, 99),
    };
}

function clampInteger(value, fallback, minimum, maximum) {
    const parsed = Number(value);
    if (!Number.isFinite(parsed)) return fallback;
    return Math.min(maximum, Math.max(minimum, Math.round(parsed)));
}

export function createHIITTimeline(rawConfig) {
    const config = normalizeHIITConfig(rawConfig);
    const segments = [];
    let cursorMs = 0;
    for (let round = 1; round <= config.rounds; round += 1) {
        const workDuration = config.workSeconds * 1000;
        segments.push({ type: 'work', round, startMs: cursorMs, endMs: cursorMs + workDuration, durationMs: workDuration });
        cursorMs += workDuration;
        if (round < config.rounds) {
            const restDuration = config.restSeconds * 1000;
            segments.push({ type: 'rest', round, startMs: cursorMs, endMs: cursorMs + restDuration, durationMs: restDuration });
            cursorMs += restDuration;
        }
    }
    return { config, segments, totalMs: cursorMs };
}

export function resolveHIITProgress(timeline, rawElapsedMs) {
    const elapsedMs = Math.max(0, Number(rawElapsedMs) || 0);
    if (elapsedMs >= timeline.totalMs) {
        return {
            complete: true,
            phase: 'complete',
            round: timeline.config.rounds,
            segmentIndex: timeline.segments.length,
            remainingMs: 0,
            totalRemainingMs: 0,
            progress: 1,
        };
    }
    let index = timeline.segments.findIndex(segment => elapsedMs < segment.endMs);
    if (index < 0) index = timeline.segments.length - 1;
    const segment = timeline.segments[index];
    const spent = Math.max(0, elapsedMs - segment.startMs);
    return {
        complete: false,
        phase: segment.type,
        round: segment.round,
        segmentIndex: index,
        remainingMs: Math.max(0, segment.endMs - elapsedMs),
        totalRemainingMs: Math.max(0, timeline.totalMs - elapsedMs),
        progress: Math.min(1, spent / segment.durationMs),
    };
}

export function formatTimer(milliseconds) {
    const totalSeconds = Math.max(0, Math.ceil(milliseconds / 1000));
    const minutes = Math.floor(totalSeconds / 60);
    return `${String(minutes).padStart(2, '0')}:${String(totalSeconds % 60).padStart(2, '0')}`;
}
