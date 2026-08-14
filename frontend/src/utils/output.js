export function formatOutput(value) {
    if (typeof value === 'string') return value;
    return JSON.stringify(value, null, 2);
}

export function errorMessage(error) {
    if (!error) return 'Error';
    if (typeof error === 'string') return error;
    return error.detail ? `${error.message}: ${error.detail}` : error.message;
}
