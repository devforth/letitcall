<script lang="ts">
	import { theme } from '$lib/stores/theme';

	function handleToggle() {
		const newTheme = $theme === 'light' ? 'dark' : 'light';
		theme.setTheme(newTheme);
	}
</script>

<button
	onclick={handleToggle}
	aria-label="Toggle theme"
	class="toggle-switch"
	class:dark={$theme === 'dark'}
>
	<div class="toggle-track">
		<div class="toggle-thumb">
			{#if $theme === 'light'}
				<svg class="icon" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
					<path d="M12 18a6 6 0 100-12 6 6 0 000 12zM12 2v6m0 6v6M4.22 4.22l4.24 4.24m0 5.08l-4.24 4.24M19.78 4.22l-4.24 4.24m0 5.08l4.24 4.24" />
				</svg>
			{:else}
				<svg class="icon" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
					<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
				</svg>
			{/if}
		</div>
	</div>
</button>

<style>
	.toggle-switch {
		position: relative;
		width: 56px;
		height: 32px;
		border: 2px solid rgba(255, 255, 255, 0.5);
		border-radius: 16px;
		cursor: pointer;
		background: rgba(255, 255, 255, 0.1);
		transition: all 0.3s ease;
		padding: 0;
		display: flex;
		align-items: center;
		backdrop-filter: blur(8px);
	}

	.toggle-switch:hover {
		background: rgba(255, 255, 255, 0.15);
		border-color: rgba(255, 255, 255, 0.7);
	}

	.toggle-switch.dark {
		background: rgba(255, 255, 255, 0.05);
		border-color: rgba(255, 255, 255, 0.2);
	}

	.toggle-switch.dark:hover {
		background: rgba(255, 255, 255, 0.1);
		border-color: rgba(255, 255, 255, 0.3);
	}

	.toggle-track {
		position: absolute;
		width: 100%;
		height: 100%;
		border-radius: 16px;
		display: flex;
		align-items: center;
	}

	.toggle-thumb {
		position: absolute;
		width: 28px;
		height: 28px;
		border-radius: 50%;
		background: rgba(255, 255, 255, 0.9);
		display: flex;
		align-items: center;
		justify-content: center;
		transition: transform 0.3s ease;
		left: 2px;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
	}

	.toggle-switch.dark .toggle-thumb {
		transform: translateX(24px);
		background: rgba(255, 255, 255, 0.8);
	}

	.icon {
		width: 16px;
		height: 16px;
		color: rgb(0, 153, 255);
	}

	.toggle-switch.dark .icon {
		color: rgb(0, 201, 80);
	}
</style>
