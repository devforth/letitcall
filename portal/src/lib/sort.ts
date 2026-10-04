/** Sort state is stored as `${key}-${direction}` so one value can back both column headers and a dropdown. */
export function nextSort(sort: string, key: string): string {
	return `${key}-${sort === `${key}-ascending` ? 'descending' : 'ascending'}`;
}
