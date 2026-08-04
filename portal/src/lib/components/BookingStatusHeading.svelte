<script lang="ts">
	let { text }: { text: string } = $props();
</script>

<p class="booking-status-heading">{text}</p>

<style>
	.booking-status-heading {
		font-size: var(--text-3xl);
		font-weight: 400;
		line-height: 1.25;
		color: rgb(var(--color-primary));
	}
</style>
