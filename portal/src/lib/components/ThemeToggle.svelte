<script lang="ts">
	import moonIcon from '@iconify-icons/tabler/moon';
	import sunIcon from '@iconify-icons/tabler/sun-high';
	import SegmentedControl from '$lib/components/ui/SegmentedControl.svelte';
	import { theme, type Theme } from '$lib/stores/theme';

	const themeOptions = [
		{ value: 'light', label: 'Light', icon: sunIcon },
		{ value: 'dark', label: 'Dark', icon: moonIcon }
	];

	function selectTheme(value: string) {
		theme.setTheme(value as Theme);
	}
</script>

<SegmentedControl
	options={themeOptions}
	value={$theme}
	label="Color theme"
	iconOnly
	onchange={selectTheme}
/>
