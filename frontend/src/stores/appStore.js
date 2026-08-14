export function createAppState() {
    return {
        tools: [],
        allTools: [],
        search: '',
        favorites: [],
        favoriteItems: [],
        history: [],
        selectedTool: null,
        currentView: 'home',
        paletteOpen: false,
        settings: {
            theme: 'dark',
            language: 'en',
            storagePath: '',
            jsonIndent: 2,
        },
    };
}
