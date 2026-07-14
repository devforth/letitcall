<script lang="ts">
	let { open = false }: { open?: boolean } = $props();
</script>

<span class:open class="menu-toggle-icon" aria-hidden="true">
	<span></span>
	<span></span>
	<span></span>
</span>

<style>
	.menu-toggle-icon {
		position: relative;
		display: block;
		width: 20px;
		height: 16px;
	}

	.menu-toggle-icon span {
		position: absolute;
		left: 0;
		width: 100%;
		height: 2px;
		border-radius: 999px;
		background: currentColor;
		transition: transform 0.2s ease, opacity 0.2s ease;
	}

	.menu-toggle-icon span:nth-child(1) {
		top: 0;
	}

	.menu-toggle-icon span:nth-child(2) {
		top: 7px;
	}

	.menu-toggle-icon span:nth-child(3) {
		top: 14px;
	}

	.menu-toggle-icon.open span:nth-child(1) {
		transform: translateY(7px) rotate(45deg);
	}

	.menu-toggle-icon.open span:nth-child(2) {
		opacity: 0;
		transform: scaleX(0);
	}

	.menu-toggle-icon.open span:nth-child(3) {
		transform: translateY(-7px) rotate(-45deg);
	}

	@media (prefers-reduced-motion: reduce) {
		.menu-toggle-icon span {
			transition: none;
		}
	}
</style>
