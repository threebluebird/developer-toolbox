export function filterTools(tools, searchText) {
    if (!searchText) return tools;
    return tools.filter(tool => [tool.name, tool.description, tool.category, tool.id]
        .some(value => value && value.toLowerCase().includes(searchText)));
}

export function groupByCategory(tools) {
    return tools.reduce((groups, tool) => {
        const category = tool.category || 'Other';
        if (!groups[category]) groups[category] = [];
        groups[category].push(tool);
        return groups;
    }, {});
}
