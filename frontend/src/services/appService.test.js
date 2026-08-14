import test from 'node:test';
import assert from 'node:assert/strict';

globalThis.window = {};

const { appService } = await import('./appService.js');

test('web preview records amount and database documentation usage', async () => {
    await appService.clearHistory();

    await appService.executeTool({ toolId: 'amount-cn', payload: { input: '123.45', action: 'upper' } });
    await appService.executeTool({ toolId: 'database-docs', payload: { input: '', database: 'postgresql' } });

    const history = await appService.getHistory();
    assert.deepEqual(history.map(item => item.toolId), ['database-docs', 'amount-cn']);
    assert.equal(history[0].input.database, 'postgresql');
    assert.equal(history[1].input.input, '123.45');
});
